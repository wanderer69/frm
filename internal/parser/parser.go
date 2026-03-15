package parser

import (
	"fmt"
	"strconv"

	"github.com/wanderer69/frm/internal/ast"
	"github.com/wanderer69/frm/internal/lexer"
	valueType "github.com/wanderer69/frm/pkg/value_types"
)

//
// ====== Парсер ======
//

type Parser struct {
	l      *lexer.Lexer
	cur    lexer.Token
	peeked bool
}

func NewParser(l *lexer.Lexer) *Parser {
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

func (p *Parser) expect(tt lexer.TokenType, lit string) {
	if p.cur.Type != tt || (lit != "" && p.cur.Lit != lit) {
		panic(fmt.Sprintf("ожидалось %v %q, а получено %v %q", tt, lit, p.cur.Type, p.cur.Lit))
	}
	p.next()
}

func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{}
	for p.cur.Type != lexer.EOF {
		if p.cur.Type == lexer.IDENT && p.cur.Lit == "функция" {
			prog.Decls = append(prog.Decls, p.parseFunction())
		} else if p.cur.Type == lexer.IDENT && p.cur.Lit == "фреймы" {
			prog.Decls = append(prog.Decls, p.parseFrames())
		} else {
			panic("ожидалась функция")
		}
	}
	return prog
}

func (p *Parser) parseFunction() *ast.FunctionDecl {
	p.expect(lexer.IDENT, "функция")
	if p.cur.Type != lexer.IDENT {
		panic("ожидалось имя функции")
	}
	name := p.cur.Lit
	p.next()
	p.expect(lexer.LPAREN, "")
	var params []string
	if p.cur.Type == lexer.SYMBOL_QMARK {
		for {
			p.expect(lexer.SYMBOL_QMARK, "")
			if p.cur.Type != lexer.IDENT {
				panic("ожидалось имя параметра")
			}
			params = append(params, p.cur.Lit)
			p.next()
			if p.cur.Type == lexer.COMMA {
				p.next()
				continue
			}
			break
		}
	}
	p.expect(lexer.RPAREN, "")
	p.expect(lexer.LBRACE, "")
	var body []ast.Stmt
	for !(p.cur.Type == lexer.RBRACE) {
		body = append(body, p.parseStmt())
		if p.cur.Type == lexer.RBRACE {
			break
		}
		p.expect(lexer.SEMICOLON, "")
	}
	p.expect(lexer.RBRACE, "")
	p.expect(lexer.SEMICOLON, "")
	return &ast.FunctionDecl{Name: name, Params: params, Body: body}
}

func (p *Parser) parseFrames() *ast.FramesDecl {
	//p.expect(IDENT, "фреймы")
	if p.cur.Type != lexer.IDENT {
		panic("ожидалось имя функции")
	}
	p.next()
	p.expect(lexer.LBRACE, "")
	var body []ast.Stmt
	for !(p.cur.Type == lexer.RBRACE) {
		body = append(body, p.parseFrame())
		p.expect(lexer.SEMICOLON, "")
	}
	p.expect(lexer.RBRACE, "")
	p.expect(lexer.SEMICOLON, "")
	return &ast.FramesDecl{Body: body}
}

func (p *Parser) parseStmt() ast.Stmt {
	if p.cur.Type == lexer.IDENT && p.cur.Lit == "печатать" {
		return p.parsePrint()
	}
	if p.cur.Type == lexer.IDENT && p.cur.Lit == "фрейм" {
		return p.parseFrame()
	}
	if p.cur.Type == lexer.IDENT && p.cur.Lit == "если" {
		return p.parseIf()
	}
	if p.cur.Type == lexer.IDENT && p.cur.Lit == "для" {
		return p.parseForEachFrames()
	}
	if p.cur.Type == lexer.LPAREN {
		return p.parseFrameOperation()
	}
	// присваивание: Expr => ?ид
	ex := p.parseExpr()
	if p.cur.Type == lexer.ARROW {
		p.next()
		//	p.expect(ARROW, "")
		p.expect(lexer.SYMBOL_QMARK, "")
		if p.cur.Type != lexer.IDENT {
			panic("ожидалось имя переменной после ?")
		}
		name := p.cur.Lit
		p.next()
		return &ast.AssignStmt{Expr: ex, Name: name}
	}
	return &ast.ExprStmt{Expr: ex}
}

func (p *Parser) parsePrint() ast.Stmt {
	p.expect(lexer.IDENT, "печатать")
	p.expect(lexer.LPAREN, "")
	var exprs []ast.Expr
	exprs = append(exprs, p.parseExpr())
	for p.cur.Type == lexer.COMMA {
		p.next()
		exprs = append(exprs, p.parseExpr())
	}
	p.expect(lexer.RPAREN, "")
	return &ast.PrintStmt{Exprs: exprs}
}

