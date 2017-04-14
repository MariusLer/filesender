package main

import (
	"fmt"
	"strings"

	"github.com/mariusler/filesender/receiver"
	"github.com/mariusler/filesender/sender"
)

func main() {
	for {
		fmt.Println("Type s/send if you want to send a file or r/receive to receive")
		var input string
		fmt.Scanln(&input)
		input = strings.ToLower(input)
		if input == "s" || input == "send" {
			sender.Sender()
		} else if input == "r" || input == "receive" {
			receiver.Receiver()
		}
	}
}
