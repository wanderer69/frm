package bytecode

import (
	"encoding/gob"
	"fmt"
	"os"
	"strings"

	"github.com/wanderer69/frm/pkg/value"
)

//
// ====== Байткод ======
//

type OpCode byte

const (
	OpConstant OpCode = iota
	OpPop
	OpLoadLocal
	OpStoreLocal
	OpPrint
	OpFrameNew
	OpJumpIfFalse
	OpJump
	OpIterBeginFrames
	OpIterNextFrames
	OpReturn
	OpFindFrames
	OpUpdateFrames
	OpCallFunc
)

func (oc OpCode) String() string {
	s := ""
	switch oc {
	case OpConstant:
		s = "OpConstant"
	case OpPop:
		s = "OpPop"
	case OpLoadLocal:
		s = "OpLoadLocal"
	case OpStoreLocal:
		s = "OpStoreLocal"
	case OpPrint:
		s = "OpPrint"
	case OpFrameNew:
		s = "OpFrameNew"
	case OpJumpIfFalse:
		s = "OpJumpIfFalse"
	case OpJump:
		s = "OpJump"
	case OpIterBeginFrames:
		s = "OpIterBeginFrames"
	case OpIterNextFrames:
		s = "OpIterNextFrames"
	case OpReturn:
		s = "OpReturn"
	case OpFindFrames:
		s = "OpFindFrames"
	case OpUpdateFrames:
		s = "OpUpdateFrames"
	case OpCallFunc:
		s = "OpCallFunc"
	}
	return s
}

type Instruction struct {
	Arg1 int32
	Arg2 int32
	Op   OpCode
}

func (in *Instruction) String() string {
	return fmt.Sprintf("%v %v %v", in.Op, in.Arg1, in.Arg2)
}

type Chunk struct {
	Name         string
	Instructions []Instruction
	Constants    []value.Value
	Locals       map[string]int
	SortedArgs   []int
	//Iters []func() (value.Value, bool)
	IterNumbers  int
	LocalNumbers int
}

func NewChunk() *Chunk {
	return &Chunk{
		Locals: make(map[string]int),
	}
}

func (c *Chunk) AddConst(v value.Value) int {
	c.Constants = append(c.Constants, v)
	return len(c.Constants) - 1
}

func (c *Chunk) Emit(op OpCode, arg1 int) int {
	c.Instructions = append(c.Instructions, Instruction{Op: op, Arg1: int32(arg1)})
	return len(c.Instructions) - 1
}

func (c *Chunk) EmitExt(op OpCode, arg1 int, arg2 int) int {
	c.Instructions = append(c.Instructions, Instruction{Op: op, Arg1: int32(arg1), Arg2: int32(arg2)})
	return len(c.Instructions) - 1
}

//
// ====== Сериализация байткода ======
//

func SaveChunks(path string, chunks []*Chunk) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	return enc.Encode(chunks)
}

func LoadChunks(path string) ([]*Chunk, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var chunks []*Chunk
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&chunks); err != nil {
		return nil, err
	}
	return chunks, nil
}

func PrintChunk(chunk *Chunk) string {
	ss := fmt.Sprintf("%v\r\n", chunk.Name)
	sl := []string{}
	for i := range chunk.Instructions {
		s := fmt.Sprintf("%d: %s %d %d", i, chunk.Instructions[i].Op, chunk.Instructions[i].Arg1, chunk.Instructions[i].Arg2)
		sl = append(sl, s)
	}
	ss += strings.Join(sl, "\r\n") + "\r\n"
	sl = []string{}
	for k, v := range chunk.Locals {
		s := fmt.Sprintf("%v: %v", k, v)
		sl = append(sl, s)
	}
	ss += strings.Join(sl, "\r\n") + "\r\n"
	sl = []string{}
	for i := range chunk.Constants {
		s := fmt.Sprintf("%v: %v", i, chunk.Constants[i])
		sl = append(sl, s)
	}
	ss += strings.Join(sl, "\r\n") + "\r\n"

	return ss
}
