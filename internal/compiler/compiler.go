package compiler

import (
	"github.com/wanderer69/frm/internal/ast"
	"github.com/wanderer69/frm/internal/bytecode"
	valueType "github.com/wanderer69/frm/pkg/value_types"
)

//
// ====== Компилятор ======
//

type Compiler struct {
	chunk     *bytecode.Chunk
	chunks    []*bytecode.Chunk
	nextLocal int
	nextIter  int
}

func NewCompiler() *Compiler {
	return &Compiler{
		chunk: bytecode.NewChunk(),
		//		locals: make(map[string]int),
	}
}

func (c *Compiler) Compile(prog *ast.Program) []*bytecode.Chunk {
	for _, d := range prog.Decls {
		c.chunk = bytecode.NewChunk()
		c.nextIter = 0
		switch d.(type) {
		case *ast.FramesDecl:
		case *ast.FunctionDecl:
			fn := d.(*ast.FunctionDecl)
			for i, p := range fn.Params {
				c.chunk.Locals[p] = i
				c.nextLocal = i + 1
				//c.chunk.sortedArgs[i] = i
			}
			c.chunk.SortedArgs = make([]int, len(c.chunk.Locals))
			i := 0
			for range c.chunk.Locals {
				c.chunk.SortedArgs[i] = i
				i++
			}
			for _, st := range fn.Body {
				c.compileStmt(st)
			}
			c.chunk.Emit(bytecode.OpReturn, 0)
			c.chunk.Name = fn.Name
			//c.chunk.Locals = c.chunk.Locals
			c.chunk.IterNumbers = c.nextIter
			c.chunk.LocalNumbers = c.nextLocal
			c.chunks = append(c.chunks, c.chunk)
		}
	}
	return c.chunks
}

func (c *Compiler) compileExpr(e ast.Expr) {
	switch v := e.(type) {
	case *ast.LiteralExpr:
		idx := c.chunk.AddConst(v.Value)
		c.chunk.Emit(bytecode.OpConstant, idx)
	case *ast.VarExpr:
		slot, ok := c.chunk.Locals[v.Name]
		if !ok {
			panic("неизвестная переменная " + v.Name)
		}
		c.chunk.Emit(bytecode.OpLoadLocal, slot)
	case *ast.FuncCallExpr:
		// вызов функции
		for i := range v.Exprs {
			c.compileExpr(v.Exprs[i])
		}
		idx := c.chunk.AddConst(&valueType.ValueSymbol{Symbol: v.Name})
		//c.chunk.Emit(bytecode.OpConstant, idx)
		c.chunk.EmitExt(bytecode.OpCallFunc, idx, len(v.Exprs))
	default:
		panic("unsupported expr")
	}
}

func (c *Compiler) compileBlock(body []ast.Stmt) {
	for _, st := range body {
		c.compileStmt(st)
	}
}

