package sender

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// The sender is the server here

func getExternalIP() []byte {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		fmt.Println("Error: ", err)
		return nil
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	n, _ := io.Copy(&buf, resp.Body)
	return buf.Bytes()[0:n]
}

func getMyIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")

	return localAddr[0:idx]
}

// Sender called when we are the server
func Sender() {
	//var ip = getMyIP()
	var ipExt = getExternalIP()
	var addr net.TCPAddr
	var ip net.IP = ipExt
	addr.IP = ip
	addr.Port = 20000
	fmt.Println("Your ip is:", string(ip))
	//var ipAndPort = ip + ":20000"
	//ln, err := net.Listen("tcp", ipAndPort)
	fmt.Println(addr)
	ln, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error", err)
			continue
		}
		go send(conn)
	}
}

func send(conn net.Conn) {
	defer conn.Close()
	var filepath string
	var file *os.File
	for {
		fmt.Println("Put in absolute filepath or filename if you have the file in the same folder as the program")
		fmt.Scanln(&filepath)

		f, err := os.Open(strings.TrimSpace(filepath)) // removing whitespaces etc
		if err != nil {
			fmt.Println("Error", err)
		} else {
			file = f
			break
		}
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error", err)
	}

	filename := fileinfo.Name()

	conn.Write([]byte(filename))

	time.Sleep(time.Second)

	n, err := io.Copy(conn, file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sending complete", n, "bytes sent")
}
