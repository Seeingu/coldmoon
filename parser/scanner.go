package parser

import (
	"errors"
	t "github.com/Seeingu/coldmoon/token"
	"unicode"
)

type Scanner struct {
	source       string
	index        int
	line         int
	currentToken t.Token
	nextToken    t.Token
}

func NewScanner(source string) *Scanner {
	s := &Scanner{
		source: source,
		index:  0,
		line:   1,
	}
	// LR(1)
	// Scan after init, make sure currentToken always exist
	s.nextToken = s.scanToken()
	return s
}

func (s *Scanner) Next() {

}

func (s *Scanner) Peek() rune {
	return rune(s.source[s.index])
}

func (s *Scanner) PeekNext() rune {
	return rune(s.source[s.index+1])
}

var (
	errorIndexOutOfSource = errors.New("index is out of source")
)

func (s *Scanner) PeekNextMany(i int) (r rune, err error) {
	ii := s.index + i
	if s.indexIsAtEnd(ii) {
		err = errorIndexOutOfSource
		return
	}
	r = rune(s.source[s.index+i])
	return
}

func (s *Scanner) error() {

}

// match will update index if matched target str
func (s *Scanner) match(str string) bool {
	end := s.index + len(str)
	if s.source[s.index:end] == str {
		s.index = end
		return true
	} else {
		return false
	}
}

// TODO: float, double
func (s *Scanner) number() t.Token {
	start := s.index
	for !s.IsAtEnd() && unicode.IsDigit(rune(s.source[s.index])) {
		s.index++
	}
	l := s.source[start:s.index]
	s.index++
	return t.NewToken(t.Number, l)
}

func (s *Scanner) keyword(v string) (tt t.Token, ok bool) {
	switch v {
	case "var":
		tt = t.NewToken(t.Var, v)
	case "const":
		tt = t.NewToken(t.Const, v)
	case "let":
		tt = t.NewToken(t.Let, v)
	case "true":
		tt = t.NewToken(t.True, v)
	case "false":
		tt = t.NewToken(t.False, v)
	case "for":
		tt = t.NewToken(t.For, v)
	case "while":
		tt = t.NewToken(t.While, v)
	case "if":
		tt = t.NewToken(t.If, v)
	case "function":
		tt = t.NewToken(t.Function, v)
	case "else":
		tt = t.NewToken(t.Else, v)
	case "null":
		tt = t.NewToken(t.Null, v)
	case "undefined":
		tt = t.NewToken(t.Undefined, v)
	default:
		// ignore
		return tt, false
	}
	return
}

func (s *Scanner) string() t.Token {
	start := s.index
	for !s.IsAtEnd() && !unicode.IsSpace(rune(s.source[s.index])) {
		s.index++
	}

	l := s.source[start:s.index]
	s.index++
	return t.NewToken(t.String, l)
}

func (s *Scanner) scanToken() t.Token {
	c := rune(s.source[s.index])
	if unicode.IsDigit(c) {
		return s.number()
	}
	if unicode.IsLetter(c) {
		token := s.string()
		if token, ok := s.keyword(token.Literal); ok {
			return token
		}
		return token
	}
	switch c {
	case '=':
		if s.match("===") {
			return t.NewToken(t.EqualEqualEqual, "===")
		} else if s.match("==") {
			return t.NewToken(t.EqualEqual, "==")
		} else {
			return t.NewToken(t.Equal, "=")
		}
	case '(':
		return t.NewToken(t.LeftParenthesis, "(")
	case ')':
		return t.NewToken(t.RightParenthesis, ")")
	case '{':
		return t.NewToken(t.LeftBracket, "{")
	case '}':
		return t.NewToken(t.RightBracket, "}")
	case '[':
		return t.NewToken(t.LeftSquareBracket, "[")
	case ']':
		return t.NewToken(t.RightSquareBracket, "]")
	case ',':
		return t.NewToken(t.Comma, ",")
	case '.':
		if s.match("...") {
			return t.NewToken(t.DotDotDot, "...")
		}
		return t.NewToken(t.Dot, ".")
	}

	panic("unreachable")
}

func (s *Scanner) CurrentToken() t.Token {
	return s.currentToken
}

func (s *Scanner) NextToken() t.Token {
	return s.nextToken
}

// Scan will find next token and return
func (s *Scanner) Scan() t.Token {
	token := s.scanToken()
	s.currentToken = s.nextToken
	s.nextToken = token
	return s.currentToken
}

func (s *Scanner) IsAtEnd() bool {
	return s.index >= len(s.source)
}

func (s *Scanner) indexIsAtEnd(index int) bool {
	return index >= 0 && index < len(s.source)
}
