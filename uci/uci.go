package uci

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

const (
	UCI_SEND_UCI          = "uci\n"
	UCI_SEND_POSITION_FEN = "position fen %s\n"
	UCI_SEND_GO_MOVETIME  = "go movetime %d\n"

	UCI_RECV_UCIOK    = "uciok"
	UCI_RECV_BESTMOVE = "bestmove"
)

type Pipe struct {
	in  *bufio.Writer
	out *bufio.Scanner
}

func CreatePipe(cmd string) (Pipe, error) {
	pipe := Pipe{}
	uciProg := exec.Command(cmd)
	var err error

	in, err := uciProg.StdinPipe()
	pipe.in = bufio.NewWriter(in)

	out, err := uciProg.StdoutPipe()
	pipe.out = bufio.NewScanner(out)

	err = uciProg.Start()
	if err != nil {
		err = fmt.Errorf("Could not communicate with %s.\n%w", cmd, err)
	}

	return pipe, err
}

func (p *Pipe) Send(cmd string) error {
	_, err := p.in.WriteString(cmd)
	if err == nil {
		p.in.Flush()
	}
	return err
}

func (p *Pipe) WaitForExpected(prefix string) string {
	// TODO: We need to handle the situation where this fails
	fullLine := ""
	for p.out.Scan() {
		t := p.out.Text()
		if strings.HasPrefix(t, prefix) {
			fullLine = t
			break
		}
	}
	return fullLine
}

func (p *Pipe) SendPositionFen(fen string) {
	cmd := fmt.Sprintf(UCI_SEND_POSITION_FEN, fen)
	p.Send(cmd)
}

func (p *Pipe) SendGoMoveTime(movetime int) {
	cmd := fmt.Sprintf(UCI_SEND_GO_MOVETIME, movetime)
	p.Send(cmd)
}

func (p *Pipe) CalculateBestMove(movetime int) string {
	cmd := fmt.Sprintf(UCI_SEND_GO_MOVETIME, movetime)
	p.Send(cmd)
	bestMoveLine := p.WaitForExpected(UCI_RECV_BESTMOVE)
	// Example: bestmove e2e4 ponder e7e5
	return bestMoveLine[9:13]
}
