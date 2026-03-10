package vm

import (
	"fmt"

	"github.com/wanderer69/frm/internal/bytecode"
	"github.com/wanderer69/frm/internal/frame"
	"github.com/wanderer69/frm/internal/list"
	"github.com/wanderer69/frm/internal/proof"
	"github.com/wanderer69/frm/internal/value"
	valueType "github.com/wanderer69/frm/internal/value_types"
)

// internal
func Cmp(args []value.Value) ([]value.Value, error) {
	if len(args) != 2 {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("unexpected len args")
	}
	return []value.Value{&valueType.ValueBool{Bool: valueType.Cmp(args[1], args[0])}}, nil
}

func Add(args []value.Value) ([]value.Value, error) {
	if len(args) != 2 {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("unexpected len args")
	}
	return []value.Value{valueType.Op("add", args[1], args[0])}, nil
}

func Sub(args []value.Value) ([]value.Value, error) {
	if len(args) != 2 {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("unexpected len args")
	}
	return []value.Value{valueType.Op("sub", args[1], args[0])}, nil
}

func Mul(args []value.Value) ([]value.Value, error) {
	if len(args) != 2 {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("unexpected len args")
	}
	return []value.Value{valueType.Op("mul", args[1], args[0])}, nil
}

func Div(args []value.Value) ([]value.Value, error) {
	if len(args) != 2 {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("unexpected len args")
	}
	return []value.Value{valueType.Op("div", args[1], args[0])}, nil
}

func To(args []value.Value) ([]value.Value, error) {
	if len(args) != 2 {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("unexpected len args")
	}

	return nil, nil
}

func Type(args []value.Value) ([]value.Value, error) {
	if len(args) != 1 {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("unexpected len args")
	}
	return []value.Value{&valueType.ValueString{String: args[0].Type().String()}}, nil
}

func Inc(args []value.Value) ([]value.Value, error) {
	if len(args) != 1 {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("unexpected len args")
	}
	return []value.Value{valueType.OpUnary("inc", args[0])}, nil
}

func Dec(args []value.Value) ([]value.Value, error) {
	if len(args) != 1 {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("unexpected len args")
	}
	return []value.Value{valueType.OpUnary("dec", args[0])}, nil
}

func NewList(args []value.Value) ([]value.Value, error) {
	if len(args) != 0 {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("unexpected len args")
	}
	return []value.Value{&valueType.ValueList{List: &list.List{}}}, nil
}

func Append(args []value.Value) ([]value.Value, error) {
	if len(args) != 2 {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("unexpected len args")
	}
	if args[1].Type() != value.ValueTypeList {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("bad type")
	}
	l := args[1].Value().(*list.List)
	l.Add(args[0])
	return []value.Value{args[0]}, nil
}

func Prove(args []value.Value) ([]value.Value, error) {
	if len(args) != 2 {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("unexpected len args")
	}
	if args[1].Type() != value.ValueTypeList {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("bad type")
	}
	if args[0].Type() != value.ValueTypeList {
		return []value.Value{valueType.ValueNil}, fmt.Errorf("bad type")
	}

	l1 := args[1].Value().(*list.List)
	kb := proof.NewKnowledgeBaseByList(l1)
	l2 := args[0].Value().(*list.List)
	queryTemplates := []*valueType.ValueFrame{}
	l2.Find(func(valueFrames value.Value) (bool, bool) {
		if valueFrames.Type() == value.ValueTypeFrame {
			queryTemplates = append(queryTemplates, valueFrames.(*valueType.ValueFrame))
		}
		return false, false
	})
	ml := kb.Prove(queryTemplates)
	result := []value.Value{}
	for i := range ml {
		f := frame.NewFrame("")
		for k, v := range ml[i] {
			f.Slots.Put(k, v)
		}
		result = append(result, &valueType.ValueFrame{Frame: f})
	}
	return result, nil
}

//
// ====== VM ======
//

type Context struct {
	chunk  *bytecode.Chunk
	ip     int
	stack  []value.Value
	locals []value.Value
	iter   []func() (value.Value, bool)
}

type IntFunc func(args []value.Value) ([]value.Value, error)

type IntFuncContext struct {
	Name      string
	ArgLen    int
	ResultLen int
	Func      IntFunc
}

type VM struct {
	// global frame storage in memory
	frames []*frame.Frame

	context      *Context
	contexts     map[string]*Context
	contextStack []*Context

	internalFunctionsByName map[string]*IntFuncContext
}

func NewContext(chunk *bytecode.Chunk) *Context {
	return &Context{
		chunk:  chunk,
		stack:  make([]value.Value, 0, 256),
		locals: make([]value.Value, chunk.LocalNumbers),
		iter:   make([]func() (value.Value, bool), chunk.IterNumbers),
	}
}

func NewVM(chunks []*bytecode.Chunk) *VM {
	vm := &VM{
		frames:                  make([]*frame.Frame, 0),
		contexts:                make(map[string]*Context),
		internalFunctionsByName: make(map[string]*IntFuncContext),
	}
	for i := range chunks {
		context := NewContext(chunks[i])
		vm.contexts[chunks[i].Name] = context
		if chunks[i].Name == "main" {
			vm.context = context
		}
	}

	return vm
}

func (vm *VM) InitIntFunc() {
	vm.AddIntFunc("cmp", 2, 1, Cmp)
	vm.AddIntFunc("add", 2, 1, Add)
	vm.AddIntFunc("sub", 2, 1, Sub)
	vm.AddIntFunc("mul", 2, 1, Mul)
	vm.AddIntFunc("div", 2, 1, Div)
	vm.AddIntFunc("to", 2, 1, To)
	vm.AddIntFunc("type", 1, 1, Type)
	vm.AddIntFunc("inc", 1, 1, Inc)
	vm.AddIntFunc("dec", 1, 1, Dec)
	vm.AddIntFunc("list", 0, 1, NewList)
	vm.AddIntFunc("append", 2, 1, Append)
	vm.AddIntFunc("prove", 2, 1, Prove)
}

