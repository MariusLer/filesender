package sender

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mariusler/filesender/config"
	"github.com/mariusler/filesender/messages"
	"github.com/mariusler/filesender/progressBar"
	"github.com/mariusler/filesender/utility"
)

// Sender called when we are the server
func Sender() {
	var externalIP = getExternalIP()
	var localIP = getLocalIP()
	var addr net.TCPAddr
	addr.Port = config.Port
	fmt.Println("Your local ip is:", localIP)
	fmt.Println("Your external ip is:", externalIP)
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
	filepaths, errr := getFilePathAndListFiles()
	for errr != nil {
		filepaths, errr = getFilePathAndListFiles()
	}
	dir := filepaths[0] // Top level folder

	// Add up sizes or file descriptions etc and ask for confirmations
	// This should prob be an own function
	var transMsg messages.TransferInfo
	folders, files := findDirectoriesAndFiles(filepaths)
	var absolutePathLenght = len(dir)
	if len(folders) > 0 {
		topFolder := strings.Split(folders[0], string(filepath.Separator))
		absolutePathLenght -= len(topFolder[len(topFolder)-2]) + 1
	} else {
		fullPath := strings.Split(files[0], string(filepath.Separator))
		absolutePathLenght -= len(fullPath[len(fullPath)-1])
	}

	fmt.Println("Listing files to be sent")
	for _, file := range files {
		relativePath := file[absolutePathLenght:] // find all subfolders after specified path
		fileInfo, err := os.Stat(file)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(relativePath)
		transMsg.Sizes = append(transMsg.Sizes, fileInfo.Size())
		transMsg.TotalSize += fileInfo.Size()
		transMsg.Files = append(transMsg.Files, relativePath)
	}
	fmt.Printf("Total size: ")
	utility.PrintBytesPrefix(transMsg.TotalSize)
	fmt.Println()

	bytes, err := json.Marshal(transMsg)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = conn.Write([]byte(strconv.Itoa(len(bytes))))
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = conn.Write(bytes)
	if err != nil {
		fmt.Println("Error reading response", err)
	}
	buf := make([]byte, 8)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading response", err)
	}
	received := string(buf[:n])
	received = strings.ToLower(received)
	if received == "y" || received == "yes" {
		sendFiles(files, transMsg, conn)
	}
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")

	return localAddr[0:idx]
}

func getExternalIP() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		fmt.Println("Error: ", err)
		return ""
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	n, _ := io.Copy(&buf, resp.Body)
	return string(buf.Bytes()[0:n])
}

func fileWalk(dir string) ([]string, error) {
	fileList := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return err
	})
	if err != nil {
		fmt.Println("Error", err)
	}
	return fileList, err
}

func sendFiles(files []string, transferInfo messages.TransferInfo, conn net.Conn) {
	var progressInfo messages.ProgressInfo
	var totalBytesSent int64
	ticker := time.NewTicker(time.Millisecond * config.ProgressBarRefreshTime)
	for index := range files {
		fileSize := transferInfo.Sizes[index]
		f, err := os.Open(files[index])
		var fileBytesSent int64
		defer f.Close()
		if err != nil {
			fmt.Println(err)
			continue
		}
		sendbuffer := make([]byte, config.ChunkSize)
		for {
			select {
			case <-ticker.C:
				progressInfo.Progresses[0] = float32(totalBytesSent) / float32(transferInfo.TotalSize) * 100
				progressInfo.Progresses[1] = float32(fileBytesSent) / float32(fileSize) * 100
				progressInfo.Currentfile = transferInfo.Files[index]
				progressBar.PrintProgressBar(progressInfo)
			default: // Skip if ticker is not out
			}
			bytesCopied, copyErr := f.Read(sendbuffer)
			if copyErr != nil && copyErr != io.EOF {
				fmt.Println(copyErr)
				break
			}
			n, err := conn.Write(sendbuffer[:bytesCopied])
			fileBytesSent += int64(n)
			totalBytesSent += int64(n)
			if err != nil {
				fmt.Println(err)
			}
			if copyErr == io.EOF {
				break
			}
		}
	}
	time.Sleep(time.Millisecond)
	progressInfo.Progresses[0] = float32(totalBytesSent) / float32(transferInfo.TotalSize) * 100
	progressInfo.Progresses[1] = float32(100)
	progressBar.PrintProgressBar(progressInfo)
	fmt.Println()
	fmt.Println("Done sending")
}

func isFolder(path string) bool {
	fileinfo, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if fileinfo.IsDir() == true {
		return true
	}
	return false
}

func findDirectoriesAndFiles(paths []string) ([]string, []string) {
	dirs := make([]string, 0)
	files := make([]string, 0)
	for _, item := range paths {
		if isFolder(item) == true {
			dirs = append(dirs, item)
		} else {
			files = append(files, item)
		}
	}
	return dirs, files
}

func getFilePathAndListFiles() ([]string, error) {
	var dir string
	var filepaths []string
	var err error
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Put in absolute filepath for file or folder")
		scanner.Scan()
		dir = filepath.Clean(scanner.Text())
		if isFolder(dir) && dir[len(dir)-1] != filepath.Separator {
			dir += string(filepath.Separator)
		}
		filepaths, err = fileWalk(dir)
		if err != nil {
			fmt.Println("Error opening files", err)
			continue
		}
		return filepaths, err
	}
}
