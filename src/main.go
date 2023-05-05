package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Requires selfPort, partnerAddr")
	}
	selfPort := os.Args[1]
	partnerAddr := os.Args[2]

	listen, err := net.Listen("tcp", "0.0.0.0:"+selfPort)
	checkErr(err)
	fmt.Printf("Up at 0.0.0.0:%s\n", selfPort)

	defer listen.Close()
	for {
		conn, err := listen.Accept()
		checkErr(err)
		go handleRequest(conn, partnerAddr)
	}
}

func handleRequest(conn net.Conn, partnerAddr string) {
	buffer := make([]byte, 32)
	_, err := conn.Read(buffer)
	checkErr(err)

	recv := strings.TrimRight(string(buffer), "\x00")
	conn.Close()
	log.Printf("Received %s", recv)
	if recv == "done" {
		os.Exit(0)
	}

	recvI, err := strconv.Atoi(recv)
	checkErr(err)

	var msg string
	if recvI == 0 {
		msg = "done"
	} else {
		msg = strconv.Itoa(recvI - 1)
	}
	// conn.Write([]byte(msg))

	tcpServer, err := net.ResolveTCPAddr("tcp", partnerAddr)
	checkErr(err)
	partnerConn, err := net.DialTCP("tcp", nil, tcpServer)
	checkErr(err)
	partnerConn.Write([]byte(msg))
	partnerConn.Close()

	if recvI == 0 {
		os.Exit(0)
	}
}
