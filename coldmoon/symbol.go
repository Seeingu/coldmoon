package coldmoon

import "fmt"

type SymbolValue struct {
	Value
	Id          uint64
	Description string
}

var _ Value = (*SymbolValue)(nil)

func (s *SymbolValue) String() string {
	panic("TypeError")
}

func (s *SymbolValue) ToBoolean() bool {
	return true
}

// 20.4.3.3.1
func (s *SymbolValue) SymbolDescriptiveString() string {
	return fmt.Sprintf("Symbol(%s)", s.Description)
}

// MARK: - SymbolObject

type SymbolObject struct {
	*Object
	Data *SymbolValue
}

func NewSymbolObject(agent *Agent, s *SymbolValue, prototype ObjectType) *SymbolObject {
	symbolObject := &SymbolObject{
		Object: NewObject(agent, prototype),
		Data:   s,
	}

	return symbolObject
}

func NewSymbolConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		description := argumentsList[0]
		if newTarget != nil {
			panic("TypeError")
		}

		var descriptionString string
		if description != nil {
			descriptionString = description.String()
		}

		return agent.CreateSymbol(descriptionString)
	}

	object := CreateBuiltinFunction(realm.Agent, behavior, 0, "Symbol", builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		isConstructor: true,
	})

	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.SymbolPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinProperty(realm.Intrinsics.SymbolPrototype, "constructor",
		NewValueFromObject(object),
	)

	return object
}

func NewSymbolPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype)
	return object
}

func thisSymbolValue(v Value) *SymbolValue {
	if symbol, ok := v.(*SymbolValue); ok {
		return symbol
	}
	if object, ok := v.(*ObjectValue); ok {
		s, ok := object.Object.(*SymbolObject)
		if ok {
			return s.Data
		}
	}

	panic("TypeError")
}
