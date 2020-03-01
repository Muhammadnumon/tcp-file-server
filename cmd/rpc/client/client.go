package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"tcp-server/pkg/rpc"
)

var (
	download = flag.String("download", "default", "download")
	upload   = flag.String("upload", "default", "upload")
	list     = flag.Bool("list", false, "list")
)

func main() {
	file, err := os.OpenFile("client-log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
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
	flag.Parse()
	var cmd, fileName string
	if *download != "default" {
		fileName = *download
		cmd = rpc.Download
	} else if *upload != "default" {
		fileName = *upload
		cmd = rpc.Upload
	} else if *list != false {
		cmd = rpc.FileList
	} else {
		return
	}

	operationsLoop(cmd, fileName)
}

func operationsLoop(cmd, fileName string) {
	log.Print("client connecting")
	address := "localhost:9999"
	conn, err := net.Dial(rpc.Tcp, address)
	if err != nil {
		log.Fatalf("can't connect to %s: %v", address, err)
	}
	defer func() {
		log.Print("closing connection")
		err := conn.Close()
		if err != nil {
			log.Printf("can't close connection: %v", err)
		}
		log.Print("connection closed")
	}()
	log.Print("client connected")
	writer := bufio.NewWriter(conn)
	line := cmd + ":" + fileName
	log.Print("command sending")
	err = rpc.WriteLine(line, writer)
	if err != nil {
		log.Fatalf("can't send command %s to server: %v", line, err)
	}
	log.Print("command sent")
	switch cmd {
	case rpc.Download:
		log.Print("download from server operations started")
		downloadFromServer(conn, fileName)
		log.Print("download from server operations ended")
	case rpc.Upload:
		log.Print("upload to server operations started")
		uploadToServer(conn, fileName)
		log.Print("upload to server operations ended")
	case rpc.FileList:
		log.Print("get file list operations started")
		listFile(conn)
		log.Print("get file list operations ended")
	}
}

func downloadFromServer(conn net.Conn, fileName string) {
	reader := bufio.NewReader(conn)
	log.Print("reading from client")
	line, err := rpc.ReadLine(reader)
	if err != nil {
		log.Printf("can't read: %v", err)
		return
	}
	if line == rpc.ResultErr + rpc.Suffix {
		log.Printf("file not found: %v", err)
		fmt.Printf("Файл \"%s\" не существует на сервере\n", fileName)
		return
	}

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		if err != io.EOF {
			log.Printf("can't read data: %v", err)
		}
		return
	}
	log.Print("saving file to 'files' path")
	err = ioutil.WriteFile(rpc.ClientFiles+ fileName, bytes, 0666)
	if err != nil {
		log.Printf("can't write file: %v", err)
		return
	}
	fmt.Printf("Файл \"%s\" успешно скачен\n", fileName)
}

func uploadToServer(conn net.Conn, fileName string) {
	options := strings.TrimSuffix(fileName, rpc.Suffix)
	log.Print("opening file")
	file, err := os.Open(rpc.ClientFiles + options)
	writer := bufio.NewWriter(conn)
	if err != nil {
		log.Print("file does not exist")
		err = rpc.WriteLine(rpc.ResultErr, writer)
		fmt.Printf("Файл \"%s\" не существует\n", fileName)
		return
	}
	defer func() {
		log.Print("closing file")
		err = file.Close()
		if err != nil {
			log.Print("can't close file")
			return
		}
		log.Print("file closed")
	}()
	err = rpc.WriteLine(rpc.ResultOK, writer)
	if err != nil {
		log.Printf("error while writing: %v", err)
		return
	}
	log.Print("start sending file")
	fileByte, err := io.Copy(writer, file)
	log.Print(fileByte)
	err = writer.Flush()
	if err != nil {
		log.Printf("can't flush: %v", err)
	}
	log.Print("file sent")
	fmt.Printf("Файл \"%s\" успешно отправлен на сервер\n", fileName)
}

func listFile(conn net.Conn) {
	reader := bufio.NewReader(conn)
	line, err := rpc.ReadLine(reader)
	if err != nil {
		log.Printf("can't read: %v", err)
		return
	}
	fmt.Println("Список доступных файлов для скачивания:")
	var list string
	for i := 0; i < len(line); i++{
		if string(line[i]) == " " || string(line[i]) == "\n"{
			fmt.Println(list)
			list = ""
		} else {
			list = list + string(line[i])
		}
	}
	_, err = ioutil.ReadAll(reader)
	if err != nil {
		if err != io.EOF {
			log.Printf("can't read data: %v", err)
		}
	}
}