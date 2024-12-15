package coldmoon

type IntrinsicName string

const (
	IntrinsicNameObjectPrototype                          IntrinsicName = "%Object.Prototype%"
	IntrinsicNameObject                                   IntrinsicName = "%Object%"
	IntrinsicNameFunctionPrototype                        IntrinsicName = "%Function.Prototype%"
	IntrinsicNameFunction                                 IntrinsicName = "%Function%"
	IntrinsicNameArray                                    IntrinsicName = "%Array%"
	IntrinsicNameArrayPrototype                           IntrinsicName = "%Array.Prototype%"
	IntrinsicNameArrayIteratorPrototype                   IntrinsicName = "%ArrayIteratorPrototype%"
	IntrinsicNameArrayPrototypeValues                     IntrinsicName = "%Array.prototype.values%"
	IntrinsicNameArrayBufferPrototype                     IntrinsicName = "%ArrayBuffer.prototype%"
	IntrinsicNameArrayBuffer                              IntrinsicName = "%ArrayBuffer%"
	IntrinsicNameBooleanPrototype                         IntrinsicName = "%Boolean.Prototype%"
	IntrinsicNameBoolean                                  IntrinsicName = "%Boolean%"
	IntrinsicNameStringPrototype                          IntrinsicName = "%String.Prototype%"
	IntrinsicNameString                                   IntrinsicName = "%String%"
	IntrinsicNameGeneratorFunction                        IntrinsicName = "%GeneratorFunction%"
	IntrinsicNameGeneratorFunctionPrototype               IntrinsicName = "%GeneratorFunction.prototype%"
	IntrinsicNameGeneratorFunctionPrototypePrototype      IntrinsicName = "%GeneratorFunction.prototype.prototype%"
	IntrinsicNameDataViewPrototype                        IntrinsicName = "%DataView.prototype%"
	IntrinsicNameDataView                                 IntrinsicName = "%DataView%"
	IntrinsicNameAsyncGeneratorFunction                   IntrinsicName = "%AsyncGeneratorFunction%"
	IntrinsicNameAsyncGeneratorFunctionPrototype          IntrinsicName = "%AsyncGeneratorFunction.prototype%"
	IntrinsicNameAsyncGeneratorFunctionPrototypePrototype IntrinsicName = "%AsyncGeneratorFunction.prototype.prototype%"
	IntrinsicNameAsyncIteratorPrototype                   IntrinsicName = "%AsyncIteratorPrototype%"
	IntrinsicNameAsyncFunction                            IntrinsicName = "%AsyncFunction%"
	IntrinsicNameAsyncFunctionPrototype                   IntrinsicName = "%AsyncFunction.prototype%"
	IntrinsicNamePromise                                  IntrinsicName = "%Promise%"
	IntrinsicNamePromisePrototype                         IntrinsicName = "%Promise.prototype%"
	IntrinsicNameDatePrototype                            IntrinsicName = "%Date.Prototype%"
	IntrinsicNameDate                                     IntrinsicName = "%Date%"
	IntrinsicNameStringIteratorPrototype                  IntrinsicName = "%StringIteratorPrototype%"
	IntrinsicNameNumberPrototype                          IntrinsicName = "%Number.Prototype%"
	IntrinsicNameNumber                                   IntrinsicName = "%Number%"
	IntrinsicNameSymbolPrototype                          IntrinsicName = "&Symbol.Prototype%"
	IntrinsicNameSymbol                                   IntrinsicName = "%Symbol%"
	IntrinsicNameBigIntPrototype                          IntrinsicName = "%BigInt.Prototype%"
	IntrinsicNameBigInt                                   IntrinsicName = "%BigInt%"
	IntrinsicNameJSON                                     IntrinsicName = "%JSON%"
	IntrinsicNameReflect                                  IntrinsicName = "%Reflect%"
	IntrinsicNameProxy                                    IntrinsicName = "%Proxy%"
	IntrinsicNameIteratorPrototype                        IntrinsicName = "%IteratorPrototype%"
	IntrinsicNameForInIteratorPrototype                   IntrinsicName = "%ForInIteratorPrototype%"
	IntrinsicNameMath                                     IntrinsicName = "%Math%"
	IntrinsicNameMap                                      IntrinsicName = "%Map%"
	IntrinsicNameMapPrototype                             IntrinsicName = "%Map.prototype%"
	IntrinsicNameMapIteratorPrototype                     IntrinsicName = "%MapIteratorPrototype%"
	IntrinsicNameSet                                      IntrinsicName = "%SetObject%"
	IntrinsicNameSetPrototype                             IntrinsicName = "%SetObject.prototype%"
	IntrinsicNameSetIteratorPrototype                     IntrinsicName = "%SetIteratorPrototype%"
	IntrinsicNameIntl                                     IntrinsicName = "%Intl%"
	IntrinsicNameTypedArray                               IntrinsicName = "%TypedArray%"
	IntrinsicNameTypedArrayPrototype                      IntrinsicName = "%TypedArray.prototype%"
	IntrinsicNameBigInt64Array                            IntrinsicName = "%BigInt64Array%"
	IntrinsicNameBigInt64ArrayPrototype                   IntrinsicName = "%BigInt64Array.prototype%"
	IntrinsicNameBigUint64Array                           IntrinsicName = "%BigUint64Array%"
	IntrinsicNameBigUint64ArrayPrototype                  IntrinsicName = "%BigUint64Array.prototype%"
	IntrinsicNameFloat32Array                             IntrinsicName = "%Float32Array%"
	IntrinsicNameFloat32ArrayPrototype                    IntrinsicName = "%Float32Array.prototype%"
	IntrinsicNameFloat64Array                             IntrinsicName = "%Float64Array%"
	IntrinsicNameFloat64ArrayPrototype                    IntrinsicName = "%Float64Array.prototype%"
	IntrinsicNameInt8Array                                IntrinsicName = "%Int8Array%"
	IntrinsicNameInt8ArrayPrototype                       IntrinsicName = "%Int8Array.prototype%"
	IntrinsicNameInt16Array                               IntrinsicName = "%Int16Array%"
	IntrinsicNameInt16ArrayPrototype                      IntrinsicName = "%Int16Array.prototype%"
	IntrinsicNameInt32Array                               IntrinsicName = "%Int32Array%"
	IntrinsicNameInt32ArrayPrototype                      IntrinsicName = "%Int32Array.prototype%"
	IntrinsicNameUint8Array                               IntrinsicName = "%Uint8Array%"
	IntrinsicNameUint8ArrayPrototype                      IntrinsicName = "%Uint8Array.prototype%"
	IntrinsicNameUint8ClampedArray                        IntrinsicName = "%Uint8ClampedArray%"
	IntrinsicNameUint8ClampedArrayPrototype               IntrinsicName = "%Uint8ClampedArray.prototype%"
	IntrinsicNameUint16Array                              IntrinsicName = "%Uint16Array%"
	IntrinsicNameUint16ArrayPrototype                     IntrinsicName = "%Uint16Array.prototype%"
	IntrinsicNameUint32Array                              IntrinsicName = "%Uint32Array%"
	IntrinsicNameUint32ArrayPrototype                     IntrinsicName = "%Uint32Array.prototype%"
	IntrinsicNameIsFinite                                 IntrinsicName = "%IsFinite%"
	IntrinsicNameIsNaN                                    IntrinsicName = "%isNaN%"
	IntrinsicNameEval                                     IntrinsicName = "%eval%"
	IntrinsicNameParseInt                                 IntrinsicName = "%parseInt%"
	IntrinsicNameParseFloat                               IntrinsicName = "%parseFloat%"
	IntrinsicNameDecodeURI                                IntrinsicName = "%decodeURI%"
	IntrinsicNameDecodeURIComponent                       IntrinsicName = "%decodeURIComponent%"
	IntrinsicNameEncodeURI                                IntrinsicName = "%encodeURI%"
	IntrinsicNameEncodeURIComponent                       IntrinsicName = "%encodeURIComponent%"
	IntrinsicNameRegExp                                   IntrinsicName = "%RegExp%"
	IntrinsicNameRegExpPrototype                          IntrinsicName = "%RegExp.prototype%"
	IntrinsicNameRegExpStringIteratorPrototype            IntrinsicName = "%RegExpStringIteratorPrototype%"
	IntrinsicNameEvalError                                IntrinsicName = "%EvalError%"
	IntrinsicNameEvalErrorPrototype                       IntrinsicName = "%EvalError.prototype%"
	IntrinsicNameRangeError                               IntrinsicName = "%RangeError%"
	IntrinsicNameRangeErrorPrototype                      IntrinsicName = "%RangeError.prototype%"
	IntrinsicNameReferenceError                           IntrinsicName = "%ReferenceError%"
	IntrinsicNameReferenceErrorPrototype                  IntrinsicName = "%ReferenceError.prototype%"
	IntrinsicNameSyntaxError                              IntrinsicName = "%SyntaxError%"
	IntrinsicNameSyntaxErrorPrototype                     IntrinsicName = "%SyntaxError.prototype%"
	IntrinsicNameTypeError                                IntrinsicName = "%TypeError%"
	IntrinsicNameTypeErrorPrototype                       IntrinsicName = "%TypeError.prototype%"
	IntrinsicNameURIError                                 IntrinsicName = "%URIError%"
	IntrinsicNameURIErrorPrototype                        IntrinsicName = "%URIError.prototype%"
	IntrinsicNameError                                    IntrinsicName = "%Error%"
	IntrinsicNameErrorPrototype                           IntrinsicName = "%Error.prototype%"
	IntrinsicNameThrowTypeError                           IntrinsicName = "%ThrowTypeError%"
	IntrinsicNameAggregateError                           IntrinsicName = "%AggregateError%"
	IntrinsicNameAggregateErrorPrototype                  IntrinsicName = "%AggregateError.prototype%"
)

