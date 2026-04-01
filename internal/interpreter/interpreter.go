package interpreter

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/superciccio/drop-lang/internal/ast"
	"github.com/superciccio/drop-lang/internal/store"
	"github.com/superciccio/drop-lang/internal/ui"
	"github.com/superciccio/drop-lang/internal/web"
)

// ReturnSignal is used to unwind the call stack on return statements.
type ReturnSignal struct {
	Value interface{}
}

func (r ReturnSignal) Error() string {
	return "return signal"
}

type Environment struct {
	vars   map[string]interface{}
	funcs  map[string]*ast.FuncDecl
	parent *Environment
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		vars:   make(map[string]interface{}),
		funcs:  make(map[string]*ast.FuncDecl),
		parent: parent,
	}
}

func (e *Environment) Get(name string) (interface{}, bool) {
	if val, ok := e.vars[name]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

func (e *Environment) Set(name string, val interface{}) {
	e.vars[name] = val
}

func (e *Environment) GetFunc(name string) (*ast.FuncDecl, bool) {
	if fn, ok := e.funcs[name]; ok {
		return fn, true
	}
	if e.parent != nil {
		return e.parent.GetFunc(name)
	}
	return nil, false
}

func (e *Environment) SetFunc(name string, fn *ast.FuncDecl) {
	e.funcs[name] = fn
}

// RespondSignal is used to send HTTP responses from route handlers.
type RespondSignal struct {
	Value  interface{}
	Status int
}

func (r RespondSignal) Error() string {
	return "respond signal"
}

type Interpreter struct {
	env    *Environment
	server *web.Server
	stores map[string]*store.Store
}

func New() *Interpreter {
	return &Interpreter{
		env:    NewEnvironment(nil),
		stores: make(map[string]*store.Store),
	}
}

// Stop shuts down the running server, if any.
func (interp *Interpreter) Stop() {
	if interp.server != nil {
		interp.server.Stop()
	}
}

// HasServer returns true if the program configured a server via "serve".
func (interp *Interpreter) HasServer() bool {
	return interp.server != nil
}

func (interp *Interpreter) Run(program *ast.Program) error {
	for _, stmt := range program.Statements {
		if _, err := interp.exec(stmt, interp.env); err != nil {
			return err
		}
	}

	// If a server was configured, start it (blocking)
	if interp.server != nil {
		return interp.server.Start()
	}
	return nil
}

func (interp *Interpreter) exec(node ast.Node, env *Environment) (interface{}, error) {
	switch n := node.(type) {
	case *ast.Assignment:
		val, err := interp.eval(n.Value, env)
		if err != nil {
			return nil, err
		}
		env.Set(n.Name, val)
		return nil, nil

	case *ast.PrintStatement:
		val, err := interp.eval(n.Value, env)
		if err != nil {
			return nil, err
		}
		fmt.Println(stringify(val))
		return nil, nil

	case *ast.FuncDecl:
		env.SetFunc(n.Name, n)
		return nil, nil

	case *ast.ReturnStatement:
		val, err := interp.eval(n.Value, env)
		if err != nil {
			return nil, err
		}
		return nil, ReturnSignal{Value: val}

	case *ast.IfStatement:
		cond, err := interp.eval(n.Condition, env)
		if err != nil {
			return nil, err
		}
		if isTruthy(cond) {
			for _, s := range n.Body {
				if _, err := interp.exec(s, env); err != nil {
					return nil, err
				}
			}
		} else if n.Else != nil {
			for _, s := range n.Else {
				if _, err := interp.exec(s, env); err != nil {
					return nil, err
				}
			}
		}
		return nil, nil

	case *ast.ServeStatement:
		port, err := interp.eval(n.Port, env)
		if err != nil {
			return nil, err
		}
		portNum, ok := toNumber(port)
		if !ok {
			return nil, fmt.Errorf("serve requires a port number")
		}
		interp.server = web.NewServer(int(portNum))
		return nil, nil

	case *ast.RouteStatement:
		if interp.server == nil {
			return nil, fmt.Errorf("must call 'serve' before defining routes")
		}
		method := n.Method
		path := n.Path
		body := n.Body
		capturedEnv := env
		interp.server.AddRoute(method, path, func(params map[string]string, reqBody interface{}) (interface{}, int, error) {
			routeEnv := NewEnvironment(capturedEnv)
			// Set route params as variables
			for k, v := range params {
				routeEnv.Set(k, v)
			}
			// Set body
			if reqBody != nil {
				routeEnv.Set("__body__", reqBody)
			}
			for _, s := range body {
				if _, err := interp.exec(s, routeEnv); err != nil {
					if resp, ok := err.(RespondSignal); ok {
						return resp.Value, resp.Status, nil
					}
					return nil, 0, err
				}
			}
			return "OK", 200, nil
		})
		return nil, nil

	case *ast.RespondStatement:
		val, err := interp.eval(n.Value, env)
		if err != nil {
			return nil, err
		}
		status := 200
		if n.Status != nil {
			s, err := interp.eval(n.Status, env)
			if err != nil {
				return nil, err
			}
			if num, ok := toNumber(s); ok {
				status = int(num)
			}
		}
		return nil, RespondSignal{Value: val, Status: status}

	case *ast.StoreStatement:
		interp.stores[n.Name] = store.New(n.Name)
		return nil, nil

	case *ast.SaveStatement:
		s, ok := interp.stores[n.Store]
		if !ok {
			return nil, fmt.Errorf("unknown store: %s", n.Store)
		}
		val, err := interp.eval(n.Value, env)
		if err != nil {
			return nil, err
		}
		result := s.Save(val)
		return result, nil

	case *ast.RemoveStatement:
		s, ok := interp.stores[n.Store]
		if !ok {
			return nil, fmt.Errorf("unknown store: %s", n.Store)
		}
		id, err := interp.eval(n.ID, env)
		if err != nil {
			return nil, err
		}
		s.Remove(stringify(id))
		return nil, nil

	case *ast.PageStatement:
		body, err := interp.renderChildren(n.Children, env)
		if err != nil {
			return nil, err
		}
		page := ui.RenderPage(n.Title, body)
		return nil, RespondSignal{Value: page, Status: 200}

	case *ast.ForStatement:
		iter, err := interp.eval(n.Iterable, env)
		if err != nil {
			return nil, err
		}
		list, ok := iter.([]interface{})
		if !ok {
			return nil, fmt.Errorf("'for' requires a list, got %T", iter)
		}
		loopEnv := NewEnvironment(env)
		for _, item := range list {
			loopEnv.Set(n.VarName, item)
			for _, s := range n.Body {
				if _, err := interp.exec(s, loopEnv); err != nil {
					return nil, err
				}
			}
		}
		return nil, nil

	default:
		// Try evaluating as expression (for function calls as statements, etc.)
		val, err := interp.eval(node, env)
		if err != nil {
			return nil, err
		}
		return val, nil
	}
}

func (interp *Interpreter) eval(node ast.Node, env *Environment) (interface{}, error) {
	switch n := node.(type) {
	case *ast.NumberLiteral:
		return n.Value, nil

	case *ast.StringLiteral:
		return interp.interpolateString(n.Value, env), nil

	case *ast.BoolLiteral:
		return n.Value, nil

	case *ast.Identifier:
		val, ok := env.Get(n.Name)
		if !ok {
			return nil, fmt.Errorf("undefined variable: %s", n.Name)
		}
		return val, nil

	case *ast.BodyExpr:
		val, ok := env.Get("__body__")
		if !ok {
			return nil, nil
		}
		return val, nil

	case *ast.LoadExpr:
		s, ok := interp.stores[n.Store]
		if !ok {
			return nil, fmt.Errorf("unknown store: %s", n.Store)
		}
		return s.Load(), nil

	case *ast.FetchExpr:
		urlVal, err := interp.eval(n.URL, env)
		if err != nil {
			return nil, err
		}
		url, ok := urlVal.(string)
		if !ok {
			return nil, fmt.Errorf("fetch requires a string URL, got %T", urlVal)
		}
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("fetch failed: %v", err)
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("fetch: error reading response: %v", err)
		}
		// Try to parse as JSON
		var jsonData interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
			return jsonData, nil
		}
		// Fall back to string
		return string(bodyBytes), nil

	case *ast.ListLiteral:
		var elems []interface{}
		for _, e := range n.Elements {
			val, err := interp.eval(e, env)
			if err != nil {
				return nil, err
			}
			elems = append(elems, val)
		}
		return elems, nil

	case *ast.MapLiteral:
		m := make(map[string]interface{})
		for i, key := range n.Keys {
			val, err := interp.eval(n.Values[i], env)
			if err != nil {
				return nil, err
			}
			m[key] = val
		}
		return m, nil

	case *ast.BinaryExpr:
		left, err := interp.eval(n.Left, env)
		if err != nil {
			return nil, err
		}
		right, err := interp.eval(n.Right, env)
		if err != nil {
			return nil, err
		}
		return interp.evalBinary(n.Operator, left, right)

	case *ast.UnaryExpr:
		operand, err := interp.eval(n.Operand, env)
		if err != nil {
			return nil, err
		}
		switch n.Operator {
		case "not":
			return !isTruthy(operand), nil
		case "-":
			num, ok := toNumber(operand)
			if !ok {
				return nil, fmt.Errorf("cannot negate %T", operand)
			}
			return -num, nil
		}
		return nil, fmt.Errorf("unknown unary operator: %s", n.Operator)

	case *ast.DotExpr:
		obj, err := interp.eval(n.Object, env)
		if err != nil {
			return nil, err
		}
		m, ok := obj.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot access property '%s' on %T", n.Property, obj)
		}
		return m[n.Property], nil

	case *ast.IndexExpr:
		obj, err := interp.eval(n.Object, env)
		if err != nil {
			return nil, err
		}
		idx, err := interp.eval(n.Index, env)
		if err != nil {
			return nil, err
		}
		switch o := obj.(type) {
		case []interface{}:
			i, ok := toNumber(idx)
			if !ok {
				return nil, fmt.Errorf("list index must be a number")
			}
			index := int(i)
			if index < 0 || index >= len(o) {
				return nil, fmt.Errorf("index %d out of range (length %d)", index, len(o))
			}
			return o[index], nil
		case map[string]interface{}:
			key, ok := idx.(string)
			if !ok {
				return nil, fmt.Errorf("map key must be a string")
			}
			return o[key], nil
		default:
			return nil, fmt.Errorf("cannot index %T", obj)
		}

	case *ast.CallExpr:
		fn, ok := env.GetFunc(n.Name)
		if !ok {
			return nil, fmt.Errorf("undefined function: %s", n.Name)
		}
		if len(n.Args) != len(fn.Params) {
			return nil, fmt.Errorf("function '%s' expects %d args, got %d", n.Name, len(fn.Params), len(n.Args))
		}
		callEnv := NewEnvironment(env)
		for i, param := range fn.Params {
			val, err := interp.eval(n.Args[i], env)
			if err != nil {
				return nil, err
			}
			callEnv.Set(param, val)
		}
		for _, s := range fn.Body {
			if _, err := interp.exec(s, callEnv); err != nil {
				if ret, ok := err.(ReturnSignal); ok {
					return ret.Value, nil
				}
				return nil, err
			}
		}
		return nil, nil

	default:
		return nil, fmt.Errorf("cannot evaluate node type %T", node)
	}
}

