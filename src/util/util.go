package util

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type PeerConn struct {
	C *net.TCPConn
	R *bufio.Reader
	W *bufio.Writer
}

func NewPeerConn(c *net.TCPConn) PeerConn {
	return PeerConn{
		C: c,
		R: bufio.NewReader(c),
		W: bufio.NewWriter(c),
	}
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

// Retry reading a line from a bufio reader, exponential wait.
// Attempt delays begin at 100ms and multiply by 2.
// Max total runtime: 1 > 100ms, 2 > 300ms, 3 > 700ms, 4 > 1.5s, 5 > 3.1s, 6 > 6.3s,
// 7 > 12.7s, 8 > 25.5s, 9 > 51.1s, 10 > 102.3s
func RetryReadLine(pc PeerConn, attempts int) ([]byte, error) {
	defer pc.C.SetReadDeadline(time.Time{})
	delay := time.Duration(100) * time.Millisecond
	for i := 0; i < attempts; i++ {
		pc.C.SetReadDeadline(time.Now().Add(delay))
		data, err := pc.R.ReadBytes(byte('\n'))
		if err == nil {
			if len(data) > 0 {
				return data[:len(data)-1], nil
			} else {
				return data, nil
			}
		} else if (errors.Is(err, io.EOF) || errors.Is(err, os.ErrDeadlineExceeded)) &&
			i != attempts-1 {
			delay *= time.Duration(2)
			continue
		} else {
			return nil, err
		}
	}
	return nil, io.EOF
}

func ResolveDialTCP(addr string) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	c, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	return c, nil
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
