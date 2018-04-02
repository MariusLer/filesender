package messages

// TransferInfo contains info about the data to be sent
type TransferInfo struct {
	Files     []string
	Sizes     []int64
	TotalSize int64
}