func (p *Parser) parseFrame() ast.Stmt {
	p.expect(lexer.IDENT, "фрейм")
	p.expect(lexer.LPAREN, "")
	var slots []ast.SlotExpr
	for {
		if p.cur.Type != lexer.IDENT {
			panic("ожидалось имя слота")
		}
		name := p.cur.Lit
		p.next()
		p.expect(lexer.POINT, "")
		if p.cur.Type != lexer.STRING && p.cur.Type != lexer.IDENT {
			panic("ожидалось значение слота (строка или идентификатор)")
		}
		var val ast.Expr
		if p.cur.Type == lexer.STRING {
			val = &ast.LiteralExpr{Value: &valueType.ValueString{String: p.cur.Lit}}
		} else {
			val = &ast.LiteralExpr{Value: &valueType.ValueSymbol{Symbol: p.cur.Lit}}
		}
		p.next()
		slots = append(slots, ast.SlotExpr{Name: name, Value: val})
		if p.cur.Type == lexer.COMMA {
			p.next()
			continue
		}
		break
	}
	p.expect(lexer.RPAREN, "")
	var target string
	if p.cur.Type == lexer.ARROW {
		p.next()
		p.expect(lexer.SYMBOL_QMARK, "")
		if p.cur.Type != lexer.IDENT {
			panic("ожидалось имя переменной после ?")
		}
		target = p.cur.Lit
		p.next()
	}
	return &ast.FrameStmt{Slots: slots, TargetVar: target}
}

func (p *Parser) parseFrameOperation() ast.Stmt {
	p.expect(lexer.LPAREN, "")
	var slots []ast.SlotExpr

	parseSlot := func(slotsIn []ast.SlotExpr) []ast.SlotExpr {
		var name ast.Expr
		//name := p.cur.Lit

		if p.cur.Type != lexer.STRING && p.cur.Type != lexer.IDENT {
			panic("ожидалось значение слота (строка или идентификатор)")
		}
		/*
			if p.cur.Type == lexer.STRING {
				name = &ast.LiteralExpr{Value: p.cur.Lit}
			} else {
				name = &ast.LiteralExpr{Value: p.cur.Lit}
			}
		*/
		if p.cur.Type == lexer.STRING {
			name = &ast.LiteralExpr{Value: &valueType.ValueString{String: p.cur.Lit}}
		} else {
			name = &ast.LiteralExpr{Value: &valueType.ValueSymbol{Symbol: p.cur.Lit}}
		}

		p.next()
		//		p.expect(lexer.POINT, "")
		var val ast.Expr
		if p.cur.Type == lexer.POINT {
			p.next()
			if p.cur.Type != lexer.STRING && p.cur.Type != lexer.IDENT {
				panic("ожидалось значение слота (строка или идентификатор)")
			}
			/*
				if p.cur.Type == lexer.STRING {
					val = &ast.LiteralExpr{Value: p.cur.Lit}
				} else {
					val = &ast.LiteralExpr{Value: p.cur.Lit}
				}
			*/
			if p.cur.Type == lexer.STRING {
				val = &ast.LiteralExpr{Value: &valueType.ValueString{String: p.cur.Lit}}
			} else {
				val = &ast.LiteralExpr{Value: &valueType.ValueSymbol{Symbol: p.cur.Lit}}
			}
			p.next()
		}
		slotsIn = append(slotsIn, ast.SlotExpr{NameVar: name, Value: val})
		return slotsIn
	}
	for {
		/*
			if p.cur.Type != lexer.IDENT {
				panic("ожидалось имя слота")
			}
		*/
		slots = parseSlot(slots)
		if p.cur.Type == lexer.COMMA {
			p.next()
			continue
		}
		break
	}
	p.expect(lexer.RPAREN, "")

	var target string
	var result ast.Stmt
	switch p.cur.Type {
	case lexer.SYMBOL_QMARK:
		p.next()
		if p.cur.Type == lexer.ARROW {
			p.next()
			p.expect(lexer.SYMBOL_QMARK, "")
			if p.cur.Type != lexer.IDENT {
				panic("ожидалось имя переменной после ?")
			}
			target = p.cur.Lit
			p.next()
		}
		result = &ast.FrameOpStmt{Op: "?", SlotsIn: slots, TargetVar: target}

	case lexer.POINT:
		p.next()
		p.expect(lexer.LPAREN, "")
		var slots []ast.SlotExpr
		var slotsOut []ast.SlotExpr

		for {
			/*
				if p.cur.Type != IDENT {
					panic("ожидалось имя слота")
				}
				name := p.cur.Lit
				p.next()
				var val Expr
				if p.cur.Type == POINT {
					p.next()
					if p.cur.Type != STRING && p.cur.Type != IDENT {
						panic("ожидалось значение слота (строка или идентификатор)")
					}
					if p.cur.Type == STRING {
						val = &LiteralExpr{Value: p.cur.Lit}
					} else {
						val = &LiteralExpr{Value: p.cur.Lit}
					}
					p.next()
				}
				slots = append(slots, SlotExpr{Name: name, Value: val})
			*/
			slots = parseSlot(slots)
			if p.cur.Type == lexer.COMMA {
				p.next()
				continue
			}
			break
		}

		p.expect(lexer.RPAREN, "")
		result = &ast.FrameOpStmt{SlotsIn: slots, SlotsOut: slotsOut, Op: "."}

	case lexer.COLON:
		p.next()
		p.expect(lexer.LPAREN, "")
		var slots []ast.SlotExpr
		var slotsOut []ast.SlotExpr
		for {
			/*
				if p.cur.Type != IDENT {
					panic("ожидалось имя слота")
				}
				name := p.cur.Lit
				p.next()
				//		p.expect(POINT, "")
				var val Expr
				if p.cur.Type == POINT {
					p.next()
					if p.cur.Type != STRING && p.cur.Type != IDENT {
						panic("ожидалось значение слота (строка или идентификатор)")
					}
					if p.cur.Type == STRING {
						val = &LiteralExpr{Value: p.cur.Lit}
					} else {
						val = &LiteralExpr{Value: p.cur.Lit}
					}
					p.next()
				}
				slots = append(slots, SlotExpr{Name: name, Value: val})
			*/
			slots = parseSlot(slots)

			if p.cur.Type == lexer.COMMA {
				p.next()
				continue
			}
			break
		}
		p.expect(lexer.RPAREN, "")

		result = &ast.FrameOpStmt{SlotsIn: slots, SlotsOut: slotsOut, Op: ":"}
	}
	return result
}

