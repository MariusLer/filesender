package utility

import (
	"fmt"
	"strconv"
)

// PreintBytesPrefix prints the size with a prefix
func PrintBytesPrefix(size int64) {
	switch {
	case size >= 1024*1024*1024:
		fmt.Printf("%.1f", float32(size)/float32(1024*1024*1024))
		fmt.Printf(" GiB")
	case size >= 1024*1024:
		fmt.Printf("%.1f", float32(size)/float32(1024*1024))
		fmt.Printf(" MiB")
	case size >= 1024:
		fmt.Printf("%.1f", float32(size)/float32(1024))
		fmt.Printf(" KiB")
	default:
		fmt.Printf(strconv.Itoa(int(size)))
		fmt.Printf(" B")
	}
}
