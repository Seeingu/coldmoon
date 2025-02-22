package coldmoon

import "github.com/samber/lo"

type (
	Exports         []string
	ModuleNamespace struct {
		*Object
		// [[Module]]
		Module ModuleRecord
		// [[Exports]]
		Exports Exports
	}
)

func ModuleNamespaceCreate(agent *Agent, module *SourceTextModule, exports []string) ObjectType {
	Assert(module == nil)

	sortedExports := exports

	object := NewObject(agent, nil, "ModuleNamespace")
	M := &ModuleNamespace{
		Object:  object,
		Module:  module,
		Exports: sortedExports,
	}
	internalMethods := object.InternalMethods()
	internalMethods.GetPrototypeOf = func(o ObjectType) ObjectType {
		return nil
	}
	internalMethods.SetPrototypeOf = func(o ObjectType, v ObjectType) bool {
		return SetImmutablePrototype(o, v)
	}
	internalMethods.IsExtensible = func(o ObjectType) bool {
		return false
	}
	internalMethods.PreventExtensions = func(o ObjectType) bool {
		return true
	}
	internalMethods.GetOwnProperty = moduleNamespaceGetOwnProperty
	internalMethods.DefineOwnProperty = func(o ObjectType, p PropertyKey, desc *PropertyDescriptor) bool {
		if _, ok := p.(SymbolPropertyKey); ok {
			return OrdinaryDefineOwnProperty(o, p, desc)
		}
		current := o.InternalMethods().GetOwnProperty(o, p)
		if current == nil {
			return false
		}
		if desc.Configurable {
			return false
		}
		if !desc.Enumerable {
			return false
		}
		if desc.IsAccessorDescriptor() {
			return false
		}
		if !desc.Writable {
			return false
		}
		if !SameValue(desc.Value, current.Value) {
			return false
		}
		return true
	}
	internalMethods.HasProperty = func(o ObjectType, p PropertyKey) bool {
		if _, ok := p.(SymbolPropertyKey); ok {
			return OrdinaryHasProperty(o, p)
		}
		_exports := o.(*ModuleNamespace).Exports
		return lo.Contains(_exports, p.ToValue().String())
	}
	internalMethods.Get = func(o ObjectType, p PropertyKey, receiver Value) CompletionValue {
		return moduleNamespaceGet(agent, o, p, receiver)
	}
	internalMethods.Set = func(o ObjectType, p PropertyKey, v Value, receiver Value) bool {
		return false
	}
	internalMethods.Delete = func(o ObjectType, p PropertyKey) bool {
		if _, ok := p.(SymbolPropertyKey); ok {
			return OrdinaryDelete(o, p)
		}
		_exports := o.(*ModuleNamespace).Exports
		if !lo.Contains(_exports, p.ToValue().String()) {
			return false
		}
		return true
	}
	internalMethods.OwnPropertyKeys = func(o ObjectType) []PropertyKey {
		_exports := o.(*ModuleNamespace).Exports
		symbolKeys := OrdinaryOwnPropertyKeys(o.(*Object))
		for _, e := range _exports {
			symbolKeys = append(symbolKeys, NewStringPropertyKey(e))
		}
		return symbolKeys
	}

	object.defineToStringTag("Module")
	module.Namespace = M
	return M
}

func moduleNamespaceGetOwnProperty(o ObjectType, p PropertyKey) *PropertyDescriptor {
	if _, ok := p.(SymbolPropertyKey); ok {
		return OrdinaryGetOwnProperty(o, p)
	}
	exports := o.(*ModuleNamespace).Exports
	if !lo.Contains(exports, p.ToValue().String()) {
		return nil
	}
	value := ReturnAssertNormal(o.InternalMethods().Get(o, p, o.ToValue()))
	return &PropertyDescriptor{
		Value:        value,
		Writable:     true,
		Enumerable:   true,
		Configurable: false,
	}
}

func moduleNamespaceGet(agent *Agent, o ObjectType, p PropertyKey, receiver Value) CompletionValue {
	if _, ok := p.(SymbolPropertyKey); ok {
		return OrdinaryGet(o, p, receiver)
	}
	exports := o.(*ModuleNamespace).Exports
	if !lo.Contains(exports, p.ToValue().String()) {
		return UndefinedValue.ToCompletion()
	}
	m := o.(*ModuleNamespace).Module
	binding := m.(*SourceTextModule).ResolveExport(p.ToValue().String(), nil)
	targetModule := binding.Module.(*SourceTextModule)
	if binding.BindingName.Namespace {
		return GetModuleNamespace(agent, targetModule).ToValue().ToCompletion()
	}
	targetEnv := targetModule.Environment
	if targetEnv == nil {
		panic("ReferenceError")
	}
	bindingName := targetEnv.GetBindingValue(agent, binding.BindingName.String, true)
	return bindingName
}
