package coldmoon

import (
	"github.com/Seeingu/coldmoon/pkg"
	lo "github.com/samber/lo"
)

type TokenType int

const (
	// TLeftBrace {
	TLeftBrace TokenType = iota
	TRightBrace
	// TLeftBracket [
	TLeftBracket
	TRightBracket
	// TLeftParen (
	TLeftParen
	TRightParen
	TDot
	TDotDotDot
	TSemicolon
	TPipe
	TPipeEquals
	TPipePipe
	TPipePipeEquals
	TComma
	TLessThan
	TGreaterThan
	TLessThanEquals
	TGreaterThanEquals
	TEquals
	TNotEquals
	TStrictEquals
	TStrictNotEquals
	TPlus
	TMinus
	TPlusPlus
	TMinusMinus
	TWave
	TStar
	TPercent
	TIncrement
	TDecrement
	TLeftShift
	TRightShift
	TUnsignedRightShift
	TAmpersand
	TNot
	TQuestion
	TQuestionQuestion
	TQuestionQuestionEquals
	TQuestionDot
	TCaret
	TCaretEquals
	TColon
	TTilde
	TSlash
	TSlashSlash
	TEqualsEquals
	TArrow
	TPlusEquals
	TThis
	TMinusEquals
	TStarEquals
	TStarStar
	TStarStarEquals
	TAmpersandEquals
	TAmpersandAmpersand
	TAmpersandAmpersandEquals
	TPercentEquals
	TLeftShiftEquals
	TRightShiftEquals
	TUnsignedRightShiftEquals
	TBitwiseAndEquals
	TBitwiseOrEquals
	TBitwiseXorEquals
	TDivideEquals
	TIdentifier
	TNumber
	TString
	TRegularExpression
	TTrue
	TFalse
	TDebugger
	TNull
	TUndefined
	TIf
	TElse
	TWhile
	TBreak
	TDo
	TReturn
	TThrow
	TVoid
	TTypeof
	TYield
	TAwait
	TFunction
	TDelete
	TComment
	TTemplateHead
	TTemplateMiddle
	TTemplateTail
	TNoSubstitutionTemplate
	TIn
	TStatic
	TInstanceof
	TNew
	TVar
	TLet
	TConst
	TTry
	TCatch
	TFinally
	TClass
	TExtends
	TSuper
	TImport
	TExport
	TDefault
	TAsync
	TFrom
	TAs
	TFor
	TOf
	TWith
	TSwitch
	TCase
	TDefaultCase
	TContinue
	TEnum
	TEOF
)

type Token struct {
	Type  TokenType
	Value string
}

type cachedState struct {
	index        int
	line         int
	currentToken Token
	nextToken    Token
}
type Tokenizer struct {
	SourceText   []rune
	Index        int
	Length       int
	line         int
	CurrentToken Token
	NextToken    Token
	cachedStates pkg.Stack[*cachedState]
	isTemplate   bool
}

func NewTokenizer(sourceText string) *Tokenizer {
	source := []rune(sourceText)
	tokenizer := &Tokenizer{
		SourceText: source,
		Index:      0,
		line:       1,
		Length:     len(source),
	}
	tokenizer.Peek()
	tokenizer.Peek()
	return tokenizer
}

func (t *Tokenizer) store() {
	t.cachedStates.Push(&cachedState{
		index:        t.Index,
		line:         t.line,
		currentToken: t.CurrentToken,
		nextToken:    t.NextToken,
	})
}

func (t *Tokenizer) popCachedState() {
	t.cachedStates.Pop()
}
func (t *Tokenizer) restore() {
	if t.cachedStates.IsEmpty() {
		panic("no cached state")
	}
	cachedState := t.cachedStates.Pop()
	t.Index = cachedState.index
	t.line = cachedState.line
	t.CurrentToken = cachedState.currentToken
	t.NextToken = cachedState.nextToken
}

