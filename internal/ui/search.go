package ui

import (
	"strings"

	tui "github.com/smasonuk/earlgray"
	"github.com/smasonuk/falken-vector-tui/internal/app"
)

func searchScreen(model app.AppModel, cbs callbacks) tui.Node {
	items := make([]string, 0, len(model.Search.Results))
	for _, result := range model.Search.Results {
		items = append(items, result.Title)
	}
	preview := ""
	if model.Search.Selected >= 0 && model.Search.Selected < len(model.Search.Results) {
		preview = model.Search.Results[model.Search.Selected].Preview
	}
	return section(
		tui.Text("Question:"),
		tui.TextArea(tui.TextAreaProps{
			Value:       model.Search.Question,
			Placeholder: "Search the indexed directory",
			OnChange: func(value string) {
				cbs.update(func(model app.AppModel) app.AppModel {
					model.Search.Question = value
					return model
				})
			},
			Style:         tui.Style{Height: tui.Cells(4), Border: tui.BorderAll},
			FocusedStyle:  tui.Style{Foreground: tui.ANSIColor(3)},
			ShowScrollbar: true,
		}),
		tui.View(tui.Style{Direction: tui.Row, Gap: 2},
			button("Search", model.Busy, cbs.startSearch),
		),
		errorText(model.Search.InlineError),
		tui.View(tui.Style{Direction: tui.Row, FlexGrow: 1, Gap: 1},
			tui.List(tui.ListProps{
				Items:         items,
				SelectedIndex: model.Search.Selected,
				OnSelect: func(index int) {
					cbs.update(func(model app.AppModel) app.AppModel {
						model.Search.Selected = index
						return model
					})
				},
				Style:        tui.Style{Width: tui.Cells(42), Border: tui.BorderRight},
				FocusedStyle: tui.Style{Foreground: tui.ANSIColor(3), Bold: true},
			}),
			tui.TextPanel(tui.TextPanelProps{
				Text:          strings.TrimSpace(preview + "\n\n" + model.Search.QueryPlan),
				WordWrap:      true,
				ShowScrollbar: true,
				Style:         tui.Style{FlexGrow: 1},
				FocusedStyle:  tui.Style{Foreground: tui.ANSIColor(6)},
			}),
		),
	)
}
