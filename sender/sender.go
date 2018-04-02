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

	"github.com/mariusler/filesender/config"
	"github.com/mariusler/filesender/messages"
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

func sendFiles(files []string, conn net.Conn) {
	for _, file := range files {
		f, err := os.Open(file)
		var bytesSent int64
		defer f.Close()
		if err != nil {
			fmt.Println(err)
			continue
		}
		for {
			n, err := io.CopyN(conn, f, config.ChunkSize)
			bytesSent += n
			if err == io.EOF {
				fmt.Println("Reached EOF")
				break
			}
		}
		fmt.Println("Sent file", file, "Size:", bytesSent)
	}
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

func send(conn net.Conn) {
	defer conn.Close()
	var dir string
	var filepaths []string
	var err error
	for {
		for {
			fmt.Println("Put in absolute filepath for file or folder")
			fmt.Scanln(&dir)
			if isFolder(dir) && dir[len(dir)-1] != filepath.Separator {
				dir += "/"
			}
			filepaths, err = fileWalk(dir)
			if err != nil {
				fmt.Println("Error opening files", err)
				continue
			}
			dir = filepaths[0]
			break
		}
		// Add up sizes or file descriptions etc and ask for confirmations
		// This should prob be an own function
		var transMsg messages.TransferInfo
		folders, files := findDirectoriesAndFiles(filepaths)
		var absolutePathLenght = len(dir)
		if len(folders) > 0 {
			topFolder := strings.Split(folders[0], string(filepath.Separator))
			fmt.Println(topFolder)
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
			continue
		}
		conn.Write(bytes)
		buf := make([]byte, 8)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("wiuuuuuuuuuuuuut", err)
		}
		received := string(buf[:n])
		received = strings.ToLower(received)
		if received == "y" || received == "yes" {
			sendFiles(files, conn)
		}
		return
	}
}
