package coldmoon

type SharedArrayBufferObject struct {
	*Object
	ArrayBufferData          DataBlock
	ArrayBufferByteLength    JSInt
	ArrayBufferMaxByteLength JSInt
}
