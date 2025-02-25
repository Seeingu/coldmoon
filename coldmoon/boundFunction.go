package coldmoon

type BoundFunctionObject struct {
	*Object
	BoundTargetFunction ObjectType
	BoundThis           Value
	BoundArguments      []Value
}

func (b *BoundFunctionObject) GetFunctionRealm() *Realm {
	return b.BoundTargetFunction.GetFunctionRealm()
}

func BoundFunctionCreate(agent *Agent, target ObjectType, this Value, args []Value) ObjectType {
	proto := target.InternalMethods().GetPrototypeOf(target)
	boundFunction := &BoundFunctionObject{
		Object:              NewObject(agent, proto, "BoundFunction"),
		BoundTargetFunction: target,
		BoundThis:           this,
		BoundArguments:      args,
	}

	call := func(o ObjectType, this Value, arguments []Value) CompletionValue {
		b := o.(*BoundFunctionObject)
		_target := b.BoundTargetFunction
		_boundThis := b.BoundThis
		_boundArgs := b.BoundArguments
		_args := append(_boundArgs, arguments...)
		return _target.Call(_boundThis, _args)
	}
	boundFunction.InternalMethods().Call = call
	if IsConstructor(target.ToValue()) {
		boundFunction.InternalMethods().Construct = func(o ObjectType, arguments []Value, newTarget ObjectType) Completion[ObjectType] {
			b := o.(*BoundFunctionObject)
			_target := b.BoundTargetFunction
			_boundArgs := b.BoundArguments
			_args := append(_boundArgs, arguments...)
			_newTarget := newTarget
			if _newTarget == o {
				_newTarget = _target
			}
			return target.Construct(_args, _newTarget)
		}
	}
	boundFunction.ref = boundFunction
	return boundFunction
}