func (p *Parser) parseIf() ast.Stmt {
	p.expect(lexer.IDENT, "если")
	p.expect(lexer.LPAREN, "")
	cond := p.parseExpr()
	p.expect(lexer.RPAREN, "")
	p.expect(lexer.LBRACE, "")
	var body []ast.Stmt
	for !(p.cur.Type == lexer.RBRACE) {
		body = append(body, p.parseStmt())
		p.expect(lexer.SEMICOLON, "")
	}
	p.expect(lexer.RBRACE, "")
	return &ast.IfStmt{Cond: cond, Body: body}
}

// очень упрощённый синтаксис:
// для каждого (фрейм) => ?x { ... }
func (p *Parser) parseForEachFrames() ast.Stmt {
	p.expect(lexer.IDENT, "для")
	if p.cur.Type != lexer.IDENT || p.cur.Lit != "каждого" {
		panic("ожидалось 'каждого'")
	}
	p.next()
	if p.cur.Type != lexer.IDENT || p.cur.Lit != "элемента" {
		panic("ожидалось 'элемента'")
	}
	p.next()
	p.expect(lexer.LPAREN, "")
	/*
		p.expect(lexer.SYMBOL_QMARK, "")

			//if p.cur.Type != lexer.IDENT || p.cur.Lit != "фрейм" {
			//	panic("ожидалось 'фрейм' в скобках")
			//}

		if p.cur.Type != lexer.IDENT {
			panic("ожидалось имя переменной цикла")
		}
		iterVarName := p.cur.Lit
		p.next()
	*/
	iterateBy := p.parseStmt()
	p.expect(lexer.RPAREN, "")
	p.expect(lexer.ARROW, "")
	p.expect(lexer.SYMBOL_QMARK, "")
	if p.cur.Type != lexer.IDENT {
		panic("ожидалось имя переменной цикла")
	}
	varName := p.cur.Lit
	p.next()
	p.expect(lexer.LBRACE, "")
	var body []ast.Stmt
	for !(p.cur.Type == lexer.RBRACE) {
		body = append(body, p.parseStmt())
		p.expect(lexer.SEMICOLON, "")
	}
	p.expect(lexer.RBRACE, "")
	return &ast.ForEachFramesStmt{IterateBy: iterateBy, VarName: varName, Body: body}
}

func (p *Parser) parseExpr() ast.Expr {
	switch p.cur.Type {
	case lexer.SYMBOL_QMARK:
		p.next()
		if p.cur.Type != lexer.IDENT {
			panic("ожидалось имя переменной после ?")
		}
		name := p.cur.Lit
		p.next()
		return &ast.VarExpr{Name: name}
	case lexer.STRING:
		v := p.cur.Lit
		p.next()
		return &ast.LiteralExpr{Value: &valueType.ValueString{String: v}}
	case lexer.NUMBER:
		n, _ := strconv.Atoi(p.cur.Lit)
		p.next()
		return &ast.LiteralExpr{Value: &valueType.ValueInt{Int: n}}
	default:
		if p.cur.Type == lexer.IDENT {
			v := p.cur.Lit
			p.next()
			if p.cur.Type == lexer.LPAREN {
				// вызов функции
				p.next()
				exprs := []ast.Expr{}
				if p.cur.Type != lexer.RPAREN {
					for {
						expr := p.parseExpr()
						exprs = append(exprs, expr)
						if p.cur.Type != lexer.COMMA {
							break
						}
						p.next()
					}
				}
				p.expect(lexer.RPAREN, "")
				return &ast.FuncCallExpr{Name: v, Exprs: exprs}
			}
			return &ast.LiteralExpr{Value: &valueType.ValueSymbol{Symbol: v}}
		}
	}
	panic(fmt.Sprintf("неожиданное выражение: %v %q", p.cur.Type, p.cur.Lit))
}
