package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"tcp-server/pkg/rpc"
)

func main() {
	file, err := os.OpenFile("server-log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		log.Print("closing file")
		err := file.Close()
		if err != nil {
			log.Printf("can't close file: %v", err)
			return
		}
		log.Print("file closed")
	}()
	log.SetOutput(file)
	address := "0.0.0.0:9999"
	err = startServer(address)
	if err != nil {
		log.Fatalf("can't start server: %v", err)
	}
}

func startServer(address string) (err error) {
	log.Printf("starting server at: %s", address)
	listener, err := net.Listen(rpc.Tcp, address)
	if err != nil {
		log.Printf("can't listen %s: %v", address, err)
		return err
	}
	log.Printf("server started at: %s", address)
	defer func() {
		log.Print("closing server")
		err := listener.Close()
		if err != nil {
			log.Fatalf("can't close server: %v", err)
		}
		log.Print("server closed")
	}()
	for {
		conn, err := listener.Accept()
		log.Print("accepting connection")
		if err != nil {
			log.Printf("can't accept: %v", err)
			continue
		}
		log.Print("connection accepted")

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer func() {
		log.Print("closing connection")
		err := conn.Close()
		if err != nil {
			log.Fatalf("can't close connection: %v", err)
		}
		log.Print("connection closed")
	}()
	reader := bufio.NewReader(conn)
	line, err := rpc.ReadLine(reader)
	if err != nil {
		log.Printf("error while reading: %v", err)
		return
	}

	index := strings.IndexByte(line, ':')
	writer := bufio.NewWriter(conn)
	if index == -1 {
		log.Printf("invalid line received %s", line)
		err := rpc.WriteLine("error: invalid line", writer)
		if err != nil {
			log.Printf("error while writing: %v", err)
			return
		}
		return
	}
	cmd, options := line[:index], line[index+1:]
	log.Printf("command received: %s", cmd)
	log.Printf("options received: %s", options)

	switch cmd {
	case rpc.Upload:
		options := strings.TrimSuffix(options, rpc.Suffix)
		line, err := rpc.ReadLine(reader)
		if err != nil {
			log.Printf("can't read: %v", err)
			return
		}
		if line == rpc.ResultErr + rpc.Suffix {
			log.Printf("file not found: %v", err)
			return
		}
		bytes, err := ioutil.ReadAll(reader)
		if err != nil {
			if err != io.EOF {
				log.Printf("can't read data: %v", err)
				return
			}
		}
		log.Print("writing to file")
		err = ioutil.WriteFile(rpc.ServerFiles+ options, bytes, 0666)
		if err != nil {
			log.Printf("can't write file: %v", err)
			return
		}
		log.Print("file written")
		err = rpc.WriteLine(rpc.ResultOK, writer)
		if err != nil {
			log.Printf("error while writing: %v", err)
			return
		}
	case rpc.Download:
		options = strings.TrimSuffix(options, rpc.Suffix)
		log.Printf("opening file: %s", options)
		file, err := os.Open(rpc.ServerFiles + options)
		if err != nil {
			log.Print("file does not exist")
			err = rpc.WriteLine(rpc.ResultErr, writer)
			return
		}
		defer func() {
			log.Print("closing file")
			err = file.Close()
			if err != nil {
				log.Printf("can't close file: %v", err)
				return
			}
			log.Print("file closed")
		}()
		err = rpc.WriteLine(rpc.ResultOK, writer)
		if err != nil {
			log.Printf("error while writing: %v", err)
			return
		}
		_, err = io.Copy(writer, file)
		err = writer.Flush()
		if err != nil {
			log.Printf("can't flush: %v", err)
			return
		}
	case rpc.FileList:
		options = strings.TrimSuffix(options, rpc.Suffix)
		log.Print("getting file list")
		fileName := rpc.GetFileList(rpc.ServerFiles)
		log.Print("file list received")

		err := rpc.WriteLine(fileName, writer)
		if err != nil {
			log.Printf("error while writing: %v", err)
			return
		}
	default:
		log.Print("invalid operation selected")
		err := rpc.WriteLine(rpc.ResultErr, writer)
		if err != nil {
			log.Printf("error while writing: %v", err)
			return
		}
	}
}