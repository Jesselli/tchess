package piece

import "fmt"

type Piece struct {
	Type  Type
	Color Color
}

type Color int

const (
	WHITE Color = 1
	BLACK Color = 2
)

func (c Color) Opposite() Color {
	return (c % 2) + 1
}

type Type int

const (
	NONE   Type = 0
	PAWN   Type = 1
	ROOK   Type = 2
	KNIGHT Type = 3
	BISHOP Type = 4
	QUEEN  Type = 5
	KING   Type = 6
)

var PieceNames = map[Type]string{
	NONE:   "",
	PAWN:   "Pawn",
	ROOK:   "Rook",
	KNIGHT: "Knight",
	BISHOP: "Bishop",
	QUEEN:  "Queen",
	KING:   "King",
}

var PieceRunesOutlined = map[Type]rune{
	NONE:   ' ',
	PAWN:   '♙',
	ROOK:   '♖',
	KNIGHT: '♘',
	BISHOP: '♗',
	QUEEN:  '♕',
	KING:   '♔',
}

var PieceRunesFilled = map[Type]rune{
	NONE:   ' ',
	PAWN:   '♟',
	ROOK:   '♜',
	KNIGHT: '♞',
	BISHOP: '♝',
	QUEEN:  '♛',
	KING:   '♚',
}

var ToFenChar = map[Piece]byte{
	PAWN_W: 'P',
	PAWN_B: 'p',
	ROOK_W: 'R',
	ROOK_B: 'r',
	KNGT_W: 'N',
	KNGT_B: 'n',
	BSHP_W: 'B',
	BSHP_B: 'b',
	QUEN_W: 'Q',
	QUEN_B: 'q',
	KING_W: 'K',
	KING_B: 'k',
}

var FromFenChar = map[byte]Piece{
	'P': PAWN_W,
	'p': PAWN_B,
	'R': ROOK_W,
	'r': ROOK_B,
	'N': KNGT_W,
	'n': KNGT_B,
	'B': BSHP_W,
	'b': BSHP_B,
	'Q': QUEN_W,
	'q': QUEN_B,
	'K': KING_W,
	'k': KING_B,
}

var EMPTYP = Piece{}
var PAWN_W = Piece{PAWN, WHITE}
var PAWN_B = Piece{PAWN, BLACK}
var ROOK_W = Piece{ROOK, WHITE}
var ROOK_B = Piece{ROOK, BLACK}
var KNGT_W = Piece{KNIGHT, WHITE}
var KNGT_B = Piece{KNIGHT, BLACK}
var BSHP_W = Piece{BISHOP, WHITE}
var BSHP_B = Piece{BISHOP, BLACK}
var QUEN_W = Piece{QUEEN, WHITE}
var QUEN_B = Piece{QUEEN, BLACK}
var KING_W = Piece{KING, WHITE}
var KING_B = Piece{KING, BLACK}

// TODO: Unnecesssary. Remove.
func (p Piece) Equals(other Piece) bool {
	return p.Color == other.Color && p.Type == other.Type
}

func (p Piece) Name() string {
	color := "White"
	if p.Color == BLACK {
		color = "Black"
	}
	return fmt.Sprintf("%s %s", color, PieceNames[p.Type])
}