func (t *Tokenizer) peek() Token {
	t.skipWhiteSpace()
	if t.Index >= t.Length {
		return Token{Type: TEOF}
	}

	ch := t.SourceText[t.Index]
	switch ch {
	case '^':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '=' {
			t.Index++
			return Token{Type: TCaretEquals, Value: "^="}
		}
		return Token{Type: TCaret, Value: "^"}
	case ':':
		t.Index++
		return Token{Type: TColon, Value: ":"}
	case '{':
		t.Index++
		return Token{Type: TLeftBrace, Value: "{"}
	case '}':
		if t.isTemplate {
			return t.templateMiddleOrTail()
		}
		t.Index++
		return Token{Type: TRightBrace, Value: "}"}
	case '[':
		t.Index++
		return Token{Type: TLeftBracket, Value: "["}
	case ']':
		t.Index++
		return Token{Type: TRightBracket, Value: "]"}
	case '(':
		t.Index++
		return Token{Type: TLeftParen, Value: "("}
	case ')':
		t.Index++
		return Token{Type: TRightParen, Value: ")"}
	case '&':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '&' {
			t.Index++
			if t.Index < t.Length && t.SourceText[t.Index] == '=' {
				t.Index++
				return Token{Type: TAmpersandAmpersandEquals, Value: "&&="}
			}
			return Token{Type: TAmpersandAmpersand, Value: "&&"}
		}
		if t.Index < t.Length && t.SourceText[t.Index] == '=' {
			t.Index++
			return Token{Type: TAmpersandEquals, Value: "&="}
		}
		return Token{Type: TAmpersand, Value: "&"}
	case '%':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '=' {
			t.Index++
			return Token{Type: TPercentEquals, Value: "%="}
		}
		return Token{Type: TPercent, Value: "%"}
	case '/':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '=' {
			t.Index++
			return Token{Type: TDivideEquals, Value: "/="}
		}
		if t.Index < t.Length && t.SourceText[t.Index] == '/' {
			t.Index++
			comment := t.comment("//")
			return Token{Type: TComment, Value: comment}
		}
		if t.Index < t.Length && t.SourceText[t.Index] == '*' {
			t.Index++
			comment := t.comment("/*")
			return Token{Type: TComment, Value: comment}
		}
		if token, ok := t.tryToMatchRegularExpression(); ok {
			return token
		}
		return Token{Type: TSlash, Value: "/"}
	case '`':
		return t.templateHead()
	case '*':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '=' {
			t.Index++
			return Token{Type: TStarEquals, Value: "*="}
		}
		if t.Index < t.Length && t.SourceText[t.Index] == '*' {
			t.Index++
			if t.Index < t.Length && t.SourceText[t.Index] == '=' {
				t.Index++
				return Token{Type: TStarStarEquals, Value: "**="}
			}
			return Token{Type: TStarStar, Value: "**"}
		}
		return Token{Type: TStar, Value: "*"}
	case '.':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '.' {
			t.Index++
			if t.Index < t.Length && t.SourceText[t.Index] == '.' {
				t.Index++
				return Token{Type: TDotDotDot, Value: "..."}
			}
		}
		return Token{Type: TDot, Value: "."}
	case ';':
		t.Index++
		return Token{Type: TSemicolon, Value: ";"}
	case ',':
		t.Index++
		return Token{Type: TComma, Value: ","}
	case '<':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '=' {
			t.Index++
			return Token{Type: TLessThanEquals, Value: "<="}
		}
		if t.Index < t.Length && t.SourceText[t.Index] == '<' {
			t.Index++
			if t.Index < t.Length && t.SourceText[t.Index] == '=' {
				t.Index++
				return Token{Type: TLeftShiftEquals, Value: "<<="}
			}
			return Token{Type: TLeftShift, Value: "<<"}
		}
		return Token{Type: TLessThan, Value: "<"}
	case '+':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '+' {
			t.Index++
			return Token{Type: TPlusPlus, Value: "++"}
		}
		if t.Index < t.Length && t.SourceText[t.Index] == '=' {
			t.Index++
			return Token{Type: TPlusEquals, Value: "+="}
		}
		return Token{Type: TPlus, Value: "+"}
	case '-':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '-' {
			t.Index++
			return Token{Type: TMinusMinus, Value: "--"}
		}
		if t.Index < t.Length && t.SourceText[t.Index] == '=' {
			t.Index++
			return Token{Type: TMinusEquals, Value: "-="}
		}
		return Token{Type: TMinus, Value: "-"}
	case '>':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '=' {
			t.Index++
			return Token{Type: TGreaterThanEquals, Value: ">="}
		}
		if t.Index < t.Length && t.SourceText[t.Index] == '>' {
			t.Index++
			if t.Index < t.Length && t.SourceText[t.Index] == '=' {
				t.Index++
				return Token{Type: TRightShiftEquals, Value: ">>="}
			}
			if t.Index < t.Length && t.SourceText[t.Index] == '>' {
				t.Index++
				if t.Index < t.Length && t.SourceText[t.Index] == '=' {
					t.Index++
					return Token{Type: TUnsignedRightShiftEquals, Value: ">>>="}
				}
				return Token{Type: TUnsignedRightShift, Value: ">>>"}
			}
			return Token{Type: TRightShift, Value: ">>"}
		}
		return Token{Type: TGreaterThan, Value: ">"}
	case '=':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '=' {
			t.Index++
			if t.Index < t.Length && t.SourceText[t.Index] == '=' {
				t.Index++
				return Token{Type: TStrictEquals, Value: "==="}
			}
			return Token{Type: TEqualsEquals, Value: "=="}
		}
		if t.Index < t.Length && t.SourceText[t.Index] == '>' {
			t.Index++
			return Token{Type: TArrow, Value: "=>"}
		}
		return Token{Type: TEquals, Value: "="}
	case '|':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '=' {
			t.Index++
			return Token{Type: TPipeEquals, Value: "|="}
		}
		if t.Index < t.Length && t.SourceText[t.Index] == '|' {
			t.Index++
			if t.Index < t.Length && t.SourceText[t.Index] == '=' {
				t.Index++
				return Token{Type: TPipePipeEquals, Value: "||="}
			}
			return Token{Type: TPipePipe, Value: "||"}
		}
		return Token{Type: TPipe, Value: "|"}
	case '!':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '=' {
			t.Index++
			if t.Index < t.Length && t.SourceText[t.Index] == '=' {
				t.Index++
				return Token{Type: TStrictNotEquals, Value: "!=="}
			} else {
				return Token{Type: TNotEquals, Value: "!="}
			}
		}
		return Token{Type: TNot, Value: "!"}
	case '?':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '.' {
			t.Index++
			return Token{Type: TQuestionDot, Value: "?."}
		}
		if t.Index < t.Length && t.SourceText[t.Index] == '?' {
			t.Index++
			if t.Index < t.Length && t.SourceText[t.Index] == '=' {
				t.Index++
				return Token{Type: TQuestionQuestionEquals, Value: "??="}
			}
			return Token{Type: TQuestionQuestion, Value: "??"}
		}
		return Token{Type: TQuestion, Value: "?"}
	case '~':
		t.Index++
		return Token{Type: TTilde, Value: "~"}
	case '\'', '"':
		return t.string()
	default:
		if lo.Contains(lo.NumbersCharset, ch) {
			return t.number()
		}
		if lo.Contains(identifierStartCharset, ch) {
			return t.identifierOrKeyword()
		}
	}
	panic("unhandled token: " + string(ch))

}

