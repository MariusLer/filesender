package receiver

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mariusler/filesender/config"
	"github.com/mariusler/filesender/messages"
)

func connectToServer() net.Conn {
	var ip string
	for {
		fmt.Println("Type in the ip address")
		fmt.Scanln(&ip)
		var addr net.TCPAddr
		addr.Port = config.Port
		addr.IP = net.ParseIP(ip)
		var counter = 0
		for {
			conn, err := net.DialTCP("tcp", nil, &addr)
			if err != nil {
				fmt.Println("Error connecting to server", err)
				counter++
				if counter == 10 {
					fmt.Println("Timed out connecting")
					break
				}
				time.Sleep(time.Second)
			} else {
				return conn
			}
		}
	}
}

func printTransferInfo(msg messages.TransferInfo) {
	fmt.Println("Listing all files to be received")
	for ind := range msg.Files {
		fmt.Println("File:", msg.Files[ind], "Size:", msg.Sizes[ind], "Bytes")
	}
	fmt.Println("Totalsize:", msg.TotalSize, "Bytes")
}

func createFolders(files []string) {
	for _, file := range files {
		hierarchy := strings.Split(file, string(filepath.Separator))
		if len(hierarchy) > 0 {
			pathLength := len(file) - len(hierarchy[len(hierarchy)-1])
			os.MkdirAll(file[:pathLength], os.ModePerm)
		}
	}
}

func receiveFiles(msg messages.TransferInfo, conn net.Conn) { // Use sizes to display progress
	for ind, file := range msg.Files {
		var receivedBytes int64
		fileSize := msg.Sizes[ind]
		f, err := os.Create(file)
		defer f.Close()
		if err != nil {
			fmt.Println(err)
		}
		var n int64
		var copyErr error
		for {
			if (fileSize - receivedBytes) < config.ChunkSize {
				n, copyErr = io.CopyN(f, conn, (fileSize - receivedBytes)) // Onle read the remaining bytes, nothing more
				if copyErr != nil {
					fmt.Println(err)
				}
				receivedBytes += n
				break
			}
			n, copyErr = io.CopyN(f, conn, config.ChunkSize)
			if copyErr != nil {
				fmt.Println(copyErr)
			}
			receivedBytes += n
		}
		fmt.Println("Received file:", file, "Size:", receivedBytes)
	}
}

// Receiver calles when we are receivging a file
func Receiver() {
	conn := connectToServer()
	defer conn.Close()

	buf := make([]byte, 10240) // 10 KiB buffer
	var msg messages.TransferInfo
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
	}
	jsonErr := json.Unmarshal(buf[:n], &msg)
	if jsonErr != nil {
		fmt.Println(jsonErr)
	}
	printTransferInfo(msg)
	fmt.Println("Enter y or yes to receive these files")
	var input string
	fmt.Scanln(&input)
	input = strings.ToLower(input)
	if input != "y" && input != "yes" {
		fmt.Println("You decided not to receive the files")
		return
	}
	createFolders(msg.Files)
	conn.Write([]byte(input))
	receiveFiles(msg, conn)
}
