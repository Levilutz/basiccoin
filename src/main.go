package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const Version = "0.1.0"

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func getCLIArgs() (string, *string) {
	if len(os.Args) == 1 {
		return "21720", nil
	} else if len(os.Args) == 2 {
		return os.Args[1], nil
	}
	return os.Args[1], &os.Args[2]
}

func main() {
	selfPort, seedPartnerAddr := getCLIArgs()

	neighbors := make([]string, 1)
	if seedPartnerAddr != nil {
		neighbors[0] = *seedPartnerAddr
	}

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

func versionHandler(conn net.Conn, lines []string) {
	if len(lines) < 2 || len(lines[1]) == 0 {
		conn.Close()
		fmt.Println("No version given")
		return
	}
	theirVersion := lines[1]
	// Send verack
	conn.Write([]byte("verack\n" + theirVersion))
	conn.Close()
}

func fallbackHandler(conn net.Conn, lines []string) {
	conn.Write([]byte("UNKNOWN"))
	conn.Close()
}

func dispatchRequest(conn net.Conn) {
	buffer := make([]byte, 128)
	_, err := conn.Read(buffer)
	checkErr(err)

	recv := strings.TrimRight(string(buffer), "\x00")
	lines := strings.Split(recv, "\n")

	dispatch := map[string]func(conn net.Conn, lines []string){
		"ping":    pingHandler,
		"version": versionHandler,
	}

	handler, ok := dispatch[lines[0]]
	if ok {
		handler(conn, lines)
	} else {
		fallbackHandler(conn, lines)
	}
}