func (t *Tokenizer) templateMiddleOrTail() Token {
	start := t.Index
	t.Index++
	for t.Index < t.Length {
		ch := t.SourceText[t.Index]
		if ch == '`' {
			t.Index++
			t.isTemplate = false
			return Token{Type: TTemplateTail, Value: string(t.SourceText[start : t.Index-1])}
		}
		if ch == '$' {
			t.Index++
			if t.Index < t.Length && t.SourceText[t.Index] == '{' {
				t.Index++
				t.isTemplate = true
				return Token{Type: TTemplateMiddle, Value: string(t.SourceText[start : t.Index-2])}
			}
		}
		t.Index++
	}
	panic("unterminated template")
}

func (t *Tokenizer) templateHead() Token {
	start := t.Index
	for t.Index < t.Length {
		ch := t.SourceText[t.Index]
		if ch == '`' {
			t.Index++
			return Token{Type: TNoSubstitutionTemplate, Value: string(t.SourceText[start:t.Index])}
		}
		if ch == '$' {
			t.Index++
			if t.Index < t.Length && t.SourceText[t.Index] == '{' {
				t.Index++
				t.isTemplate = true
				return Token{Type: TTemplateHead, Value: string(t.SourceText[start : t.Index-2])}
			}
		}
		t.Index++
	}
	panic("unterminated template")
}

