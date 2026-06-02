package main

import (
	"fmt"
	"os"

	tui "github.com/smasonuk/earlgray"
	tuiconfig "github.com/smasonuk/falken-vector-tui/internal/config"
	"github.com/smasonuk/falken-vector-tui/internal/ui"
)

func main() {
	cfg, err := tuiconfig.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := tui.RunWithOptions(ui.NewRoot(cfg), tui.RunOptions{DisableCtrlCQuit: true}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
