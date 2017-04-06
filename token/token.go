// Copyright 2017 Alejandro Sirgo Rica
//
// This file is part of Modbus_exporter.
//
//     Modbus_exporter is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     Modbus_exporter is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with Modbus_exporter.  If not, see <http://www.gnu.org/licenses/>.

// Package token provides the base lexical tokens for parsing files.
package token

import "strconv"

//Token is the token representation for the GryphOn uration files
type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF           // EOF
	COMMENT       // #

	literalBeg

	IDENT // Slave1
	INT   // 200
	FLOAT // 123.45
	literalEnd

	operatorBeg
	// Operators and delimiters
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	AND // &
	OR  // |

	DAND   // &&
	DOR    // ||
	ASSIGN // =
	NOT    // !

	comparisonOpBeg
	EQL // ==
	LSS // <
	GTR // >
	NEQ // !=
	LEQ // <=
	GEQ // >=
	comparisonOpEnd

	LPAREN   // (
	LBRACK   // [
	LBRACE   // {
	COMMA    // ,
	PERIOD   // .
	ELLIPSIS // ...

	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	COLON     // :

	ITERATOR // $
	operatorEnd

	keywordBeg

	OUTPUT // Output registers
	INPUT  // Input registers
	keywordEnd
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",

	IDENT: "IDENT",
	INT:   "INT",
	FLOAT: "FLOAT",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "%",

	AND:  "&",
	OR:   "|",
	DAND: "&&",
	DOR:  "||",

	EQL:    "==",
	LSS:    "<",
	GTR:    ">",
	ASSIGN: "=",
	NOT:    "!",

	NEQ: "!=",
	LEQ: "<=",
	GEQ: ">=",

	LPAREN:   "(",
	LBRACK:   "[",
	LBRACE:   "{",
	COMMA:    ",",
	PERIOD:   ".",
	ELLIPSIS: "...",

	RPAREN:    ")",
	RBRACK:    "]",
	RBRACE:    "}",
	SEMICOLON: ";",
	COLON:     ":",

	ITERATOR: "$",

	OUTPUT: "OUTPUT",
	INPUT:  "INPUT",
}

func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

// Precedence indicates the priority of the defined operators
func (tok Token) Precedence() int {
	switch tok {
	case DOR:
		return 1
	case DAND:
		return 2
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return 3
	case ADD, SUB:
		return 4
	case MUL, QUO, REM:
		return 5
	}
	return 0
}

// Predicates

// IsLiteral returns true for tokens corresponding to identifiers
// and basic type literals; it returns false otherwise.
func (tok Token) IsLiteral() bool { return literalBeg < tok && tok < literalEnd }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
func (tok Token) IsOperator() bool { return operatorBeg < tok && tok < operatorEnd }

// IsComparisonOperator returns true for tokens corresponding to comparison
// operators; it returns false otherwise.
func (tok Token) IsComparisonOperator() bool {
	return comparisonOpBeg < tok && tok < comparisonOpEnd
}

// IsKeyword returns true for tokens corresponding to keywords;
// it returns false otherwise.
func (tok Token) IsKeyword() bool { return keywordBeg < tok && tok < keywordEnd }