func (vm *VM) AddIntFunc(name string, argLen int, resultLen int, fn IntFunc) {
	inc := &IntFuncContext{
		Name:      name,
		ArgLen:    argLen,
		ResultLen: resultLen,
		Func:      fn,
	}
	vm.internalFunctionsByName[name] = inc
}

func (vm *VM) CallIntFunc(name string) bool {
	fn, ok := vm.internalFunctionsByName[name]
	if !ok {
		return false
	}
	args := make([]value.Value, fn.ArgLen)
	for i := 0; i < fn.ArgLen; i++ {
		args[i] = vm.pop()
	}
	results, err := fn.Func(args)
	if err != nil {

	}
	for i := range results {
		vm.push(results[i])
	}
	return true
}

func (vm *VM) push(v value.Value) {
	vm.context.stack = append(vm.context.stack, v)
}

func (vm *VM) pop() value.Value {
	n := len(vm.context.stack)
	v := vm.context.stack[n-1]
	vm.context.stack = vm.context.stack[:n-1]
	return v
}

func (vm *VM) pushContext() {
	vm.contextStack = append(vm.contextStack, vm.context)
}

func (vm *VM) popContext() {
	n := len(vm.contextStack)
	v := vm.contextStack[n-1]
	vm.contextStack = vm.contextStack[:n-1]
	vm.context = v
}

func truthy(v value.Value) bool {
	if v == nil {
		return false
	}
	switch v.Type() {
	case value.ValueTypeBool:
		return v.Value().(bool)
	default:
		return false
	}
}

func (vm *VM) Run() error {
	for vm.context.ip < len(vm.context.chunk.Instructions) {
		inst := vm.context.chunk.Instructions[vm.context.ip]
		fmt.Printf("%v: %v %v %v\r\n", vm.context.ip, inst.Op, inst.Arg1, inst.Arg2)
		vm.context.ip++
		switch inst.Op {
		case bytecode.OpConstant:
			vm.push(vm.context.chunk.Constants[inst.Arg1])
		case bytecode.OpPop:
			vm.pop()
		case bytecode.OpLoadLocal:
			vm.push(vm.context.locals[inst.Arg1])
		case bytecode.OpStoreLocal:
			vm.context.locals[inst.Arg1] = vm.pop()
		case bytecode.OpPrint:
			v := vm.pop()
			vv, err := v.To(value.ValueTypeString)
			if err != nil {
				fmt.Printf("Error %v\r\n", err)
			}
			fmt.Println(vv.Value().(string))
		case bytecode.OpFrameNew:
			n := int(inst.Arg1)
			f := frame.NewFrame("") // Frame{}
			for i := 0; i < n; i++ {
				val := vm.pop()
				name := vm.pop().Value().(string)
				//f[name] = val
				f.Slots.Put(name, val)
			}
			vm.frames = append(vm.frames, f)
			vm.push(&valueType.ValueFrame{Frame: f})
		case bytecode.OpJumpIfFalse:
			v := vm.pop()
			if truthy(v) {
				vm.context.ip = int(inst.Arg1)
			}
		case bytecode.OpJump:
			vm.context.ip = int(inst.Arg1)
		case bytecode.OpIterBeginFrames:
			//vm.iters[int(inst.Arg1)] = 0
			vv := vm.context.locals[inst.Arg1]
			vm.context.iter[inst.Arg2] = vv.Iterate()
			//vm.context.locals[inst.Arg1]
		case bytecode.OpIterNextFrames:
			vv := vm.context.iter[inst.Arg1]
			vvv, isFinished := vv()
			if isFinished {
				vm.push(&valueType.ValueBool{Bool: isFinished})
			} else {
				vm.push(vvv)
				vm.push(&valueType.ValueBool{Bool: isFinished})
			}
		case bytecode.OpFindFrames:
			n := int(inst.Arg2)
			f := frame.NewFrame("")
			for i := 0; i < n; i++ {
				val := vm.pop()
				name := vm.pop().Value().(string)
				//f[name] = val
				f.Slots.Put(name, val)
			}
			/*
				n = inst.Arg1

				for i := 0; i < n; i++ {
					val := vm.pop()
					name := vm.pop().(string)
					//f[name] = val
					f.Slots.Put(name, val)
				}

				for i := 0; i < n; i++ {
					val := vm.pop()
					name := vm.pop().(string)
					f[name] = val
				}
			*/
			vm.frames = append(vm.frames, f)
			vm.push(&valueType.ValueFrame{Frame: f})

		case bytecode.OpReturn:
			if len(vm.contextStack) == 0 {
				return nil
			}
			vm.popContext()

		case bytecode.OpCallFunc:
			name := vm.context.chunk.Constants[inst.Arg1]
			// find internal function
			isInt := vm.CallIntFunc(name.Value().(string))
			if isInt {
				continue
			}
			//			fmt.Printf("-> %v\r\n", name)
			c, ok := vm.contexts[name.Value().(string)]
			if !ok {
				return fmt.Errorf("unknown func name %v", name)
			}
			for i := 0; i < int(inst.Arg2); i++ {
				val := vm.pop()
				//				fmt.Printf("--> %v %v\r\n", c.chunk.sortedArgs[i], val)
				c.locals[c.chunk.SortedArgs[i]] = val
			}
			vm.pushContext()
			vm.context = c
		default:
			return fmt.Errorf("unknown opcode %d", inst.Op)
		}
	}
	return nil
}
