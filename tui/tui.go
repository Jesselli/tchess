package tui

import (
	"fmt"
	"os"
	"strings"
)

type RGB [3]int

var (
	WHITE = RGB{255, 255, 255}
	BLACK = RGB{0, 0, 0}
	GRAY  = RGB{100, 100, 100}
	BLUE  = RGB{0, 0, 255}
)

func DrawMsgBox(msg string, x, y int, fg RGB, bg RGB, border bool) {
	lines := strings.Split(msg, "\n")
	width := 0
	for _, line := range lines {
		lineLen := len(line)
		if lineLen > width {
			width = lineLen
		}
	}

	var sb strings.Builder
	sb.WriteString("╔")
	for range width + 2 {
		sb.WriteString("═")
	}
	sb.WriteString("╗")
	top := sb.String()

	sb.Reset()
	sb.WriteString("╚")
	for range width + 2 {
		sb.WriteString("═")
	}
	sb.WriteString("╝")
	bot := sb.String()

	SaveCursorPos()
	defer RestoreCursorPos()

	SetFgBg(fg, bg)
	defer ResetStyle()

	MoveCursorTo(y, x)
	if border {
		fmt.Fprint(os.Stdout, top)
		for i, line := range lines {
			MoveCursorTo(y+i+1, x)
			row := fmt.Sprintf("║ %-*s ║", width, line)
			fmt.Fprint(os.Stdout, row)
		}
		MoveCursorTo(y+len(lines)+1, x)
		fmt.Fprint(os.Stdout, bot)
	} else {
		for i, line := range lines {
			MoveCursorTo(y+i+1, x)
			row := fmt.Sprintf("%-*s", width, line)
			fmt.Fprint(os.Stdout, row)
		}
	}
}

func ResetStyle() {
	fmt.Fprintf(os.Stdout, "\033[0m")
}

func MoveCursorTo(row, col int) {
	fmt.Fprintf(os.Stdout, "\033[%d;%dH", row, col)
}

func SaveCursorPos() {
	fmt.Fprint(os.Stdout, "\033[s")
}

func RestoreCursorPos() {
	fmt.Fprint(os.Stdout, "\033[u")
}

func EraseLine() {
	fmt.Fprintf(os.Stdout, "\033[2K")
}

func SetFgBg(fg, bg RGB) {
	fgStr := fmt.Sprintf("\033[38;2;%d;%d;%dm", fg[0], fg[1], fg[2])
	fmt.Fprint(os.Stdout, fgStr)

	bgStr := fmt.Sprintf("\033[48;2;%d;%d;%dm", bg[0], bg[1], bg[2])
	fmt.Fprint(os.Stdout, bgStr)
}

func CursorVisible(visible bool) {
	if visible {
		fmt.Print("\033[?25h")
	} else {
		fmt.Print("\033[?25l")
	}
}

func SetAlternateBuffer(on bool) {
	if on {
		fmt.Fprintf(os.Stdout, "\033[?1049h")
	} else {
		fmt.Fprintf(os.Stdout, "\033[?1049l")
	}
}