type Intrinsics struct {
	// %Object.Prototype%
	ObjectPrototype ObjectType
	// %Object%
	ObjectConstructor ObjectType
	// %Function.Prototype%
	FunctionPrototype ObjectType
	// %Function%
	FunctionConstructor ObjectType
	// %Array%
	ArrayConstructor ObjectType
	// %Array.Prototype%
	ArrayPrototype ObjectType
	// %ArrayIteratorPrototype%
	ArrayIteratorPrototype ObjectType
	// %Array.prototype.values%
	ArrayPrototypeValues ObjectType
	// %ArrayBuffer.prototype%
	ArrayBufferPrototype ObjectType
	// %ArrayBuffer%
	ArrayBufferConstructor ObjectType
	// %Boolean.Prototype%
	BooleanPrototype *BooleanObject
	// %Boolean%
	BooleanConstructor ObjectType
	// %String.Prototype%
	StringPrototype *StringObject
	// %String%
	StringConstructor ObjectType
	// %GeneratorFunction%
	GeneratorFunctionConstructor ObjectType
	// %GeneratorFunction.prototype%
	GeneratorFunctionPrototype ObjectType
	// %GeneratorFunction.prototype.prototype%
	GeneratorFunctionPrototypePrototype ObjectType
	// %DataView.prototype%
	DataViewPrototype ObjectType
	// %DataView%
	DataViewConstructor ObjectType
	// %AsyncGeneratorFunction%
	AsyncGeneratorFunction ObjectType
	// %AsyncGeneratorFunction.prototype%
	AsyncGeneratorFunctionPrototype ObjectType
	// %AsyncGeneratorFunction.prototype.prototype%
	AsyncGeneratorFunctionPrototypePrototype ObjectType
	// %AsyncIteratorPrototype%
	AsyncIteratorPrototype ObjectType
	// %AsyncFunction%
	AsyncFunctionConstructor ObjectType
	// %AsyncFunction.prototype%
	AsyncFunctionPrototype ObjectType
	// %Promise%
	Promise ObjectType
	// %Promise.prototype%
	PromisePrototype ObjectType
	// %Date.Prototype%
	DatePrototype ObjectType
	// %Date%
	DateConstructor ObjectType
	// %StringIteratorPrototype%
	StringIteratorPrototype ObjectType
	// %Number.Prototype%
	NumberPrototype *NumberObject
	// %Number%
	NumberConstructor ObjectType
	// &Symbol.Prototype%
	SymbolPrototype ObjectType
	// %Symbol%
	SymbolConstructor ObjectType
	// %BigInt.Prototype%
	BigIntPrototype ObjectType
	// %BigInt%
	BigIntConstructor ObjectType
	// %JSON%
	JSON ObjectType
	// %Reflect%
	Reflect ObjectType
	// %Proxy%
	Proxy ObjectType
	// %IteratorPrototype%
	IteratorPrototype ObjectType
	// %ForInIteratorPrototype%
	ForInIteratorPrototype ObjectType
	// %Math%
	Math ObjectType
	// %Map%
	Map ObjectType
	// %Map.prototype%
	MapPrototype ObjectType
	// %MapIteratorPrototype%
	MapIteratorPrototype ObjectType
	// %SetObject%
	Set ObjectType
	// %SetObject.prototype%
	SetPrototype ObjectType
	// %SetIteratorPrototype%
	SetIteratorPrototype ObjectType
	// %Intl%
	Intl ObjectType
	// %TypedArray%
	TypedArrayConstructor ObjectType
	// %TypedArray.prototype%
	TypedArrayPrototype ObjectType
	// %BigInt64Array%
	BigInt64ArrayConstructor ObjectType
	// %BigInt64Array.prototype%
	BigInt64ArrayPrototype ObjectType
	// %BigUint64Array%
	BigUint64ArrayConstructor ObjectType
	// %BigUint64Array.prototype%
	BigUint64ArrayPrototype ObjectType
	// %Float32Array%
	Float32ArrayConstructor ObjectType
	// %Float32Array.prototype%
	Float32ArrayPrototype ObjectType
	// %Float64Array%
	Float64ArrayConstructor ObjectType
	// %Float64Array.prototype%
	Float64ArrayPrototype ObjectType
	// %Int8Array%
	Int8ArrayConstructor ObjectType
	// %Int8Array.prototype%
	Int8ArrayPrototype ObjectType
	// %Int16Array%
	Int16ArrayConstructor ObjectType
	// %Int16Array.prototype%
	Int16ArrayPrototype ObjectType
	// %Int32Array%
	Int32ArrayConstructor ObjectType
	// %Int32Array.prototype%
	Int32ArrayPrototype ObjectType
	// %Uint8Array%
	Uint8ArrayConstructor ObjectType
	// %Uint8Array.prototype%
	Uint8ArrayPrototype ObjectType
	// %Uint8ClampedArray%
	Uint8ClampedArrayConstructor ObjectType
	// %Uint8ClampedArray.prototype%
	Uint8ClampedArrayPrototype ObjectType
	// %Uint16Array%
	Uint16ArrayConstructor ObjectType
	// %Uint16Array.prototype%
	Uint16ArrayPrototype ObjectType
	// %Uint32Array%
	Uint32ArrayConstructor ObjectType
	// %Uint32Array.prototype%
	Uint32ArrayPrototype ObjectType
	// %IsFinite%
	IsFinite ObjectType
	// %isNaN%
	IsNaN ObjectType
	// %eval%
	Eval ObjectType
	// %parseInt%
	ParseInt ObjectType
	// %parseFloat%
	ParseFloat ObjectType
	// %decodeURI%
	DecodeURI ObjectType
	// %decodeURIComponent%
	DecodeURIComponent ObjectType
	// %encodeURI%
	EncodeURI ObjectType
	// %encodeURIComponent%
	EncodeURIComponent ObjectType
	// %RegExp%
	RegExpConstructor ObjectType
	// %RegExp.prototype%
	RegExpPrototype ObjectType
	// %RegExpStringIteratorPrototype%
	RegExpStringIteratorPrototype ObjectType
	// %EvalError%
	EvalErrorConstructor ObjectType
	// %EvalError.prototype%
	EvalErrorPrototype ObjectType
	// %RangeError%
	RangeErrorConstructor ObjectType
	// %RangeError.prototype%
	RangeErrorPrototype ObjectType
	// %ReferenceError%
	ReferenceErrorConstructor ObjectType
	// %ReferenceError.prototype%
	ReferenceErrorPrototype ObjectType
	// %SyntaxError%
	SyntaxErrorConstructor ObjectType
	// %SyntaxError.prototype%
	SyntaxErrorPrototype ObjectType
	// %TypeError%
	TypeErrorConstructor ObjectType
	// %TypeError.prototype%
	TypeErrorPrototype ObjectType
	// %URIError%
	URIErrorConstructor ObjectType
	// %URIError.prototype%
	URIErrorPrototype ObjectType
	// %Error%
	ErrorConstructor ObjectType
	// %Error.prototype%
	ErrorPrototype ObjectType
	// %ThrowTypeError%
	ThrowTypeError ObjectType
	// %AggregateError%
	AggregateErrorConstructor ObjectType
	// %AggregateError.prototype%
	AggregateErrorPrototype ObjectType
}

