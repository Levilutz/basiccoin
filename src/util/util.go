package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
)

// Aggregate channels
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

// Generates base64(json(content))
func JsonB64(content any) ([]byte, error) {
	bodyJson, err := json.Marshal(content)
	if err != nil {
		return []byte{}, err
	}
	body := make([]byte, base64.StdEncoding.EncodedLen(len(bodyJson)))
	base64.StdEncoding.Encode(body, bodyJson)
	return body, nil
}

// Recovers unjson(unb64(body))
func UnJsonB64[R any](body []byte) (R, error) {
	var content R
	bodyJson := make([]byte, base64.StdEncoding.DecodedLen(len(body)))
	n, err := base64.StdEncoding.Decode(bodyJson, body)
	if err != nil {
		return content, err
	}
	err = json.Unmarshal(bodyJson[:n], &content)
	return content, err
}

// Pretty print json-able content
func PrettyPrint(content any) {
	b, err := json.MarshalIndent(content, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
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
