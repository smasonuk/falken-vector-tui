package app

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type ClipboardCommandRunner func(name string, args []string, input string) error

type clipboardCommand struct {
	name string
	args []string
}

func CopyToClipboard(text string) error {
	return copyToClipboardWith(runtime.GOOS, text, runClipboardCommand)
}

func copyToClipboardWith(goos, text string, run ClipboardCommandRunner) error {
	if text == "" {
		return nil
	}
	if run == nil {
		return errors.New("clipboard command runner is nil")
	}
	commands := clipboardCommands(goos)
	if len(commands) == 0 {
		return fmt.Errorf("unsupported clipboard platform %q", goos)
	}
	var errs []string
	for _, command := range commands {
		if err := run(command.name, command.args, text); err == nil {
			return nil
		} else {
			errs = append(errs, fmt.Sprintf("%s: %v", command.name, err))
		}
	}
	return fmt.Errorf("clipboard command not found or failed (%s)", strings.Join(errs, "; "))
}

func clipboardCommands(goos string) []clipboardCommand {
	switch goos {
	case "darwin":
		return []clipboardCommand{{name: "pbcopy"}}
	case "linux":
		return []clipboardCommand{
			{name: "wl-copy"},
			{name: "xclip", args: []string{"-selection", "clipboard"}},
			{name: "xsel", args: []string{"--clipboard", "--input"}},
		}
	case "windows":
		return []clipboardCommand{{name: "clip"}}
	default:
		return nil
	}
}

func runClipboardCommand(name string, args []string, input string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(input)
	if output, err := cmd.CombinedOutput(); err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return err
		}
		return fmt.Errorf("%w: %s", err, message)
	}
	return nil
}