func (t *Tokenizer) comment(commentType string) string {
	startIndex := t.Index
	if commentType == "//" {
		for t.Index < t.Length {
			ch := t.SourceText[t.Index]
			if lo.Contains(lineTerminators, ch) {
				t.Index++
				t.line++
				return string(t.SourceText[startIndex:t.Index])
			}
			t.Index++
		}
	}
	if commentType == "/*" {
		for t.Index < t.Length {
			ch := t.SourceText[t.Index]
			if ch == '\n' {
				t.line++
			}
			if ch == '*' {
				t.Index++
				if t.Index < t.Length && t.SourceText[t.Index] == '/' {
					t.Index++
					return string(t.SourceText[startIndex : t.Index-2])
				}
			}
			t.Index++
		}
	}
	return ""
}

// MARK: - String
func (t *Tokenizer) string() Token {
	// TODO: handle escape
	start := t.Index + 1
	quote := t.SourceText[t.Index]
	t.Index++
	for t.Index < t.Length {
		ch := t.SourceText[t.Index]
		if ch == quote {
			t.Index++
			break
		}
		t.Index++
	}
	value := string(t.SourceText[start : t.Index-1])
	return Token{Type: TString, Value: value}
}

// MARK: - Number
func (t *Tokenizer) number() Token {
	start := t.Index
	for t.Index < t.Length {
		ch := t.SourceText[t.Index]
		if lo.Contains(lo.NumbersCharset, ch) {
			t.Index++
		} else {
			break
		}
	}
	if t.Index < t.Length && t.SourceText[t.Index] == '.' {
		t.Index++
		for t.Index < t.Length {
			ch := t.SourceText[t.Index]
			if lo.Contains(lo.NumbersCharset, ch) {
				t.Index++
			} else {
				break
			}
		}
	}
	value := string(t.SourceText[start:t.Index])
	return Token{Type: TNumber, Value: value}
}

// MARK: - Identifier, Keyword
var identifierStartCharset = append(lo.LettersCharset, []rune{'$', '_'}...)
var identifierCharset = append(identifierStartCharset, lo.NumbersCharset...)

func (t *Tokenizer) identifierOrKeyword() Token {
	start := t.Index
	for t.Index < t.Length {
		ch := t.SourceText[t.Index]
		if lo.Contains(identifierCharset, ch) {
			t.Index++
		} else {
			break
		}
	}
	value := string(t.SourceText[start:t.Index])
	if lo.Contains(lo.Keys(keywordsMap), value) {
		return Token{Type: keywordsMap[value], Value: value}
	}
	return Token{Type: TIdentifier, Value: value}
}

