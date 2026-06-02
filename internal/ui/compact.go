package ui

import (
	tui "github.com/smasonuk/earlgray"
	"github.com/smasonuk/falken-vector-tui/internal/app"
)

func compactScreen(model app.AppModel, cbs callbacks) tui.Node {
	result := ""
	if model.Compact.LastResult != nil {
		result = app.FormatCompactResult(*model.Compact.LastResult)
	}
	confirm := ""
	if model.Compact.Confirming {
		confirm = "Compact will rebuild the vector database for this index. Press Compact now again to continue."
	}
	return section(
		tui.Text("Compact rebuilds the vector database from active manifest chunks."),
		tui.View(tui.Style{Direction: tui.Row, Gap: 2},
			tui.Text("Batch size:"),
			input(model.Compact.BatchSizeInput, "100", func(value string) {
				cbs.updatePersistent(func(model app.AppModel) app.AppModel {
					model.Compact.BatchSizeInput = value
					return model
				})
			}),
			tui.Checkbox(tui.CheckboxProps{
				Label: "Keep backup",
				Value: model.Compact.KeepBackup,
				OnChange: func(value bool) {
					cbs.updatePersistent(func(model app.AppModel) app.AppModel {
						model.Compact.KeepBackup = value
						return model
					})
				},
				FocusedStyle: tui.Style{Foreground: tui.ANSIColor(3)},
			}),
		),
		tui.View(tui.Style{Direction: tui.Row, Gap: 2},
			button("Dry run compact", model.Busy, cbs.startCompactDry),
			button("Compact now", model.Busy, cbs.startCompact),
			button("Cancel", !model.Busy, cbs.cancel),
		),
		tui.Text(confirm, tui.WithTextStyle(tui.Style{Foreground: tui.ANSIColor(3)})),
		errorText(model.Compact.InlineError),
		tui.TextPanel(tui.TextPanelProps{
			Text:          result,
			WordWrap:      true,
			ShowScrollbar: true,
			Style:         tui.Style{FlexGrow: 1, Border: tui.BorderTop},
		}),
	)
}
