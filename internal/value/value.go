package value

type ValueType byte

const (
	ValueTypeInt ValueType = iota
	ValueTypeFloat
	ValueTypeString
	ValueTypeSymbol
	ValueTypeFrame
	ValueTypeList
	ValueTypeFunc
	ValueTypeNil
	ValueTypeBool
)

func (vt ValueType) String() string {
	switch vt {
	case ValueTypeInt:
		return "Int"
	case ValueTypeFloat:
		return "Float"
	case ValueTypeString:
		return "String"
	case ValueTypeSymbol:
		return "Symbol"
	case ValueTypeFrame:
		return "Frame"
	case ValueTypeList:
		return "List"
	case ValueTypeFunc:
		return "Func"
	case ValueTypeNil:
		return "Nil"
	case ValueTypeBool:
		return "Bool"

	}
	return "error"
}

type Value interface {
	Type() ValueType
	Value() interface{}
	To(vt ValueType) (Value, error)
	Iterate() func() (Value, bool)
	Op(string, Value) (Value, bool)
}
