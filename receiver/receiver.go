package receiver

import (
	"fmt"
	"net"
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

}