func (interp *Interpreter) evalBinary(op string, left, right interface{}) (interface{}, error) {
	// String concatenation with +
	if op == "+" {
		ls, lok := left.(string)
		rs, rok := right.(string)
		if lok && rok {
			return ls + rs, nil
		}
	}

	// Numeric operations
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		switch op {
		case "+":
			return ln + rn, nil
		case "-":
			return ln - rn, nil
		case "*":
			return ln * rn, nil
		case "/":
			if rn == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return ln / rn, nil
		case "%":
			if rn == 0 {
				return nil, fmt.Errorf("modulo by zero")
			}
			return float64(int(ln) % int(rn)), nil
		case "<":
			return ln < rn, nil
		case "<=":
			return ln <= rn, nil
		case ">":
			return ln > rn, nil
		case ">=":
			return ln >= rn, nil
		case "==":
			return ln == rn, nil
		case "!=":
			return ln != rn, nil
		}
	}

	// Generic equality
	switch op {
	case "==":
		return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right), nil
	case "!=":
		return fmt.Sprintf("%v", left) != fmt.Sprintf("%v", right), nil
	}

	// Boolean operators
	switch op {
	case "and":
		return isTruthy(left) && isTruthy(right), nil
	case "or":
		return isTruthy(left) || isTruthy(right), nil
	}

	return nil, fmt.Errorf("unsupported operation: %T %s %T", left, op, right)
}

