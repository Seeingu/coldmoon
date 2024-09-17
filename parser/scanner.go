package parser

import (
	"errors"
	t "github.com/Seeingu/coldmoon/token"
	"github.com/samber/lo"
	"log"
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

// MARK: Scanner utils

func (s *Scanner) error(m string) {
	log.Printf("scan failed: %s\n", m)
	log.Fatalf("scan error")
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

// TODO: Rename: it's a little bit confusing
// matchUntil match until f(rune) not satisfy
func (s *Scanner) matchUntil(f func(c rune) bool) string {
	start := s.index
	for !s.IsAtEnd() && f(s.Peek()) {
		s.index++
	}
	l := s.source[start:s.index]
	s.index++
	return l
}

func (s *Scanner) nextIndex() {
	s.index++
}

func (s *Scanner) matchUntilChar(c rune) string {
	return s.matchUntil(func(cc rune) bool {
		return c == cc
	})
}

// MARK: token generation

func (s *Scanner) multilineComment() string {
	return s.matchUntil(func(c rune) bool {
		return s.Peek() != '*' && s.Peek() != '/'
	})
}

// TODO: float, double
func (s *Scanner) number() t.Token {
	l := s.matchUntil(func(c rune) bool {
		return unicode.IsDigit(c)
	})
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

func (s *Scanner) string(endChar rune) t.Token {
	s.index++
	l := s.matchUntil(func(c rune) bool {
		return c != endChar
	})
	return t.NewToken(t.String, l)
}

func (s *Scanner) identifier() t.Token {
	start := s.index
	supportSpecialChars := []rune{'_'}
	for !s.IsAtEnd() {
		c := rune(s.source[s.index])
		if unicode.IsSpace(c) || c == '(' {
			break
		}
		if unicode.IsLetter(c) || unicode.IsDigit(c) || lo.Contains(supportSpecialChars, c) {
			s.index++
		} else {
			s.error("identifier, unexpected char: " + string(c))
		}
	}

	l := s.source[start:s.index]
	return t.NewToken(t.Identifier, l)
}

func (s *Scanner) scanToken() t.Token {
	c := s.Peek()
	if unicode.IsDigit(c) {
		return s.number()
	}
	if unicode.IsLetter(c) {
		return s.identifier()
	}
	switch c {
	case '"', '\'':
		token := s.string(c)
		if token, ok := s.keyword(token.Literal); ok {
			return token
		}
		return token
	case '+':
		if s.match("+=") {
			return t.NewToken(t.PlusEqual, "+=")
		} else if s.match("++") {
			return t.NewToken(t.PlusPlus, "++")
		}
		s.nextIndex()
		return t.NewToken(t.Plus, "+")
	case '-':
		if s.match("-=") {
			return t.NewToken(t.MinusEqual, "-=")
		} else if s.match("--") {
			return t.NewToken(t.MinusMinus, "--")
		}
		s.nextIndex()
		return t.NewToken(t.Minus, "-")
	case '*':
		if s.match("*=") {
			return t.NewToken(t.StarEqual, "*=")
		} else if s.match("**") {
			return t.NewToken(t.StarStar, "**")
		}
		s.nextIndex()
		return t.NewToken(t.Star, "*")
	case '/':
		if s.match("//") {
			comment := s.matchUntilChar('\n')
			return t.NewToken(t.SlashSlash, comment)
		} else if s.match("/*") {
			comment := s.multilineComment()
			return t.NewToken(t.SlashStar, comment)
		} else if s.match("/=") {
			return t.NewToken(t.SlashEqual, "/=")
		}
		s.nextIndex()
		return t.NewToken(t.Slash, "/")
	case '!':
		if s.match("!=") {
			return t.NewToken(t.BangEqual, "!=")
		}
		s.nextIndex()
		return t.NewToken(t.Bang, "!")
	case '=':
		if s.match("===") {
			return t.NewToken(t.EqualEqualEqual, "===")
		} else if s.match("==") {
			return t.NewToken(t.EqualEqual, "==")
		} else if s.match("=>") {
			return t.NewToken(t.EqualGreater, "=>")
		}

		s.nextIndex()
		return t.NewToken(t.Equal, "=")
	case '>':
		if s.match(">=") {
			return t.NewToken(t.GreaterEqual, ">=")
		} else if s.match(">>=") {
			return t.NewToken(t.GreaterGreaterEqual, ">>=")
		} else if s.match(">>") {
			return t.NewToken(t.GreaterGreater, ">>")
		} else if s.match(">>>") {
			return t.NewToken(t.GreaterGreaterGreater, ">>>")
		}
		s.nextIndex()
		return t.NewToken(t.Greater, ">")
	case '<':
		if s.match("<=") {
			return t.NewToken(t.LessEqual, "<=")
		} else if s.match("<<=") {
			return t.NewToken(t.LessEqual, "<<=")
		} else if s.match("<<") {
			return t.NewToken(t.LessLess, "<<")
		} else if s.match("<<<") {
			return t.NewToken(t.LessLessLess, "<<<")
		}
		s.nextIndex()
		return t.NewToken(t.Less, "<")
	case '?':
		s.nextIndex()
		return t.NewToken(t.Question, "?")
	case '(':
		s.nextIndex()
		return t.NewToken(t.LeftParenthesis, "(")
	case ')':
		s.nextIndex()
		return t.NewToken(t.RightParenthesis, ")")
	case '{':
		s.nextIndex()
		return t.NewToken(t.LeftBracket, "{")
	case '}':
		s.nextIndex()
		return t.NewToken(t.RightBracket, "}")
	case '[':
		s.nextIndex()
		return t.NewToken(t.LeftSquareBracket, "[")
	case ']':
		s.nextIndex()
		return t.NewToken(t.RightSquareBracket, "]")
	case ',':
		s.nextIndex()
		return t.NewToken(t.Comma, ",")
	case ':':
		s.nextIndex()
		return t.NewToken(t.Colon, ":")
	case ';':
		s.nextIndex()
		return t.NewToken(t.Semicolon, ";")
	case '.':
		if s.match("...") {
			return t.NewToken(t.DotDotDot, "...")
		}
		s.nextIndex()
		return t.NewToken(t.Dot, ".")
	case '~':
		s.nextIndex()
		return t.NewToken(t.Tilde, "~")
	}

	panic("unreachable")
}

// MARK: Public

func (s *Scanner) CurrentToken() t.Token {
	return s.currentToken
}

func (s *Scanner) NextToken() t.Token {
	return s.nextToken
}

// Scan will find next token and return
func (s *Scanner) Scan() t.Token {
	if s.IsAtEnd() {
		s.currentToken = s.nextToken
		s.nextToken = t.NewToken(t.EOF, "")
		return s.currentToken
	}
	token := s.scanToken()
	s.currentToken = s.nextToken
	s.nextToken = token
	return s.currentToken
}

func (s *Scanner) IsAtEnd() bool {
	return s.index >= len(s.source)
}

func (s *Scanner) HasNextToken() bool {
	return !s.nextToken.Is(t.EOF)
}

func (s *Scanner) indexIsAtEnd(index int) bool {
	return index >= 0 && index < len(s.source)
}
