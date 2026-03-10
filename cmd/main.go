package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/wanderer69/frm/internal/bytecode"
	"github.com/wanderer69/frm/internal/compiler"
	"github.com/wanderer69/frm/internal/lexer"
	"github.com/wanderer69/frm/internal/parser"
	"github.com/wanderer69/frm/internal/vm"
)

/*
//
// ====== Лексер ======
//

type TokenType int

const (

	ILLEGAL TokenType = iota
	EOF
	IDENT
	NUMBER
	STRING
	SYMBOL_QMARK // ?
	ARROW        // =>
	LBRACE       // {
	RBRACE       // }
	LPAREN       // (
	RPAREN       // )
	COMMA        // ,
	SEMICOLON    // ;

)

	type Token struct {
		Type TokenType
		Lit  string
	}

	type Lexer struct {
		input []rune
		pos   int
	}

	func NewLexer(src string) *Lexer {
		return &Lexer{input: []rune(src)}
	}

	func (l *Lexer) next() rune {
		if l.pos >= len(l.input) {
			return 0
		}
		ch := l.input[l.pos]
		l.pos++
		return ch
	}

	func (l *Lexer) peek() rune {
		if l.pos >= len(l.input) {
			return 0
		}
		return l.input[l.pos]
	}

	func (l *Lexer) NextToken() Token {
		for {
			ch := l.next()
			switch {
			case ch == 0:
				return Token{Type: EOF}
			case ch == '#':
				for ch != '\n' && ch != 0 {
					ch = l.next()
				}
			case unicode.IsSpace(ch):
				continue
			case ch == '{':
				return Token{Type: LBRACE, Lit: "{"}
			case ch == '}':
				return Token{Type: RBRACE, Lit: "}"}
			case ch == '(':
				return Token{Type: LPAREN, Lit: "("}
			case ch == ')':
				return Token{Type: RPAREN, Lit: ")"}
			case ch == ',':
				return Token{Type: COMMA, Lit: ","}
			case ch == ';':
				return Token{Type: SEMICOLON, Lit: ";"}
			case ch == '?':
				return Token{Type: SYMBOL_QMARK, Lit: "?"}
			case ch == '=' && l.peek() == '>':
				l.next()
				return Token{Type: ARROW, Lit: "=>"}
			case ch == '"':
				var sb strings.Builder
				for {
					c := l.next()
					if c == '"' || c == 0 {
						break
					}
					sb.WriteRune(c)
				}
				return Token{Type: STRING, Lit: sb.String()}
			default:
				if unicode.IsLetter(ch) || ch == '_' || ch >= 128 {
					var sb strings.Builder
					sb.WriteRune(ch)
					for {
						c := l.peek()
						if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' || c >= 128 {
							sb.WriteRune(c)
							l.next()
						} else {
							break
						}
					}
					return Token{Type: IDENT, Lit: sb.String()}
				}
				if unicode.IsDigit(ch) {
					var sb strings.Builder
					sb.WriteRune(ch)
					for {
						c := l.peek()
						if unicode.IsDigit(c) {
							sb.WriteRune(c)
							l.next()
						} else {
							break
						}
					}
					return Token{Type: NUMBER, Lit: sb.String()}
				}
				return Token{Type: ILLEGAL, Lit: string(ch)}
			}
		}
	}

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

	type Stmt interface {
		isStmt()
	}

	type AssignStmt struct {
		Expr Expr
		Name string
	}

func (*AssignStmt) isStmt() {}

	type PrintStmt struct {
		Exprs []Expr
	}

func (*PrintStmt) isStmt() {}

	type FrameStmt struct {
		Slots     []SlotExpr
		TargetVar string
	}

func (*FrameStmt) isStmt() {}

	type IfStmt struct {
		Cond Expr
		Body []Stmt
	}

func (*IfStmt) isStmt() {}

	type ForEachFramesStmt struct {
		VarName string
		Body    []Stmt
	}

func (*ForEachFramesStmt) isStmt() {}

	type Expr interface {
		isExpr()
	}

	type VarExpr struct {
		Name string
	}

func (*VarExpr) isExpr() {}

	type LiteralExpr struct {
		Value any
	}

func (*LiteralExpr) isExpr() {}

	type SlotExpr struct {
		Name  string
		Value Expr
	}

//
// ====== Парсер ======
//

	type Parser struct {
		l      *Lexer
		cur    Token
		peeked bool
	}

	func NewParser(l *Lexer) *Parser {
		p := &Parser{l: l}
		p.next()
		return p
	}

	func (p *Parser) next() {
		if p.peeked {
			p.peeked = false
			return
		}
		p.cur = p.l.NextToken()
	}

	func (p *Parser) expect(tt TokenType, lit string) {
		if p.cur.Type != tt || (lit != "" && p.cur.Lit != lit) {
			panic(fmt.Sprintf("ожидалось %v %q, а получено %v %q", tt, lit, p.cur.Type, p.cur.Lit))
		}
		p.next()
	}

	func (p *Parser) ParseProgram() *Program {
		prog := &Program{}
		for p.cur.Type != EOF {
			if p.cur.Type == IDENT && p.cur.Lit == "функция" {
				prog.Decls = append(prog.Decls, p.parseFunction())
			} else {
				panic("ожидалась функция")
			}
		}
		return prog
	}

	func (p *Parser) parseFunction() *FunctionDecl {
		p.expect(IDENT, "функция")
		if p.cur.Type != IDENT {
			panic("ожидалось имя функции")
		}
		name := p.cur.Lit
		p.next()
		p.expect(LPAREN, "")
		var params []string
		if p.cur.Type == SYMBOL_QMARK {
			for {
				p.expect(SYMBOL_QMARK, "")
				if p.cur.Type != IDENT {
					panic("ожидалось имя параметра")
				}
				params = append(params, p.cur.Lit)
				p.next()
				if p.cur.Type == COMMA {
					p.next()
					continue
				}
				break
			}
		}
		p.expect(RPAREN, "")
		p.expect(LBRACE, "")
		var body []Stmt
		for !(p.cur.Type == RBRACE) {
			body = append(body, p.parseStmt())
			p.expect(SEMICOLON, "")
		}
		p.expect(RBRACE, "")
		return &FunctionDecl{Name: name, Params: params, Body: body}
	}

	func (p *Parser) parseStmt() Stmt {
		if p.cur.Type == IDENT && p.cur.Lit == "печатать" {
			return p.parsePrint()
		}
		if p.cur.Type == IDENT && p.cur.Lit == "фрейм" {
			return p.parseFrame()
		}
		if p.cur.Type == IDENT && p.cur.Lit == "если" {
			return p.parseIf()
		}
		if p.cur.Type == IDENT && p.cur.Lit == "для" {
			return p.parseForEachFrames()
		}
		// присваивание: Expr => ?ид
		ex := p.parseExpr()
		p.expect(ARROW, "")
		p.expect(SYMBOL_QMARK, "")
		if p.cur.Type != IDENT {
			panic("ожидалось имя переменной после ?")
		}
		name := p.cur.Lit
		p.next()
		return &AssignStmt{Expr: ex, Name: name}
	}

	func (p *Parser) parsePrint() Stmt {
		p.expect(IDENT, "печатать")
		p.expect(LPAREN, "")
		var exprs []Expr
		exprs = append(exprs, p.parseExpr())
		for p.cur.Type == COMMA {
			p.next()
			exprs = append(exprs, p.parseExpr())
		}
		p.expect(RPAREN, "")
		return &PrintStmt{Exprs: exprs}
	}

	func (p *Parser) parseFrame() Stmt {
		p.expect(IDENT, "фрейм")
		p.expect(LPAREN, "")
		var slots []SlotExpr
		for {
			if p.cur.Type != IDENT {
				panic("ожидалось имя слота")
			}
			name := p.cur.Lit
			p.next()
			if p.cur.Type != STRING && p.cur.Type != IDENT {
				panic("ожидалось значение слота (строка или идентификатор)")
			}
			var val Expr
			if p.cur.Type == STRING {
				val = &LiteralExpr{Value: p.cur.Lit}
			} else {
				val = &LiteralExpr{Value: p.cur.Lit}
			}
			p.next()
			slots = append(slots, SlotExpr{Name: name, Value: val})
			if p.cur.Type == COMMA {
				p.next()
				continue
			}
			break
		}
		p.expect(RPAREN, "")
		var target string
		if p.cur.Type == ARROW {
			p.next()
			p.expect(SYMBOL_QMARK, "")
			if p.cur.Type != IDENT {
				panic("ожидалось имя переменной после ?")
			}
			target = p.cur.Lit
			p.next()
		}
		return &FrameStmt{Slots: slots, TargetVar: target}
	}

	func (p *Parser) parseIf() Stmt {
		p.expect(IDENT, "если")
		p.expect(LPAREN, "")
		cond := p.parseExpr()
		p.expect(RPAREN, "")
		p.expect(LBRACE, "")
		var body []Stmt
		for !(p.cur.Type == RBRACE) {
			body = append(body, p.parseStmt())
			p.expect(SEMICOLON, "")
		}
		p.expect(RBRACE, "")
		return &IfStmt{Cond: cond, Body: body}
	}

// очень упрощённый синтаксис:
// для каждого (фрейм) => ?x { ... }

	func (p *Parser) parseForEachFrames() Stmt {
		p.expect(IDENT, "для")
		if p.cur.Type != IDENT || p.cur.Lit != "каждого" {
			panic("ожидалось 'каждого'")
		}
		p.next()
		if p.cur.Type != IDENT || p.cur.Lit != "элемента" {
			panic("ожидалось 'элемента'")
		}
		p.next()
		p.expect(LPAREN, "")
		if p.cur.Type != IDENT || p.cur.Lit != "фрейм" {
			panic("ожидалось 'фрейм' в скобках")
		}
		p.next()
		p.expect(RPAREN, "")
		p.expect(ARROW, "")
		p.expect(SYMBOL_QMARK, "")
		if p.cur.Type != IDENT {
			panic("ожидалось имя переменной цикла")
		}
		varName := p.cur.Lit
		p.next()
		p.expect(LBRACE, "")
		var body []Stmt
		for !(p.cur.Type == RBRACE) {
			body = append(body, p.parseStmt())
			p.expect(SEMICOLON, "")
		}
		p.expect(RBRACE, "") // <-- исправь пробел, если IDE ругнётся
		return &ForEachFramesStmt{VarName: varName, Body: body}
	}

	func (p *Parser) parseExpr() Expr {
		switch p.cur.Type {
		case SYMBOL_QMARK:
			p.next()
			if p.cur.Type != IDENT {
				panic("ожидалось имя переменной после ?")
			}
			name := p.cur.Lit
			p.next()
			return &VarExpr{Name: name}
		case STRING:
			v := p.cur.Lit
			p.next()
			return &LiteralExpr{Value: v}
		case NUMBER:
			n, _ := strconv.Atoi(p.cur.Lit)
			p.next()
			return &LiteralExpr{Value: n}
		default:
			if p.cur.Type == IDENT {
				v := p.cur.Lit
				p.next()
				return &LiteralExpr{Value: v}
			}
		}
		panic(fmt.Sprintf("неожиданное выражение: %v %q", p.cur.Type, p.cur.Lit))
	}

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

)

	type Instruction struct {
		Op   OpCode
		Arg1 int
	}

	type Chunk struct {
		Instructions []Instruction
		Constants    []any
	}

	func (c *Chunk) addConst(v any) int {
		c.Constants = append(c.Constants, v)
		return len(c.Constants) - 1
	}

	func (c *Chunk) emit(op OpCode, arg1 int) int {
		c.Instructions = append(c.Instructions, Instruction{Op: op, Arg1: arg1})
		return len(c.Instructions) - 1
	}

//
// ====== Компилятор ======
//

	type Compiler struct {
		chunk     *Chunk
		locals    map[string]int
		nextLocal int
	}

	func NewCompiler() *Compiler {
		return &Compiler{
			chunk:  &Chunk{},
			locals: make(map[string]int),
		}
	}

	func (c *Compiler) Compile(prog *Program) *Chunk {
		for _, d := range prog.Decls {
			fn := d.(*FunctionDecl)
			for i, p := range fn.Params {
				c.locals[p] = i
				c.nextLocal = i + 1
			}
			for _, st := range fn.Body {
				c.compileStmt(st)
			}
			c.chunk.emit(OpReturn, 0)
		}
		return c.chunk
	}

	func (c *Compiler) compileExpr(e Expr) {
		switch v := e.(type) {
		case *LiteralExpr:
			idx := c.chunk.addConst(v.Value)
			c.chunk.emit(OpConstant, idx)
		case *VarExpr:
			slot, ok := c.locals[v.Name]
			if !ok {
				panic("неизвестная переменная " + v.Name)
			}
			c.chunk.emit(OpLoadLocal, slot)
		default:
			panic("unsupported expr")
		}
	}

	func (c *Compiler) compileBlock(body []Stmt) {
		for _, st := range body {
			c.compileStmt(st)
		}
	}

	func (c *Compiler) compileStmt(s Stmt) {
		switch st := s.(type) {
		case *AssignStmt:
			c.compileExpr(st.Expr)
			slot, ok := c.locals[st.Name]
			if !ok {
				slot = c.nextLocal
				c.locals[st.Name] = slot
				c.nextLocal++
			}
			c.chunk.emit(OpStoreLocal, slot)
		case *PrintStmt:
			for _, e := range st.Exprs {
				c.compileExpr(e)
				c.chunk.emit(OpPrint, 0)
			}
		case *FrameStmt:
			for _, sl := range st.Slots {
				nameIdx := c.chunk.addConst(sl.Name)
				c.chunk.emit(OpConstant, nameIdx)
				c.compileExpr(sl.Value)
			}
			c.chunk.emit(OpFrameNew, len(st.Slots))
			if st.TargetVar != "" {
				slot, ok := c.locals[st.TargetVar]
				if !ok {
					slot = c.nextLocal
					c.locals[st.TargetVar] = slot
					c.nextLocal++
				}
				c.chunk.emit(OpStoreLocal, slot)
			} else {
				c.chunk.emit(OpPop, 0)
			}
		case *IfStmt:
			c.compileExpr(st.Cond)
			jumpPos := c.chunk.emit(OpJumpIfFalse, 0)
			c.compileBlock(st.Body)
			end := len(c.chunk.Instructions)
			c.chunk.Instructions[jumpPos].Arg1 = end
		case *ForEachFramesStmt:
			iterLocal := c.nextLocal
			c.nextLocal++
			c.chunk.emit(OpIterBeginFrames, iterLocal)
			loopStart := len(c.chunk.Instructions)
			c.chunk.emit(OpIterNextFrames, iterLocal)
			jumpEnd := c.chunk.emit(OpJumpIfFalse, 0)
			slot, ok := c.locals[st.VarName]
			if !ok {
				slot = c.nextLocal
				c.locals[st.VarName] = slot
				c.nextLocal++
			}
			c.chunk.emit(OpStoreLocal, slot)
			c.compileBlock(st.Body)
			c.chunk.emit(OpJump, loopStart)
			end := len(c.chunk.Instructions)
			c.chunk.Instructions[jumpEnd].Arg1 = end
		default:
			panic("unsupported stmt")
		}
	}

//
// ====== VM ======
//

type Frame map[string]any

	type VM struct {
		chunk  *Chunk
		ip     int
		stack  []any
		locals []any
		frames []Frame
		iters  map[int]int // iterLocal -> index
	}

	func NewVM(chunk *Chunk) *VM {
		return &VM{
			chunk:  chunk,
			stack:  make([]any, 0, 256),
			locals: make([]any, 256),
			frames: make([]Frame, 0),
			iters:  make(map[int]int),
		}
	}

	func (vm *VM) push(v any) {
		vm.stack = append(vm.stack, v)
	}

	func (vm *VM) pop() any {
		n := len(vm.stack)
		v := vm.stack[n-1]
		vm.stack = vm.stack[:n-1]
		return v
	}

	func truthy(v any) bool {
		if v == nil {
			return false
		}
		switch t := v.(type) {
		case bool:
			return t
		case int:
			return t != 0
		case string:
			return t != ""
		default:
			return true
		}
	}

	func (vm *VM) Run() error {
		for vm.ip < len(vm.chunk.Instructions) {
			inst := vm.chunk.Instructions[vm.ip]
			vm.ip++
			switch inst.Op {
			case OpConstant:
				vm.push(vm.chunk.Constants[inst.Arg1])
			case OpPop:
				vm.pop()
			case OpLoadLocal:
				vm.push(vm.locals[inst.Arg1])
			case OpStoreLocal:
				vm.locals[inst.Arg1] = vm.pop()
			case OpPrint:
				v := vm.pop()
				fmt.Println(v)
			case OpFrameNew:
				n := inst.Arg1
				f := Frame{}
				for i := 0; i < n; i++ {
					val := vm.pop()
					name := vm.pop().(string)
					f[name] = val
				}
				vm.frames = append(vm.frames, f)
				vm.push(f)
			case OpJumpIfFalse:
				v := vm.pop()
				if !truthy(v) {
					vm.ip = inst.Arg1
				}
			case OpJump:
				vm.ip = inst.Arg1
			case OpIterBeginFrames:
				vm.iters[inst.Arg1] = 0
			case OpIterNextFrames:
				idx := vm.iters[inst.Arg1]
				if idx >= len(vm.frames) {
					vm.push(false)
				} else {
					vm.push(vm.frames[idx])
					vm.iters[inst.Arg1] = idx + 1
					vm.push(true)
				}
			case OpReturn:
				return nil
			default:
				return fmt.Errorf("unknown opcode %d", inst.Op)
			}
		}
		return nil
	}

//
// ====== Сериализация байткода ======
//

	func SaveChunk(path string, chunk *Chunk) error {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		enc := gob.NewEncoder(f)
		return enc.Encode(chunk)
	}

	func LoadChunk(path string) (*Chunk, error) {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		var chunk Chunk
		dec := gob.NewDecoder(f)
		if err := dec.Decode(&chunk); err != nil {
			return nil, err
		}
		return &chunk, nil
	}
*/
//
// ====== CLI ======
//