func (c *Compiler) compileStmt(s ast.Stmt) {
	switch st := s.(type) {
	case *ast.AssignStmt:
		c.compileExpr(st.Expr)
		slot, ok := c.chunk.Locals[st.Name]
		if !ok {
			slot = c.nextLocal
			c.chunk.Locals[st.Name] = slot
			c.nextLocal++
		}
		c.chunk.Emit(bytecode.OpStoreLocal, slot)
	case *ast.PrintStmt:
		for _, e := range st.Exprs {
			c.compileExpr(e)
			c.chunk.Emit(bytecode.OpPrint, 0)
		}
	case *ast.FrameStmt:
		for _, sl := range st.Slots {
			nameIdx := c.chunk.AddConst(&valueType.ValueSymbol{Symbol: sl.Name})
			c.chunk.Emit(bytecode.OpConstant, nameIdx)
			c.compileExpr(sl.Value)
		}
		c.chunk.Emit(bytecode.OpFrameNew, len(st.Slots))
		if st.TargetVar != "" {
			slot, ok := c.chunk.Locals[st.TargetVar]
			if !ok {
				slot = c.nextLocal
				c.chunk.Locals[st.TargetVar] = slot
				c.nextLocal++
			}
			c.chunk.Emit(bytecode.OpStoreLocal, slot)
		} else {
			c.chunk.Emit(bytecode.OpPop, 0)
		}

	case *ast.FrameOpStmt:
		switch st.Op {
		case ".":
			for _, sl := range st.SlotsIn {
				//nameIdx := c.chunk.addConst(sl.Name)
				//c.chunk.Emit(bytecode.OpConstant, nameIdx)
				c.compileExpr(sl.NameVar)
				c.compileExpr(sl.Value)
			}
		case "?":
			for _, sl := range st.SlotsIn {
				//nameIdx := c.chunk.addConst(sl.Name)
				//c.chunk.Emit(bytecode.OpConstant, nameIdx)
				c.compileExpr(sl.NameVar)
				c.compileExpr(sl.Value)
			}
			c.chunk.Emit(bytecode.OpFindFrames, len(st.SlotsIn))
			if st.TargetVar != "" {
				slot, ok := c.chunk.Locals[st.TargetVar]
				if !ok {
					slot = c.nextLocal
					c.chunk.Locals[st.TargetVar] = slot
					c.nextLocal++
				}
				c.chunk.Emit(bytecode.OpStoreLocal, slot)
			} else {
				c.chunk.Emit(bytecode.OpPop, 0)
			}
		case ":":
			for _, sl := range st.SlotsIn {
				c.compileExpr(sl.NameVar)
				c.compileExpr(sl.Value)
			}

			for _, sl := range st.SlotsOut {
				c.compileExpr(sl.NameVar)
				c.compileExpr(sl.Value)
			}

			c.chunk.EmitExt(bytecode.OpFindFrames, len(st.SlotsIn), len(st.SlotsOut))
		}
		/*
			c.chunk.Emit(bytecode.OpFrameNew, len(st.Slots))
			if st.TargetVar != "" {
				slot, ok := c.locals[st.TargetVar]
				if !ok {
					slot = c.nextLocal
					c.locals[st.TargetVar] = slot
					c.nextLocal++
				}
				c.chunk.Emit(bytecode.OpStoreLocal, slot)
			} else {
				c.chunk.Emit(bytecode.OpPop, 0)
			}
		*/
	case *ast.IfStmt:
		c.compileExpr(st.Cond)
		jumpPos := c.chunk.Emit(bytecode.OpJumpIfFalse, 0)
		c.compileBlock(st.Body)
		end := len(c.chunk.Instructions)
		c.chunk.Instructions[jumpPos].Arg1 = int32(end)
	case *ast.ForEachFramesStmt:

		c.compileStmt(st.IterateBy)

		iterLocal := c.nextLocal
		c.nextLocal++
		c.chunk.Emit(bytecode.OpStoreLocal, iterLocal)

		iterIter := c.nextIter
		c.nextIter++
		c.chunk.EmitExt(bytecode.OpIterBeginFrames, iterLocal, iterIter)
		loopStart := len(c.chunk.Instructions)
		c.chunk.Emit(bytecode.OpIterNextFrames, iterIter)
		jumpEnd := c.chunk.Emit(bytecode.OpJumpIfFalse, 0)

		/*
			slot, ok := c.chunk.Locals[st.IterateVarName]
			if !ok {
				slot = c.nextLocal
				c.chunk.Locals[st.VarName] = slot
				c.nextLocal++
			}
		*/

		slot, ok := c.chunk.Locals[st.VarName]
		if !ok {

		}
		c.chunk.Locals[st.VarName] = slot
		c.nextLocal++
		c.chunk.Emit(bytecode.OpStoreLocal, slot)
		c.compileBlock(st.Body)
		c.chunk.Emit(bytecode.OpJump, loopStart)
		end := len(c.chunk.Instructions)
		c.chunk.Instructions[jumpEnd].Arg1 = int32(end)

		/*
			iterLocal := c.nextLocal
			c.nextLocal++
			c.chunk.Emit(bytecode.OpIterBeginFrames, iterLocal)
			loopStart := len(c.chunk.Instructions)
			c.chunk.Emit(bytecode.OpIterNextFrames, iterLocal)
			jumpEnd := c.chunk.Emit(bytecode.OpJumpIfFalse, 0)
			slot, ok := c.chunk.Locals[st.IterateVarName]
			if !ok {
				slot = c.nextLocal
				//c.locals[st.VarName] = slot
				//c.nextLocal++
			}
			c.chunk.Locals[st.VarName] = slot
			c.nextLocal++
			c.chunk.Emit(bytecode.OpStoreLocal, slot)
			c.compileBlock(st.Body)
			c.chunk.Emit(bytecode.OpJump, loopStart)
			end := len(c.chunk.Instructions)
			c.chunk.Instructions[jumpEnd].Arg1 = int32(end)
		*/
	case *ast.ExprStmt:
		c.compileExpr(st.Expr)
		/*
			slot, ok := c.locals[st.Name]
			if !ok {
				slot = c.nextLocal
				c.locals[st.Name] = slot
				c.nextLocal++
			}
		*/
		//		c.chunk.Emit(bytecode.OpStoreLocal, slot)

	default:
		panic("unsupported stmt")
	}
}
