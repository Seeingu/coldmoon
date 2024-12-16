package coldmoon

type DataBlock []byte

func (db DataBlock) Size() JSInt {
	return JSInt(len(db))
}

func (db DataBlock) Equal(other DataBlock) bool {
	if len(db) != len(other) {
		return false
	}
	for i, v := range db {
		if v != other[i] {
			return false
		}
	}
	return true
}

// 6.2.9.1
func CreateByteDataBlock(agent *Agent, size int) DataBlock {
	db := make(DataBlock, size)
	return db
}

// 6.2.9.2
func CreateSharedByteDataBlock(agent *Agent, size JSInt) DataBlock {
	db := make(DataBlock, size)
	return db
}

// 6.2.9.3
func CopyDataBlockBytes(
	toBlock DataBlock,
	toIndex JSInt,
	fromBlock DataBlock,
	fromIndex JSInt,
	count JSInt,
) {
	copy(toBlock[toIndex:], fromBlock[fromIndex:fromIndex+count])
}
