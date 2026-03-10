package ast

import "github.com/wanderer69/frm/internal/value"

//
// ====== AST ======
//

type Program struct {
	Decls []Decl
}

type Decl interface {
	isDecl()
}

type FunctionDecl struct {
	Name   string
	Params []string
	Body   []Stmt
}

func (*FunctionDecl) isDecl() {}

type FramesDecl struct {
	Body []Stmt
}

func (*FramesDecl) isDecl() {}

type Stmt interface {
	isStmt()
}

type AssignStmt struct {
	Expr Expr
	Name string
}

func (*AssignStmt) isStmt() {}

type ExprStmt struct {
	Expr Expr
}

func (*ExprStmt) isStmt() {}

type PrintStmt struct {
	Exprs []Expr
}

func (*PrintStmt) isStmt() {}

type FrameStmt struct {
	Slots     []SlotExpr
	TargetVar string
}

func (*FrameStmt) isStmt() {}

type FrameOpStmt struct {
	Op        string
	SlotsIn   []SlotExpr
	SlotsOut  []SlotExpr
	TargetVar string
}

func (*FrameOpStmt) isStmt() {}

type IfStmt struct {
	Cond Expr
	Body []Stmt
}

func (*IfStmt) isStmt() {}

type ForEachFramesStmt struct {
	IterateVarName string
	VarName        string
	Body           []Stmt
	IterateBy      Stmt
}

func (*ForEachFramesStmt) isStmt() {}

type Expr interface {
	isExpr()
}

type VarExpr struct {
	Name string
}

func (*VarExpr) isExpr() {}

type FuncCallExpr struct {
	Name  string
	Exprs []Expr
}

func (*FuncCallExpr) isExpr() {}

type LiteralExpr struct {
	Value value.Value
}

func (*LiteralExpr) isExpr() {}

type SlotExpr struct {
	Name    string
	NameVar Expr
	Value   Expr
}
