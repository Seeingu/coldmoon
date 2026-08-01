package coldmoon

import (
	"math/big"
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
	index         int
	line          int
	previousToken Token
	currentToken  Token
	nextToken     Token
}
type Tokenizer struct {
	SourceText    []rune
	Index         int
	Length        int
	line          int
	PreviousToken Token
	CurrentToken  Token
	NextToken     Token
	cachedStates  pkg.Stack[*cachedState]
	isTemplate    bool
	sourceName    string
}

func (t *Tokenizer) parseFailure(message string) *parseFailure {
	start := t.Index
	end := start
	incomplete := t.atEnd()
	if !incomplete {
		end++
	}
	return &parseFailure{
		message:    message,
		start:      start,
		end:        end,
		incomplete: incomplete,
		sourceName: t.sourceName,
		sourceText: string(t.SourceText),
	}
}

func (t *Tokenizer) currentParseFailure(message string) *parseFailure {
	token := t.CurrentToken
	return &parseFailure{
		message:    message,
		start:      token.StartIndex,
		end:        token.EndIndex,
		incomplete: token.Type == TEOF,
		sourceName: t.sourceName,
		sourceText: string(t.SourceText),
	}
}

func NewTokenizer(sourceText string) *Tokenizer {
	return newTokenizer(sourceText, "")
}

