package progressBar

import (
	"fmt"
	"time"

	"github.com/mariusler/filesender/config"
	"github.com/mariusler/filesender/messages"
)

const numPoints = 20

// PrintProgressBarTicker prints a prog bar with interval
func PrintProgressBarTicker(progressesCh <-chan messages.ProgressInfo, doneSendingCh <-chan bool, donePrintingCh chan<- bool) {
	var progressInfo messages.ProgressInfo
	ticker := time.NewTicker(config.ProgressBarRefreshTime)
	for {
		select {
		case msg := <-progressesCh:
			if msg.Currentfile != "" {
				progressInfo = msg
			} else {
				progressInfo.Progresses = msg.Progresses
			}
		case <-ticker.C:
			printProgressBar(progressInfo)
		case <-doneSendingCh:
			printProgressBar(progressInfo)
			donePrintingCh <- true
			ticker.Stop()
			return
		}
	}
}

func printProgressBar(progressInfo messages.ProgressInfo) {
	fmt.Print("Total progress:")
	for index := range progressInfo.Progresses {
		prog := int(float32(progressInfo.Progresses[index]/100*numPoints) + float32(0.5))
		fmt.Print("[")
		for i := 0; i < numPoints; i++ {
			if i+1 <= prog {
				fmt.Print("#")
			} else {
				fmt.Print(".")
			}
		}
		fmt.Print("] ")
		fmt.Print((int(progressInfo.Progresses[index])))
		fmt.Print("%" + " ")
		if index == 0 {
			fmt.Print("File:")
		} else if index == 1 {
			fmt.Print("name: " + progressInfo.Currentfile)
		}
	}
	fmt.Print("            \r")
}
