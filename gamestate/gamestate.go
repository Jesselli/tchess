package gamestate

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jesselli/tchess/board"
	"github.com/Jesselli/tchess/parser"
	"github.com/Jesselli/tchess/piece"
	"github.com/Jesselli/tchess/tui"
)

var clockUpdateTicker = time.NewTicker(100 * time.Millisecond)

type Status string
type GameState struct {
	Board                board.Board
	ActiveColor          piece.Color
	enPassantSq          int // Square that can be taken en passant
	HalfMoveClock        int
	FullMoveCount        int
	Message              string
	Status               Status
	MoveHistory          []board.Move
	WhiteTimeRemainingMs int // In milliseconds
	BlackTimeRemainingMs int
	Increment            int
	WhiteIsHuman         bool
	BlackIsHuman         bool
	BoardHistory         []board.Board
	DrawMutex            sync.Mutex
}

const (
	STATUS_NOT_STARTED          Status = "Ready"
	STATUS_PLAYING              Status = "Playing..."
	STATUS_CHECKMATE_WHITE_WINS Status = "Checkmate! White wins."
	STATUS_CHECKMATE_BLACK_WINS Status = "Checkmate! Black wins."
	STATUS_TIMEOUT_WHITE_WINS   Status = "White wins on time!"
	STATUS_TIMEOUT_BLACK_WINS   Status = "Black wins on time!"
	STATUS_DRAW_INSUFFICIENT    Status = "Draw! Insufficient material."
	STATUS_DRAW_STALEMATE       Status = "Draw! Stalemate."
	STATUS_DRAW_REPETITION      Status = "Draw! Threefold repetition."
	STATUS_DRAW_FIFTY_MOVES     Status = "Draw! Fifty move rule."
	STATUS_DRAW_AGREEMENT       Status = "Draw by agreement!"
	STATUS_QUIT                 Status = "Quitting..."
)

func CreateDefault() *GameState {
	gs := GameState{}
	gs.ActiveColor = piece.WHITE
	gs.Board = board.CreateDefault()
	gs.enPassantSq = -1
	gs.WhiteIsHuman = true
	gs.BlackIsHuman = true
	gs.BoardHistory = make([]board.Board, 0)
	gs.Status = STATUS_NOT_STARTED
	return &gs
}

func (gs *GameState) UpdateAndDrawClocks(boardRotated bool) {
	for {
		<-clockUpdateTicker.C
		gs.DrawMutex.Lock()
		if gs.WhiteTimeRemainingMs <= 0 {
			gs.Status = STATUS_TIMEOUT_BLACK_WINS
			gs.Draw()
			break
		} else if gs.BlackTimeRemainingMs <= 0 {
			gs.Status = STATUS_TIMEOUT_WHITE_WINS
			gs.Draw()
			break
		}

		wFg := tui.GRAY
		bFg := tui.GRAY
		if gs.ActiveColor == piece.WHITE {
			gs.WhiteTimeRemainingMs -= 100
			wFg = tui.WHITE
		} else {
			gs.BlackTimeRemainingMs -= 100
			bFg = tui.WHITE
		}
		wTimeMs := gs.WhiteTimeRemainingMs
		wMin := wTimeMs / 1000 / 60
		wSec := wTimeMs - (wMin * 60 * 1000)
		wTimeMsg := fmt.Sprintf("%02d:%02d", int(wMin), int(wSec/1000))

		bTimeMs := gs.BlackTimeRemainingMs
		bMin := bTimeMs / 1000 / 60
		bSec := bTimeMs - (bMin * 60 * 1000)
		bTimeMsg := fmt.Sprintf("%02d:%02d", int(bMin), int(bSec/1000))

		if boardRotated {
			tui.DrawMsgBox(wTimeMsg, 22, 2, wFg, tui.BLACK, true)
			tui.DrawMsgBox(bTimeMsg, 22, 7, bFg, tui.BLACK, true)
		} else {
			tui.DrawMsgBox(wTimeMsg, 22, 7, wFg, tui.BLACK, true)
			tui.DrawMsgBox(bTimeMsg, 22, 2, bFg, tui.BLACK, true)
		}
		gs.DrawMutex.Unlock()
	}
}

