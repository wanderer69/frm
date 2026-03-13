package valuetypes

import (
	"fmt"
	"strconv"

	"github.com/wanderer69/frm/internal/bytecode"
	"github.com/wanderer69/frm/pkg/frame"
	"github.com/wanderer69/frm/pkg/list"
	"github.com/wanderer69/frm/pkg/value"
)

func Cmp(a value.Value, b value.Value) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch a.Type() {
	case value.ValueTypeInt:
		if a.Value().(int) == b.Value().(int) {
			return true
		}
	case value.ValueTypeFloat:
		if a.Value().(float64) == b.Value().(float64) {
			return true
		}
	case value.ValueTypeString:
		if a.Value().(string) == b.Value().(string) {
			return true
		}
	case value.ValueTypeSymbol:
		if a.Value().(string) == b.Value().(string) {
			return true
		}
	//case value.ValueTypeFrame:
	//case value.ValueTypeList:
	case value.ValueTypeNil:
		if a.Value().(string) == b.Value().(string) {
			return true
		}
	case value.ValueTypeBool:
		if a.Value().(bool) == b.Value().(bool) {
			return true
		}
	}
	return false
}

func OpUnary(op string, a value.Value) value.Value {
	switch a.Type() {
	case value.ValueTypeInt:
		switch op {
		case "inc":
			valueInt := a.Value().(ValueInt)
			valueInt.Int += 1
		case "dec":
			valueInt := a.Value().(ValueInt)
			valueInt.Int -= 1
		}
	case value.ValueTypeFloat:
		switch op {
		case "inc":
			valueFloat := a.Value().(ValueFloat)
			valueFloat.Float += 1
		case "dec":
			valueFloat := a.Value().(ValueFloat)
			valueFloat.Float -= 1
		}
	}
	return ValueNil
}

func Op(op string, a value.Value, b value.Value) value.Value {
	if a.Type() != b.Type() {
		return ValueNil
	}
	opsInt := func(op string, a int, b int) int {
		switch op {
		case "add":
			return a + b
		case "sub":
			return a - b
		case "mul":
			return a * b
		case "div":
			return a / b
		}
		return 0
	}
	opsFloat := func(op string, a float64, b float64) float64 {
		switch op {
		case "add":
			return a + b
		case "sub":
			return a - b
		case "mul":
			return a * b
		case "div":
			return a / b
		}
		return 0
	}
	switch a.Type() {
	case value.ValueTypeInt:
		switch b.Type() {
		case value.ValueTypeInt:
			return &ValueInt{Int: opsInt(op, a.Value().(int), b.Value().(int))}
		case value.ValueTypeFloat:
			return &ValueFloat{Float: opsFloat(op, float64(a.Value().(int)), b.Value().(float64))}
		}
	case value.ValueTypeFloat:
		switch b.Type() {
		case value.ValueTypeInt:
			return &ValueInt{Int: opsInt(op, int(a.Value().(float64)), b.Value().(int))}
		case value.ValueTypeFloat:
			return &ValueFloat{Float: opsFloat(op, float64(a.Value().(int)), b.Value().(float64))}
		}

	case value.ValueTypeString:
		if op == "add" {
			return &ValueString{String: a.Value().(string) + b.Value().(string)}
		}
	case value.ValueTypeSymbol:
		if op == "add" {
			return &ValueSymbol{Symbol: a.Value().(string) + b.Value().(string)}
		}
		//case value.ValueTypeFrame:
		//case value.ValueTypeList:
		//case value.ValueTypeNil:
		//case value.ValueTypeBool:
	}
	return ValueNil
}

type valueNil struct {
}

var ValueNil = &valueNil{}

func (v *valueNil) Type() value.ValueType {
	return value.ValueTypeNil
}

func (v *valueNil) Value() interface{} {
	return nil
}

func (v *valueNil) To(vt value.ValueType) (value.Value, error) {
	switch vt {
	//case value.ValueTypeInt:
	//case value.ValueTypeFloat:
	case value.ValueTypeString:
		return &ValueString{String: "nil"}, nil
	case value.ValueTypeSymbol:
		return &ValueSymbol{Symbol: "nil"}, nil
	//case value.ValueTypeFrame:
	//case value.ValueTypeList:
	case value.ValueTypeNil:
		return v, nil
		//case value.ValueTypeBool:
	}
	return nil, fmt.Errorf("bad value type")
}

