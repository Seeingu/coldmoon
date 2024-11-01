package coldmoon

import lo "github.com/samber/lo"

type TokenType int

const (
	TLeftBrace TokenType = iota
	TRightBrace
	TLeftBracket
	TRightBracket
	TLeftParen
	TRightParen
	TPeriod
	TSemicolon
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
	TStar
	TPercent
	TIncrement
	TDecrement
	TLeftShift
	TRightShift
	TUnsignedRightShift
	TBitwiseAnd
	TBitwiseOr
	TBitwiseXor
	TNot
	TAnd
	TOr
	TQuestion
	TColon
	TTilde
	TSlash
	TEqualsEquals
	TNotEqualsEquals
	TPlusEquals
	TThis
	TMinusEquals
	TStarEquals
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
	TIf
	TElse
	TWhile
	TBreak
	TReturn
	TEOF
)

type Token struct {
	Type  TokenType
	Value string
}

type Tokenizer struct {
	SourceText   []rune
	Index        int
	Length       int
	CurrentToken Token
	NextToken    Token
}

func NewTokenizer(sourceText string) *Tokenizer {
	source := []rune(sourceText)
	tokenizer := &Tokenizer{
		SourceText: source,
		Index:      0,
		Length:     len(source),
	}
	tokenizer.Peek()
	tokenizer.Peek()
	return tokenizer
}

func (t *Tokenizer) peek() Token {
	t.skipWhiteSpace()
	if t.Index >= t.Length {
		return Token{Type: TEOF}
	}

	ch := t.SourceText[t.Index]
	switch ch {
	case '{':
		t.Index++
		return Token{Type: TLeftBrace, Value: "{"}
	case '}':
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
	case '.':
		t.Index++
		return Token{Type: TPeriod, Value: "."}
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
		return Token{Type: TLessThan, Value: "<"}
	case '>':
		t.Index++
		if t.Index < t.Length && t.SourceText[t.Index] == '=' {
			t.Index++
			return Token{Type: TGreaterThanEquals, Value: ">="}
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
		return Token{Type: TEquals, Value: "="}
	default:
		if lo.Contains(identifierStartCharset, ch) {
			return t.identifierOrKeyword()
		}
	}
	panic("unhandled token: " + string(ch))

}

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
	"if":       TIf,
	"else":     TElse,
	"true":     TTrue,
	"false":    TFalse,
	"null":     TNull,
	"debugger": TDebugger,
	"this":     TThis,
	"break":    TBreak,
	"while":    TWhile,
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

func (t *Tokenizer) Peek() Token {
	token := t.peek()
	t.CurrentToken = t.NextToken
	t.NextToken = token
	return t.CurrentToken
}

// TODO: remove
func (t *Tokenizer) Next() Token {
	t.skipWhiteSpace()
	token := t.Peek()
	return token
}

// MustMatch consumes the current token if it matches the given type.
// Will throw an error if the current token does not match the given type.
func (t *Tokenizer) MustMatch(tokenType TokenType) {
	if t.CurrentToken.Type == tokenType {
		t.Next()
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
			t.Index++
		} else {
			break
		}
	}
}
