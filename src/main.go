package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Requires self and partner ports")
	}
	selfPort := os.Args[1]
	// partnerPort := os.Args[2]

	listen, err := net.Listen("tcp", "0.0.0.0:"+selfPort)
	fmt.Printf("Up at 0.0.0.0:%s\n", selfPort)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	buffer := make([]byte, 32)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
		conn.Close()
		os.Exit(1)
	}

	recv := strings.TrimRight(string(buffer), "\x00")
	log.Printf("Received %s", recv)
	if recv == "done" {
		conn.Close()
		os.Exit(0)
	}

	recvI, err := strconv.Atoi(recv)
	if err != nil {
		log.Fatal(err)
		conn.Close()
		os.Exit(1)
	}

	var msg string
	if recvI == 0 {
		msg = "done"
	} else {
		msg = strconv.Itoa(recvI - 1)
	}
	conn.Write([]byte(msg))

	conn.Close()
}
