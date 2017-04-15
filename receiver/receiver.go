package receiver

import (
	"fmt"
	"io"
	"net"
	"os"
)

func connectToServer() net.Conn {
	var ip string
	for {
		fmt.Println("Type in the ip address")
		fmt.Scanln(&ip)
		var ipAndPort = ip + ":20000"
		conn, err := net.Dial("tcp", ipAndPort)
		if err != nil {
			fmt.Println("Error connecting to server")
		} else {
			return conn
		}
	}
}

// Receiver calles when we are receivging a file
func Receiver() {
	conn := connectToServer()
	defer conn.Close()

	buf := make([]byte, 25)
	n, _ := conn.Read(buf)

	filename := string(buf[0:n])

	fmt.Println(filename)
	newfile, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error", err)
	}
	defer newfile.Close()

	nb, _ := io.Copy(newfile, conn)
	fmt.Println("File received", nb, "Bytes received")
}
