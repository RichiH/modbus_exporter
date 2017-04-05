// Copyright 2017 Alejandro Sirgo Rica
//
// This file is part of GryphOn.
//
//     GryphOn is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     GryphOn is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with GryphOn.  If not, see <http://www.gnu.org/licenses/>.

// Package lexer manages the imported configuration in order to generate secuences
// of tokens.
package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/lupoDharkael/modbus_exporter/token"

	multierror "github.com/hashicorp/go-multierror"
)

// Scanner is the scanning tool which contains de internal estate of analysis.
// Must be initialized via Init before use.
type Scanner struct {
	context string            // extra information for error notification
	src     []byte            // source
	err     *multierror.Error // error reporting

	// scanning state
	ch         rune // current character
	offset     int  // character offset
	rdOffset   int  // reading offset (position after current character)
	lineOffset int  // current line offset
	insertSemi bool // defines if a separator has to be added
}

// Init prepares the scanner s to tokenize the text src by setting the
// scanner at the beginning of src.
func (s *Scanner) Init(context string, src []byte) {
	s.context = context
	s.src = src
	s.err = new(multierror.Error)
	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
	s.insertSemi = false
	s.next()
}

// helper method for informative error generation
func (s *Scanner) error(msg string) {
	line := string(s.src[s.lineOffset:s.lookupLine()])
	if len(line) == 64 {
		line = line + "..."
	}
	newErr := fmt.Errorf("[%s] %s in line: \"%s\"", s.context, msg, line)
	s.err = multierror.Append(s.err, newErr)
}

// lookupLine gets the raw string of the actual line being analyzed. If the line is
// too long, it gets shortened.
func (s *Scanner) lookupLine() int {
	i := s.lineOffset
	for i < s.lineOffset+64 && s.src[i] != '\n' && i < len(s.src)-1 {
		i++
	}
	return i
}

// GetReport returns the errors of the scanner as a common error interface
func (s *Scanner) GetReport() error {
	return s.err.ErrorOrNil()
}

const bom = 0xFEFF // byte order mark, only permitted as very first character

func (s *Scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
		}
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r == 0:
			s.error("illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error("illegal UTF-8 encoding")
			} else if r == bom && s.offset > 0 {
				s.error("illegal byte order mark")
			}
		}
		s.rdOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		s.ch = -1 // eof
	}
}

