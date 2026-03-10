package lexer

import (
	"strings"
	"unicode"
)

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
	POINT
	COLON
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
		case ch == '.':
			return Token{Type: POINT, Lit: "."}
		case ch == ':':
			return Token{Type: COLON, Lit: ":"}
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
