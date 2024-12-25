package board

import (
	"fmt"
	"io"

	"github.com/Jesselli/tchess/piece"
	"github.com/Jesselli/tchess/tui"
)

const (
	CASTLE_WHITE_SHORT uint8 = 0b0001
	CASTLE_WHITE_LONG  uint8 = 0b0010
	CASTLE_BLACK_SHORT uint8 = 0b0100
	CASTLE_BLACK_LONG  uint8 = 0b1000
)

type Board struct {
	Pieces         [64]piece.Piece
	CastleRights   uint8
	CapturedPieces []piece.Piece
	HiglightSq     int
}

func (b Board) Display(out io.Writer, rotated bool) {
	tui.MoveCursorTo(0, 0)

	for i := 0; i < 64; i++ {
		var p piece.Piece
		if rotated {
			p = b.Pieces[63-i]
		} else {
			p = b.Pieces[i]
		}

		if i%8 == 0 && rotated {
			fmt.Fprintln(out)
			fmt.Fprintf(out, "%d  ", 1+i/8)
		} else if i%8 == 0 {
			fmt.Fprintln(out)
			fmt.Fprintf(out, "%d  ", 8-i/8)
		}

		isWhiteSq := (i+i/8)%2 == 0
		isHighlight := i == b.HiglightSq
		if rotated {
			isHighlight = 63-i == b.HiglightSq
		}
		pieceRune := piece.PieceRunesFilled[p.Type]
		if isWhiteSq {
			// TODO: Use coloring from tui package
			if p.Color == piece.BLACK {
				fmt.Fprintf(out, "\u001b[38;5;16m")
			} else {
				fmt.Fprintf(out, "\u001b[38;5;231m")
			}

			if isHighlight {
				fmt.Fprintf(out, "\u001b[48;5;81m")
			} else {
				fmt.Fprintf(out, "\u001b[48;5;74m")
			}
		} else {
			if p.Color == piece.BLACK {
				fmt.Fprintf(out, "\u001b[38;5;16m")
			} else {
				fmt.Fprintf(out, "\u001b[38;5;231m")
			}

			if isHighlight {
				fmt.Fprintf(out, "\u001b[48;5;81m")
			} else {
				fmt.Fprintf(out, "\u001b[48;5;24m")
			}

		}

		fmt.Fprintf(out, "%c ", pieceRune)
		tui.ResetStyle()
	}

	fmt.Fprintf(out, "\n")

	if rotated {
		fmt.Fprint(out, "   h g f e d c b a\n")
	} else {
		fmt.Fprint(out, "   a b c d e f g h\n")
	}
}

func (b Board) CanCastleShort(c piece.Color) bool {
	if c == piece.WHITE {
		return b.CastleRights&CASTLE_WHITE_SHORT == CASTLE_WHITE_SHORT
	} else {
		return b.CastleRights&CASTLE_BLACK_SHORT == CASTLE_BLACK_SHORT
	}
}

func (b Board) CanCastleLong(c piece.Color) bool {
	if c == piece.WHITE {
		return b.CastleRights&CASTLE_WHITE_LONG == CASTLE_WHITE_LONG
	} else {
		return b.CastleRights&CASTLE_BLACK_LONG == CASTLE_BLACK_LONG
	}
}

func (b Board) AllMoves(player piece.Color) []Move {
	allMoves := []Move{}
	allMoves = append(allMoves, b.PawnMoves(player)...)
	allMoves = append(allMoves, b.KingMoves(player)...)
	allMoves = append(allMoves, b.RookMoves(player)...)
	allMoves = append(allMoves, b.QueenMoves(player)...)
	allMoves = append(allMoves, b.BishopMoves(player)...)
	allMoves = append(allMoves, b.KnightMoves(player)...)
	return allMoves
}

func (b Board) AllValidMoves(c piece.Color) []Move {
	allMoves := b.AllMoves(c)
	validMoves := make([]Move, 0)
	for _, mv := range allMoves {
		if ok, _ := b.ValidateMove(mv, c); ok {
			validMoves = append(validMoves, mv)
		}
	}
	return validMoves
}

func SquareIsLight(sqNum int) bool {
	return (sqNum+sqNum/8)%2 == 0
}

