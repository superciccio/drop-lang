package parser

import (
	"fmt"
	"strconv"

	"github.com/superciccio/drop-lang/internal/ast"
	"github.com/superciccio/drop-lang/internal/lexer"
)

type Parser struct {
	tokens  []lexer.Token
	pos     int
	current lexer.Token
}

func New(tokens []lexer.Token) *Parser {
	// Filter out newlines and comments (keep indent/dedent for block structure)
	filtered := make([]lexer.Token, 0, len(tokens))
	for _, t := range tokens {
		if t.Type != lexer.TokenNewline && t.Type != lexer.TokenComment {
			filtered = append(filtered, t)
		}
	}
	p := &Parser{tokens: filtered}
	if len(filtered) > 0 {
		p.current = filtered[0]
	}
	return p
}

func (p *Parser) Parse() (*ast.Program, error) {
	program := &ast.Program{}
	for p.current.Type != lexer.TokenEOF {
		// Skip stray dedents at top level
		if p.current.Type == lexer.TokenDedent {
			p.advance()
			continue
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}
	return program, nil
}

func (p *Parser) advance() {
	p.pos++
	if p.pos < len(p.tokens) {
		p.current = p.tokens[p.pos]
	} else {
		p.current = lexer.Token{Type: lexer.TokenEOF}
	}
}

func (p *Parser) peek() lexer.Token {
	if p.pos+1 < len(p.tokens) {
		return p.tokens[p.pos+1]
	}
	return lexer.Token{Type: lexer.TokenEOF}
}

func (p *Parser) expect(t lexer.TokenType) error {
	if p.current.Type != t {
		return fmt.Errorf("line %d: expected %s, got %s (%q)", p.current.Line, t, p.current.Type, p.current.Value)
	}
	p.advance()
	return nil
}

func (p *Parser) parseStatement() (ast.Node, error) {
	switch p.current.Type {
	case lexer.TokenPrint:
		return p.parsePrint()
	case lexer.TokenDo:
		return p.parseFuncDecl()
	case lexer.TokenIf:
		return p.parseIf()
	case lexer.TokenFor:
		return p.parseFor()
	case lexer.TokenReturn:
		return p.parseReturn()
	case lexer.TokenServe:
		return p.parseServe()
	case lexer.TokenGet, lexer.TokenPost, lexer.TokenPut, lexer.TokenDelete:
		return p.parseRoute()
	case lexer.TokenRespond:
		return p.parseRespond()
	case lexer.TokenStore:
		return p.parseStore()
	case lexer.TokenSave:
		return p.parseSave()
	case lexer.TokenRemove:
		return p.parseRemove()
	case lexer.TokenPage:
		return p.parsePage()
	case lexer.TokenText:
		return p.parseText()
	case lexer.TokenRow:
		return p.parseRow()
	case lexer.TokenEach:
		return p.parseEach()
	case lexer.TokenButton:
		return p.parseButton()
	case lexer.TokenForm:
		return p.parseForm()
	case lexer.TokenInput:
		return p.parseInput()
	case lexer.TokenSubmit:
		return p.parseSubmit()
	case lexer.TokenLink:
		return p.parseLink()
	case lexer.TokenImage:
		return p.parseImage()
	case lexer.TokenIdentifier:
		// Could be assignment (name = ...) or function call (name(...))
		if p.peek().Type == lexer.TokenAssign {
			return p.parseAssignment()
		}
		return p.parseExpression()
	default:
		return p.parseExpression()
	}
}

func (p *Parser) parsePrint() (*ast.PrintStatement, error) {
	p.advance() // skip 'print'
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return &ast.PrintStatement{Value: expr}, nil
}

func (p *Parser) parseFuncDecl() (*ast.FuncDecl, error) {
	p.advance() // skip 'do'
	if p.current.Type != lexer.TokenIdentifier {
		return nil, fmt.Errorf("line %d: expected function name after 'do'", p.current.Line)
	}
	name := p.current.Value
	p.advance()

	if err := p.expect(lexer.TokenLeftParen); err != nil {
		return nil, err
	}

	var params []string
	for p.current.Type != lexer.TokenRightParen {
		if p.current.Type != lexer.TokenIdentifier {
			return nil, fmt.Errorf("line %d: expected parameter name", p.current.Line)
		}
		params = append(params, p.current.Value)
		p.advance()
		if p.current.Type == lexer.TokenComma {
			p.advance()
		}
	}
	p.advance() // skip ')'

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &ast.FuncDecl{Name: name, Params: params, Body: body}, nil
}

func (p *Parser) parseIf() (*ast.IfStatement, error) {
	p.advance() // skip 'if'
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	var elseBody []ast.Node
	if p.current.Type == lexer.TokenElse {
		p.advance()
		elseBody, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}

	return &ast.IfStatement{Condition: cond, Body: body, Else: elseBody}, nil
}

func (p *Parser) parseFor() (*ast.ForStatement, error) {
	p.advance() // skip 'for'
	if p.current.Type != lexer.TokenIdentifier {
		return nil, fmt.Errorf("line %d: expected variable name after 'for'", p.current.Line)
	}
	varName := p.current.Value
	p.advance()

	if err := p.expect(lexer.TokenIn); err != nil {
		return nil, err
	}

	iterable, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &ast.ForStatement{VarName: varName, Iterable: iterable, Body: body}, nil
}

func (p *Parser) parseReturn() (*ast.ReturnStatement, error) {
	p.advance() // skip 'return'
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return &ast.ReturnStatement{Value: expr}, nil
}

func (p *Parser) parseAssignment() (*ast.Assignment, error) {
	name := p.current.Value
	p.advance() // skip identifier
	p.advance() // skip '='
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return &ast.Assignment{Name: name, Value: value}, nil
}

func (p *Parser) parseServe() (*ast.ServeStatement, error) {
	p.advance() // skip 'serve'
	port, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return &ast.ServeStatement{Port: port}, nil
}

func (p *Parser) parseRoute() (*ast.RouteStatement, error) {
	method := p.current.Value
	p.advance()

	if p.current.Type != lexer.TokenString {
		return nil, fmt.Errorf("line %d: expected route path string after '%s'", p.current.Line, method)
	}
	path := p.current.Value
	p.advance()

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &ast.RouteStatement{Method: method, Path: path, Body: body}, nil
}

func (p *Parser) parseRespond() (*ast.RespondStatement, error) {
	p.advance() // skip 'respond'
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Optional status code
	var status ast.Node
	if p.current.Type == lexer.TokenNumber {
		status, err = p.parsePrimary()
		if err != nil {
			return nil, err
		}
	}

	return &ast.RespondStatement{Value: value, Status: status}, nil
}

func (p *Parser) parseStore() (*ast.StoreStatement, error) {
	p.advance() // skip 'store'
	if p.current.Type != lexer.TokenString {
		return nil, fmt.Errorf("line %d: expected store name string", p.current.Line)
	}
	name := p.current.Value
	p.advance()
	return &ast.StoreStatement{Name: name}, nil
}

func (p *Parser) parseSave() (*ast.SaveStatement, error) {
	p.advance() // skip 'save'
	if p.current.Type != lexer.TokenString {
		return nil, fmt.Errorf("line %d: expected store name string after 'save'", p.current.Line)
	}
	store := p.current.Value
	p.advance()

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return &ast.SaveStatement{Store: store, Value: value}, nil
}

func (p *Parser) parseRemove() (*ast.RemoveStatement, error) {
	p.advance() // skip 'remove'
	if p.current.Type != lexer.TokenString {
		return nil, fmt.Errorf("line %d: expected store name string after 'remove'", p.current.Line)
	}
	store := p.current.Value
	p.advance()

	id, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return &ast.RemoveStatement{Store: store, ID: id}, nil
}

func (p *Parser) parseBlock() ([]ast.Node, error) {
	if p.current.Type != lexer.TokenIndent {
		return nil, fmt.Errorf("line %d: expected indented block", p.current.Line)
	}
	p.advance() // skip INDENT

	var stmts []ast.Node
	for p.current.Type != lexer.TokenDedent && p.current.Type != lexer.TokenEOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	if p.current.Type == lexer.TokenDedent {
		p.advance()
	}

	return stmts, nil
}

// --- UI parsing ---

func (p *Parser) parsePage() (*ast.PageStatement, error) {
	p.advance() // skip 'page'
	if p.current.Type != lexer.TokenString {
		return nil, fmt.Errorf("line %d: expected page title string", p.current.Line)
	}
	title := p.current.Value
	p.advance()

	children, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.PageStatement{Title: title, Children: children}, nil
}

func (p *Parser) parseText() (*ast.TextElement, error) {
	p.advance() // skip 'text'
	val, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return &ast.TextElement{Value: val}, nil
}

func (p *Parser) parseRow() (*ast.RowElement, error) {
	p.advance() // skip 'row'
	children, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.RowElement{Children: children}, nil
}

func (p *Parser) parseEach() (*ast.EachElement, error) {
	p.advance() // skip 'each'
	if p.current.Type != lexer.TokenIdentifier {
		return nil, fmt.Errorf("line %d: expected variable name after 'each'", p.current.Line)
	}
	varName := p.current.Value
	p.advance()

	if err := p.expect(lexer.TokenIn); err != nil {
		return nil, err
	}

	iterable, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	children, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.EachElement{VarName: varName, Iterable: iterable, Children: children}, nil
}

func (p *Parser) parseButton() (*ast.ButtonElement, error) {
	p.advance() // skip 'button'
	if p.current.Type != lexer.TokenString {
		return nil, fmt.Errorf("line %d: expected button label string", p.current.Line)
	}
	label := p.current.Value
	p.advance()

	method, action, err := p.parseAction()
	if err != nil {
		return nil, err
	}
	return &ast.ButtonElement{Label: label, Method: method, Action: action}, nil
}

func (p *Parser) parseForm() (*ast.FormElement, error) {
	p.advance() // skip 'form'
	if p.current.Type != lexer.TokenString {
		return nil, fmt.Errorf("line %d: expected form title string", p.current.Line)
	}
	title := p.current.Value
	p.advance()

	children, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.FormElement{Title: title, Children: children}, nil
}

func (p *Parser) parseInput() (*ast.InputElement, error) {
	p.advance() // skip 'input'
	if p.current.Type != lexer.TokenString {
		return nil, fmt.Errorf("line %d: expected input name string", p.current.Line)
	}
	name := p.current.Value
	p.advance()
	return &ast.InputElement{Name: name}, nil
}

func (p *Parser) parseSubmit() (*ast.SubmitElement, error) {
	p.advance() // skip 'submit'
	if p.current.Type != lexer.TokenString {
		return nil, fmt.Errorf("line %d: expected submit label string", p.current.Line)
	}
	label := p.current.Value
	p.advance()

	method, action, err := p.parseAction()
	if err != nil {
		return nil, err
	}
	return &ast.SubmitElement{Label: label, Method: method, Action: action}, nil
}

func (p *Parser) parseLink() (*ast.LinkElement, error) {
	p.advance() // skip 'link'
	if p.current.Type != lexer.TokenString {
		return nil, fmt.Errorf("line %d: expected link text string", p.current.Line)
	}
	text := p.current.Value
	p.advance()

	if p.current.Type != lexer.TokenString {
		return nil, fmt.Errorf("line %d: expected link href string", p.current.Line)
	}
	href := p.current.Value
	p.advance()
	return &ast.LinkElement{Text: text, Href: href}, nil
}

func (p *Parser) parseImage() (*ast.ImageElement, error) {
	p.advance() // skip 'image'
	if p.current.Type != lexer.TokenString {
		return nil, fmt.Errorf("line %d: expected image src string", p.current.Line)
	}
	src := p.current.Value
	p.advance()

	alt := ""
	if p.current.Type == lexer.TokenString {
		alt = p.current.Value
		p.advance()
	}
	return &ast.ImageElement{Src: src, Alt: alt}, nil
}

// parseAction parses `-> method "path"` (e.g. `-> delete "/todos/1"`)
func (p *Parser) parseAction() (string, string, error) {
	if p.current.Type != lexer.TokenArrow {
		return "", "", fmt.Errorf("line %d: expected '->' for action", p.current.Line)
	}
	p.advance() // skip ->

	// method keyword: get, post, put, delete
	var method string
	switch p.current.Type {
	case lexer.TokenGet:
		method = "GET"
	case lexer.TokenPost:
		method = "POST"
	case lexer.TokenPut:
		method = "PUT"
	case lexer.TokenDelete:
		method = "DELETE"
	default:
		return "", "", fmt.Errorf("line %d: expected HTTP method after '->'", p.current.Line)
	}
	p.advance()

	if p.current.Type != lexer.TokenString {
		return "", "", fmt.Errorf("line %d: expected URL string after method", p.current.Line)
	}
	action := p.current.Value
	p.advance()

	return method, action, nil
}

// --- Expression parsing (precedence climbing) ---

func (p *Parser) parseExpression() (ast.Node, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (ast.Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.current.Type == lexer.TokenOr {
		op := p.current.Value
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Left: left, Operator: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAnd() (ast.Node, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for p.current.Type == lexer.TokenAnd {
		op := p.current.Value
		p.advance()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Left: left, Operator: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseComparison() (ast.Node, error) {
	left, err := p.parseAddSub()
	if err != nil {
		return nil, err
	}
	for p.current.Type == lexer.TokenEqual || p.current.Type == lexer.TokenNotEqual ||
		p.current.Type == lexer.TokenLess || p.current.Type == lexer.TokenLessEqual ||
		p.current.Type == lexer.TokenGreater || p.current.Type == lexer.TokenGreaterEqual {
		op := p.current.Value
		p.advance()
		right, err := p.parseAddSub()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Left: left, Operator: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAddSub() (ast.Node, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return nil, err
	}
	for p.current.Type == lexer.TokenPlus || p.current.Type == lexer.TokenMinus {
		op := p.current.Value
		p.advance()
		right, err := p.parseMulDiv()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Left: left, Operator: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseMulDiv() (ast.Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.current.Type == lexer.TokenStar || p.current.Type == lexer.TokenSlash || p.current.Type == lexer.TokenModulo {
		op := p.current.Value
		p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Left: left, Operator: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseUnary() (ast.Node, error) {
	if p.current.Type == lexer.TokenNot || p.current.Type == lexer.TokenMinus {
		op := p.current.Value
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Operator: op, Operand: operand}, nil
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() (ast.Node, error) {
	node, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		if p.current.Type == lexer.TokenDot {
			p.advance()
			if p.current.Type != lexer.TokenIdentifier {
				return nil, fmt.Errorf("line %d: expected property name after '.'", p.current.Line)
			}
			node = &ast.DotExpr{Object: node, Property: p.current.Value}
			p.advance()
		} else if p.current.Type == lexer.TokenLeftBracket {
			p.advance()
			index, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			if err := p.expect(lexer.TokenRightBracket); err != nil {
				return nil, err
			}
			node = &ast.IndexExpr{Object: node, Index: index}
		} else {
			break
		}
	}

	return node, nil
}

func (p *Parser) parsePrimary() (ast.Node, error) {
	switch p.current.Type {
	case lexer.TokenNumber:
		val, err := strconv.ParseFloat(p.current.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid number %q", p.current.Line, p.current.Value)
		}
		p.advance()
		return &ast.NumberLiteral{Value: val}, nil

	case lexer.TokenString:
		val := p.current.Value
		p.advance()
		return &ast.StringLiteral{Value: val}, nil

	case lexer.TokenTrue:
		p.advance()
		return &ast.BoolLiteral{Value: true}, nil

	case lexer.TokenFalse:
		p.advance()
		return &ast.BoolLiteral{Value: false}, nil

	case lexer.TokenBody:
		p.advance()
		return &ast.BodyExpr{}, nil

	case lexer.TokenLoad:
		p.advance()
		if p.current.Type != lexer.TokenString {
			return nil, fmt.Errorf("line %d: expected store name after 'load'", p.current.Line)
		}
		store := p.current.Value
		p.advance()
		return &ast.LoadExpr{Store: store}, nil

	case lexer.TokenFetch:
		p.advance() // skip 'fetch'
		url, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ast.FetchExpr{URL: url}, nil

	case lexer.TokenIdentifier:
		name := p.current.Value
		p.advance()
		// Check if it's a function call
		if p.current.Type == lexer.TokenLeftParen {
			p.advance()
			var args []ast.Node
			for p.current.Type != lexer.TokenRightParen {
				arg, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if p.current.Type == lexer.TokenComma {
					p.advance()
				}
			}
			p.advance() // skip ')'
			return &ast.CallExpr{Name: name, Args: args}, nil
		}
		return &ast.Identifier{Name: name}, nil

	case lexer.TokenLeftParen:
		p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.expect(lexer.TokenRightParen); err != nil {
			return nil, err
		}
		return expr, nil

	case lexer.TokenLeftBracket:
		return p.parseList()

	case lexer.TokenLeftBrace:
		return p.parseMap()

	default:
		return nil, fmt.Errorf("line %d: unexpected token %s (%q)", p.current.Line, p.current.Type, p.current.Value)
	}
}

func (p *Parser) parseList() (ast.Node, error) {
	p.advance() // skip '['
	var elements []ast.Node
	for p.current.Type != lexer.TokenRightBracket {
		elem, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
		if p.current.Type == lexer.TokenComma {
			p.advance()
		}
	}
	p.advance() // skip ']'
	return &ast.ListLiteral{Elements: elements}, nil
}

func (p *Parser) parseMap() (ast.Node, error) {
	p.advance() // skip '{'
	var keys []string
	var values []ast.Node
	for p.current.Type != lexer.TokenRightBrace {
		if p.current.Type != lexer.TokenIdentifier && p.current.Type != lexer.TokenString {
			return nil, fmt.Errorf("line %d: expected map key", p.current.Line)
		}
		key := p.current.Value
		p.advance()
		if err := p.expect(lexer.TokenColon); err != nil {
			return nil, err
		}
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
		values = append(values, val)
		if p.current.Type == lexer.TokenComma {
			p.advance()
		}
	}
	p.advance() // skip '}'
	return &ast.MapLiteral{Keys: keys, Values: values}, nil
}
