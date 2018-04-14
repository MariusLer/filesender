package messages

// TransferInfo contains info about the data to be sent
type TransferInfo struct {
	Files     []string
	Sizes     []int64
	TotalSize int64
}

// ProgressInfo used to print prog bar
type ProgressInfo struct {
	Progresses  [2]float32
	Currentfile string
}