func newTokenizer(sourceText string, sourceName string) *Tokenizer {
	source := []rune(sourceText)
	tokenizer := &Tokenizer{
		SourceText: source,
		Index:      0,
		line:       1,
		Length:     len(source),
		sourceName: sourceName,
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
		index:         t.Index,
		line:          t.line,
		previousToken: t.PreviousToken,
		currentToken:  t.CurrentToken,
		nextToken:     t.NextToken,
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
	t.PreviousToken = cachedState.previousToken
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

func (t *Tokenizer) newTokenAt(startIndex int, tokenType TokenType, value string) Token {
	return Token{
		Type:       tokenType,
		Line:       t.line,
		StartIndex: startIndex,
		EndIndex:   t.Index,
		Value:      value,
	}
}

func (t *Tokenizer) peek() Token {
	t.skipWhiteSpace()
	if t.Index >= t.Length {
		return Token{
			Type:       TEOF,
			Line:       t.line,
			StartIndex: t.Index,
			EndIndex:   t.Index,
		}
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
		if t.Index+1 < t.Length && lo.Contains(lo.NumbersCharset, t.SourceText[t.Index+1]) {
			return t.number()
		}
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
	case '#':
		return t.privateIdentifier()
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
	panic(t.parseFailure("unhandled token: " + string(ch)))
}

func (t *Tokenizer) privateIdentifier() Token {
	start := t.Index
	t.step()
	if t.atEnd() || !lo.Contains(identifierStartCharset, t.SourceText[t.Index]) {
		panic(t.parseFailure("private identifier requires an identifier name"))
	}
	for !t.atEnd() && lo.Contains(identifierCharset, t.SourceText[t.Index]) {
		t.step()
	}
	return t.newTokenAt(start, TPrivateIdentifier, string(t.SourceText[start:t.Index]))
}

func (t *Tokenizer) templateMiddleOrTail() Token {
	start := t.Index
	t.step()
	for !t.atEnd() {
		if t.SourceText[t.Index] == '\\' {
			t.skipTemplateEscape()
			continue
		}
		if t.match('`') {
			t.isTemplate = false
			return t.newTokenAt(start, TTemplateTail, string(t.SourceText[start:t.Index-1]))
		}
		if t.match('$') {
			if t.match('{') {
				t.isTemplate = true
				return t.newTokenAt(start, TTemplateMiddle, string(t.SourceText[start:t.Index-2]))
			}
			continue
		}
		t.step()
	}
	panic(t.parseFailure("unterminated template"))
}

func (t *Tokenizer) templateHead() Token {
	start := t.Index
	t.step()
	for !t.atEnd() {
		if t.SourceText[t.Index] == '\\' {
			t.skipTemplateEscape()
			continue
		}
		if t.match('`') {
			return t.newTokenAt(start, TNoSubstitutionTemplate, string(t.SourceText[start:t.Index]))
		}
		if t.match('$') {
			if t.match('{') {
				t.isTemplate = true
				return t.newTokenAt(start, TTemplateHead, string(t.SourceText[start:t.Index-2]))
			}
			continue
		}
		t.step()
	}
	panic(t.parseFailure("unterminated template"))
}

// skipTemplateEscape keeps escaped backticks and dollar signs inside the
// current template token instead of treating them as lexical delimiters. The
// escape is validated later because tagged templates permit malformed escape
// sequences and expose an undefined cooked value.
func (t *Tokenizer) skipTemplateEscape() {
	t.step()
	if t.atEnd() {
		panic(t.parseFailure("unterminated template escape"))
	}
	if t.SourceText[t.Index] == '\r' {
		t.step()
		if !t.atEnd() && t.SourceText[t.Index] == '\n' {
			t.step()
		}
		return
	}
	t.step()
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
			if t.SourceText[t.Index] == '*' && t.Index+1 < t.Length && t.SourceText[t.Index+1] == '/' {
				endIndex := t.Index
				t.step()
				t.step()
				return string(t.SourceText[startIndex:endIndex])
			}
			t.step()
		}
		panic(t.parseFailure("unterminated block comment"))
	}
	return ""
}

// MARK: - String
func (t *Tokenizer) string() Token {
	start := t.Index
	quote := t.SourceText[t.Index]
	t.step()
	var value strings.Builder
	for !t.atEnd() {
		ch := t.SourceText[t.Index]
		if ch == quote {
			t.step()
			return t.newTokenAt(start, TString, value.String())
		}
		if lo.Contains(lineTerminators, ch) {
			panic(t.parseFailure("unterminated string"))
		}
		if ch != '\\' {
			value.WriteRune(ch)
			t.step()
			continue
		}

		t.step()
		if t.atEnd() {
			panic(t.parseFailure("unterminated string escape"))
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
		case '\n', '\u2028', '\u2029':
			// A line continuation contributes no character.
		case '\r':
			// CRLF is a single LineTerminatorSequence.
			t.match('\n')
		default:
			value.WriteRune(escaped)
		}
	}
	panic(t.parseFailure("unterminated string"))
}

// MARK: - Number
func (t *Tokenizer) number() Token {
	start := t.Index
	base := 10
	isDecimalInteger := false
	var value string

	switch {
	case t.matchPrefix("0x", "0X"):
		base = 16
		value = t.parseDigits(hexDigitCharset, true)
	case t.matchPrefix("0b", "0B"):
		base = 2
		value = t.parseDigits([]rune{'0', '1'}, true)
	case t.matchPrefix("0o", "0O"):
		base = 8
		value = t.parseDigits([]rune{'0', '1', '2', '3', '4', '5', '6', '7'}, true)
	default:
		value, isDecimalInteger = t.parseDecimalNumber()
	}

	isBigInt := false
	if base != 10 || isDecimalInteger {
		isBigInt = t.match('n')
	}
	t.ensureNumericLiteralBoundary()

	if base != 10 {
		integer, ok := new(big.Int).SetString(value, base)
		if !ok {
			panic(t.parseFailure("invalid numeric literal"))
		}
		value = integer.Text(10)
	}

	if isBigInt {
		if base == 10 && len(value) > 1 && value[0] == '0' {
			panic(t.parseFailure("invalid decimal BigInt literal with a leading zero"))
		}
		return t.newTokenAt(start, TBigInt, value)
	}

	return t.newTokenAt(start, TNumber, value)
}

func (t *Tokenizer) parseDigits(validDigits []rune, required bool) string {
	var value strings.Builder
	digitCount := 0
	for !t.atEnd() {
		ch := t.SourceText[t.Index]
		if lo.Contains(validDigits, ch) {
			value.WriteRune(ch)
			digitCount++
			t.step()
			continue
		}
		if ch != '_' {
			break
		}
		if digitCount == 0 || t.Index+1 >= t.Length || !lo.Contains(validDigits, t.SourceText[t.Index+1]) {
			panic(t.parseFailure("invalid numeric separator"))
		}
		t.step()
	}
	if required && digitCount == 0 {
		panic(t.parseFailure("numeric literal requires at least one digit"))
	}
	return value.String()
}

func (t *Tokenizer) parseDecimalNumber() (value string, isInteger bool) {
	start := t.Index
	integerDigits := ""
	if t.SourceText[t.Index] == '.' {
		t.step()
		t.parseDigits(lo.NumbersCharset, true)
	} else {
		integerDigits = t.parseDigits(lo.NumbersCharset, true)
		isInteger = true
		if t.match('.') {
			isInteger = false
			t.parseDigits(lo.NumbersCharset, false)
		}
	}
	if t.matchCharset([]rune{'e', 'E'}) {
		isInteger = false
		t.matchCharset([]rune{'-', '+'})
		t.parseDigits(lo.NumbersCharset, true)
	}
	value = strings.ReplaceAll(string(t.SourceText[start:t.Index]), "_", "")
	if isInteger {
		value = integerDigits
	}
	return value, isInteger
}

func (t *Tokenizer) ensureNumericLiteralBoundary() {
	if t.atEnd() {
		return
	}
	if lo.Contains(identifierCharset, t.SourceText[t.Index]) {
		panic(t.parseFailure("identifier cannot immediately follow a numeric literal"))
	}
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
	"continue":   TContinue,
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
	if tokenCanEndExpression(t.NextToken.Type) {
		return
	}
	if !t.hasRegularExpressionTerminator() {
		return
	}
	return t.regularExpression(), true
}

func (t *Tokenizer) hasRegularExpressionTerminator() bool {
	index := t.Index
	inCharacterClass := false
	for index < t.Length {
		ch := t.SourceText[index]
		if lo.Contains(lineTerminators, ch) {
			return false
		}
		if ch == '\\' {
			index++
			if index >= t.Length || lo.Contains(lineTerminators, t.SourceText[index]) {
				return false
			}
			index++
			continue
		}
		if ch == '[' {
			inCharacterClass = true
		} else if ch == ']' {
			inCharacterClass = false
		} else if ch == '/' && !inCharacterClass {
			return true
		}
		index++
	}
	return false
}

func tokenCanEndExpression(tokenType TokenType) bool {
	switch tokenType {
	case TIdentifier,
		TNumber,
		TBigInt,
		TString,
		TRegularExpression,
		TTrue,
		TFalse,
		TNull,
		TUndefined,
		TThis,
		TAsync,
		TRightParen,
		TRightBracket,
		TRightBrace,
		TPlusPlus,
		TMinusMinus,
		TTemplateTail,
		TNoSubstitutionTemplate:
		return true
	default:
		return false
	}
}

// ReinterpretCurrentSlashAsRegularExpression applies the parser's
// InputElementRegExp lexical goal to an otherwise ambiguous slash. A right
// brace can end an expression (object, class, or function expression), so the
// tokenizer initially treats a following slash as division. When the parser
// has instead completed a statement block and starts a new statement, it uses
// this method to rescan that slash as a regular expression literal.
func (t *Tokenizer) ReinterpretCurrentSlashAsRegularExpression() {
	if t.CurrentToken.Type != TSlash {
		panic("regular expression rescan requires a slash token")
	}
	t.Index = t.CurrentToken.StartIndex
	t.line = t.CurrentToken.Line
	t.step()
	regularExpression := t.regularExpression()
	t.CurrentToken = regularExpression
	// peek consults NextToken as the token immediately preceding the text it
	// scans, so publish the rescanned literal before rebuilding lookahead.
	t.NextToken = regularExpression
	t.NextToken = t.peek()
}

func (t *Tokenizer) regularExpression() Token {
	tokenStart := t.Index - 1
	patternStart := t.Index
	inCharacterClass := false
	terminated := false
	for !t.atEnd() {
		ch := t.SourceText[t.Index]
		if lo.Contains(lineTerminators, ch) {
			panic(t.parseFailure("unterminated regular expression literal"))
		}
		if ch == '\\' {
			t.step()
			if t.atEnd() || lo.Contains(lineTerminators, t.SourceText[t.Index]) {
				panic(t.parseFailure("unterminated regular expression escape"))
			}
			t.step()
			continue
		}
		if ch == '[' {
			inCharacterClass = true
		} else if ch == ']' {
			inCharacterClass = false
		} else if ch == '/' && !inCharacterClass {
			t.step()
			terminated = true
			break
		}
		t.step()
	}
	if !terminated {
		panic(t.parseFailure("unterminated regular expression literal"))
	}
	value := string(t.SourceText[patternStart : t.Index-1])
	return t.newTokenAt(tokenStart, TRegularExpression, value)
}

func (t *Tokenizer) Peek() Token {
	token := t.peek()
	t.PreviousToken = t.CurrentToken
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
	panic(t.currentParseFailure("unexpected token: " + t.CurrentToken.Value))
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
	if lo.Contains(lineTerminators, ch) && !(ch == '\n' && t.Index > 0 && t.SourceText[t.Index-1] == '\r') {
		t.nextLine()
	}
	t.Index++
}

func (t *Tokenizer) atEnd() bool {
	return t.Index >= t.Length
}

func (t *Tokenizer) matchPrefix(prefixes ...string) bool {
	for _, prefix := range prefixes {
		if t.Index+len(prefix) <= t.Length && string(t.SourceText[t.Index:t.Index+len(prefix)]) == prefix {
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