func (v *valueNil) Iterate() func() (value.Value, bool) {
	return func() (value.Value, bool) { return nil, false }
}

func (v *valueNil) Op(op string, arg value.Value) (value.Value, bool) {
	switch op {
	case "cmp":
		if Cmp(v, arg) {
			return arg, true
		}
	}
	return ValueNil, false
}

type ValueBool struct {
	Bool bool
}

func (v *ValueBool) Type() value.ValueType {
	return value.ValueTypeBool
}

func (v *ValueBool) Value() interface{} {
	return v.Bool
}

func (v *ValueBool) To(vt value.ValueType) (value.Value, error) {
	switch vt {
	case value.ValueTypeInt:
		p := 0
		if v.Bool {
			p = 1
		}
		return &ValueInt{Int: p}, nil
	//case value.ValueTypeFloat:
	case value.ValueTypeString:
		return &ValueString{String: fmt.Sprintf("%v", v.Bool)}, nil
	case value.ValueTypeSymbol:
		return &ValueSymbol{Symbol: fmt.Sprintf("%v", v.Bool)}, nil
	//case value.ValueTypeFrame:
	//case value.ValueTypeList:
	//case value.ValueTypeNil:
	case value.ValueTypeBool:
		return v, nil
	}
	return nil, fmt.Errorf("bad value type")
}

func (v *ValueBool) Iterate() func() (value.Value, bool) {
	return func() (value.Value, bool) { return nil, false }
}

func (v *ValueBool) Op(op string, arg value.Value) (value.Value, bool) {
	switch op {
	case "cmp":
		if Cmp(v, arg) {
			return arg, true
		}
	}
	return ValueNil, false
}

type ValueInt struct {
	Int int
}

func (v *ValueInt) Type() value.ValueType {
	return value.ValueTypeInt
}

func (v *ValueInt) Value() interface{} {
	return v.Int
}

func (v *ValueInt) To(vt value.ValueType) (value.Value, error) {
	switch vt {
	case value.ValueTypeInt:
		return v, nil
	case value.ValueTypeFloat:
		return &ValueFloat{Float: float64(v.Int)}, nil
		//	case value.ValueTypeString:
		//	case value.ValueTypeSymbol:
	case value.ValueTypeString:
		return &ValueString{String: fmt.Sprintf("%v", v.Int)}, nil
	//case value.ValueTypeFrame:
	//case value.ValueTypeList:
	//case value.ValueTypeNil:
	case value.ValueTypeBool:
		if v.Int == 0 {
			return &ValueBool{Bool: false}, nil
		}
		return &ValueBool{Bool: true}, nil
	}
	return nil, fmt.Errorf("bad value type")
}

func (v *ValueInt) Iterate() func() (value.Value, bool) {
	pos := 0
	n := v.Int
	return func() (value.Value, bool) {
		isFinished := false
		if pos < n {
			isFinished = true
		}
		v := &ValueInt{
			Int: pos,
		}
		pos += 1
		return v, isFinished
	}
}

func (v *ValueInt) Op(op string, arg value.Value) (value.Value, bool) {
	switch op {
	case "cmp":
		if Cmp(v, arg) {
			return arg, true
		}
	}
	return ValueNil, false
}

type ValueFloat struct {
	Float float64
}

func (v *ValueFloat) Type() value.ValueType {
	return value.ValueTypeFloat
}

func (v *ValueFloat) Value() interface{} {
	return v.Float
}

func (v *ValueFloat) To(vt value.ValueType) (value.Value, error) {
	switch vt {
	case value.ValueTypeInt:
		return &ValueInt{Int: int(v.Float)}, nil
	case value.ValueTypeFloat:
		return v, nil
	case value.ValueTypeString:
		return &ValueString{String: fmt.Sprintf("%v", v.Float)}, nil
		//	case value.ValueTypeSymbol:
		//	case value.ValueTypeFrame:
		//	case value.ValueTypeList:
		//	case value.ValueTypeNil:
		//	case value.ValueTypeBool:
	}
	return nil, fmt.Errorf("bad value type")
}

