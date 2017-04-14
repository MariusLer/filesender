package sender

import (
	"fmt"
	"log"
	"net"
	"strings"
)

// The sender is the server here

func connListener(ip string) {

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
	var ip = getMyIP()
	var ipAndPort = ip + ":20000"
	fmt.Println(ipAndPort)
	ln, err := net.Listen("tcp", ipAndPort)
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}
	defer ln.Close()

}
