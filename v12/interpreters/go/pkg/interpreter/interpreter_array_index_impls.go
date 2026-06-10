package interpreter

func (i *Interpreter) noteIndexImplementation(interfaceName, typeName string, generic bool) {
	if i == nil {
		return
	}
	if !generic && normalizeKernelAliasName(typeName) != "Array" {
		return
	}
	switch interfaceName {
	case "Index", "able.core.interfaces.Index":
		i.arrayIndexImpls = true
	case "IndexMut", "able.core.interfaces.IndexMut":
		i.arrayIndexMutImpls = true
	}
}

func (i *Interpreter) canUseDirectArrayIndexGetFastPath() bool {
	return i != nil && !i.arrayIndexImpls
}

func (i *Interpreter) canUseDirectArrayIndexSetFastPath() bool {
	return i != nil && !i.arrayIndexMutImpls
}
