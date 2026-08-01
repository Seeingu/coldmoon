package coldmoon

import "sync"

type DataBlock struct {
	mu   sync.RWMutex
	data []byte
}

func (db *DataBlock) Size() JSInt {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return JSInt(len(db.data))
}

func (db *DataBlock) Equal(other *DataBlock) bool {
	return db == other
}

func (db *DataBlock) Slice(start JSInt, end JSInt) []byte {
	db.mu.RLock()
	defer db.mu.RUnlock()
	result := make([]byte, end-start)
	copy(result, db.data[start:end])
	return result
}

func (db *DataBlock) Set(index JSInt, value []byte) {
	db.mu.Lock()
	defer db.mu.Unlock()
	copy(db.data[index:], value)
}

// AtomicModify applies one indivisible read-modify-write operation and
// returns a copy of the bytes observed before the update.
func (db *DataBlock) AtomicModify(index JSInt, size JSInt, modify func([]byte) []byte) []byte {
	db.mu.Lock()
	defer db.mu.Unlock()
	previous := make([]byte, size)
	copy(previous, db.data[index:index+size])
	replacement := modify(previous)
	copy(db.data[index:index+size], replacement)
	return previous
}

// AtomicCompareExchange compares and conditionally replaces one byte range,
// returning a copy of the bytes observed before the operation.
func (db *DataBlock) AtomicCompareExchange(index JSInt, expected, replacement []byte) []byte {
	db.mu.Lock()
	defer db.mu.Unlock()
	previous := make([]byte, len(expected))
	copy(previous, db.data[index:index+JSInt(len(expected))])
	matched := true
	for i := range expected {
		if previous[i] != expected[i] {
			matched = false
			break
		}
	}
	if matched {
		copy(db.data[index:index+JSInt(len(replacement))], replacement)
	}
	return previous
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
	bytes := fromBlock.Slice(fromIndex, fromIndex+count)
	toBlock.Set(toIndex, bytes)
}