func (v *ValueFloat) Iterate() func() (value.Value, bool) {
	pos := 0
	n := int(v.Float)
	return func() (value.Value, bool) {
		isFinished := false
		if pos < n {
			isFinished = true
		}
		v := &ValueInt{
			Int: pos,
		}
		pos += 1
		return v, isFinished
	}
}

func (v *ValueFloat) Op(op string, arg value.Value) (value.Value, bool) {
	switch op {
	case "cmp":
		if Cmp(v, arg) {
			return arg, true
		}
	}
	return ValueNil, false
}

type ValueString struct {
	String string
}

func (v *ValueString) Type() value.ValueType {
	return value.ValueTypeString
}

func (v *ValueString) Value() interface{} {
	return v.String
}

func (v *ValueString) To(vt value.ValueType) (value.Value, error) {
	switch vt {
	case value.ValueTypeInt:
		i, err := strconv.ParseInt(v.String, 10, 64)
		if err != nil {
			return nil, err
		}
		return &ValueInt{Int: int(i)}, nil
	case value.ValueTypeFloat:
		i, err := strconv.ParseFloat(v.String, 64)
		if err != nil {
			return nil, err
		}
		return &ValueFloat{Float: float64(i)}, nil
	case value.ValueTypeString:
		return v, nil
	case value.ValueTypeSymbol:
		return &ValueSymbol{Symbol: v.String}, nil
	case value.ValueTypeFrame:
	case value.ValueTypeList:
	case value.ValueTypeNil:
		if v.String != "nil" {
			return nil, fmt.Errorf("failed convert %v to nil", v.String)
		}
		return ValueNil, nil
	case value.ValueTypeBool:
		i, err := strconv.ParseBool(v.String)
		if err != nil {
			return nil, err
		}
		return &ValueBool{Bool: i}, nil
	}
	return nil, fmt.Errorf("bad value type")
}

func (v *ValueString) Iterate() func() (value.Value, bool) {
	pos := 0
	runes := []rune(v.String)
	n := len(runes)
	return func() (value.Value, bool) {
		isFinished := false
		if pos < n {
			isFinished = true
		}

		v := &ValueString{
			String: string(runes[pos]),
		}
		pos += 1
		return v, isFinished
	}
}

func (v *ValueString) Op(op string, arg value.Value) (value.Value, bool) {
	switch op {
	case "cmp":
		if Cmp(v, arg) {
			return arg, true
		}
	}
	return ValueNil, false
}

type ValueSymbol struct {
	Symbol string
}

func (v *ValueSymbol) Type() value.ValueType {
	return value.ValueTypeSymbol
}

func (v *ValueSymbol) Value() interface{} {
	return v.Symbol
}

func (v *ValueSymbol) To(vt value.ValueType) (value.Value, error) {
	switch vt {
	//	case value.ValueTypeInt:
	//	case value.ValueTypeFloat:
	case value.ValueTypeString:
		return &ValueString{String: v.Symbol}, nil
	case value.ValueTypeSymbol:
		return v, nil
		//	case value.ValueTypeFrame:
		//	case value.ValueTypeList:
	case value.ValueTypeNil:
		if v.Symbol != "nil" {
			return nil, fmt.Errorf("failed convert %v to nil", v.Symbol)
		}
		return ValueNil, nil
		//	case value.ValueTypeBool:
	}
	return nil, fmt.Errorf("bad value type")
}

func (v *ValueSymbol) Iterate() func() (value.Value, bool) {
	pos := 0
	runes := []rune(v.Symbol)
	n := len(runes)
	return func() (value.Value, bool) {
		isFinished := false
		if pos < n {
			isFinished = true
		}

		v := &ValueString{
			String: string(runes[pos]),
		}
		pos += 1
		return v, isFinished
	}
}

func (v *ValueSymbol) Op(op string, arg value.Value) (value.Value, bool) {
	switch op {
	case "cmp":
		if Cmp(v, arg) {
			return arg, true
		}
	}
	return ValueNil, false
}

type ValueFrame struct {
	Frame *frame.Frame
}

func (v *ValueFrame) Type() value.ValueType {
	return value.ValueTypeFrame
}

func (v *ValueFrame) Value() interface{} {
	return v.Frame
}