func (interp *Interpreter) interpolateString(s string, env *Environment) string {
	var result strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '{' {
			j := i + 1
			for j < len(runes) && runes[j] != '}' {
				j++
			}
			if j < len(runes) {
				expr := string(runes[i+1 : j])
				val := interp.resolveInterpolation(expr, env)
				result.WriteString(stringify(val))
				i = j
				continue
			}
		}
		result.WriteRune(runes[i])
	}
	return result.String()
}

func (interp *Interpreter) resolveInterpolation(expr string, env *Environment) interface{} {
	// Handle dot access like "user.name"
	parts := strings.Split(expr, ".")
	val, ok := env.Get(parts[0])
	if !ok {
		return "{" + expr + "}"
	}
	for _, part := range parts[1:] {
		m, ok := val.(map[string]interface{})
		if !ok {
			return "{" + expr + "}"
		}
		val = m[part]
	}
	return val
}

// --- UI rendering ---

func (interp *Interpreter) renderChildren(nodes []ast.Node, env *Environment) (string, error) {
	var b strings.Builder
	for _, node := range nodes {
		s, err := interp.renderNode(node, env)
		if err != nil {
			return "", err
		}
		b.WriteString(s)
	}
	return b.String(), nil
}

func (interp *Interpreter) renderNode(node ast.Node, env *Environment) (string, error) {
	switch n := node.(type) {
	case *ast.TextElement:
		val, err := interp.evalExpr(n.Value, env)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`<div class="text-block">%s</div>`, html.EscapeString(stringify(val))), nil

	case *ast.RowElement:
		var b strings.Builder
		b.WriteString(`<div class="row">`)
		for _, child := range n.Children {
			s, err := interp.renderNode(child, env)
			if err != nil {
				return "", err
			}
			b.WriteString(s)
		}
		b.WriteString(`</div>`)
		return b.String(), nil

	case *ast.EachElement:
		iter, err := interp.evalExpr(n.Iterable, env)
		if err != nil {
			return "", err
		}
		list, ok := iter.([]interface{})
		if !ok {
			return "", fmt.Errorf("'each' requires a list")
		}
		var b strings.Builder
		for _, item := range list {
			loopEnv := NewEnvironment(env)
			loopEnv.Set(n.VarName, item)
			for _, child := range n.Children {
				s, err := interp.renderNode(child, loopEnv)
				if err != nil {
					return "", err
				}
				b.WriteString(s)
			}
		}
		return b.String(), nil

	case *ast.ButtonElement:
		action := interp.interpolateString(n.Action, env)
		isDanger := n.Method == "DELETE"
		class := ""
		if isDanger {
			class = ` class="danger"`
		}
		return fmt.Sprintf(
			`<form method="POST" action="%s" style="display:inline">`+
				`<input type="hidden" name="_method" value="%s">`+
				`<button type="submit"%s>%s</button>`+
				`</form>`,
			html.EscapeString(action),
			html.EscapeString(n.Method),
			class,
			html.EscapeString(n.Label),
		), nil

	case *ast.FormElement:
		var b strings.Builder
		b.WriteString(`<div class="form-section">`)
		b.WriteString(fmt.Sprintf(`<div class="form-title">%s</div>`, html.EscapeString(n.Title)))
		// Find submit to get the action
		method, action := "POST", ""
		for _, child := range n.Children {
			if sub, ok := child.(*ast.SubmitElement); ok {
				method = sub.Method
				action = sub.Action
			}
		}
		b.WriteString(fmt.Sprintf(`<form method="POST" action="%s">`, html.EscapeString(action)))
		if method != "POST" {
			b.WriteString(fmt.Sprintf(`<input type="hidden" name="_method" value="%s">`, html.EscapeString(method)))
		}
		for _, child := range n.Children {
			s, err := interp.renderNode(child, env)
			if err != nil {
				return "", err
			}
			b.WriteString(s)
		}
		b.WriteString(`</form></div>`)
		return b.String(), nil

	case *ast.InputElement:
		return fmt.Sprintf(
			`<input type="text" name="%s" placeholder="%s">`,
			html.EscapeString(n.Name),
			html.EscapeString(n.Name),
		), nil

	case *ast.SubmitElement:
		return fmt.Sprintf(`<button type="submit">%s</button>`, html.EscapeString(n.Label)), nil

	case *ast.LinkElement:
		return fmt.Sprintf(
			`<a href="%s">%s</a>`,
			html.EscapeString(n.Href),
			html.EscapeString(n.Text),
		), nil

	case *ast.ImageElement:
		alt := n.Alt
		if alt == "" {
			alt = "image"
		}
		return fmt.Sprintf(
			`<img src="%s" alt="%s">`,
			html.EscapeString(n.Src),
			html.EscapeString(alt),
		), nil

	default:
		return "", fmt.Errorf("unsupported UI element: %T", node)
	}
}

// evalExpr is a wrapper around eval for use in UI rendering contexts.
func (interp *Interpreter) evalExpr(node ast.Node, env *Environment) (interface{}, error) {
	return interp.eval(node, env)
}

func isTruthy(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case string:
		return val != ""
	case []interface{}:
		return len(val) > 0
	case map[string]interface{}:
		return len(val) > 0
	default:
		return true
	}
}

func toNumber(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	default:
		return 0, false
	}
}

func stringify(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int(val)) {
			return fmt.Sprintf("%d", int(val))
		}
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []interface{}:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = stringify(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]interface{}:
		parts := make([]string, 0, len(val))
		for k, v := range val {
			parts = append(parts, k+": "+stringify(v))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	default:
		return fmt.Sprintf("%v", val)
	}
}
