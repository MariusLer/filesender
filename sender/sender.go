package sender

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// The sender is the server here

/*
func getExternalIP() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		fmt.Println("Error: ", err)
		return ""
	}
	defer resp.Body.Close()
	var ip string
	io.Copy(ip, resp.Body)
}
*/

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
	var ip = getMyIP()
	fmt.Println("Your ip is:", ip)
	var ipAndPort = ip + ":20000"
	ln, err := net.Listen("tcp", ipAndPort)
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
		fmt.Println("Put in absolute filepath")
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
