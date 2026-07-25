package coldmoon

import (
	"strconv"
	"strings"

	"github.com/Seeingu/coldmoon/pkg"
	lo "github.com/samber/lo"
)

var (
	identifierStartCharset = append(lo.LettersCharset, []rune{'$', '_'}...)
	identifierCharset      = append(identifierStartCharset, lo.NumbersCharset...)
	hexDigitCharset        = append(lo.NumbersCharset, []rune("abcdefABCDEF")...)
)

type Token struct {
	Type       TokenType
	Value      string
	Line       int
	StartIndex int
	EndIndex   int
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

func (t *Tokenizer) CurrentStartIndex() int {
	if t.CurrentToken.Type == TEOF {
		return t.Index
	}
	return t.CurrentToken.StartIndex
}

func (t *Tokenizer) CurrentEndIndex() int {
	return t.CurrentToken.EndIndex
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

func (t *Tokenizer) newToken(tokenType TokenType, value string) Token {
	return Token{
		Type:       tokenType,
		Line:       t.line,
		StartIndex: t.Index - len(value),
		EndIndex:   t.Index,
		Value:      value,
	}
}

func (t *Tokenizer) peek() Token {
	t.skipWhiteSpace()
	if t.Index >= t.Length {
		return Token{Type: TEOF}
	}

	ch := t.SourceText[t.Index]
	switch ch {
	case '^':
		t.step()
		if t.match('=') {
			return t.newToken(TCaretEquals, "^=")
		}
		return t.newToken(TCaret, "^")
	case ':':
		t.step()
		return t.newToken(TColon, ":")
	case '{':
		t.step()
		return t.newToken(TLeftBrace, "{")
	case '}':
		if t.isTemplate {
			return t.templateMiddleOrTail()
		}
		t.step()
		return t.newToken(TRightBrace, "}")
	case '[':
		t.step()
		return t.newToken(TLeftBracket, "[")
	case ']':
		t.step()
		return t.newToken(TRightBracket, "]")
	case '(':
		t.step()
		return t.newToken(TLeftParen, "(")
	case ')':
		t.step()
		return t.newToken(TRightParen, ")")
	case '&':
		t.step()
		if t.match('&') {
			if t.match('=') {
				return t.newToken(TAmpersandAmpersandEquals, "&&=")
			}
			return t.newToken(TAmpersandAmpersand, "&&")
		}
		if t.match('=') {
			return t.newToken(TAmpersandEquals, "&=")
		}
		return t.newToken(TAmpersand, "&")
	case '%':
		t.step()
		if t.match('=') {
			return t.newToken(TPercentEquals, "%=")
		}
		return t.newToken(TPercent, "%")
	case '/':
		t.step()
		if t.match('=') {
			return t.newToken(TDivideEquals, "/=")
		}
		if t.match('/') {
			t.comment("//")
			return t.peek()
		}
		if t.match('*') {
			t.comment("/*")
			return t.peek()
		}
		if token, ok := t.tryToMatchRegularExpression(); ok {
			return token
		}
		return t.newToken(TSlash, "/")
	case '`':
		return t.templateHead()
	case '*':
		t.step()
		if t.match('=') {
			return t.newToken(TStarEquals, "*=")
		}
		if t.match('*') {
			if t.match('=') {
				return t.newToken(TStarStarEquals, "**=")
			}
			return t.newToken(TStarStar, "**")
		}
		return t.newToken(TStar, "*")
	case '.':
		t.step()
		if t.match('.') {
			if t.match('.') {
				return t.newToken(TDotDotDot, "...")
			}
		}
		return t.newToken(TDot, ".")
	case ';':
		t.step()
		return t.newToken(TSemicolon, ";")
	case ',':
		t.step()
		return t.newToken(TComma, ",")
	case '<':
		t.step()
		if t.match('=') {
			return t.newToken(TLessThanEquals, "<=")
		}
		if t.match('<') {
			if t.match('=') {
				return t.newToken(TLeftShiftEquals, "<<=")
			}
			return t.newToken(TLeftShift, "<<")
		}
		return t.newToken(TLessThan, "<")
	case '+':
		t.step()
		if t.match('+') {
			return t.newToken(TPlusPlus, "++")
		}
		if t.match('=') {
			return t.newToken(TPlusEquals, "+=")
		}
		return t.newToken(TPlus, "+")
	case '-':
		t.step()
		if t.match('-') {
			return t.newToken(TMinusMinus, "--")
		}
		if t.match('=') {
			return t.newToken(TMinusEquals, "-=")
		}
		return t.newToken(TMinus, "-")
	case '>':
		t.step()
		if t.match('=') {
			return t.newToken(TGreaterThanEquals, ">=")
		}
		if t.match('>') {
			if t.match('=') {
				return t.newToken(TRightShiftEquals, ">>=")
			}
			if t.match('>') {
				if t.match('=') {
					return t.newToken(TUnsignedRightShiftEquals, ">>>=")
				}
				return t.newToken(TUnsignedRightShift, ">>>")
			}
			return t.newToken(TRightShift, ">>")
		}
		return t.newToken(TGreaterThan, ">")
	case '=':
		t.step()
		if t.match('=') {
			if t.match('=') {
				return t.newToken(TStrictEquals, "===")
			}
			return t.newToken(TEqualsEquals, "==")
		}
		if t.match('>') {
			return t.newToken(TArrow, "=>")
		}
		return t.newToken(TEquals, "=")
	case '|':
		t.step()
		if t.match('=') {
			return t.newToken(TPipeEquals, "|=")
		}
		if t.match('|') {
			if t.match('=') {
				return t.newToken(TPipePipeEquals, "||=")
			}
			return t.newToken(TPipePipe, "||")
		}
		return t.newToken(TPipe, "|")
	case '!':
		t.step()
		if t.match('=') {
			if t.match('=') {
				return t.newToken(TStrictNotEquals, "!==")
			}
			return t.newToken(TNotEquals, "!=")
		}
		return t.newToken(TNot, "!")
	case '?':
		t.step()
		if t.match('.') {
			return t.newToken(TQuestionDot, "?.")
		}
		if t.match('?') {
			if t.match('=') {
				return t.newToken(TQuestionQuestionEquals, "??=")
			}
			return t.newToken(TQuestionQuestion, "??")
		}
		return t.newToken(TQuestion, "?")
	case '~':
		t.step()
		return t.newToken(TTilde, "~")
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
	t.step()
	for !t.atEnd() {
		if t.match('`') {
			t.isTemplate = false
			return t.newToken(TTemplateTail, string(t.SourceText[start:t.Index-1]))
		}
		if t.match('$') {
			if t.match('{') {
				t.isTemplate = true
				return t.newToken(TTemplateMiddle, string(t.SourceText[start:t.Index-2]))
			}
		}
		t.step()
	}
	panic("unterminated template")
}

func (t *Tokenizer) templateHead() Token {
	start := t.Index
	t.step()
	for !t.atEnd() {
		if t.match('`') {
			return t.newToken(TNoSubstitutionTemplate, string(t.SourceText[start:t.Index]))
		}
		if t.match('$') {
			if t.match('{') {
				t.isTemplate = true
				return t.newToken(TTemplateHead, string(t.SourceText[start:t.Index-2]))
			}
		}
		t.step()
	}
	panic("unterminated template")
}

func (t *Tokenizer) comment(commentType string) string {
	startIndex := t.Index
	if commentType == "//" {
		for !t.atEnd() {
			ch := t.SourceText[t.Index]
			if lo.Contains(lineTerminators, ch) {
				t.step()
				return string(t.SourceText[startIndex:t.Index])
			}
			t.step()
		}
	}
	if commentType == "/*" {
		for !t.atEnd() {
			if t.match('*') {
				if t.match('/') {
					t.step()
					return string(t.SourceText[startIndex : t.Index-2])
				}
			}
			t.step()
		}
	}
	return ""
}

// MARK: - String
func (t *Tokenizer) string() Token {
	quote := t.SourceText[t.Index]
	t.step()
	var value strings.Builder
	for !t.atEnd() {
		ch := t.SourceText[t.Index]
		if ch == quote {
			t.step()
			return t.newToken(TString, value.String())
		}
		if ch != '\\' {
			value.WriteRune(ch)
			t.step()
			continue
		}

		t.step()
		if t.atEnd() {
			panic("unterminated string escape")
		}
		escaped := t.SourceText[t.Index]
		t.step()
		switch escaped {
		case 'b':
			value.WriteRune('\b')
		case 'f':
			value.WriteRune('\f')
		case 'n':
			value.WriteRune('\n')
		case 'r':
			value.WriteRune('\r')
		case 't':
			value.WriteRune('\t')
		case 'v':
			value.WriteRune('\v')
		case '\n':
			// A line continuation contributes no character.
		default:
			value.WriteRune(escaped)
		}
	}
	panic("unterminated string")
}

// MARK: - Number
// TODO: parse later, use StringToBigInt
func (t *Tokenizer) number() Token {
	var value string

	switch {
	case t.matchPrefix("0x", "0X"):
		value = t.parseNumber(16, hexDigitCharset)
	case t.matchPrefix("0b", "0B"):
		value = t.parseNumber(2, []rune{'0', '1'})
	case t.matchPrefix("0o", "0O"):
		value = t.parseNumber(8, []rune{'0', '1', '2', '3', '4', '5', '6', '7'})
	default:
		value = t.parseDecimalNumber()
	}

	if t.matchCharset([]rune{'n', 'N'}) {
		return t.newToken(TBigInt, value)
	}

	return t.newToken(TNumber, value)
}

func (t *Tokenizer) parseNumber(base int, validDigits []rune) string {
	start := t.Index
	for t.matchCharset(validDigits) {
		t.Index++
	}
	value, err := strconv.ParseInt(string(t.SourceText[start:t.Index]), base, 64)
	if err != nil {
		panic(err)
	}
	return strconv.FormatInt(value, 10)
}

func (t *Tokenizer) parseDecimalNumber() string {
	start := t.Index
	// TODO(XXX): is there a better way to handle empty loop/condition
	for t.matchCharset(lo.NumbersCharset) {
	}
	if t.match('.') {
		for t.matchCharset(lo.NumbersCharset) {
		}
	}
	if t.matchCharset([]rune{'e', 'E'}) {
		if t.matchCharset([]rune{'-', '+'}) {
		}
		for t.matchCharset(lo.NumbersCharset) {
		}
	}
	return string(t.SourceText[start:t.Index])
}

// MARK: - Identifier, Keyword

func (t *Tokenizer) identifierOrKeyword() Token {
	start := t.Index
	for !t.atEnd() {
		ch := t.SourceText[t.Index]
		if lo.Contains(identifierCharset, ch) {
			t.step()
		} else {
			break
		}
	}
	value := string(t.SourceText[start:t.Index])
	if lo.Contains(lo.Keys(keywordsMap), value) {
		return t.newToken(keywordsMap[value], value)
	}
	return t.newToken(TIdentifier, value)
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
	"delete":     TDelete,
}

func (t *Tokenizer) keyword() (token Token, ok bool) {
	for keyword, tokenType := range keywordsMap {
		if t.matchString(keyword) {
			ok = true
			token = t.newToken(tokenType, keyword)
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
	for !t.atEnd() {
		if t.match('/') {
			break
		}
		if t.match('\\') {
		}
		t.step()
	}
	value := string(t.SourceText[start : t.Index-1])
	return t.newToken(TRegularExpression, value)
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
var (
	lineTerminators = []rune{'\n', '\r', '\u2028', '\u2029'}
	whitespace      = []rune{' ', '\t', '\v', '\f', '\u00A0', '\uFEFF', '\u1680', '\u180E', '\u2000', '\u2001', '\u2002', '\u2003', '\u2004', '\u2005', '\u2006', '\u2007', '\u2008', '\u2009', '\u200A', '\u202F', '\u205F', '\u3000', '\u200B', '\u200C', '\u200D', '\u2060', '\uFEFF'}
)

func (t *Tokenizer) skipWhiteSpace() {
	for t.Index < t.Length {
		ch := t.SourceText[t.Index]
		if lo.Contains(whitespace, ch) || lo.Contains(lineTerminators, ch) {
			t.step()
		} else {
			break
		}
	}
}

// step increments the index and line number.
// - step will not check if the index is at the end of the source text.
func (t *Tokenizer) step() {
	ch := t.SourceText[t.Index]
	if ch == '\n' {
		t.nextLine()
	}
	t.Index++
}

func (t *Tokenizer) atEnd() bool {
	return t.Index >= t.Length
}

func (t *Tokenizer) matchPrefix(prefixes ...string) bool {
	for _, prefix := range prefixes {
		if t.Index+len(prefix) < t.Length && string(t.SourceText[t.Index:t.Index+len(prefix)]) == prefix {
			t.Index += len(prefix)
			return true
		}
	}
	return false
}

// match consumes the current character if it matches the given character.
func (t *Tokenizer) match(ch rune) bool {
	if t.atEnd() {
		return false
	}
	if t.SourceText[t.Index] == ch {
		t.step()
		return true
	}
	return false
}

func (t *Tokenizer) matchCharset(chars []rune) bool {
	if t.atEnd() {
		return false
	}
	if lo.Contains(chars, t.SourceText[t.Index]) {
		t.step()
		return true
	}
	return false
}

func (t *Tokenizer) nextLine() {
	t.line++
}