func (gs *GameState) DrawCaptures(boardRotated bool) {
	var wCapSb strings.Builder // Pieces white has captured
	var bCapSb strings.Builder
	for _, v := range gs.Board.CapturedPieces {
		if v.Color == piece.WHITE {
			fmt.Fprintf(&bCapSb, "%c", piece.PieceRunesFilled[v.Type])
		} else {
			fmt.Fprintf(&wCapSb, "%c", piece.PieceRunesFilled[v.Type])
		}
	}

	wY := 5
	bY := 4
	if boardRotated {
		wY = 4
		bY = 5
	}
	tui.DrawMsgBox(wCapSb.String(), 22, wY, tui.WHITE, tui.BLACK, false)
	tui.DrawMsgBox(bCapSb.String(), 22, bY, tui.WHITE, tui.BLACK, false)
}

func (gs *GameState) DrawMoveHistory() {
	var sb strings.Builder
	moveNum := 1
	numRows := 8
	for i := 0; i < len(gs.MoveHistory); i += 2 {
		if len(gs.MoveHistory) > numRows*2 && i < len(gs.MoveHistory)-numRows*2 {
			moveNum++
			continue
		}
		mv := gs.MoveHistory[i]
		nextMv := board.Move{}
		if len(gs.MoveHistory) > i+1 {
			nextMv = gs.MoveHistory[i+1]
		}

		fmt.Fprint(&sb, fmt.Sprintf("%d. %s   %s\n", moveNum, mv.ToShortStr(), nextMv.ToShortStr()))
		moveNum++
	}
	tui.DrawMsgBox(sb.String(), 38, 1, tui.WHITE, tui.BLACK, false)
}

func (gs *GameState) ParseTimeControlFlag(tcFlag string) error {
	var err error
	var clock time.Duration
	var increment time.Duration
	tcSplit := strings.Split(tcFlag, "|")
	clock, err = time.ParseDuration(tcSplit[0])
	increment, err = time.ParseDuration(tcSplit[1])
	clockMs := int(clock.Milliseconds())
	incrementMs := int(increment.Milliseconds())
	gs.WhiteTimeRemainingMs = clockMs
	gs.BlackTimeRemainingMs = clockMs
	gs.Increment = incrementMs
	return err
}

func (gs *GameState) StartGame() {
	rotatedBoard := gs.BlackIsHuman && !gs.WhiteIsHuman
	go gs.UpdateAndDrawClocks(rotatedBoard)
	gs.Status = STATUS_PLAYING
}

func (gs *GameState) ParseAndExecuteAlgebraicNotation(cmd string) error {
	var err error
	var wantedMove board.Move
	wantedMove, err = parser.AlgebraicNotationToMove(cmd)
	if err == nil {
		var matchingMove board.Move
		matchingMove, err = gs.Board.FindMatchingMove(wantedMove, gs.ActiveColor)
		if err == nil {
			gs.UpdateStateAfterMove(matchingMove)
		}
	}

	if err != nil {
		gs.Message = err.Error()
	}

	return err
}

func (gs *GameState) Draw() {
	gs.DrawMutex.Lock()
	defer gs.DrawMutex.Unlock()

	rotatedBoard := gs.BlackIsHuman && !gs.WhiteIsHuman
	gs.Board.Display(os.Stdout, rotatedBoard)
	gs.DrawCaptures(rotatedBoard)
	gs.DrawMoveHistory()

	if gs.Status > STATUS_PLAYING {
		clockUpdateTicker.Stop()
		gs.Message = string(gs.Status)
	}
}

func (gs *GameState) ActivePlayerIsHuman() bool {
	return (gs.ActiveColor == piece.WHITE && gs.WhiteIsHuman) ||
		(gs.ActiveColor == piece.BLACK && gs.BlackIsHuman)
}

func (gs *GameState) AddIncrement() {
	if gs.ActiveColor == piece.WHITE {
		gs.WhiteTimeRemainingMs += gs.Increment
	} else {
		gs.BlackTimeRemainingMs += gs.Increment
	}
}

func (gs *GameState) SwitchTurn() {
	if gs.ActiveColor == piece.WHITE {
		gs.ActiveColor = piece.BLACK
	} else {
		gs.ActiveColor = piece.WHITE
	}
}