func (v *ValueFrame) To(vt value.ValueType) (value.Value, error) {
	switch vt {
	//	case value.ValueTypeInt:
	//	case value.ValueTypeFloat:
	case value.ValueTypeString:
		return &ValueString{String: v.Frame.String()}, nil
	//	case value.ValueTypeSymbol:
	case value.ValueTypeFrame:
		return v, nil
		//	case value.ValueTypeList:
		//	case value.ValueTypeNil:
		//	case value.ValueTypeBool:
	}
	return nil, fmt.Errorf("bad value type")
}

func (v *ValueFrame) Iterate() func() (value.Value, bool) {
	pos := 0
	//ss := v.Frame.Slots
	n := v.Frame.Len()
	return func() (value.Value, bool) {
		var vv *ValueList
		for {
			isFinished := true
			if pos < n {
				isFinished = false
				e := v.Frame.Item(pos)
				pos += 1
				if e == nil {
					continue
				}
				vv = &ValueList{
					List: e.Values,
				}
				return vv, isFinished
			}
			return vv, isFinished
		}
		/*
			for {
				isFinished := true
				if pos < n {
					isFinished = false
					e := v.Frame.Item(pos)
					pos += 1
					if e == nil {
						continue
					}
					vv = &ValueList{
						List: e.Values,
					}
					return vv, isFinished
				}
				return nil, isFinished
			}
		*/
	}
}

func (v *ValueFrame) Op(op string, arg value.Value) (value.Value, bool) {
	switch op {
	case "cmp":
		if Cmp(v, arg) {
			return arg, true
		}
	}
	return ValueNil, false
}

type ValueList struct {
	List *list.List
}

func (v *ValueList) Type() value.ValueType {
	return value.ValueTypeList
}

func (v *ValueList) Value() interface{} {
	return v.List
}

func (v *ValueList) To(vt value.ValueType) (value.Value, error) {
	switch vt {
	//	case value.ValueTypeInt:
	//	case value.ValueTypeFloat:
	case value.ValueTypeString:
		vv, err := v.List.String()
		if err != nil {
			return nil, err
		}
		return &ValueString{String: vv}, nil
		//	case value.ValueTypeSymbol:
		//	case value.ValueTypeFrame:
	case value.ValueTypeList:
		return v, nil
		//	case value.ValueTypeNil:
		//	case value.ValueTypeBool:
	}
	return nil, fmt.Errorf("bad value type")
}

func (v *ValueList) Iterate() func() (value.Value, bool) {
	vv := v.List.Root
	return func() (value.Value, bool) {
		isFinished := false
		if vv != nil {
			isFinished = true
		}
		vn := vv.Value
		vv = vv.Next
		return vn, isFinished
	}
}

func (v *ValueList) Op(op string, arg value.Value) (value.Value, bool) {
	switch op {
	case "cmp":
		if Cmp(v, arg) {
			return arg, true
		}
	}
	return ValueNil, false
}

type ValueFunc struct {
	Func *bytecode.Chunk
}

func (v *ValueFunc) Type() value.ValueType {
	return value.ValueTypeFunc
}

func (v *ValueFunc) Value() interface{} {
	return v.Func
}

func (v *ValueFunc) To(vt value.ValueType) (value.Value, error) {
	switch vt {
	case value.ValueTypeInt:
	case value.ValueTypeFloat:
	case value.ValueTypeString:
	case value.ValueTypeSymbol:
	case value.ValueTypeFrame:
	case value.ValueTypeList:
	case value.ValueTypeNil:
	case value.ValueTypeBool:
	}
	return nil, fmt.Errorf("bad value type")
}

func (v *ValueFunc) Iterate() func() (value.Value, bool) {
	pos := 0
	n := len(v.Func.Instructions)
	return func() (value.Value, bool) {
		isFinished := false
		if pos < n {
			isFinished = true
		}
		v := &ValueInt{
			Int: pos,
		}
		pos += 1
		return v, isFinished
	}
}

func (v *ValueFunc) Op(op string, arg value.Value) (value.Value, bool) {
	switch op {
	case "cmp":
		if Cmp(v, arg) {
			return arg, true
		}
	}
	return ValueNil, false
}
