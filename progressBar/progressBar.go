package progressBar

import (
	"fmt"

	"github.com/mariusler/filesender/messages"
)

const numPoints = 20

// PrintProgressBar used to print a progress bar with prog and filename
func PrintProgressBar(progressInfo messages.ProgressInfo) {
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
	fmt.Print("                                 \r")
}
