package coldmoon

type DataBlock struct {
	data []byte
}

func (db *DataBlock) Size() JSInt {
	return JSInt(len(db.data))
}

func (db *DataBlock) Equal(other *DataBlock) bool {
	return db == other
}

func (db *DataBlock) Slice(start JSInt, end JSInt) []byte {
	return db.data[start:end]
}

func (db *DataBlock) Set(index JSInt, value []byte) {
	copy(db.data[index:], value)
}

// 6.2.9.1
func CreateByteDataBlock(agent *Agent, size JSInt) *DataBlock {
	return &DataBlock{
		data: make([]byte, size),
	}
}

// 6.2.9.2
func CreateSharedByteDataBlock(agent *Agent, size JSInt) *DataBlock {
	return &DataBlock{
		data: make([]byte, size),
	}
}

// 6.2.9.3
func CopyDataBlockBytes(
	toBlock *DataBlock,
	toIndex JSInt,
	fromBlock *DataBlock,
	fromIndex JSInt,
	count JSInt,
) {
	copy(toBlock.data[toIndex:], fromBlock.data[fromIndex:fromIndex+count])
}
