package coldmoon

import (
	"sort"
	"unicode/utf16"

	"github.com/samber/lo"
)

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
	Assert(module != nil)
	Assert(module.Namespace == nil)

	sortedExports := append(Exports(nil), exports...)
	sortModuleExportNames(sortedExports)

	object := NewObject(agent, nil, "ModuleNamespace")
	object.defineToStringTag("Module")
	object.SetExtensible(false)
	M := &ModuleNamespace{
		Object:  object,
		Module:  module,
		Exports: sortedExports,
	}
	object.ref = M
	internalMethods := object.internalMethods()
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
	internalMethods.DefineOwnProperty = func(o ObjectType, p PropertyKey, desc *PropertyDescriptor) (co Completion[bool]) {
		if _, ok := p.(SymbolPropertyKey); ok {
			co.value = OrdinaryDefineOwnProperty(o.(*ModuleNamespace).Object, p, desc)
			return
		}
		current := o.internalMethods().GetOwnProperty(o, p)
		if current == nil {
			return
		}
		if desc.Configurable {
			return
		}
		if desc.EnumerableSet && !desc.Enumerable {
			return
		}
		if desc.IsAccessorDescriptor() {
			return
		}
		if desc.WritableSet && !desc.Writable {
			return
		}
		if desc.Value != nil && !SameValue(desc.Value, current.Value) {
			return
		}
		co.value = true
		return
	}
	internalMethods.HasProperty = func(o ObjectType, p PropertyKey) (co Completion[bool]) {
		if _, ok := p.(SymbolPropertyKey); ok {
			return OrdinaryHasProperty(o.(*ModuleNamespace).Object, p)
		}
		_exports := o.(*ModuleNamespace).Exports
		co.value = lo.Contains(_exports, p.ToValue().String())
		return
	}
	internalMethods.Get = func(o ObjectType, p PropertyKey, receiver Value) CompletionValue {
		return moduleNamespaceGet(agent, o, p, receiver)
	}
	internalMethods.Set = func(o ObjectType, p PropertyKey, v Value, receiver Value) (co Completion[bool]) {
		return
	}
	internalMethods.Delete = func(o ObjectType, p PropertyKey) (co Completion[bool]) {
		if _, ok := p.(SymbolPropertyKey); ok {
			co.value = OrdinaryDelete(o.(*ModuleNamespace).Object, p)
			return
		}
		_exports := o.(*ModuleNamespace).Exports
		if lo.Contains(_exports, p.ToValue().String()) {
			return
		}
		co.value = true
		return
	}
	internalMethods.OwnPropertyKeys = func(o ObjectType) []PropertyKey {
		_exports := o.(*ModuleNamespace).Exports
		keys := make([]PropertyKey, 0, len(_exports)+1)
		for _, e := range _exports {
			keys = append(keys, NewStringPropertyKey(e))
		}
		for _, key := range OrdinaryOwnPropertyKeys(o.(*ModuleNamespace).Object) {
			if _, ok := key.(SymbolPropertyKey); ok {
				keys = append(keys, key)
			}
		}
		return keys
	}

	module.Namespace = M
	return M
}

func sortModuleExportNames(exports []string) {
	sort.Slice(exports, func(i, j int) bool {
		left := utf16.Encode([]rune(exports[i]))
		right := utf16.Encode([]rune(exports[j]))
		limit := min(len(left), len(right))
		for index := 0; index < limit; index++ {
			if left[index] != right[index] {
				return left[index] < right[index]
			}
		}
		return len(left) < len(right)
	})
}

func moduleNamespaceGetOwnProperty(o ObjectType, p PropertyKey) *PropertyDescriptor {
	if _, ok := p.(SymbolPropertyKey); ok {
		return OrdinaryGetOwnProperty(o.(*ModuleNamespace).Object, p)
	}
	exports := o.(*ModuleNamespace).Exports
	if !lo.Contains(exports, p.ToValue().String()) {
		return nil
	}
	value := ReturnAssertNormal(o.internalMethods().Get(o, p, o.ToValue()))
	return &PropertyDescriptor{
		Value:           value,
		Writable:        true,
		WritableSet:     true,
		Enumerable:      true,
		EnumerableSet:   true,
		Configurable:    false,
		ConfigurableSet: true,
	}
}

func moduleNamespaceGet(agent *Agent, o ObjectType, p PropertyKey, receiver Value) CompletionValue {
	if _, ok := p.(SymbolPropertyKey); ok {
		return OrdinaryGet(o.(*ModuleNamespace).Object, p, receiver)
	}
	exports := o.(*ModuleNamespace).Exports
	if !lo.Contains(exports, p.ToValue().String()) {
		return UndefinedValue.ToCompletion()
	}
	m := o.(*ModuleNamespace).Module
	binding := m.(*SourceTextModule).ResolveExport(p.ToValue().String(), nil)
	Assert(binding != nil)
	Assert(!binding.IsNull())
	Assert(!binding.IsAmbiguous())
	Assert(binding.BindingName != nil)
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