func (gs *GameState) UpdateStatus() {
	validMvs := gs.Board.AllValidMoves(gs.ActiveColor)
	inCheck := gs.Board.IsInCheck(gs.ActiveColor)
	if len(validMvs) == 0 && inCheck {
		if gs.ActiveColor == piece.WHITE {
			gs.Status = STATUS_CHECKMATE_BLACK_WINS
		} else {
			gs.Status = STATUS_CHECKMATE_WHITE_WINS
		}
	} else if len(validMvs) == 0 {
		gs.Status = STATUS_DRAW_STALEMATE
	} else if inCheck {
	}

	pCnt := gs.Board.PieceCounts()
	if !inCheck && pCnt[piece.EMPTYP] == 62 {
		// king vs. king
		gs.Status = STATUS_DRAW_INSUFFICIENT
	} else if !inCheck && pCnt[piece.EMPTYP] == 61 &&
		(pCnt[piece.BSHP_W] == 1 || pCnt[piece.BSHP_B] == 1) {
		// king & bishop vs. king
		gs.Status = STATUS_DRAW_INSUFFICIENT
	} else if !inCheck && pCnt[piece.EMPTYP] == 61 &&
		(pCnt[piece.KNGT_W] == 1 || pCnt[piece.KNGT_B] == 1) {
		// king & knight vs. king
		gs.Status = STATUS_DRAW_INSUFFICIENT
	} else if !inCheck && pCnt[piece.EMPTYP] == 60 &&
		(pCnt[piece.BSHP_W] == 1 && pCnt[piece.BSHP_B] == 1) {
		// king & bishop vs. king & bishop -- bishops on same sq color
		var wBishopOnLight bool
		var bBishopOnLight bool
		for sq, p := range gs.Board.Pieces {
			if p == piece.BSHP_W {
				wBishopOnLight = board.SquareIsLight(sq)
			} else if p == piece.BSHP_B {
				bBishopOnLight = board.SquareIsLight(sq)
			}
		}
		if wBishopOnLight == bBishopOnLight {
			gs.Status = STATUS_DRAW_INSUFFICIENT
		}
	}

	if gs.HalfMoveClock == 50 {
		gs.Status = STATUS_DRAW_FIFTY_MOVES
	}

	// Three-fold repetition
	positionCount := make(map[[64]piece.Piece]int, 0)
	for _, board := range gs.BoardHistory {
		positionCount[board.Pieces]++
		if positionCount[board.Pieces] >= 3 {
			gs.Status = STATUS_DRAW_REPETITION
			break
		}
	}
}

func (gs *GameState) UpdateMoveCounts(mv board.Move, c piece.Color) {
	if c == piece.BLACK {
		gs.HalfMoveClock++
		gs.FullMoveCount++
	} else {
		gs.HalfMoveClock++
	}

	if mv.Piece == piece.PAWN || mv.Capture {
		gs.HalfMoveClock = 0
	}
}

func (gs *GameState) ToFen() string {
	var sb strings.Builder

	// Board state
	emptyCount := 0
	for i, p := range gs.Board.Pieces {
		if i > 0 && i%8 == 0 {
			if emptyCount > 0 {
				fmt.Fprintf(&sb, "%d", emptyCount)
				emptyCount = 0
			}
			fmt.Fprint(&sb, "/")
		}

		if p.Type == piece.NONE {
			emptyCount++
		} else {
			if emptyCount > 0 {
				fmt.Fprintf(&sb, "%d", emptyCount)
				emptyCount = 0
			}
			fmt.Fprintf(&sb, "%c", piece.ToFenChar[p])
		}
	}
	if emptyCount > 0 {
		fmt.Fprintf(&sb, "%d", emptyCount)
	}

	// Player turn
	if gs.ActiveColor == piece.WHITE {
		fmt.Fprintf(&sb, " %c ", 'w')
	} else {
		fmt.Fprintf(&sb, " %c ", 'b')
	}

	// Castle availability
	if gs.Board.CastleRights&board.CASTLE_WHITE_SHORT == board.CASTLE_WHITE_SHORT {
		fmt.Fprintf(&sb, "%c", 'K')
	}
	if gs.Board.CastleRights&board.CASTLE_WHITE_LONG == board.CASTLE_WHITE_LONG {
		fmt.Fprintf(&sb, "%c", 'Q')
	}
	if gs.Board.CastleRights&board.CASTLE_BLACK_SHORT == board.CASTLE_BLACK_SHORT {
		fmt.Fprintf(&sb, "%c", 'k')
	}
	if gs.Board.CastleRights&board.CASTLE_BLACK_LONG == board.CASTLE_BLACK_LONG {
		fmt.Fprintf(&sb, "%c", 'q')
	}
	if gs.Board.CastleRights == 0 {
		fmt.Fprintf(&sb, "%c", '-')
	}

	// En passant sq
	if gs.enPassantSq >= 0 && gs.enPassantSq < 64 {
		fmt.Fprintf(&sb, " %s ", board.SqNumToStr(gs.enPassantSq))
	} else {
		fmt.Fprint(&sb, " - ")
	}

	// Half move clock
	fmt.Fprintf(&sb, "%d ", gs.HalfMoveClock)

	// Full move count
	fmt.Fprintf(&sb, "%d", gs.FullMoveCount)

	return sb.String()
}

