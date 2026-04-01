package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

type Lexer struct {
	input   []rune
	pos     int
	line    int
	col     int
	tokens  []Token
	indents []int // stack of indent levels
}

func New(input string) *Lexer {
	return &Lexer{
		input:   []rune(input),
		pos:     0,
		line:    1,
		col:     1,
		indents: []int{0},
	}
}

func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.input) {
		// Handle beginning of line: measure indentation
		if l.col == 1 {
			l.handleIndentation()
		}

		ch := l.input[l.pos]

		switch {
		case ch == '\n':
			l.emit(TokenNewline, "\n")
			l.advance()
			l.line++
			l.col = 1

		case ch == ' ' || ch == '\t':
			l.advance()

		case ch == '-' && l.peek() == '-':
			l.skipComment()

		case ch == '-' && l.peek() == '>':
			l.emit(TokenArrow, "->")
			l.advance()
			l.advance()

		case ch == '"':
			if err := l.readString(); err != nil {
				return nil, err
			}

		case unicode.IsDigit(ch):
			l.readNumber()

		case unicode.IsLetter(ch) || ch == '_':
			l.readIdentifier()

		case ch == '=':
			if l.peek() == '=' {
				l.emit(TokenEqual, "==")
				l.advance()
				l.advance()
			} else {
				l.emit(TokenAssign, "=")
				l.advance()
			}

		case ch == '!':
			if l.peek() == '=' {
				l.emit(TokenNotEqual, "!=")
				l.advance()
				l.advance()
			} else {
				return nil, fmt.Errorf("line %d, col %d: unexpected character '!'", l.line, l.col)
			}

		case ch == '<':
			if l.peek() == '=' {
				l.emit(TokenLessEqual, "<=")
				l.advance()
				l.advance()
			} else {
				l.emit(TokenLess, "<")
				l.advance()
			}

		case ch == '>':
			if l.peek() == '=' {
				l.emit(TokenGreaterEqual, ">=")
				l.advance()
				l.advance()
			} else {
				l.emit(TokenGreater, ">")
				l.advance()
			}

		case ch == '+':
			l.emit(TokenPlus, "+")
			l.advance()

		case ch == '-':
			l.emit(TokenMinus, "-")
			l.advance()

		case ch == '*':
			l.emit(TokenStar, "*")
			l.advance()

		case ch == '/':
			l.emit(TokenSlash, "/")
			l.advance()

		case ch == '%':
			l.emit(TokenModulo, "%")
			l.advance()

		case ch == '.':
			l.emit(TokenDot, ".")
			l.advance()

		case ch == '|':
			l.emit(TokenPipe, "|")
			l.advance()

		case ch == '(':
			l.emit(TokenLeftParen, "(")
			l.advance()

		case ch == ')':
			l.emit(TokenRightParen, ")")
			l.advance()

		case ch == '[':
			l.emit(TokenLeftBracket, "[")
			l.advance()

		case ch == ']':
			l.emit(TokenRightBracket, "]")
			l.advance()

		case ch == '{':
			l.emit(TokenLeftBrace, "{")
			l.advance()

		case ch == '}':
			l.emit(TokenRightBrace, "}")
			l.advance()

		case ch == ',':
			l.emit(TokenComma, ",")
			l.advance()

		case ch == ':':
			l.emit(TokenColon, ":")
			l.advance()

		default:
			return nil, fmt.Errorf("line %d, col %d: unexpected character %q", l.line, l.col, ch)
		}
	}

	// Emit remaining dedents
	for len(l.indents) > 1 {
		l.emit(TokenDedent, "")
		l.indents = l.indents[:len(l.indents)-1]
	}

	l.emit(TokenEOF, "")
	return l.tokens, nil
}

func (l *Lexer) advance() {
	l.pos++
	l.col++
}

func (l *Lexer) peek() rune {
	if l.pos+1 < len(l.input) {
		return l.input[l.pos+1]
	}
	return 0
}

func (l *Lexer) emit(tokenType TokenType, value string) {
	l.tokens = append(l.tokens, Token{
		Type:   tokenType,
		Value:  value,
		Line:   l.line,
		Column: l.col,
	})
}

func (l *Lexer) handleIndentation() {
	spaces := 0
	for l.pos < len(l.input) && l.input[l.pos] == ' ' {
		spaces++
		l.advance()
	}

	// Skip blank lines
	if l.pos >= len(l.input) || l.input[l.pos] == '\n' {
		return
	}
	// Skip comment-only lines for indentation purposes
	if l.pos+1 < len(l.input) && l.input[l.pos] == '-' && l.input[l.pos+1] == '-' {
		return
	}

	current := l.indents[len(l.indents)-1]
	if spaces > current {
		l.indents = append(l.indents, spaces)
		l.emit(TokenIndent, "")
	} else if spaces < current {
		for len(l.indents) > 1 && l.indents[len(l.indents)-1] > spaces {
			l.indents = l.indents[:len(l.indents)-1]
			l.emit(TokenDedent, "")
		}
	}
}

func (l *Lexer) skipComment() {
	// skip --
	l.advance()
	l.advance()
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.advance()
	}
}

func (l *Lexer) readString() error {
	l.advance() // skip opening "
	var b strings.Builder
	for l.pos < len(l.input) && l.input[l.pos] != '"' {
		if l.input[l.pos] == '\n' {
			return fmt.Errorf("line %d: unterminated string", l.line)
		}
		b.WriteRune(l.input[l.pos])
		l.advance()
	}
	if l.pos >= len(l.input) {
		return fmt.Errorf("line %d: unterminated string", l.line)
	}
	l.advance() // skip closing "
	l.emit(TokenString, b.String())
	return nil
}

func (l *Lexer) readNumber() {
	var b strings.Builder
	hasDecimal := false
	for l.pos < len(l.input) && (unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
		if l.input[l.pos] == '.' {
			if hasDecimal {
				break
			}
			hasDecimal = true
		}
		b.WriteRune(l.input[l.pos])
		l.advance()
	}
	l.emit(TokenNumber, b.String())
}

func (l *Lexer) readIdentifier() {
	var b strings.Builder
	for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
		b.WriteRune(l.input[l.pos])
		l.advance()
	}
	word := b.String()
	tokenType := LookupKeyword(word)
	l.emit(tokenType, word)
}