var keywordsMap = map[string]TokenType{
	"if":         TIf,
	"else":       TElse,
	"true":       TTrue,
	"false":      TFalse,
	"null":       TNull,
	"undefined":  TUndefined,
	"debugger":   TDebugger,
	"this":       TThis,
	"break":      TBreak,
	"while":      TWhile,
	"do":         TDo,
	"throw":      TThrow,
	"return":     TReturn,
	"void":       TVoid,
	"typeof":     TTypeof,
	"yield":      TYield,
	"await":      TAwait,
	"function":   TFunction,
	"in":         TIn,
	"instanceof": TInstanceof,
	"new":        TNew,
	"var":        TVar,
	"let":        TLet,
	"const":      TConst,
	"try":        TTry,
	"catch":      TCatch,
	"finally":    TFinally,
	"class":      TClass,
	"extends":    TExtends,
	"super":      TSuper,
	"import":     TImport,
	"export":     TExport,
	"default":    TDefault,
	"from":       TFrom,
	"as":         TAs,
	"for":        TFor,
	"of":         TOf,
	"with":       TWith,
	"switch":     TSwitch,
	"case":       TCase,
	"enum":       TEnum,
	"async":      TAsync,
	"static":     TStatic,
}

func (t *Tokenizer) keyword() (token Token, ok bool) {
	for keyword, tokenType := range keywordsMap {
		if t.matchString(keyword) {
			ok = true
			token = Token{Type: tokenType, Value: keyword}
			return
		}
	}
	return

}
func (t *Tokenizer) matchString(s string) bool {
	endIndex := t.Index + len(s)
	if endIndex > t.Length {
		return false
	}
	if string(t.SourceText[t.Index:endIndex]) == s {
		t.Index += len(s)
		return true
	}
	return false
}

func (t *Tokenizer) tryToMatchRegularExpression() (token Token, ok bool) {
	// TODO: auto insert semicolon
	if t.NextToken.Type == TNumber || t.NextToken.Type == TString {
		return
	}
	isRegExp := false
	index := t.Index
	for index < t.Length {
		ch := t.SourceText[index]
		if ch == '\n' {
			break
		}
		if ch == '/' {
			isRegExp = true
			break
		}
		index++
	}

	if isRegExp {
		return t.regularExpression(), true
	}
	return
}

func (t *Tokenizer) regularExpression() Token {
	start := t.Index
	for t.Index < t.Length {
		ch := t.SourceText[t.Index]
		if ch == '/' {
			t.Index++
			break
		}
		if ch == '\\' {
			t.Index++
		}
		t.Index++
	}
	value := string(t.SourceText[start : t.Index-1])
	return Token{Type: TRegularExpression, Value: value}
}

func (t *Tokenizer) Peek() Token {
	token := t.peek()
	t.CurrentToken = t.NextToken
	t.NextToken = token
	return t.CurrentToken
}

func (t *Tokenizer) Next() Token {
	t.skipWhiteSpace()
	token := t.Peek()
	return token
}

func (t *Tokenizer) Match(tokenType TokenType) bool {
	if t.CurrentToken.Type == tokenType {
		t.Next()
		return true
	}
	return false
}

// MustMatch consumes the current token if it matches the given type.
// Will throw an error if the current token does not match the given type.
func (t *Tokenizer) MustMatch(tokenType TokenType) {
	if t.Match(tokenType) {
		return
	}
	panic("unexpected token: " + t.CurrentToken.Value)
}

// 12.3
var lineTerminators = []rune{'\n', '\r', '\u2028', '\u2029'}
var whitespace = []rune{' ', '\t', '\v', '\f', '\u00A0', '\uFEFF', '\u1680', '\u180E', '\u2000', '\u2001', '\u2002', '\u2003', '\u2004', '\u2005', '\u2006', '\u2007', '\u2008', '\u2009', '\u200A', '\u202F', '\u205F', '\u3000', '\u200B', '\u200C', '\u200D', '\u2060', '\uFEFF'}

func (t *Tokenizer) skipWhiteSpace() {
	for t.Index < t.Length {
		ch := t.SourceText[t.Index]
		if lo.Contains(whitespace, ch) || lo.Contains(lineTerminators, rune(ch)) {
			if ch == '\n' {
				t.line++
			}
			t.Index++
		} else {
			break
		}
	}
}
