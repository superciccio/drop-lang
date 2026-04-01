package lexer

type TokenType int

const (
	// Literals
	TokenIdentifier TokenType = iota
	TokenNumber
	TokenString
	TokenTrue
	TokenFalse

	// Operators
	TokenAssign       // =
	TokenPlus         // +
	TokenMinus        // -
	TokenStar         // *
	TokenSlash        // /
	TokenEqual        // ==
	TokenNotEqual     // !=
	TokenLess         // <
	TokenLessEqual    // <=
	TokenGreater      // >
	TokenGreaterEqual // >=
	TokenAnd          // and
	TokenOr           // or
	TokenNot          // not
	TokenDot          // .
	TokenPipe         // |
	TokenArrow        // ->
	TokenModulo       // %

	// Delimiters
	TokenLeftParen    // (
	TokenRightParen   // )
	TokenLeftBracket  // [
	TokenRightBracket // ]
	TokenLeftBrace    // {
	TokenRightBrace   // }
	TokenComma        // ,
	TokenColon        // :

	// Keywords
	TokenDo
	TokenReturn
	TokenIf
	TokenElse
	TokenFor
	TokenIn
	TokenServe
	TokenGet
	TokenPost
	TokenPut
	TokenDelete
	TokenRespond
	TokenBody
	TokenStore
	TokenSave
	TokenLoad
	TokenRemove
	TokenPage
	TokenText
	TokenRow
	TokenEach
	TokenButton
	TokenForm
	TokenInput
	TokenSubmit
	TokenLink
	TokenImage
	TokenFetch
	TokenPrint

	// Structure
	TokenNewline
	TokenIndent
	TokenDedent
	TokenEOF
	TokenComment
)

var tokenNames = map[TokenType]string{
	TokenIdentifier:   "IDENTIFIER",
	TokenNumber:       "NUMBER",
	TokenString:       "STRING",
	TokenTrue:         "TRUE",
	TokenFalse:        "FALSE",
	TokenAssign:       "=",
	TokenPlus:         "+",
	TokenMinus:        "-",
	TokenStar:         "*",
	TokenSlash:        "/",
	TokenEqual:        "==",
	TokenNotEqual:     "!=",
	TokenLess:         "<",
	TokenLessEqual:    "<=",
	TokenGreater:      ">",
	TokenGreaterEqual: ">=",
	TokenAnd:          "AND",
	TokenOr:           "OR",
	TokenNot:          "NOT",
	TokenDot:          ".",
	TokenPipe:         "|",
	TokenArrow:        "->",
	TokenModulo:       "%",
	TokenLeftParen:    "(",
	TokenRightParen:   ")",
	TokenLeftBracket:  "[",
	TokenRightBracket: "]",
	TokenLeftBrace:    "{",
	TokenRightBrace:   "}",
	TokenComma:        ",",
	TokenColon:        ":",
	TokenDo:           "DO",
	TokenReturn:       "RETURN",
	TokenIf:           "IF",
	TokenElse:         "ELSE",
	TokenFor:          "FOR",
	TokenIn:           "IN",
	TokenServe:        "SERVE",
	TokenGet:          "GET",
	TokenPost:         "POST",
	TokenPut:          "PUT",
	TokenDelete:       "DELETE",
	TokenRespond:      "RESPOND",
	TokenBody:         "BODY",
	TokenStore:        "STORE",
	TokenSave:         "SAVE",
	TokenLoad:         "LOAD",
	TokenRemove:       "REMOVE",
	TokenPage:         "PAGE",
	TokenText:         "TEXT",
	TokenRow:          "ROW",
	TokenEach:         "EACH",
	TokenButton:       "BUTTON",
	TokenForm:         "FORM",
	TokenInput:        "INPUT",
	TokenSubmit:       "SUBMIT",
	TokenLink:         "LINK",
	TokenImage:        "IMAGE",
	TokenFetch:        "FETCH",
	TokenPrint:        "PRINT",
	TokenNewline:      "NEWLINE",
	TokenIndent:       "INDENT",
	TokenDedent:       "DEDENT",
	TokenEOF:          "EOF",
	TokenComment:      "COMMENT",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Column  int
}

var keywords = map[string]TokenType{
	"do":      TokenDo,
	"return":  TokenReturn,
	"if":      TokenIf,
	"else":    TokenElse,
	"for":     TokenFor,
	"in":      TokenIn,
	"serve":   TokenServe,
	"get":     TokenGet,
	"post":    TokenPost,
	"put":     TokenPut,
	"delete":  TokenDelete,
	"respond": TokenRespond,
	"body":    TokenBody,
	"store":   TokenStore,
	"save":    TokenSave,
	"load":    TokenLoad,
	"remove":  TokenRemove,
	"page":    TokenPage,
	"text":    TokenText,
	"row":     TokenRow,
	"each":    TokenEach,
	"button":  TokenButton,
	"form":    TokenForm,
	"input":   TokenInput,
	"submit":  TokenSubmit,
	"link":    TokenLink,
	"image":   TokenImage,
	"fetch":   TokenFetch,
	"print":   TokenPrint,
	"true":    TokenTrue,
	"false":   TokenFalse,
	"and":     TokenAnd,
	"or":      TokenOr,
	"not":     TokenNot,
}

func LookupKeyword(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokenIdentifier
}
