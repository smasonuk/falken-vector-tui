package ui

import (
	tui "github.com/smasonuk/earlgray"
	"github.com/smasonuk/falken-vector-tui/internal/app"
)

func askScreen(model app.AppModel, cbs callbacks) tui.Node {
	page := section(
		tui.Text("Question:"),
		tui.TextArea(tui.TextAreaProps{
			Value:       model.Ask.Question,
			Placeholder: "Ask a question against the indexed directory",
			OnChange: func(value string) {
				cbs.update(func(model app.AppModel) app.AppModel {
					model.Ask.Question = value
					return model
				})
			},
			Style:         tui.Style{Height: tui.Cells(4), Border: tui.BorderAll},
			FocusedStyle:  tui.Style{Foreground: tui.ANSIColor(3)},
			ShowScrollbar: true,
		}),
		askControls(model, cbs),
		errorText(model.Ask.InlineError),
		tui.View(tui.Style{Direction: tui.Row, FlexGrow: 1, Gap: 1},
			tui.TextArea(tui.TextAreaProps{
				Value:         model.Ask.Answer,
				ReadOnly:      true,
				ShowScrollbar: true,
				OnCopy:        cbs.copySelection,
				Style:         tui.Style{FlexGrow: 2, Border: tui.BorderRight},
				FocusedStyle:  tui.Style{Foreground: tui.ANSIColor(6)},
			}),
			askSourcePanel(model, cbs),
		),
	)
	if !model.Ask.SourceDialogOpen {
		return page
	}
	return tui.Overlay(
		page,
		tui.Dialog(tui.DialogProps{
			CloseOnEsc: true,
			OnClose:    cbs.closeAskSource,
			Style: tui.Style{
				Border:    tui.BorderAll,
				Padding:   tui.All(1),
				Direction: tui.Column,
				FlexGrow:  1,
			},
		}, askSourceDialog(model, cbs)),
	)
}

func askControls(model app.AppModel, cbs callbacks) tui.Node {
	return tui.View(tui.Style{Direction: tui.Row, Gap: 2},
		button("Ask", model.Busy, cbs.startAsk),
	)
}

func askSourcePanel(model app.AppModel, cbs callbacks) tui.Node {
	sourceItems := make([]tui.ScrollableListItem, 0, len(model.Ask.SourceItems))
	for _, source := range model.Ask.SourceItems {
		sourceItems = append(sourceItems, tui.ScrollableListItem{
			ID:    source.Label,
			Label: source.Label,
		})
	}
	selected := model.Ask.SelectedSource
	if selected < 0 || selected >= len(sourceItems) {
		selected = 0
	}
	notes := app.FormatAskNotesPanel(model.Ask)
	children := []tui.Node{}
	if len(sourceItems) != 0 {
		children = append(children,
			tui.Text("Sources:"),
			tui.ScrollableList(tui.ScrollableListProps{
				Items:         sourceItems,
				SelectedIndex: selected,
				OnSelect:      cbs.selectAskSource,
				OnActivate:    cbs.openAskSource,
				OnClick:       cbs.openAskSource,
				VisibleRows:   8,
				ShowFooter:    true,
				Style:         tui.Style{Border: tui.BorderBottom, FlexGrow: 1, MinHeight: 3},
				FocusedStyle:  tui.Style{Foreground: tui.ANSIColor(3), Bold: true},
			}),
		)
	}
	if notes != "" {
		children = append(children, tui.TextPanel(tui.TextPanelProps{
			Text:          notes,
			WordWrap:      true,
			ShowScrollbar: true,
			Style:         tui.Style{FlexGrow: 1},
			FocusedStyle:  tui.Style{Foreground: tui.ANSIColor(6)},
		}))
	}
	if len(children) == 0 {
		children = append(children, tui.TextPanel(tui.TextPanelProps{
			Text:          "",
			WordWrap:      true,
			ShowScrollbar: true,
			Style:         tui.Style{FlexGrow: 1},
			FocusedStyle:  tui.Style{Foreground: tui.ANSIColor(6)},
		}))
	}
	return tui.View(tui.Style{FlexGrow: 1, Direction: tui.Column, Gap: 1}, children...)
}

func askSourceDialog(model app.AppModel, cbs callbacks) tui.Node {
	text := model.Ask.SourceDialogText
	if model.Ask.SourceDialogError != "" {
		text = model.Ask.SourceDialogError
	}
	return tui.View(tui.Style{Direction: tui.Column, Gap: 1, FlexGrow: 1},
		tui.Text(model.Ask.SourceDialogTitle, tui.WithTextStyle(tui.Style{Bold: true})),
		tui.TextArea(tui.TextAreaProps{
			Value:         text,
			ReadOnly:      true,
			NoWordWrap:    true,
			ShowScrollbar: true,
			OnCopy:        cbs.copySelection,
			AutoFocus:     true,
			Style:         tui.Style{FlexGrow: 1, Border: tui.BorderAll},
			FocusedStyle:  tui.Style{Foreground: tui.ANSIColor(6)},
		}),
		tui.View(tui.Style{Direction: tui.Row, Gap: 2},
			button("Close", false, cbs.closeAskSource),
		),
	)
}