func (gs *GameState) LoadFen(fen string) error {
	// Board state
	i := 0
	boardIdx := 0
	var b [64]piece.Piece
	var c byte = fen[0]
	for c != ' ' {
		if c >= '1' && c <= '8' {
			boardIdx += int(c-'1') + 1
		} else if c == '/' {
			// Do nothing. End of rank.
		} else {
			p := piece.FromFenChar[c]
			b[boardIdx] = p
			boardIdx++
		}

		i++
		c = fen[i]
	}
	gs.Board.Pieces = b

	// Active player
	// Current character should be a space, so advance the index
	i++
	c = fen[i]
	if c == 'w' {
		gs.ActiveColor = piece.WHITE
	} else if c == 'b' {
		gs.ActiveColor = piece.BLACK
	}

	// Castling
	// Next character should be a space, so advance the index twice
	i += 2
	var castleStatus uint8 = 0b0000
	c = fen[i]
	for c != ' ' {
		switch c {
		case '-':
			break
		case 'K':
			castleStatus |= board.CASTLE_WHITE_SHORT
		case 'Q':
			castleStatus |= board.CASTLE_WHITE_LONG
		case 'k':
			castleStatus |= board.CASTLE_BLACK_SHORT
		case 'q':
			castleStatus |= board.CASTLE_BLACK_LONG
		}
		i++
		c = fen[i]
	}
	gs.Board.CastleRights = castleStatus

	// En passant square
	i++
	c = fen[i]
	if c == '-' {
		gs.enPassantSq = -1
	} else {
		file := fen[i]
		rank := fen[i+1]
		alphaNum := fmt.Sprintf("%c%c", file, rank)
		gs.enPassantSq = board.StrToSqNum(alphaNum)
		i++
	}

	// TODO: Use fields for the rest of parsing
	fields := strings.Fields(fen)

	var retErr error

	// Half move clock
	halfMoves, err := strconv.Atoi(fields[4])
	if err != nil {
		retErr = fmt.Errorf("Error parsing half move clock: %w", err)
	}
	gs.HalfMoveClock = halfMoves

	// Full move count
	fullMoves, err := strconv.Atoi(fields[5])
	if err != nil {
		retErr = fmt.Errorf("Error parsing full move count: %w", err)
	}

	gs.FullMoveCount = fullMoves
	return retErr
}

// Fully updates the GameState after executing the specified Move. This includes
// updating the board, incrementing move counters, checking for win conditions,
// and switching the player turn.
func (gs *GameState) UpdateStateAfterMove(mv board.Move) {
	gs.BoardHistory = append(gs.BoardHistory, gs.Board)
	gs.Board.UpdateBoardWithMove(mv)
	gs.Board.HiglightSq = mv.TrgSqNum()
	gs.MoveHistory = append(gs.MoveHistory, mv)
	gs.Board.UpdateCastleRightsWithMove(mv, gs.ActiveColor)

	gs.UpdateMoveCounts(mv, gs.ActiveColor)
	gs.AddIncrement()
	gs.SwitchTurn()
	gs.UpdateStatus()
	gs.Message = mv.ToStr()
}
