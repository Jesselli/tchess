package gamestate

import (
	"fmt"
	"testing"

	"github.com/Jesselli/tchess/board"
	"github.com/Jesselli/tchess/piece"
)

func TestStalemate(t *testing.T) {
	gs := CreateDefault()
	gs.LoadFen("6k1/b7/8/8/5p2/7p/7P/7K w - - 0 54")
	gs.UpdateStatus()
	if gs.Status != STATUS_DRAW_STALEMATE {
		t.Fatalf("Expected status: %s Actual status: %s", STATUS_DRAW_STALEMATE, gs.Status)
	}
}

func TestPawnPromotion(t *testing.T) {
	gs := CreateDefault()
	gs.LoadFen("6k1/P7/8/4b3/5p2/7p/7P/7K w - - 1 54")
	mv := board.Move{}
	mv.Piece = piece.PAWN
	mv.SetSrcFromAlphaNum("a7")
	mv.SetTrgFromAlphaNum("a8")
	mv.Promote = piece.QUEEN
	gs.UpdateStateAfterMove(mv)
	actualFen := gs.ToFen()
	expectedFen := "Q5k1/8/8/4b3/5p2/7p/7P/7K b - - 0 54"
	if actualFen != expectedFen {
		msg := fmt.Sprintf("Expected: %s, Actual:%s", expectedFen, actualFen)
		t.Fatalf(msg)
	}
}
