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
	"github.com/mariusler/filesender/progressBar"
)

// Receiver calles when we are receivging a file
func Receiver() {
	conn := connectToServer()
	defer conn.Close()
	fmt.Println("Waiting for server")

	buf := make([]byte, 1048576) // 1 MiB buffer
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

func receiveFiles(msg messages.TransferInfo, conn net.Conn) { // Use sizes to display progress
	var progressInfo messages.ProgressInfo
	var totalReceivedBytes int64
	ticker := time.NewTicker(time.Millisecond * config.ProgressBarRefreshTime)
	for ind, file := range msg.Files {
		var receivedBytes int64
		fileSize := msg.Sizes[ind]
		f, err := os.Create(file)
		defer f.Close()
		if err != nil {
			fmt.Println(err)
			break
		}
		var n int64
		var copyErr error
		for {
			select {
			case <-ticker.C:
				progressInfo.Progresses[0] = float32(totalReceivedBytes) / float32(msg.TotalSize) * 100
				progressInfo.Progresses[1] = float32(receivedBytes) / float32(fileSize) * 100
				progressInfo.Currentfile = msg.Files[ind]
				go progressBar.PrintProgressBar(progressInfo)
			default: // Skip if ticker is not out
			}
			if (fileSize - receivedBytes) < config.ChunkSize {
				n, copyErr = io.CopyN(f, conn, (fileSize - receivedBytes)) // Onle read the remaining bytes, nothing more
				receivedBytes += n
				totalReceivedBytes += n
				if copyErr != nil {
					fmt.Println(err)
					break
				}
				break
			}
			n, copyErr = io.CopyN(f, conn, config.ChunkSize)
			receivedBytes += n
			totalReceivedBytes += n
			if copyErr != nil {
				fmt.Println(copyErr)
				break
			}
		}
	}
	time.Sleep(time.Millisecond)
	progressInfo.Progresses[0] = float32(totalReceivedBytes) / float32(msg.TotalSize) * 100
	progressInfo.Progresses[1] = float32(100)
	progressBar.PrintProgressBar(progressInfo)
	fmt.Println()
	fmt.Println("Done")
}

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

func sendProgressMessage(totalProgress float32, fileProgress float32, file string, progressCh chan<- messages.ProgressInfo) {
	var progressInfo messages.ProgressInfo
	progressInfo.Progresses[0] = totalProgress
	progressInfo.Progresses[1] = fileProgress
	progressInfo.Currentfile = file
	progressCh <- progressInfo
}
