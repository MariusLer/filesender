package sender

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// The sender is the server here

func connListener(ip string) {

}

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
	var ipAndPort = ip + ":20000"

	var filepath string
	fmt.Println("Put in absolute filepath")
	fmt.Scanln(&filepath)

	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Error opening file ", err)
		return
	}

	ln, err := net.Listen("tcp", ipAndPort)
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}
	defer ln.Close()

	for {
		conn, err = ln.Accept()
	}
}

func sendfile(conn net.Conn, file os.File) {

}
