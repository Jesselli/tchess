package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Jesselli/tchess/gamestate"
	"github.com/Jesselli/tchess/tui"
	"github.com/Jesselli/tchess/uci"
)

const (
	whitePlayerDefault = ""
	whitePlayerHelp    = "Name of UCI executable on PATH. If empty, player is human"
	blackPlayerDefault = ""
	blackPlayerHelp    = "Name of UCI executable on PATH. If empty, player is human"
	timeControlDefault = "15m|5s"
	timeControlHelp    = "5m|5s would be 5mins with a 5sec increment"
)

func parseFlags(gs *gamestate.GameState) error {
	var whitePlayer = flag.String("wp", whitePlayerDefault, whitePlayerHelp)
	var blackPlayer = flag.String("bp", blackPlayerDefault, blackPlayerHelp)
	var timeControl = flag.String("tc", timeControlDefault, timeControlHelp)
	flag.Parse()

	var err error
	if *whitePlayer != "" {
		// TODO: Use this value as the engine executable
		gs.WhiteIsHuman = false
	}
	if *blackPlayer != "" {
		gs.BlackIsHuman = false
	}

	err = gs.ParseTimeControlFlag(*timeControl)
	if err != nil {
		err = fmt.Errorf("Time control should be of the format 5m|5s. %w", err)
	}
	return err
}

func DrawMessagePrompt(gs *gamestate.GameState) {
	msg := gs.Message
	if gs.Status != gamestate.STATUS_PLAYING {
		msg = string(gs.Status)
	}

	tui.MoveCursorTo(11, 0)
	tui.EraseLine()
	fmt.Fprintf(os.Stdout, "%s\n", msg)
	tui.EraseLine()
	fmt.Fprintf(os.Stdout, "> ")
	gs.Message = ""
}

func main() {
	gs := gamestate.CreateDefault()
	err := parseFlags(gs)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var uciPipe uci.Pipe
	if !gs.WhiteIsHuman || !gs.BlackIsHuman {
		// TODO: Remove hard-coded 'stockfish' as the engine
		uciPipe, err = uci.CreatePipe("stockfish")
		if err == nil {
			uciPipe.Send(uci.UCI_SEND_UCI)
			uciPipe.WaitForExpected(uci.UCI_RECV_UCIOK)
		} else {
			fmt.Println(err.Error())
			return
		}
	}

	tui.CursorVisible(false)
	defer tui.CursorVisible(true)

	tui.SetAlternateBuffer(true)
	defer tui.SetAlternateBuffer(false)

	gs.StartGame()
	for {
		if gs.Status == gamestate.STATUS_QUIT {
			break
		}

		gs.Draw()
		DrawMessagePrompt(gs)

		if gs.ActivePlayerIsHuman() {
			PromptAndProcessUserInput(gs)
		} else if !gs.ActivePlayerIsHuman() && gs.Status == gamestate.STATUS_PLAYING {
			// Issue the game state to the engine and ask for its move
			uciPipe.SendPositionFen(gs.ToFen())
			bestMoveNotation := uciPipe.CalculateBestMove(10)
			gs.ParseAndExecuteAlgebraicNotation(bestMoveNotation)
		} else if gs.Status != gamestate.STATUS_PLAYING {
			PromptAndProcessUserInput(gs)
		}
	}
}

func PromptAndProcessUserInput(gs *gamestate.GameState) {
	var cmd string
	fmt.Scanln(&cmd)

	if cmd == "quit" || cmd == "exit" || cmd == "q" {
		gs.Status = gamestate.STATUS_QUIT
	} else if cmd == "help" {
		gs.Message = "Enter a move using algebraic notation. Or 'quit'."
	} else if gs.ActivePlayerIsHuman() && gs.Status == gamestate.STATUS_PLAYING {
		// Assume that we are issuing a move
		gs.ParseAndExecuteAlgebraicNotation(cmd)
	}
}