// Scan scans the next token and returns the token position, the token,
// and its literal string if applicable. The source end is indicated by
// token.EOF.
func (s *Scanner) Scan() (tok token.Token, lit string, pos int) {
scanAgain:
	s.skipWhitespace()
	pos = s.lineOffset
	// determine token value
	switch ch := s.ch; {
	case isLetter(ch):
		s.insertSemi = true
		lit = s.scanIdentifier()
		switch lit {
		case "INPUT":
			tok = token.INPUT
		case "OUTPUT":
			tok = token.OUTPUT
		case "AND":
			tok = token.AND
		case "OR":
			tok = token.OR
		default:
			tok = token.IDENT
		}
	case '0' <= ch && ch <= '9':
		s.insertSemi = true
		tok, lit = s.scanNumber(false)
	default:
		s.next() // always make progress
		switch ch {
		case -1:
			tok = token.EOF
		case '\n':
			//ignore end of statement if the next char is a tab or if the actual line
			// has no text
			if s.ch == '\t' || !s.insertSemi {
				goto scanAgain
			} else {
				s.insertSemi = false
				return token.SEMICOLON, "\n", s.lineOffset
			}
		case ':':
			s.insertSemi = true
			tok = token.COLON
		case '.':
			s.insertSemi = true
			if '0' <= s.ch && s.ch <= '9' {
				tok, lit = s.scanNumber(true)
			} else if s.ch == '.' {
				s.next()
				if s.ch == '.' {
					s.next()
					tok = token.ELLIPSIS
				}
			} else {
				tok = token.PERIOD
			}
		case ',':
			s.insertSemi = true
			tok = token.COMMA
		case ';':
			s.insertSemi = true
			tok = token.SEMICOLON
			lit = ";"
		case '(':
			s.insertSemi = true
			tok = token.LPAREN
		case ')':
			s.insertSemi = true
			tok = token.RPAREN
		case '[':
			s.insertSemi = true
			tok = token.LBRACK
		case ']':
			s.insertSemi = true
			tok = token.RBRACK
		case '{':
			s.insertSemi = true
			tok = token.LBRACE
		case '}':
			s.insertSemi = true
			tok = token.RBRACE
		case '+':
			s.insertSemi = true
			tok = token.ADD
		case '-':
			s.insertSemi = true
			tok = token.SUB
		case '*':
			s.insertSemi = true
			tok = token.MUL
		case '/':
			s.insertSemi = true
			tok = token.QUO
		case '#':
			s.skipComment()
			goto scanAgain
		case '%':
			s.insertSemi = true
			tok = token.REM
		case '<':
			s.insertSemi = true
			tok = s.switch2(token.LSS, token.LEQ)
		case '>':
			s.insertSemi = true
			tok = s.switch2(token.GTR, token.GEQ)
		case '=':
			s.insertSemi = true
			tok = s.switch2(token.ASSIGN, token.EQL)
		case '!':
			s.insertSemi = true
			tok = s.switch2(token.NOT, token.NEQ)
		case '&':
			s.insertSemi = true
			if s.ch == '&' {
				tok = token.DAND
			} else {
				tok = token.AND
			}
		case '|':
			s.insertSemi = true
			if s.ch == '|' {
				tok = token.DOR
			} else {
				tok = token.OR
			}
		case '$':
			s.insertSemi = true
			tok = token.ITERATOR
		default:
			// next reports unexpected BOMs - don't repeat
			if ch != bom {
				s.error(fmt.Sprintf("illegal character %#U", ch))
			}
			tok = token.ILLEGAL
			lit = string(ch)
		}
	}
	return
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\r' { //|| s.ch == '\n'
		s.next()
	}
}

func (s *Scanner) skipComment() {
	s.next()
	for s.ch != '\n' && s.ch >= 0 {
		s.next()
	}
}

func (s *Scanner) scanIdentifier() string {
	offs := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	return string(s.src[offs:s.offset])
}

func (s *Scanner) scanNumber(seenDecimalPoint bool) (token.Token, string) {
	// digitVal(s.ch) < 10
	offs := s.offset
	tok := token.INT

	if seenDecimalPoint {
		offs--
		tok = token.FLOAT
		s.scanMantissa(10)
		goto exponent
	}
	if s.ch == '0' {
		// int or float
		offs := s.offset
		s.next()
		if s.ch == 'x' || s.ch == 'X' {
			// hexadecimal int
			s.next()
			s.scanMantissa(16)
			if s.offset-offs <= 2 {
				// only scanned "0x" or "0X"
				s.error("illegal hexadecimal number")
			}
		} else {
			// octal int or float
			seenDecimalDigit := false
			s.scanMantissa(8)
			if s.ch == '8' || s.ch == '9' {
				// illegal octal int or float
				seenDecimalDigit = true
				s.scanMantissa(10)
			}
			if s.ch == '.' || s.ch == 'e' || s.ch == 'E' || s.ch == 'i' {
				goto fraction
			}
			// octal int
			if seenDecimalDigit {
				s.error("illegal octal number")
			}
		}
		goto exit
	}

	// decimal int or float
	s.scanMantissa(10)
fraction:
	if s.ch == '.' {
		tok = token.FLOAT
		s.next()
		s.scanMantissa(10)
	}

exponent:
	if s.ch == 'e' || s.ch == 'E' {
		tok = token.FLOAT
		s.next()
		if s.ch == '-' || s.ch == '+' {
			s.next()
		}
		s.scanMantissa(10)
	}

exit:
	return tok, string(s.src[offs:s.offset])
}

func (s *Scanner) scanMantissa(base int) {
	for digitVal(s.ch) < base {
		s.next()
	}
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}
	return 16 // larger than any legal digit val
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' ||
		ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

// Helper functions for scanning multi-byte tokens such as >=
func (s *Scanner) switch2(tok0, tok1 token.Token) token.Token {
	if s.ch == '=' {
		s.next()
		return tok1
	}
	return tok0
}
