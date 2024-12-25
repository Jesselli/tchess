package parser

import (
	"fmt"

	"github.com/Jesselli/tchess/board"
	"github.com/Jesselli/tchess/piece"
)

var PieceChars = map[byte]piece.Type{
	'R': piece.ROOK,
	'N': piece.KNIGHT,
	'B': piece.BISHOP,
	'Q': piece.QUEEN,
	'K': piece.KING,
}

type TokenType int

const (
	TOKEN_PIECE        = 0
	TOKEN_CAPTURE      = 1
	TOKEN_SQUARE       = 2
	TOKEN_PROMOTE      = 3
	TOKEN_SHORT_CASTLE = 4
	TOKEN_LONG_CASTLE  = 5
	TOKEN_RANK         = 6
	TOKEN_FILE         = 7
)

type Token struct {
	Type    TokenType
	Literal string
}

func AlgebraicNotationToMove(notation string) (board.Move, error) {
	tokens, err := tokenizeCommand(notation)
	if err != nil {
		return board.Move{}, err
	}
	return parseTokensIntoMove(tokens)
}

func tokenizeCommand(command string) ([]Token, error) {
	var idx int = 0
	tokens := []Token{}

	var err error
	for idx < len(command) {
		new_idx, token, tokenErr := getNextToken(idx, command)
		if tokenErr != nil {
			err = tokenErr
			break
		}
		tokens = append(tokens, token)
		idx = new_idx
	}

	return tokens, err
}

func getNextToken(idx int, command string) (int, Token, error) {
	ch := command[idx]
	token := Token{}
	var err error
	if 'a' <= ch && ch <= 'h' {
		var peekCh byte = 0
		if len(command) > idx+1 {
			peekCh = command[idx+1]
		}

		if '1' <= peekCh && peekCh <= '8' {
			token = Token{TOKEN_SQUARE, command[idx : idx+2]}
			idx += 1
		} else {
			token = Token{TOKEN_FILE, string(ch)}
		}
	} else if '1' <= ch && ch <= '8' {
		token = Token{TOKEN_RANK, string(ch)}
	} else if ch == 'x' {
		token = Token{TOKEN_CAPTURE, "x"}
	} else if ch == '=' {
		token = Token{TOKEN_PROMOTE, "="}
	} else if ch == 'o' {
		if len(command) > idx+4 && command[idx:idx+5] == "o-o-o" {
			token = Token{TOKEN_LONG_CASTLE, command[idx : idx+5]}
			idx += 4
		} else if len(command) > idx+2 && command[idx:idx+3] == "o-o" {
			token = Token{TOKEN_SHORT_CASTLE, command[idx : idx+3]}
			idx += 2
		}
	} else if _, ok := PieceChars[ch]; ok {
		token = Token{TOKEN_PIECE, string(command[idx])}
	} else {
		err = fmt.Errorf("Unrecognized notation '%c'", ch)
	}

	idx += 1

	return idx, token, err
}

func parseTokensIntoMove(tokens []Token) (board.Move, error) {
	// TODO: Add assertions and errors
	mv := board.Move{}
	var err error

	if len(tokens) == 0 {
		err = fmt.Errorf("No input.")
	}

	if len(tokens) == 1 {
		token := tokens[0]
		if token.Type == TOKEN_SQUARE {
			// Pawn move (e4)
			mv.Piece = piece.PAWN
			mv.SrcFile = token.Literal[0]
			mv.SetTrgFromAlphaNum(token.Literal)
		} else if token.Type == TOKEN_SHORT_CASTLE {
			// Short castle (o-o)
			mv.Piece = piece.KING
			mv.SrcFile = 'e'
			mv.TrgFile = 'g'
		} else if token.Type == TOKEN_LONG_CASTLE {
			// Long castle (o-o-o)
			mv.Piece = piece.KING
			mv.SrcFile = 'e'
			mv.TrgFile = 'c'
		}
	}

	if len(tokens) == 2 {
		token1 := tokens[0]
		token2 := tokens[1]
		if token1.Type == TOKEN_PIECE &&
			token2.Type == TOKEN_SQUARE {
			// Piece move (Ra6)
			mv.Piece = PieceChars[token1.Literal[0]]
			mv.SetTrgFromAlphaNum(token2.Literal)
		} else if token1.Type == TOKEN_SQUARE &&
			token2.Type == TOKEN_SQUARE {
			// Long algebraic notation (e2e4)
			mv.SetSrcFromAlphaNum(token1.Literal)
			mv.SetTrgFromAlphaNum(token2.Literal)
		} else if token1.Type == TOKEN_SQUARE &&
			token2.Type == TOKEN_PIECE {
			// Pawn promotion (e8Q)
			mv.Piece = piece.PAWN
			mv.SetTrgFromAlphaNum(token1.Literal)
			mv.Promote = PieceChars[token2.Literal[0]]
		}
	}

	if len(tokens) == 3 {
		token1 := tokens[0]
		token2 := tokens[1]
		token3 := tokens[2]
		if token3.Type == TOKEN_SQUARE {
			mv.SetTrgFromAlphaNum(token3.Literal)

			if token1.Type == TOKEN_FILE {
				// Pawn captures square (exf6)
				mv.Piece = piece.PAWN
				mv.SrcFile = token1.Literal[0]
			} else if token1.Type == TOKEN_PIECE {
				// Piece captures square (Nxf3)
				mv.Piece = PieceChars[token1.Literal[0]]
			}

			if token2.Type == TOKEN_RANK {
				// Disambiguate -- Rdf8
				mv.SrcRank = token2.Literal[0]
			} else if token2.Type == TOKEN_FILE {
				// Disambiguate -- R1a3
				mv.SrcFile = token2.Literal[0]
			}
		} else if token3.Type == TOKEN_PIECE {
			mv.Piece = piece.PAWN
			if token1.Type == TOKEN_SQUARE &&
				token2.Type == TOKEN_SQUARE {
				// Long notation pawn promotion (e7e8Q)
				mv.SetSrcFromAlphaNum(token1.Literal)
				mv.SetTrgFromAlphaNum(token2.Literal)
				mv.Promote = PieceChars[token3.Literal[0]]
			} else if token1.Type == TOKEN_SQUARE {
				// Pawn promotion (e8=Q)
				mv.SetTrgFromAlphaNum(token1.Literal)
				mv.Promote = PieceChars[token3.Literal[0]]
			}
		}
	}

	if len(tokens) == 4 {
		token1 := tokens[0]
		token2 := tokens[1]
		token4 := tokens[3]

		mv.Piece = PieceChars[token1.Literal[0]]
		mv.SetTrgFromAlphaNum(token4.Literal)

		if token2.Type == TOKEN_SQUARE {
			mv.SetSrcFromAlphaNum(token2.Literal)
		} else if token2.Type == TOKEN_RANK {
			mv.SrcRank = token2.Literal[0]
		} else if token2.Type == TOKEN_FILE {
			mv.SrcFile = token2.Literal[0]
		}
	}

	return mv, err
}