func SqNumPlusDelta(srcSqNum int, delta [2]int) (trgSqNum int, ok bool) {
	ok = true
	deltaX := delta[0]
	deltaY := delta[1]
	trgSqNum = srcSqNum + (deltaY * 8) + deltaX
	if (srcSqNum%8+deltaX != trgSqNum%8) || (srcSqNum/8+deltaY != trgSqNum/8) {
		ok = false
	} else if trgSqNum < 0 || trgSqNum >= 64 {
		ok = false
	}
	return trgSqNum, ok
}

func (b Board) PawnMoves(color piece.Color) []Move {
	up := 1
	if color == piece.BLACK {
		up = -1
	}

	var deltas = [4][2]int{
		{0, up}, {0, up * 2}, {-1, up}, {1, up},
	}

	moves := []Move{}
	for i, p := range b.Pieces {
		if p.Type != piece.PAWN || p.Color != color {
			continue
		}

		for _, d := range deltas {
			mv := Move{}
			mv.Piece = piece.PAWN
			mv.SetSrcFromSqNum(i)
			if mv.SetTrgDeltaAndCheckBounds(d[0], d[1]) {
				mv.Capture = b.Pieces[mv.TrgSqNum()] != piece.EMPTYP
				moves = append(moves, mv)
			}
		}
	}
	return moves
}

func (b Board) KnightMoves(color piece.Color) []Move {
	var deltas = [8][2]int{
		{1, 2}, {1, -2}, {-1, 2}, {-1, -2},
		{2, 1}, {2, -1}, {-2, 1}, {-2, -1},
	}

	moves := []Move{}
	for i, p := range b.Pieces {
		if p.Type != piece.KNIGHT || p.Color != color {
			continue
		}

		for _, d := range deltas {
			m := Move{}
			m.Piece = piece.KNIGHT
			m.SetSrcFromSqNum(i)
			if m.SetTrgDeltaAndCheckBounds(d[0], d[1]) {
				m.Capture = b.Pieces[m.TrgSqNum()] != piece.EMPTYP
				moves = append(moves, m)
			}
		}

	}
	return moves
}

func (b Board) KingMoves(color piece.Color) []Move {
	var deltas = [8][2]int{
		{0, 1}, {1, 1}, {1, 0}, {1, -1},
		{0, -1}, {-1, -1}, {-1, 0}, {-1, 1},
	}

	moves := []Move{}
	for i, p := range b.Pieces {
		if p.Type != piece.KING || p.Color != color {
			continue
		}

		for _, d := range deltas {
			m := Move{}
			m.Piece = piece.KING
			m.SetSrcFromSqNum(i)
			if m.SetTrgDeltaAndCheckBounds(d[0], d[1]) {
				m.Capture = b.Pieces[m.TrgSqNum()] != piece.EMPTYP
				moves = append(moves, m)
			}
		}

	}

	// Short castle
	m := Move{}
	m.Piece = piece.KING
	if color == piece.WHITE {
		m.SetSrcFromAlphaNum("e1")
		m.SetTrgFromAlphaNum("g1")
	} else {
		m.SetSrcFromAlphaNum("e8")
		m.SetTrgFromAlphaNum("g8")
	}
	moves = append(moves, m)

	// Long castle
	m = Move{}
	m.Piece = piece.KING
	if color == piece.WHITE {
		m.SetSrcFromAlphaNum("e1")
		m.SetTrgFromAlphaNum("c1")
	} else {
		m.SetSrcFromAlphaNum("e8")
		m.SetTrgFromAlphaNum("c8")
	}
	moves = append(moves, m)

	return moves
}

func (b Board) RookMoves(color piece.Color) []Move {
	var slideDeltas = [4][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}

	moves := []Move{}
	for i, p := range b.Pieces {
		if p.Type != piece.ROOK || p.Color != color {
			continue
		}

		for _, delta := range slideDeltas {
			m := Move{}
			m.Piece = piece.ROOK
			m.SetSrcFromSqNum(i)
			moves = append(moves, m.SlideMoves(b, delta)...)
		}

	}
	return moves
}

