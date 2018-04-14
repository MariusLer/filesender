package sender

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mariusler/filesender/config"
	"github.com/mariusler/filesender/messages"
	"github.com/mariusler/filesender/progressBar"
)

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
	var addr net.TCPAddr
	addr.IP = net.ParseIP(ip)
	addr.Port = config.Port
	fmt.Println("Your ip is:", ip)
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

func sendProgressMessage(totalProgress float32, fileProgress float32, file string, progressCh chan<- messages.ProgressInfo) {
	var progressInfo messages.ProgressInfo
	progressInfo.Progresses[0] = totalProgress
	progressInfo.Progresses[1] = fileProgress
	progressInfo.Currentfile = file
	progressCh <- progressInfo
}

func sendFiles(files []string, transferInfo messages.TransferInfo, conn net.Conn) {
	progressCh := make(chan messages.ProgressInfo)
	doneSendch := make(chan bool)
	donePrintCh := make(chan bool)
	var totalBytesSent int64
	go progressBar.PrintProgressBarTicker(progressCh, doneSendch, donePrintCh)
	for index := range files {
		fileSize := transferInfo.Sizes[index]
		f, err := os.Open(files[index])
		var fileBytesSent int64
		defer f.Close()
		if err != nil {
			fmt.Println(err)
			continue
		}
		var n int64
		var copyErr error
		var counter int
		for {
			counter++
			if (fileSize - fileBytesSent) < config.ChunkSize {
				n, copyErr = io.CopyN(conn, f, (fileSize - fileBytesSent)) // Onle write remaining bytes
				fileBytesSent += n
				totalBytesSent += n
				if copyErr != nil {
					fmt.Println(copyErr)
				}
				break
			}
			n, copyErr = io.CopyN(conn, f, config.ChunkSize)
			fileBytesSent += n
			totalBytesSent += n
			if copyErr != nil {
				fmt.Println(copyErr)
			}
			if counter == 40 {
				counter = 0
				sendProgressMessage(float32(totalBytesSent)/float32(transferInfo.TotalSize)*100, float32(fileBytesSent)/float32(fileSize)*100, transferInfo.Files[index], progressCh)
			}
		}
		sendProgressMessage(float32(totalBytesSent)/float32(transferInfo.TotalSize)*100, float32(100), transferInfo.Files[index], progressCh)
	}
	sendProgressMessage(float32(totalBytesSent)/float32(transferInfo.TotalSize)*100, float32(100), "", progressCh)
	doneSendch <- true
	time.Sleep(time.Millisecond * 5)
	<-donePrintCh
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
	for {
		fmt.Println("Put in absolute filepath for file or folder")
		fmt.Scanln(&dir)
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
	bytes, err := json.Marshal(transMsg)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.Write(bytes)
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
