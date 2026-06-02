package ui

import (
	tui "github.com/smasonuk/earlgray"
	"github.com/smasonuk/falken-vector-tui/internal/app"
	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

func indexScreen(model app.AppModel, cbs callbacks) tui.Node {
	result := ""
	if model.Index.LastResult != nil {
		result = app.FormatIngestResult(*model.Index.LastResult)
	}
	return section(
		tui.Text("Directory: "+model.Directory),
		tui.View(tui.Style{Direction: tui.Row, Gap: 2},
			tui.Text("Extensions:"),
			input(model.Index.ExtensionsInput, "go,md,txt", func(value string) {
				cbs.updatePersistent(func(model app.AppModel) app.AppModel {
					model.Index.ExtensionsInput = value
					return model
				})
			}),
		),
		tui.View(tui.Style{Direction: tui.Row, Gap: 2},
			tui.Text("Exclude extensions:"),
			input(model.Index.ExcludeExtensionsInput, "tmp,log", func(value string) {
				cbs.updatePersistent(func(model app.AppModel) app.AppModel {
					model.Index.ExcludeExtensionsInput = value
					return model
				})
			}),
			tui.Text("Exclude directories:"),
			input(model.Index.ExcludeDirsInput, "fixtures,dist", func(value string) {
				cbs.updatePersistent(func(model app.AppModel) app.AppModel {
					model.Index.ExcludeDirsInput = value
					return model
				})
			}),
		),
		tui.View(tui.Style{Direction: tui.Row, Gap: 2},
			tui.Text("Chunk size:"),
			input(model.Index.ChunkSizeInput, "1200", func(value string) {
				cbs.updatePersistent(func(model app.AppModel) app.AppModel {
					model.Index.ChunkSizeInput = value
					return model
				})
			}),
			tui.Text("Overlap:"),
			input(model.Index.ChunkOverlapInput, "200", func(value string) {
				cbs.updatePersistent(func(model app.AppModel) app.AppModel {
					model.Index.ChunkOverlapInput = value
					return model
				})
			}),
		),
		chunkerRow(model.Index.Chunker, cbs),
		tui.Checkbox(tui.CheckboxProps{
			Label: "Sync deleted files",
			Value: model.Index.SyncSource,
			OnChange: func(value bool) {
				cbs.updatePersistent(func(model app.AppModel) app.AppModel { model.Index.SyncSource = value; return model })
			},
			FocusedStyle: tui.Style{Foreground: tui.ANSIColor(3)},
		}),
		tui.View(tui.Style{Direction: tui.Row, Gap: 2},
			button("Dry run", model.Busy, cbs.startDryRun),
			button("Index this directory", model.Busy, cbs.startIngest),
			button("Cancel", !model.Busy, cbs.cancel),
		),
		errorText(model.Index.InlineError),
		tui.TextPanel(tui.TextPanelProps{
			Text:          result,
			WordWrap:      true,
			ShowScrollbar: true,
			Style:         tui.Style{FlexGrow: 1, Border: tui.BorderTop},
		}),
	)
}

func chunkerRow(selected falkenvector.ChunkerMode, cbs callbacks) tui.Node {
	modes := []falkenvector.ChunkerMode{
		falkenvector.ChunkerAuto,
		falkenvector.ChunkerFixed,
		falkenvector.ChunkerMarkdown,
		falkenvector.ChunkerText,
		falkenvector.ChunkerCode,
	}
	nodes := []tui.Node{tui.Text("Chunker:")}
	for _, mode := range modes {
		mode := mode
		label := string(mode)
		style := tui.Style{}
		if mode == selected {
			style = tui.Style{Bold: true, Foreground: tui.ANSIColor(3)}
		}
		nodes = append(nodes, tui.Keyed("chunker-"+label, tui.Button(tui.ButtonProps{
			Label:        label,
			Style:        style,
			FocusedStyle: tui.Style{Foreground: tui.ANSIColor(2), Bold: true},
			OnPress: func() {
				cbs.updatePersistent(func(model app.AppModel) app.AppModel {
					model.Index.Chunker = mode
					return model
				})
			},
		})))
	}
	return tui.View(tui.Style{Direction: tui.Row, Gap: 2}, nodes...)
}
