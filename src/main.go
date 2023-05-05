package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func getCLIArgs() (string, string) {
	if len(os.Args) == 1 {
		log.Fatal("Requires seedPartnerAddr, [selfPort]")
		os.Exit(1)
	} else if len(os.Args) == 2 {
		return os.Args[1], "21720"
	}
	return os.Args[1], os.Args[2]
}

func main() {
	seedPartnerAddr, selfPort := getCLIArgs()

	neighbors := make([]string, 1)
	neighbors[0] = seedPartnerAddr

	listen, err := net.Listen("tcp", "0.0.0.0:"+selfPort)
	checkErr(err)
	fmt.Printf("Up at 0.0.0.0:%s\n", selfPort)

	defer listen.Close()
	for {
		conn, err := listen.Accept()
		checkErr(err)
		go dispatchRequest(conn)
	}
}

func pingHandler(conn net.Conn, lines []string) {
	conn.Write([]byte("pong"))
	conn.Close()
}

func abcHandler(conn net.Conn, lines []string) {
	conn.Write([]byte("123"))
	conn.Close()
}

func fallbackHandler(conn net.Conn, lines []string) {
	conn.Write([]byte("UNKNOWN"))
	conn.Close()
}

func dispatchRequest(conn net.Conn) {
	buffer := make([]byte, 32)
	_, err := conn.Read(buffer)
	checkErr(err)

	recv := strings.TrimRight(string(buffer), "\x00")
	lines := strings.Split(recv, "\n")

	dispatch := map[string]func(conn net.Conn, lines []string){
		"ping": pingHandler,
		"abc":  abcHandler,
	}

	handler, ok := dispatch[lines[0]]
	if ok {
		handler(conn, lines)
	} else {
		fallbackHandler(conn, lines)
	}
}
