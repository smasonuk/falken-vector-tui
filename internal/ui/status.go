package ui

import (
	tui "github.com/smasonuk/earlgray"
	"github.com/smasonuk/falken-vector-tui/internal/app"
)

func statusScreen(model app.AppModel, cbs callbacks) tui.Node {
	if model.Status.NoIndex {
		return section(
			tui.Text("This directory has not been indexed yet.", tui.WithTextStyle(tui.Style{Bold: true})),
			tui.Text("Directory: "+model.Directory),
			tui.View(tui.Style{Direction: tui.Row, Gap: 2},
				button("Dry run ingest", model.Busy, cbs.startDryRun),
				button("Index this directory", model.Busy, cbs.startIngest),
			),
			tui.Text("CLI equivalent: falkengo ingest ."),
		)
	}
	text := model.Status.Text
	if text == "" {
		text = "Directory: " + model.Directory
	}
	return section(
		tui.TextPanel(tui.TextPanelProps{
			Text:          text,
			WordWrap:      true,
			Style:         tui.Style{FlexGrow: 1},
			FocusedStyle:  tui.Style{Foreground: tui.ANSIColor(6)},
			ShowScrollbar: true,
		}),
		tui.View(tui.Style{Direction: tui.Row, Gap: 2},
			button("Refresh", model.Busy, cbs.refreshStatus),
		),
	)
}
