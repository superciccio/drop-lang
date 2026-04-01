package ast

// Node is the interface all AST nodes implement.
type Node interface {
	nodeType() string
}

// Program is the root node.
type Program struct {
	Statements []Node
}

func (p *Program) nodeType() string { return "Program" }

// --- Expressions ---

type NumberLiteral struct {
	Value float64
}

func (n *NumberLiteral) nodeType() string { return "NumberLiteral" }

type StringLiteral struct {
	Value string
}

func (s *StringLiteral) nodeType() string { return "StringLiteral" }

type BoolLiteral struct {
	Value bool
}

func (b *BoolLiteral) nodeType() string { return "BoolLiteral" }

type Identifier struct {
	Name string
}

func (i *Identifier) nodeType() string { return "Identifier" }

type BinaryExpr struct {
	Left     Node
	Operator string
	Right    Node
}

func (b *BinaryExpr) nodeType() string { return "BinaryExpr" }

type UnaryExpr struct {
	Operator string
	Operand  Node
}

func (u *UnaryExpr) nodeType() string { return "UnaryExpr" }

type ListLiteral struct {
	Elements []Node
}

func (l *ListLiteral) nodeType() string { return "ListLiteral" }

type MapLiteral struct {
	Keys   []string
	Values []Node
}

func (m *MapLiteral) nodeType() string { return "MapLiteral" }

type DotExpr struct {
	Object   Node
	Property string
}

func (d *DotExpr) nodeType() string { return "DotExpr" }

type IndexExpr struct {
	Object Node
	Index  Node
}

func (i *IndexExpr) nodeType() string { return "IndexExpr" }

type CallExpr struct {
	Name string
	Args []Node
}

func (c *CallExpr) nodeType() string { return "CallExpr" }

type LoadExpr struct {
	Store string
}

func (l *LoadExpr) nodeType() string { return "LoadExpr" }

type BodyExpr struct{}

func (b *BodyExpr) nodeType() string { return "BodyExpr" }

type FetchExpr struct {
	URL Node
}

func (f *FetchExpr) nodeType() string { return "FetchExpr" }

// --- Statements ---

type Assignment struct {
	Name  string
	Value Node
}

func (a *Assignment) nodeType() string { return "Assignment" }

type PrintStatement struct {
	Value Node
}

func (p *PrintStatement) nodeType() string { return "PrintStatement" }

type FuncDecl struct {
	Name   string
	Params []string
	Body   []Node
}

func (f *FuncDecl) nodeType() string { return "FuncDecl" }

type ReturnStatement struct {
	Value Node
}

func (r *ReturnStatement) nodeType() string { return "ReturnStatement" }

type IfStatement struct {
	Condition Node
	Body      []Node
	Else      []Node
}

func (i *IfStatement) nodeType() string { return "IfStatement" }

type ForStatement struct {
	VarName  string
	Iterable Node
	Body     []Node
}

func (f *ForStatement) nodeType() string { return "ForStatement" }

type ServeStatement struct {
	Port Node
}

func (s *ServeStatement) nodeType() string { return "ServeStatement" }

type RouteStatement struct {
	Method string
	Path   string
	Body   []Node
}

func (r *RouteStatement) nodeType() string { return "RouteStatement" }

type RespondStatement struct {
	Value  Node
	Status Node // optional, can be nil
}

func (r *RespondStatement) nodeType() string { return "RespondStatement" }

type StoreStatement struct {
	Name string
}

func (s *StoreStatement) nodeType() string { return "StoreStatement" }

type SaveStatement struct {
	Store string
	Value Node
}

func (s *SaveStatement) nodeType() string { return "SaveStatement" }

type RemoveStatement struct {
	Store string
	ID    Node
}

func (r *RemoveStatement) nodeType() string { return "RemoveStatement" }

// --- UI Nodes ---

type PageStatement struct {
	Title    string
	Children []Node
}

func (p *PageStatement) nodeType() string { return "PageStatement" }

type TextElement struct {
	Value Node
}

func (t *TextElement) nodeType() string { return "TextElement" }

type RowElement struct {
	Children []Node
}

func (r *RowElement) nodeType() string { return "RowElement" }

type EachElement struct {
	VarName  string
	Iterable Node
	Children []Node
}

func (e *EachElement) nodeType() string { return "EachElement" }

type ButtonElement struct {
	Label  string
	Method string // HTTP method for action
	Action string // URL for action
}

func (b *ButtonElement) nodeType() string { return "ButtonElement" }

type FormElement struct {
	Title    string
	Children []Node
}

func (f *FormElement) nodeType() string { return "FormElement" }

type InputElement struct {
	Name string
}

func (i *InputElement) nodeType() string { return "InputElement" }

type SubmitElement struct {
	Label  string
	Method string
	Action string
}

func (s *SubmitElement) nodeType() string { return "SubmitElement" }

type LinkElement struct {
	Text string
	Href string
}

func (l *LinkElement) nodeType() string { return "LinkElement" }

type ImageElement struct {
	Src string
	Alt string
}

func (i *ImageElement) nodeType() string { return "ImageElement" }
