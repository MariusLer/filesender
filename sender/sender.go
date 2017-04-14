package sender

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
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

func connListener(newConnCh chan<- net.Conn) {
	var ip = getMyIP()
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
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			fmt.Println(err)
		}
		newConnCh <- conn
	}
}

func closeServer(closeServerCh chan<- bool) {
	for {
		var closeServer string
		fmt.Println("Type close to close connection")
		fmt.Scanln(&closeServer)
		closeServer = strings.ToLower(closeServer)
		if closeServer == "close" {
			closeServerCh <- true
			return
		}
	}
}

// Sender called when we are the server
func Sender() {
	newConnCh := make(chan net.Conn)
	closeServerCh := make(chan bool)

	go connListener(newConnCh)
	go closeServer(closeServerCh)

	for {
		select {
		case conn := <-newConnCh:
			go send(conn)
		case <-closeServerCh:
			return
		}
	}
}

func send(conn net.Conn) {
	defer conn.Close()
	var filepath string
	var file *os.File
	for {
		fmt.Println("Put in absolute filepath")
		fmt.Scanln(&filepath)

		f, err := os.Open(filepath)
		if err != nil {
			fmt.Println("Error opening file ", err)
		} else {
			file = f
			break
		}
	}
	fmt.Println(file)
}