func (i *Intrinsics) Get(key IntrinsicName) ObjectType {
	switch key {
	case IntrinsicNameObjectPrototype:
		return i.ObjectPrototype
	case IntrinsicNameObject:
		return i.ObjectConstructor
	case IntrinsicNameFunctionPrototype:
		return i.FunctionPrototype
	case IntrinsicNameFunction:
		return i.FunctionConstructor
	case IntrinsicNameArray:
		return i.ArrayConstructor
	case IntrinsicNameArrayPrototype:
		return i.ArrayPrototype
	case IntrinsicNameArrayIteratorPrototype:
		return i.ArrayIteratorPrototype
	case IntrinsicNameArrayPrototypeValues:
		return i.ArrayPrototypeValues
	case IntrinsicNameArrayBufferPrototype:
		return i.ArrayBufferPrototype
	case IntrinsicNameArrayBuffer:
		return i.ArrayBufferConstructor
	case IntrinsicNameBooleanPrototype:
		return i.BooleanPrototype
	case IntrinsicNameBoolean:
		return i.BooleanConstructor
	case IntrinsicNameStringPrototype:
		return i.StringPrototype
	case IntrinsicNameString:
		return i.StringConstructor
	case IntrinsicNameGeneratorFunction:
		return i.GeneratorFunctionConstructor
	case IntrinsicNameGeneratorFunctionPrototype:
		return i.GeneratorFunctionPrototype
	case IntrinsicNameGeneratorFunctionPrototypePrototype:
		return i.GeneratorFunctionPrototypePrototype
	case IntrinsicNameDataViewPrototype:
		return i.DataViewPrototype
	case IntrinsicNameDataView:
		return i.DataViewConstructor
	case IntrinsicNameAsyncGeneratorFunction:
		return i.AsyncGeneratorFunction
	case IntrinsicNameAsyncGeneratorFunctionPrototype:
		return i.AsyncGeneratorFunctionPrototype
	case IntrinsicNameAsyncGeneratorFunctionPrototypePrototype:
		return i.AsyncGeneratorFunctionPrototypePrototype
	case IntrinsicNameAsyncIteratorPrototype:
		return i.AsyncIteratorPrototype
	case IntrinsicNameAsyncFunction:
		return i.AsyncFunctionConstructor
	case IntrinsicNameAsyncFunctionPrototype:
		return i.AsyncFunctionPrototype
	case IntrinsicNamePromise:
		return i.Promise
	case IntrinsicNamePromisePrototype:
		return i.PromisePrototype
	case IntrinsicNameDatePrototype:
		return i.DatePrototype
	case IntrinsicNameDate:
		return i.DateConstructor
	case IntrinsicNameStringIteratorPrototype:
		return i.StringIteratorPrototype
	case IntrinsicNameNumberPrototype:
		return i.NumberPrototype
	case IntrinsicNameNumber:
		return i.NumberConstructor
	case IntrinsicNameSymbolPrototype:
		return i.SymbolPrototype
	case IntrinsicNameSymbol:
		return i.SymbolConstructor
	case IntrinsicNameBigIntPrototype:
		return i.BigIntPrototype
	case IntrinsicNameBigInt:
		return i.BigIntConstructor
	case IntrinsicNameJSON:
		return i.JSON
	case IntrinsicNameReflect:
		return i.Reflect
	case IntrinsicNameProxy:
		return i.Proxy
	case IntrinsicNameIteratorPrototype:
		return i.IteratorPrototype
	case IntrinsicNameForInIteratorPrototype:
		return i.ForInIteratorPrototype
	case IntrinsicNameMath:
		return i.Math
	case IntrinsicNameMap:
		return i.Map
	case IntrinsicNameMapPrototype:
		return i.MapPrototype
	case IntrinsicNameMapIteratorPrototype:
		return i.MapIteratorPrototype
	case IntrinsicNameSet:
		return i.Set
	case IntrinsicNameSetPrototype:
		return i.SetPrototype
	case IntrinsicNameSetIteratorPrototype:
		return i.SetIteratorPrototype
	case IntrinsicNameIntl:
		return i.Intl
	case IntrinsicNameTypedArray:
		return i.TypedArrayConstructor
	case IntrinsicNameTypedArrayPrototype:
		return i.TypedArrayPrototype
	case IntrinsicNameBigInt64Array:
		return i.BigInt64ArrayConstructor
	case IntrinsicNameBigInt64ArrayPrototype:
		return i.BigInt64ArrayPrototype
	case IntrinsicNameBigUint64Array:
		return i.BigUint64ArrayConstructor
	case IntrinsicNameBigUint64ArrayPrototype:
		return i.BigUint64ArrayPrototype
	case IntrinsicNameFloat32Array:
		return i.Float32ArrayConstructor
	case IntrinsicNameFloat32ArrayPrototype:
		return i.Float32ArrayPrototype
	case IntrinsicNameFloat64Array:
		return i.Float64ArrayConstructor
	case IntrinsicNameFloat64ArrayPrototype:
		return i.Float64ArrayPrototype
	case IntrinsicNameInt8Array:
		return i.Int8ArrayConstructor
	case IntrinsicNameInt8ArrayPrototype:
		return i.Int8ArrayPrototype
	case IntrinsicNameInt16Array:
		return i.Int16ArrayConstructor
	case IntrinsicNameInt16ArrayPrototype:
		return i.Int16ArrayPrototype
	case IntrinsicNameInt32Array:
		return i.Int32ArrayConstructor
	case IntrinsicNameInt32ArrayPrototype:
		return i.Int32ArrayPrototype
	case IntrinsicNameUint8Array:
		return i.Uint8ArrayConstructor
	case IntrinsicNameUint8ArrayPrototype:
		return i.Uint8ArrayPrototype
	case IntrinsicNameUint8ClampedArray:
		return i.Uint8ClampedArrayConstructor
	case IntrinsicNameUint8ClampedArrayPrototype:
		return i.Uint8ClampedArrayPrototype
	case IntrinsicNameUint16Array:
		return i.Uint16ArrayConstructor
	case IntrinsicNameUint16ArrayPrototype:
		return i.Uint16ArrayPrototype
	case IntrinsicNameUint32Array:
		return i.Uint32ArrayConstructor
	case IntrinsicNameUint32ArrayPrototype:
		return i.Uint32ArrayPrototype
	case IntrinsicNameIsFinite:
		return i.IsFinite
	case IntrinsicNameIsNaN:
		return i.IsNaN
	case IntrinsicNameEval:
		return i.Eval
	case IntrinsicNameParseInt:
		return i.ParseInt
	case IntrinsicNameParseFloat:
		return i.ParseFloat
	case IntrinsicNameDecodeURI:
		return i.DecodeURI
	case IntrinsicNameDecodeURIComponent:
		return i.DecodeURIComponent
	case IntrinsicNameEncodeURI:
		return i.EncodeURI
	case IntrinsicNameEncodeURIComponent:
		return i.EncodeURIComponent
	case IntrinsicNameRegExp:
		return i.RegExpConstructor
	case IntrinsicNameRegExpPrototype:
		return i.RegExpPrototype
	case IntrinsicNameRegExpStringIteratorPrototype:
		return i.RegExpStringIteratorPrototype
	case IntrinsicNameEvalError:
		return i.EvalErrorConstructor
	case IntrinsicNameEvalErrorPrototype:
		return i.EvalErrorPrototype
	case IntrinsicNameRangeError:
		return i.RangeErrorConstructor
	case IntrinsicNameRangeErrorPrototype:
		return i.RangeErrorPrototype
	case IntrinsicNameReferenceError:
		return i.ReferenceErrorConstructor
	case IntrinsicNameReferenceErrorPrototype:
		return i.ReferenceErrorPrototype
	case IntrinsicNameSyntaxError:
		return i.SyntaxErrorConstructor
	case IntrinsicNameSyntaxErrorPrototype:
		return i.SyntaxErrorPrototype
	case IntrinsicNameTypeError:
		return i.TypeErrorConstructor
	case IntrinsicNameTypeErrorPrototype:
		return i.TypeErrorPrototype
	case IntrinsicNameURIError:
		return i.URIErrorConstructor
	case IntrinsicNameURIErrorPrototype:
		return i.URIErrorPrototype
	case IntrinsicNameError:
		return i.ErrorConstructor
	case IntrinsicNameErrorPrototype:
		return i.ErrorPrototype
	case IntrinsicNameThrowTypeError:
		return i.ThrowTypeError
	case IntrinsicNameAggregateError:
		return i.AggregateErrorConstructor
	case IntrinsicNameAggregateErrorPrototype:
		return i.AggregateErrorPrototype
	}
	panic("unreachable")
}
