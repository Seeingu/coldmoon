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

	call := func(o ObjectType, this Value, arguments []Value) Value {
		b := o.(*BoundFunctionObject)
		_target := b.BoundTargetFunction
		_boundThis := b.BoundThis
		_boundArgs := b.BoundArguments
		_args := append(_boundArgs, arguments...)
		return (_target).ToValue().CallAssumeCallable(_boundThis, _args)
	}
	boundFunction.InternalMethods().Call = call
	if IsConstructor((target).ToValue()) {
		construct := func(o ObjectType, arguments []Value, newTarget ObjectType) ObjectType {
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
		boundFunction.InternalMethods().Construct = construct
	}
	return boundFunction
}
