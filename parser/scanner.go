package parser

import (
	"errors"
	t "github.com/Seeingu/coldmoon/token"
	"github.com/samber/lo"
	"log"
	"unicode"
)

type Scanner struct {
	source string
	index  int
	// line starts with 1
	line         uint
	col          uint
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
	s.Scan()
	s.Scan()
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
	log.Printf("line: %d, col: %d\n", s.line, s.col)
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

func (s *Scanner) newLine() {
	s.line++
	s.col = 1
}

func (s *Scanner) newToken(tokenType t.TokenType, literal string) t.Token {
	return t.NewToken(tokenType, literal, s.line, s.col)
}

// TODO: Rename: it's a little bit confusing
// matchUntil match until f(rune) not satisfy
// if matched, move index to next position
func (s *Scanner) matchUntil(f func(c rune) bool) string {
	start := s.index
	for !s.IsAtEnd() && f(s.Peek()) {
		if s.Peek() == '\n' {
			s.newLine()
		}
		s.nextIndex()
	}
	l := s.source[start:s.index]
	s.nextIndex()
	return l
}

func (s *Scanner) nextIndex() {
	s.col++
	s.index++
}

func (s *Scanner) matchUntilCharMatched(c rune) string {
	return s.matchUntil(func(cc rune) bool {
		return c != cc
	})
}

// MARK: token generation

func (s *Scanner) multilineComment() string {
	comment := s.matchUntil(func(c rune) bool {
		return s.Peek() != '*' && s.PeekNext() != '/'
	})
	s.nextIndex()
	return comment
}

// TODO: float, double
func (s *Scanner) number() t.Token {
	l := s.matchUntil(func(c rune) bool {
		return unicode.IsDigit(c)
	})
	return s.newToken(t.Number, l)
}

func (s *Scanner) keyword(v string) (tt t.Token, ok bool) {
	switch v {
	case "var":
		tt = s.newToken(t.Var, v)
	case "const":
		tt = s.newToken(t.Const, v)
	case "return":
		tt = s.newToken(t.Return, v)
	case "let":
		tt = s.newToken(t.Let, v)
	case "true":
		tt = s.newToken(t.True, v)
	case "false":
		tt = s.newToken(t.False, v)
	case "for":
		tt = s.newToken(t.For, v)
	case "while":
		tt = s.newToken(t.While, v)
	case "if":
		tt = s.newToken(t.If, v)
	case "function":
		tt = s.newToken(t.Function, v)
	case "else":
		tt = s.newToken(t.Else, v)
	case "null":
		tt = s.newToken(t.Null, v)
	case "undefined":
		tt = s.newToken(t.Undefined, v)
	case "throw":
		tt = s.newToken(t.Throw, v)
	case "new":
		tt = s.newToken(t.New, v)
	case "error":
		tt = s.newToken(t.Error, v)
	case "this":
		tt = s.newToken(t.This, v)
	case "super":
		tt = s.newToken(t.Super, v)
	default:
		// ignore
		return tt, false
	}
	return tt, true
}

func (s *Scanner) string(endChar rune) t.Token {
	s.nextIndex()
	l := s.matchUntil(func(c rune) bool {
		return c != endChar
	})
	return s.newToken(t.String, l)
}

func (s *Scanner) identifier() t.Token {
	start := s.index
	supportSpecialChars := []rune{'_'}
	for !s.IsAtEnd() {
		c := rune(s.source[s.index])
		if unicode.IsSpace(c) || c == '(' || c == ')' || c == '.' || c == ';' {
			break
		}
		if unicode.IsLetter(c) || unicode.IsDigit(c) || lo.Contains(supportSpecialChars, c) {
			s.index++
		} else {
			s.error("identifier, unexpected char: " + string(c))
		}
	}

	l := s.source[start:s.index]

	if token, ok := s.keyword(l); ok {
		return token
	}
	return s.newToken(t.Identifier, l)
}

func (s *Scanner) scanToken() t.Token {
	for unicode.IsSpace(s.Peek()) {
		if s.Peek() == '\n' {
			s.newLine()
		}
		s.nextIndex()
	}
	c := s.Peek()
	if unicode.IsDigit(c) {
		return s.number()
	}
	if unicode.IsLetter(c) || c == '_' {
		return s.identifier()
	}
	switch c {
	case '"', '\'':
		token := s.string(c)
		return token
	case '&':
		if s.match("&&") {
			return s.newToken(t.AmpersandAmpersand, "&&")
		}
		s.nextIndex()
		return s.newToken(t.AmpersandAmpersand, "&")
	case '|':
		if s.match("||") {
			return s.newToken(t.BarBar, "||")
		}
		s.nextIndex()
		return s.newToken(t.BarBar, "|")
	case '+':
		if s.match("+=") {
			return s.newToken(t.PlusEqual, "+=")
		} else if s.match("++") {
			return s.newToken(t.PlusPlus, "++")
		}
		s.nextIndex()
		return s.newToken(t.Plus, "+")
	case '-':
		if s.match("-=") {
			return s.newToken(t.MinusEqual, "-=")
		} else if s.match("--") {
			return s.newToken(t.MinusMinus, "--")
		}
		s.nextIndex()
		return s.newToken(t.Minus, "-")
	case '*':
		if s.match("*=") {
			return s.newToken(t.StarEqual, "*=")
		} else if s.match("**") {
			return s.newToken(t.StarStar, "**")
		}
		s.nextIndex()
		return s.newToken(t.Star, "*")
	case '/':
		if s.match("//") {
			comment := s.matchUntilCharMatched('\n')
			s.newLine()
			return s.newToken(t.SlashSlash, comment)
		} else if s.match("/*") {
			comment := s.multilineComment()
			return s.newToken(t.SlashStar, comment)
		} else if s.match("/=") {
			return s.newToken(t.SlashEqual, "/=")
		}
		s.nextIndex()
		return s.newToken(t.Slash, "/")
	case '!':
		if s.match("!=") {
			return s.newToken(t.BangEqual, "!=")
		}
		s.nextIndex()
		return s.newToken(t.Bang, "!")
	case '=':
		if s.match("===") {
			return s.newToken(t.EqualEqualEqual, "===")
		} else if s.match("==") {
			return s.newToken(t.EqualEqual, "==")
		} else if s.match("=>") {
			return s.newToken(t.EqualGreater, "=>")
		}

		s.nextIndex()
		return s.newToken(t.Equal, "=")
	case '>':
		if s.match(">=") {
			return s.newToken(t.GreaterEqual, ">=")
		} else if s.match(">>=") {
			return s.newToken(t.GreaterGreaterEqual, ">>=")
		} else if s.match(">>") {
			return s.newToken(t.GreaterGreater, ">>")
		} else if s.match(">>>") {
			return s.newToken(t.GreaterGreaterGreater, ">>>")
		}
		s.nextIndex()
		return s.newToken(t.Greater, ">")
	case '<':
		if s.match("<=") {
			return s.newToken(t.LessEqual, "<=")
		} else if s.match("<<=") {
			return s.newToken(t.LessEqual, "<<=")
		} else if s.match("<<") {
			return s.newToken(t.LessLess, "<<")
		} else if s.match("<<<") {
			return s.newToken(t.LessLessLess, "<<<")
		}
		s.nextIndex()
		return s.newToken(t.Less, "<")
	case '?':
		s.nextIndex()
		return s.newToken(t.Question, "?")
	case '(':
		s.nextIndex()
		return s.newToken(t.LeftParenthesis, "(")
	case ')':
		s.nextIndex()
		return s.newToken(t.RightParenthesis, ")")
	case '{':
		s.nextIndex()
		return s.newToken(t.LeftBracket, "{")
	case '}':
		s.nextIndex()
		return s.newToken(t.RightBracket, "}")
	case '[':
		s.nextIndex()
		return s.newToken(t.LeftSquareBracket, "[")
	case ']':
		s.nextIndex()
		return s.newToken(t.RightSquareBracket, "]")
	case ',':
		s.nextIndex()
		return s.newToken(t.Comma, ",")
	case ':':
		s.nextIndex()
		return s.newToken(t.Colon, ":")
	case ';':
		s.nextIndex()
		return s.newToken(t.Semicolon, ";")
	case '.':
		if s.match("...") {
			return s.newToken(t.DotDotDot, "...")
		}
		s.nextIndex()
		return s.newToken(t.Dot, ".")
	case '~':
		s.nextIndex()
		return s.newToken(t.Tilde, "~")
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

// Scan will return current token after move cursor
func (s *Scanner) Scan() t.Token {
	if s.IsAtEnd() {
		s.currentToken = s.nextToken
		s.nextToken = s.newToken(t.EOF, "")
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
