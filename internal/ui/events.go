package ui

import (
	"strconv"
	"strings"

	tui "github.com/smasonuk/earlgray"
	"github.com/smasonuk/falken-vector-tui/internal/app"
)

func eventLog(model app.AppModel) tui.Node {
	lines := make([]string, 0, len(model.Events)+2)
	lines = append(lines, "Events:")
	if model.LastError != "" {
		lines = append(lines, "error "+model.LastError)
	}
	for _, event := range model.Events {
		lines = append(lines, event.Text)
	}
	return tui.View(
		tui.Style{Height: tui.Cells(7), Border: tui.BorderTop, Padding: tui.Insets{Left: 1}, Direction: tui.Column},
		tui.TextPanel(tui.TextPanelProps{
			Text:           strings.Join(lines, "\n"),
			WordWrap:       true,
			ShowScrollbar:  true,
			ResetScrollKey: strconv.Itoa(len(model.Events)) + ":" + model.LastError,
			InitialScrollY: 1 << 30,
			Style:          tui.Style{FlexGrow: 1},
			FocusedStyle:   tui.Style{Foreground: tui.ANSIColor(6)},
		}),
	)
}
