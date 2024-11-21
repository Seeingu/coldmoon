package coldmoon

type DataBlock []byte

// 6.2.9.1
func CreateByteDataBlock(agent *Agent, size int) DataBlock {
	db := make(DataBlock, size)
	return db
}

// 6.2.9.2
func CreateSharedByteDataBlock(agent *Agent, size int) DataBlock {
	db := make(DataBlock, size)
	return db
}

// 6.2.9.3
func CopyDataBlockBytes(
	toBlock *DataBlock,
	toIndex int,
	fromBlock *DataBlock,
	fromIndex int,
	count int,
) {
	copy((*toBlock)[toIndex:], (*fromBlock)[fromIndex:fromIndex+count])
}