func compileFile(srcPath, outPath string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	l := lexer.NewLexer(string(data))
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	c := compiler.NewCompiler()
	chunk := c.Compile(prog)
	return bytecode.SaveChunks(outPath, chunk)
}

func runFile(bytecodePath string) error {
	chunk, err := bytecode.LoadChunks(bytecodePath)
	if err != nil {
		return err
	}
	vm := vm.NewVM(chunk)
	return vm.Run()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("режимы:")
		fmt.Println("  compile <input.fl> <output.fbc>")
		fmt.Println("  run <program.fbc>")
		fmt.Println("  repl")
		return
	}
	switch os.Args[1] {
	case "compile":
		if len(os.Args) != 4 {
			fmt.Println("usage: compile input.fl output.fbc")
			return
		}
		if err := compileFile(os.Args[2], os.Args[3]); err != nil {
			fmt.Println("error:", err)
		}
	case "run":
		if len(os.Args) != 3 {
			fmt.Println("usage: run program.fbc")
			return
		}
		if err := runFile(os.Args[2]); err != nil {
			fmt.Println("error:", err)
		}
	case "repl":
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print(">> ")
			line, _ := reader.ReadString('\n')
			if strings.TrimSpace(line) == "" {
				continue
			}
			src := "функция main() {" + line + ";}"
			l := lexer.NewLexer(src)
			p := parser.NewParser(l)
			prog := p.ParseProgram()
			c := compiler.NewCompiler()
			chunks := c.Compile(prog)
			vm := vm.NewVM(chunks)
			if err := vm.Run(); err != nil {
				fmt.Println("error:", err)
			}
		}
	default:
		fmt.Println("unknown command")
	}
}
