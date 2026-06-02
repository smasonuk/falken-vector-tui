package ui

import (
	tui "github.com/smasonuk/earlgray"
	"github.com/smasonuk/falken-vector-tui/internal/app"
)

func header(model app.AppModel) tui.Node {
	children := []tui.Node{
		tui.Text(" falken-vector-tui - "+model.Directory, tui.WithTextStyle(tui.Style{FlexGrow: 1})),
	}
	if model.Busy {
		children = append(children, tui.Spinner(tui.SpinnerProps{
			Active: true,
			Label:  busyLabel(model),
			Style:  tui.Style{Foreground: tui.ANSIColor(6), Bold: true},
		}))
	}
	return tui.View(
		tui.Style{Height: tui.Cells(1), Border: tui.BorderBottom, Direction: tui.Row, Gap: 1},
		children...,
	)
}

func nav(model app.AppModel, cbs callbacks) tui.Node {
	items := []struct {
		label  string
		screen app.Screen
	}{
		{"1 Ask", app.ScreenAsk},
		{"2 Status", app.ScreenStatus},
		{"3 Index", app.ScreenIndex},
		{"4 Search", app.ScreenSearch},
		{"5 Compact", app.ScreenCompact},
	}
	nodes := make([]tui.Node, 0, len(items))
	for _, item := range items {
		item := item
		style := tui.Style{}
		if model.ActiveScreen == item.screen {
			style = tui.Style{Bold: true, Foreground: tui.ANSIColor(3)}
		}
		nodes = append(nodes, tui.Keyed(string(item.screen), tui.Button(tui.ButtonProps{
			Label:        item.label,
			OnPress:      func() { cbs.setScreen(item.screen) },
			Style:        style,
			FocusedStyle: tui.Style{Bold: true, Foreground: tui.ANSIColor(2)},
		})))
	}
	return tui.View(tui.Style{Height: tui.Cells(1), Direction: tui.Row, Gap: 2, Border: tui.BorderBottom}, nodes...)
}
