package util

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
)

// Get keys from map.
func MapKeys[K comparable, V any](in map[K]V) []K {
	out := make([]K, len(in))
	i := 0
	for k := range in {
		out[i] = k
		i++
	}
	return out
}

// Prepend into slice.
func Prepend[K any](ls []K, items ...K) []K {
	for _, item := range items {
		if len(ls) == 0 {
			ls = []K{item}
		} else {
			ls = append(ls, item)
			copy(ls[1:], ls)
			ls[0] = item
		}
	}
	return ls
}

// Aggregate channels.
func Aggregate[K any](chans []chan K) <-chan K {
	out := make(chan K)
	for _, ch := range chans {
		in := ch
		go func() {
			for {
				out <- <-in
			}
		}()
	}
	return out
}

// Shallow copy a map.
func CopyMap[K comparable, V any](m map[K]V) map[K]V {
	out := make(map[K]V, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// Flatten a double list.
func FlattenLists[K comparable](in [][]K) []K {
	out := make([]K, 0)
	for _, inL := range in {
		out = append(out, inL...)
	}
	return out
}

// Generate UUID
func UUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Generate UUID, panic on failure
func AssertUUID() string {
	uuid, err := UUID()
	if err != nil {
		panic(err)
	}
	return uuid
}

// Pretty print json-able content
func PrettyPrint(content any) string {
	b, err := json.MarshalIndent(content, "", "    ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func ListenTCP(conns chan<- *net.TCPConn) {
	addr, err := net.ResolveTCPAddr("tcp", Constants.LocalAddr)
	PanicErr(err)
	listen, err := net.ListenTCP("tcp", addr)
	PanicErr(err)
	defer listen.Close()
	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			fmt.Println("Failure to connect:", err.Error())
			continue
		}
		conns <- conn
	}
}
