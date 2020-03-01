package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"tcp-server/pkg/rpc"
	"testing"
	"time"
)

func Test_DownloadFromServer(t *testing.T) {
	address := "localhost:9999"
	go func() {
		err := startServer(address)
		if err != nil {
			t.Fatalf("can't start server: %v", err)
		}
	}()
	time.Sleep(time.Millisecond)
	conn, err := net.Dial(rpc.Tcp, address)
	if err != nil {
		t.Fatalf("can't connect to server: %v", err)
	}
	defer conn.Close()
	writer := bufio.NewWriter(conn)
	options := "secondfile.txt"
	line := rpc.Download + ":" + options
	err = rpc.WriteLine(line, writer)
	if err != nil {
		t.Fatalf("can't send command %s to server: %v", line, err)
	}
	reader := bufio.NewReader(conn)
	line, err = rpc.ReadLine(reader)
	src, err := ioutil.ReadFile("files/" +options)
	if err != nil {
		log.Fatalf("can't read file: %v", err)
	}
	dst, err := ioutil.ReadFile("files/" + options)
	if err != nil {
		log.Fatalf("can't read file: %v", err)
	}
	if !bytes.Equal(src, dst) {
		t.Fatalf("files are not same: %v", err)
	}
}

func Test_UploadToServer(t *testing.T) {
	address := "localhost:9999"
	conn, err := net.Dial(rpc.Tcp, address)
	if err != nil {
		t.Fatalf("can't connect to server: %v", err)
	}
	defer conn.Close()
	writer := bufio.NewWriter(conn)
	options := "secondfile.txt"
	line := rpc.Upload + ":" + options
	err = rpc.WriteLine(line, writer)
	if err != nil {
		t.Fatalf("can't send command %s to server: %v", line, err)
	}
	src, err := ioutil.ReadFile("files/" +options)
	if err != nil {
		log.Fatalf("can't read file: %v",err)
	}
	_, err = writer.Write(src)
	if err != nil {
		log.Fatalf("can't write: %v", err)
	}
	err = writer.Flush()
	if err != nil {
		log.Fatalf("can't flush: %v", err)
	}
	err = conn.Close()
	if err != nil {
		log.Fatalf("can't close conn: %v", err)
	}
	dst, err := ioutil.ReadFile("files/" + options)
	if err != nil {
		log.Fatalf("can't Read file: %v",err)
	}
	if !bytes.Equal(src, dst) {
		t.Fatalf("files are not same: %v", err)
	}
}

func Test_GetFileListFromServer(t *testing.T)  {
	address := "localhost:9999"
	conn, err := net.Dial(rpc.Tcp, address)
	if err != nil {
		t.Fatalf("can't connect to server: %v", err)
	}
	writer := bufio.NewWriter(conn)
	options := ""
	line := rpc.FileList + ":" + options
	err = rpc.WriteLine(line, writer)
	if err != nil {
		t.Fatalf("can't send command %s to server: %v", line, err)
	}
	reader := bufio.NewReader(conn)
	line, err = rpc.ReadLine(reader)
	if line != "firstfile.txt secondfile.txt\n" {
		t.Fatalf("result not ok: %s %v", line, err)
	}
}