package coldmoon

type ForInIterator struct {
	*Object
	// [[Object]]
	Obj ObjectType
	// [[ObjectWasVisited]]
	ObjectWasVisited bool
	// [[VisitedKeys]]
	VisitedKeys map[PropertyKey]bool
	// [[RemainingKeys]]
	RemainingKeys []PropertyKey
	Done          bool
}

// MARK: - Dispatch

// 14.7.5.10.1
func CreateForInIterator(agent *Agent, obj ObjectType) *ForInIterator {
	realm := agent.CurrentRealm()
	iterator := &ForInIterator{
		Object:      NewObject(agent, realm.Intrinsics.ForInIteratorPrototype, "ForInIterator"),
		Obj:         obj,
		VisitedKeys: make(map[PropertyKey]bool),
	}
	iterator.ref = iterator
	return iterator
}

func NewForInIteratorPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.IteratorPrototype, "ForInIteratorPrototype")
	var next BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		iterator := MustGetObject(thisValue).(*ForInIterator)
		if iterator.Done {
			return (CreateIterResultObject(agent, UndefinedValue, true)).ToValue()
		}
		obj := iterator.Obj
		for {
			if !iterator.ObjectWasVisited {
				keys := obj.InternalMethods().OwnPropertyKeys(obj)
				for _, key := range keys {
					if _, ok := key.(SymbolPropertyKey); !ok {
						iterator.RemainingKeys = append(iterator.RemainingKeys, key)
					}
				}
				iterator.ObjectWasVisited = true
			}

			for len(iterator.RemainingKeys) != 0 {
				r := iterator.RemainingKeys[0]
				iterator.RemainingKeys = iterator.RemainingKeys[1:]
				if _, ok := iterator.VisitedKeys[r]; !ok {
					// Let desc be ? object.[[GetOwnProperty]](r).
					desc := obj.InternalMethods().GetOwnProperty(obj, r)
					if desc != nil {
						iterator.VisitedKeys[r] = true
						if desc.Enumerable {
							return (CreateIterResultObject(agent, r.ToValue(), false)).ToValue()
						}
					}
				}
			}

			obj = obj.InternalMethods().GetPrototypeOf(obj)
			if obj == nil {
				iterator.Done = true
				return (CreateIterResultObject(agent, UndefinedValue, true)).ToValue()
			}
			iterator.Obj = obj
			iterator.ObjectWasVisited = false
		}
	}
	DefineBuiltinFunction(object, "next", next, 0, realm)

	return object
}
