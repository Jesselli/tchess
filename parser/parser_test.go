package parser

import (
	"testing"

	"github.com/Jesselli/tchess/board"
	"github.com/Jesselli/tchess/piece"
)

func checkResult(cmd string, expectedMv, actualMv board.Move, err error, t *testing.T) {
	if err != nil {
		t.Fatalf("Expected command %s to be parsed successfully", cmd)
	}
	if !actualMv.Equals(expectedMv) {
		msg := "Unexpected result from %s\nExpected: %+v\nActual: %+v"
		t.Fatalf(msg, cmd, expectedMv, actualMv)
	}
}

func TestE1(t *testing.T) {
	cmd := "e1"
	expectedMv := board.Move{}
	expectedMv.Piece = piece.PAWN
	expectedMv.SrcFile = 'e'
	expectedMv.TrgFile = 'e'
	expectedMv.TrgRank = '1'
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}

func TestShortCastle(t *testing.T) {
	cmd := "o-o"
	expectedMv := board.Move{}
	expectedMv.Piece = piece.KING
	expectedMv.SrcFile = 'e'
	expectedMv.TrgFile = 'g'
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}

func TestLongCastle(t *testing.T) {
	cmd := "o-o-o"
	expectedMv := board.Move{}
	expectedMv.Piece = piece.KING
	expectedMv.SrcFile = 'e'
	expectedMv.TrgFile = 'c'
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}

func TestPieceMoveBf5(t *testing.T) {
	cmd := "Bf5"
	expectedMv := board.Move{}
	expectedMv.Piece = piece.BISHOP
	expectedMv.TrgFile = 'f'
	expectedMv.TrgRank = '5'
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}

func TestPawnPromoteE8Q(t *testing.T) {
	cmd := "e8Q"
	expectedMv := board.Move{}
	expectedMv.Piece = piece.PAWN
	expectedMv.TrgFile = 'e'
	expectedMv.TrgRank = '8'
	expectedMv.Promote = piece.QUEEN
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}

func TestLongNotationMoveE2e4(t *testing.T) {
	cmd := "e2e4"
	expectedMv := board.Move{}
	expectedMv.SrcFile = 'e'
	expectedMv.SrcRank = '2'
	expectedMv.TrgFile = 'e'
	expectedMv.TrgRank = '4'
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}

func TestLongNotationPromote(t *testing.T) {
	cmd := "e7e8Q"
	expectedMv := board.Move{}
	expectedMv.Piece = piece.PAWN
	expectedMv.SrcFile = 'e'
	expectedMv.SrcRank = '7'
	expectedMv.TrgFile = 'e'
	expectedMv.TrgRank = '8'
	expectedMv.Promote = piece.QUEEN
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}

func TestPawnCaptures(t *testing.T) {
	cmd := "cxd7"
	expectedMv := board.Move{}
	expectedMv.Piece = piece.PAWN
	expectedMv.SrcFile = 'c'
	expectedMv.TrgFile = 'd'
	expectedMv.TrgRank = '7'
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}

func TestPieceCaptures(t *testing.T) {
	cmd := "Rxf8"
	expectedMv := board.Move{}
	expectedMv.Piece = piece.ROOK
	expectedMv.TrgFile = 'f'
	expectedMv.TrgRank = '8'
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}

func TestPromotionEqualsSign(t *testing.T) {
	cmd := "g8=R"
	expectedMv := board.Move{}
	expectedMv.Piece = piece.PAWN
	expectedMv.TrgFile = 'g'
	expectedMv.TrgRank = '8'
	expectedMv.Promote = piece.ROOK
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}

func TestDisambiguateByFile(t *testing.T) {
	cmd := "Rdf8"
	expectedMv := board.Move{}
	expectedMv.Piece = piece.ROOK
	expectedMv.SrcFile = 'd'
	expectedMv.TrgFile = 'f'
	expectedMv.TrgRank = '8'
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}

func TestDisambiguateByRank(t *testing.T) {
	cmd := "R1a3"
	expectedMv := board.Move{}
	expectedMv.Piece = piece.ROOK
	expectedMv.SrcRank = '1'
	expectedMv.TrgFile = 'a'
	expectedMv.TrgRank = '3'
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}

func TestDisambiguateRankCapture(t *testing.T) {
	cmd := "N3xc5"
	expectedMv := board.Move{}
	expectedMv.Piece = piece.KNIGHT
	expectedMv.SrcRank = '3'
	expectedMv.TrgFile = 'c'
	expectedMv.TrgRank = '5'
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}

func TestFullDisambiguateCapture(t *testing.T) {
	cmd := "Qh4xe1"
	expectedMv := board.Move{}
	expectedMv.Piece = piece.QUEEN
	expectedMv.SrcRank = '4'
	expectedMv.SrcFile = 'h'
	expectedMv.TrgFile = 'e'
	expectedMv.TrgRank = '1'
	actualMv, err := AlgebraicNotationToMove(cmd)
	checkResult(cmd, expectedMv, actualMv, err, t)
}