func (b Board) BishopMoves(color piece.Color) []Move {
	var slideDeltas = [4][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

	moves := []Move{}
	for i, p := range b.Pieces {
		if p.Type != piece.BISHOP || p.Color != color {
			continue
		}

		for _, delta := range slideDeltas {
			m := Move{}
			m.Piece = piece.BISHOP
			m.SetSrcFromSqNum(i)
			moves = append(moves, m.SlideMoves(b, delta)...)
		}

	}
	return moves
}

func (b Board) QueenMoves(color piece.Color) []Move {
	var slideDeltas = [8][2]int{
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
		{0, 1}, {0, -1}, {1, 0}, {-1, 0},
	}

	moves := []Move{}
	for i, p := range b.Pieces {
		if p.Type != piece.QUEEN || p.Color != color {
			continue
		}

		for _, delta := range slideDeltas {
			m := Move{}
			m.Piece = piece.QUEEN
			m.SetSrcFromSqNum(i)
			moves = append(moves, m.SlideMoves(b, delta)...)
		}

	}
	return moves
}

func (b Board) ValidateMove(mv Move, color piece.Color) (ok bool, msg string) {
	ok = true
	trgSq := mv.TrgSqNum()
	srcSq := mv.SrcSqNum()

	// TODO: Verify that castling doesn't move through or into check
	shortCastle := mv.IsShortCastle()
	longCastle := mv.IsLongCastle()
	if shortCastle && !b.CanCastleShort(color) {
		msg = "Can no longer short castle"
		ok = false
	} else if longCastle && !b.CanCastleLong(color) {
		msg = "Can no longer long castle"
		ok = false
	}

	if ok && b.Pieces[mv.SrcSqNum()].Color == b.Pieces[trgSq].Color {
		sqAlphaNum := SqNumToStr(trgSq)
		msg = fmt.Sprintf("%s is occupied by your own piece", sqAlphaNum)
		ok = false
	}

	if ok && mv.Piece != piece.KNIGHT {
		if p, sq := b.pieceBlockingPath(mv); !p.Equals(piece.EMPTYP) {
			sqAlphaNum := SqNumToStr(sq)
			msg = fmt.Sprintf("A %s on %s is blocking your path", p.Name(), sqAlphaNum)
			ok = false
		}
	}

	if ok && mv.Piece == piece.PAWN {
		dx, dy := mv.MoveDelta()
		if b.Pieces[srcSq].Color == piece.WHITE && mv.SrcRank != '2' && dy == 2 {
			msg = "A pawn can only move two spaces on its first move"
			ok = false
		} else if b.Pieces[srcSq].Color == piece.BLACK && mv.SrcRank != '7' && dy == -2 {
			msg = "A pawn can only move two spaces on its first move"
			ok = false
		} else if b.Pieces[trgSq].Type != piece.NONE && dx == 0 {
			msg = "Pawns cannot move into an occupied square"
			ok = false
		} else if b.Pieces[trgSq].Type == piece.NONE && dx != 0 {
			msg = "Pawns can only move diagonally to capture"
			ok = false
		}
		// else if mv.TrgRank == '1' || mv.TrgRank == '8' && mv.Promote == piece.NONE {
		// 	msg = "Must specify the piece for pawn promotion"
		// 	ok = false
		// }
	}

	// Check if the move results in a check
	// Don't consider moves where we actually take the king
	if ok && b.Pieces[mv.TrgSqNum()].Type != piece.KING {
		b.UpdateBoardWithMove(mv)
		if ok && b.IsInCheck(color) {
			msg = "Your king would be in check"
			ok = false
		}
	}

	return ok, msg
}

func (b Board) FindKing(color piece.Color) int {
	kingSq := -1
	for sqNum, p := range b.Pieces {
		if p.Type == piece.KING && p.Color == color {
			kingSq = sqNum
			break
		}
	}

	if kingSq == -1 {
		panic("King missing from board")
	}

	return kingSq
}

func (b Board) moveRevealsCheck(color piece.Color, mv Move) bool {
	revealsCheck := false
	kingSq := b.FindKing(color)
	srcSq := mv.SrcSqNum()
	trgSq := mv.TrgSqNum()
	dx := (srcSq % 8) - (kingSq % 8)
	dy := (kingSq / 8) - (srcSq / 8)
	normVec := [2]int{0, 0}
	if dx != 0 {
		normVec[0] = Abs(dx) / dx
	}
	if dy != 0 {
		normVec[1] = Abs(dy) / dy
	}

	b.UpdateBoardWithMove(mv)
	p, pieceSq := b.FirstPieceInDirection(kingSq, normVec)

	horizontalVerticalCheck := dx == 0 || dy == 0
	queenOrRook := p.Type == piece.ROOK || p.Type == piece.QUEEN

	diagonalCheck := dx == dy || dx == -dy
	queenOrBishop := p.Type == piece.BISHOP || p.Type == piece.QUEEN

	if pieceSq == trgSq || p.Color == color {
		// This means that the piece it found is the one that will be captured
		// OR it's the player's own piece
		revealsCheck = false
	} else if horizontalVerticalCheck && queenOrRook {
		revealsCheck = true
	} else if diagonalCheck && queenOrBishop {
		revealsCheck = true
	}

	return revealsCheck
}

func Abs(i int) int {
	if i < 0 {
		return -i
	} else {
		return i
	}
}

func (b Board) pieceBlockingPath(mv Move) (piece.Piece, int) {
	dx, dy := mv.MoveDelta()

	normDx := 0
	if dx > 0 {
		normDx = 1
	} else if dx < 0 {
		normDx = -1
	}

	normDy := 0
	if dy > 0 {
		normDy = 1
	} else if dy < 0 {
		normDy = -1
	}

	mvSrcSq := mv.SrcSqNum()
	mvTrgSq := mv.TrgSqNum()
	p, pieceSq := b.FirstPieceInDirection(mv.SrcSqNum(), [2]int{normDx, normDy})

	if mvSrcSq < mvTrgSq && pieceSq < mvTrgSq {
		return p, pieceSq
	} else if mvSrcSq > mvTrgSq && pieceSq > mvTrgSq {
		return p, pieceSq
	} else {
		return piece.EMPTYP, -1
	}
}

func (b Board) FirstPieceInDirection(srcSq int, normVector [2]int) (piece.Piece, int) {
	sqNum := srcSq
	pieceInPath := piece.EMPTYP
	for pieceInPath.Equals(piece.EMPTYP) && sqNum < 64 {
		sqNum += normVector[0] - (normVector[1] * 8)
		if sqNum > 0 && sqNum < 64 && b.Pieces[sqNum].Type != piece.NONE {
			pieceInPath = b.Pieces[sqNum]
			break
		} else if sqNum < 0 || sqNum >= 64 {
			break
		}
	}
	return pieceInPath, sqNum
}

// TODO: Make receivers uniform
func (b *Board) UpdateBoardWithMove(mv Move) {
	trg := mv.TrgSqNum()
	src := mv.SrcSqNum()

	if b.Pieces[trg] != piece.EMPTYP {
		b.CapturedPieces = append(b.CapturedPieces, b.Pieces[trg])
	}

	b.Pieces[trg] = b.Pieces[src]
	b.Pieces[src] = piece.EMPTYP

	// Special case -- castling also moves Rook
	if mv.Piece == piece.KING && src == 60 && trg == 62 {
		b.Pieces[63] = piece.EMPTYP
		b.Pieces[61] = piece.ROOK_W
	} else if mv.Piece == piece.KING && src == 60 && trg == 58 {
		b.Pieces[56] = piece.EMPTYP
		b.Pieces[59] = piece.ROOK_W
	} else if mv.Piece == piece.KING && src == 4 && trg == 6 {
		b.Pieces[7] = piece.EMPTYP
		b.Pieces[5] = piece.ROOK_B
	} else if mv.Piece == piece.KING && src == 4 && trg == 2 {
		b.Pieces[0] = piece.EMPTYP
		b.Pieces[3] = piece.ROOK_B
	}

	// Special case -- pawn promotion
	if mv.Piece == piece.PAWN && (mv.TrgRank == '1' || mv.TrgRank == '8') {
		promoPiece := piece.Piece{}
		if mv.Promote == piece.NONE {
			promoPiece.Type = piece.QUEEN
			promoPiece.Color = b.Pieces[trg].Color
		} else {
			promoPiece.Type = mv.Promote
			promoPiece.Color = b.Pieces[trg].Color
		}
		b.Pieces[trg] = promoPiece
	}
}

func CreateDefault() Board {
	b := Board{}
	b.Pieces = DefaultBoard
	b.CastleRights = 0b1111
	return b
}

var DefaultBoard = [64]piece.Piece{
	piece.ROOK_B, piece.KNGT_B, piece.BSHP_B, piece.QUEN_B, piece.KING_B, piece.BSHP_B, piece.KNGT_B, piece.ROOK_B,
	piece.PAWN_B, piece.PAWN_B, piece.PAWN_B, piece.PAWN_B, piece.PAWN_B, piece.PAWN_B, piece.PAWN_B, piece.PAWN_B,
	piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP,
	piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP,
	piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP,
	piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP, piece.EMPTYP,
	piece.PAWN_W, piece.PAWN_W, piece.PAWN_W, piece.PAWN_W, piece.PAWN_W, piece.PAWN_W, piece.PAWN_W, piece.PAWN_W,
	piece.ROOK_W, piece.KNGT_W, piece.BSHP_W, piece.QUEN_W, piece.KING_W, piece.BSHP_W, piece.KNGT_W, piece.ROOK_W,
}

// MOVE ///////////////////////////////////////////////////////////////////////

type Move struct {
	Piece   piece.Type
	SrcRank byte
	SrcFile byte
	TrgRank byte
	TrgFile byte
	Promote piece.Type
	Capture bool
}

func (m *Move) SetSrcFromAlphaNum(alphaNum string) {
	m.SrcFile = alphaNum[0]
	m.SrcRank = alphaNum[1]
}

func (m *Move) SrcSqNum() int {
	return int(m.SrcFile - 'a' + ('8'-m.SrcRank)*8)
}

func (m *Move) TrgSqNum() int {
	return int(m.TrgFile - 'a' + ('8'-m.TrgRank)*8)
}

func (m *Move) SetSrcFromSqNum(sqNum int) {
	m.SrcRank = '8' - byte(sqNum/8)
	m.SrcFile = 'a' + byte(sqNum%8)
}

func (m *Move) SetTrgFromAlphaNum(alphaNum string) {
	m.TrgFile = alphaNum[0]
	m.TrgRank = alphaNum[1]
}

func (m *Move) SetTrgFromSqNum(sqNum int) {
	m.TrgRank = '8' - byte(sqNum/8)
	m.TrgFile = 'a' + byte(sqNum%8)
}

func (m *Move) MoveDelta() (int, int) {
	deltaX := int(m.TrgFile) - int(m.SrcFile)
	deltaY := int(m.TrgRank) - int(m.SrcRank)
	return deltaX, deltaY
}

func (m *Move) IsShortCastle() bool {
	isKing := m.Piece == piece.KING
	whiteShortCastle := m.SrcSqNum() == 60 && m.TrgSqNum() == 62
	blackShortCastle := m.SrcSqNum() == 4 && m.TrgSqNum() == 6
	isShortCastle := whiteShortCastle || blackShortCastle
	return isKing && isShortCastle
}

func (m *Move) IsLongCastle() bool {
	isKing := m.Piece == piece.KING
	whiteLongCastle := m.SrcSqNum() == 60 && m.TrgSqNum() == 58
	blackLongCastle := m.SrcSqNum() == 4 && m.TrgSqNum() == 2
	isLongCastle := whiteLongCastle || blackLongCastle
	return isKing && isLongCastle
}

func (m *Move) SlideMoves(b Board, d [2]int) []Move {
	moves := []Move{}
	for i := 1; i < 8; i++ {
		move := Move{}
		move.Piece = m.Piece
		move.SrcRank = m.SrcRank
		move.SrcFile = m.SrcFile
		if move.SetTrgDeltaAndCheckBounds(d[0]*i, d[1]*i) {
			move.Capture = b.Pieces[move.TrgSqNum()] != piece.EMPTYP
			moves = append(moves, move)
		} else {
			break
		}
	}
	return moves
}

func (m *Move) SetTrgDeltaAndCheckBounds(deltaX, deltaY int) bool {
	m.TrgRank = m.SrcRank + byte(deltaY)
	m.TrgFile = m.SrcFile + byte(deltaX)

	withinBoardY := '1' <= m.TrgRank && m.TrgRank <= '8'
	withinBoardX := 'a' <= m.TrgFile && m.TrgFile <= 'h'
	return withinBoardX && withinBoardY
}

func (m *Move) Equals(other Move) bool {
	isEqual := m.Piece == other.Piece
	isEqual = isEqual && (m.SrcRank == other.SrcRank)
	isEqual = isEqual && (m.SrcFile == other.SrcFile)
	isEqual = isEqual && (m.TrgRank == other.TrgRank)
	isEqual = isEqual && (m.TrgFile == other.TrgFile)
	isEqual = isEqual && (m.Promote == other.Promote)
	return isEqual
}

func (m *Move) ToStr() string {
	return fmt.Sprintf("%s from %c%c to %c%c", piece.PieceNames[m.Piece], m.SrcFile, m.SrcRank, m.TrgFile, m.TrgRank)
}

func (m *Move) ToShortStr() string {
	if m.Capture {
		return fmt.Sprintf("%cx%c%c", piece.PieceRunesFilled[m.Piece], m.TrgFile, m.TrgRank)
	} else {
		return fmt.Sprintf("%c %c%c", piece.PieceRunesFilled[m.Piece], m.TrgFile, m.TrgRank)
	}
}

func (m *Move) Matches(wantedMv Move) bool {
	pieceMatches := m.Piece == wantedMv.Piece
	if m.TrgSqNum() == wantedMv.TrgSqNum() && m.SrcSqNum() == wantedMv.SrcSqNum() {
		// This is for long algebraic notation where a source square and
		// target square are specified.
		return true
	} else if pieceMatches && m.TrgSqNum() == wantedMv.TrgSqNum() {
		// We have to make sure we match any disambiguating data
		if wantedMv.SrcFile != 0 && m.SrcFile != wantedMv.SrcFile {
			return false
		} else if wantedMv.SrcRank != 0 && m.SrcRank != wantedMv.SrcRank {
			return false
		}
		return true
	} else if pieceMatches && m.Piece == piece.KING && m.SrcFile == wantedMv.SrcFile && m.TrgFile == wantedMv.TrgFile {
		// Castling
		return true
	}

	return false
}

func SqNumToStr(sqNum int) string {
	rank := '8' - byte(sqNum/8)
	file := 'a' + byte(sqNum%8)
	return fmt.Sprintf("%c%c", file, rank)
}

func StrToSqNum(alphaNum string) int {
	file := alphaNum[0]
	rank := alphaNum[1]
	return int(('8'-rank)*8 + (file - 'a'))
}

func (b Board) IsInCheck(c piece.Color) bool {
	// Just check if the moved piece is putting the king in check
	// OR if the moved piece reveals a check on the enemy.
	inCheck := false
	enemyMoves := b.AllMoves(c.Opposite())
	kingSq := b.FindKing(c)
	for _, mv := range enemyMoves {
		if mv.TrgSqNum() == kingSq {
			if ok, _ := b.ValidateMove(mv, c.Opposite()); ok {
				inCheck = true
				break
			}
		}
	}
	return inCheck
}

func (b Board) PieceCounts() map[piece.Piece]int {
	counts := make(map[piece.Piece]int, 12)
	for _, p := range b.Pieces {
		counts[p] += 1
	}
	return counts
}

func (b Board) FindMatchingMove(wantedMv Move, c piece.Color) (Move, error) {
	move := Move{}
	var err error

	allMoves := b.AllMoves(c)
	matchingMoves := []Move{}
	for _, mv := range allMoves {
		if mv.Matches(wantedMv) {
			if ok, msg := b.ValidateMove(mv, c); ok {
				mv.Promote = wantedMv.Promote
				matchingMoves = append(matchingMoves, mv)
			} else {
				err = fmt.Errorf(msg)
			}
		}
	}

	if len(matchingMoves) == 1 {
		move = matchingMoves[0]
		err = nil
	} else if len(matchingMoves) > 1 {
		err = fmt.Errorf("Ambiguous move")
		// TODO: Ask which piece they meant to move
	} else if err == nil {
		// No move candidates were found
		err = fmt.Errorf("Illegal move")
	}

	return move, err
}

// TODO: Make receivers uniform
func (b *Board) UpdateCastleRightsWithMove(mv Move, c piece.Color) {
	if mv.Piece == piece.KING && c == piece.WHITE {
		b.CastleRights &= ^(CASTLE_WHITE_SHORT | CASTLE_WHITE_LONG)
	} else if mv.Piece == piece.KING && c == piece.BLACK {
		b.CastleRights &= ^(CASTLE_BLACK_SHORT | CASTLE_BLACK_LONG)
	} else if mv.Piece == piece.ROOK && c == piece.WHITE {
		if mv.SrcSqNum() == 63 {
			b.CastleRights &= ^CASTLE_WHITE_SHORT
		} else if mv.SrcSqNum() == 56 {
			b.CastleRights &= ^CASTLE_WHITE_LONG
		}
	} else if mv.Piece == piece.ROOK && c == piece.BLACK {
		if mv.SrcSqNum() == 7 {
			b.CastleRights &= ^CASTLE_BLACK_SHORT
		} else if mv.SrcSqNum() == 0 {
			b.CastleRights &= ^CASTLE_BLACK_LONG
		}
	}
}
