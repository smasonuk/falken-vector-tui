package ui

import (
	"fmt"
	"os"
	"time"

	tui "github.com/smasonuk/earlgray"
	"github.com/smasonuk/falken-vector-tui/internal/app"
	tuiconfig "github.com/smasonuk/falken-vector-tui/internal/config"
	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

type callbacks struct {
	setScreen        func(app.Screen)
	update           func(func(app.AppModel) app.AppModel)
	updatePersistent func(func(app.AppModel) app.AppModel)
	refreshStatus    func()
	startDryRun      func()
	startIngest      func()
	startSearch      func()
	startAsk         func()
	startCompactDry  func()
	startCompact     func()
	copySelection    func(string)
	selectAskSource  func(int)
	openAskSource    func(int)
	closeAskSource   func()
	cancel           func()
}

func NewRoot(cfg tuiconfig.Config) func() tui.Node {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	factory := app.EngineFactoryFromConfig(cfg)
	preferencesPath := app.ResolvePreferencesPath(cfg, os.Getenv, cwd)
	preferences, preferencesErr := app.LoadPreferences(preferencesPath)
	return func() tui.Node {
		return tui.ComponentWithKey("falken-vector-tui-root", func() tui.Node {
			return appRoot(cwd, factory, preferencesPath, preferences, preferencesErr)
		})
	}
}

func appRoot(directory string, factory app.EngineFactory, preferencesPath string, preferences app.Preferences, preferencesErr error) tui.Node {
	initialModel := app.ApplyPreferences(app.NewModel(directory), preferences)
	if preferencesErr != nil {
		initialModel.LastError = preferencesErr.Error()
		initialModel = app.WithEventLine(initialModel, app.EventLine{Text: "warning " + preferencesErr.Error()})
	}
	model, _, updateModel := tui.UseStateWithUpdater(initialModel)
	runnerRef := tui.UseRef(app.NewOperationRunner())
	runner := *runnerRef
	tuiApp := tui.UseApp()
	updatePersistent := func(update func(app.AppModel) app.AppModel) {
		updateModel(func(model app.AppModel) app.AppModel {
			model = update(model)
			if err := app.SavePreferences(preferencesPath, app.PreferencesFromModel(model)); err != nil {
				model.LastError = err.Error()
			}
			return model
		})
	}

	setBusy := func(kind app.OperationKind) {
		updateModel(func(model app.AppModel) app.AppModel {
			model.Busy = true
			model.BusyKind = kind
			model.LastError = ""
			return model
		})
	}
	setError := func(err error) {
		if err == nil {
			return
		}
		updateModel(func(model app.AppModel) app.AppModel {
			model.LastError = err.Error()
			return model
		})
	}
	startStatus := func() {
		if err := app.StartStatus(runner, factory, directory); err != nil {
			setError(err)
			return
		}
		setBusy(app.OperationStatus)
	}

	tui.UseEffect(func() func() {
		startStatus()
		return nil
	})

	tui.UseChannel(runner.Results(), func(result app.OperationResult) {
		updateModel(func(model app.AppModel) app.AppModel {
			model.Busy = false
			model.BusyKind = app.OperationNone
			if result.Err != nil {
				model.LastError = result.Err.Error()
				switch result.Kind {
				case app.OperationSearch:
					if app.IsNoIndexError(result.Err) {
						model.Search.InlineError = app.NoIndexSearchMessage()
					} else {
						model.Search.InlineError = result.Err.Error()
					}
				case app.OperationAsk:
					if app.IsNoIndexError(result.Err) {
						model.Ask.InlineError = app.NoIndexSearchMessage()
					} else {
						model.Ask.InlineError = result.Err.Error()
					}
				case app.OperationIngest:
					model.Index.InlineError = result.Err.Error()
				case app.OperationCompact:
					model.Compact.InlineError = result.Err.Error()
				}
				return model
			}
			model.LastError = ""
			switch {
			case result.Status != nil:
				model.Status = *result.Status
			case result.Ingest != nil:
				model.Index.LastResult = result.Ingest
				model.Index.InlineError = ""
				if result.Ingest.ChunksEmbedded == 0 {
					model.Index.LastDryRunComplete = true
				}
			case result.Search != nil:
				model.Search = *result.Search
				model.Search.InlineError = ""
			case result.Ask != nil:
				model.Ask = *result.Ask
				model.Ask.InlineError = ""
			case result.Compact != nil:
				model.Compact.LastResult = result.Compact
				model.Compact.InlineError = ""
				if result.Compact.ReembeddedChunks == 0 && result.Compact.UpdatedChunks == 0 {
					model.Compact.DryRunComplete = true
				}
			}
			return model
		})
		if result.RefreshStatus {
			startStatus()
		}
	}, runner)

	tui.UseChannel(runner.Events(), func(event falkenvector.Event) {
		line := app.EventLineFromSDK(event)
		updateModel(func(model app.AppModel) app.AppModel {
			return app.WithEventLine(model, line)
		})
	}, runner)

	cbs := callbacks{
		setScreen: func(screen app.Screen) {
			updateModel(func(model app.AppModel) app.AppModel {
				model.ActiveScreen = screen
				return model
			})
		},
		update:           updateModel,
		updatePersistent: updatePersistent,
		refreshStatus:    startStatus,
		startDryRun: func() {
			if err := app.StartIngest(runner, factory, model.Index, true); err != nil {
				updateModel(func(model app.AppModel) app.AppModel {
					model.Index.InlineError = err.Error()
					return model
				})
				return
			}
			setBusy(app.OperationIngest)
		},
		startIngest: func() {
			if err := app.StartIngest(runner, factory, model.Index, false); err != nil {
				updateModel(func(model app.AppModel) app.AppModel {
					model.Index.InlineError = err.Error()
					return model
				})
				return
			}
			setBusy(app.OperationIngest)
		},
		startSearch: func() {
			if err := app.StartSearch(runner, factory, model.Search); err != nil {
				updateModel(func(model app.AppModel) app.AppModel {
					model.Search.InlineError = err.Error()
					return model
				})
				return
			}
			setBusy(app.OperationSearch)
		},
		startAsk: func() {
			if err := app.StartAsk(runner, factory, directory, model.Ask); err != nil {
				updateModel(func(model app.AppModel) app.AppModel {
					model.Ask.InlineError = err.Error()
					return model
				})
				return
			}
			setBusy(app.OperationAsk)
		},
		startCompactDry: func() {
			if err := app.StartCompact(runner, factory, model.Compact, true); err != nil {
				updateModel(func(model app.AppModel) app.AppModel {
					model.Compact.InlineError = err.Error()
					return model
				})
				return
			}
			setBusy(app.OperationCompact)
		},
		startCompact: func() {
			if !model.Compact.Confirming {
				updateModel(func(model app.AppModel) app.AppModel {
					model.Compact.Confirming = true
					return model
				})
				return
			}
			if err := app.StartCompact(runner, factory, model.Compact, false); err != nil {
				updateModel(func(model app.AppModel) app.AppModel {
					model.Compact.InlineError = err.Error()
					return model
				})
				return
			}
			setBusy(app.OperationCompact)
		},
		copySelection: func(text string) {
			if text == "" {
				return
			}
			if err := app.CopyToClipboard(text); err != nil {
				updateModel(func(model app.AppModel) app.AppModel {
					model.LastError = "copy failed: " + err.Error()
					return model
				})
				return
			}
			updateModel(func(model app.AppModel) app.AppModel {
				model.LastError = ""
				return app.WithEventLine(model, app.EventLine{At: time.Now(), Text: "copied selection"})
			})
		},
		selectAskSource: func(index int) {
			updateModel(func(model app.AppModel) app.AppModel {
				if index >= 0 && index < len(model.Ask.SourceItems) {
					model.Ask.SelectedSource = index
				}
				return model
			})
		},
		openAskSource: func(index int) {
			if index < 0 || index >= len(model.Ask.SourceItems) {
				return
			}
			source := model.Ask.SourceItems[index]
			text, err := app.LoadAskSourcePreview(model.Directory, source)
			updateModel(func(model app.AppModel) app.AppModel {
				model.Ask.SelectedSource = index
				model.Ask.SourceDialogOpen = true
				model.Ask.SourceDialogTitle = source.Label
				model.Ask.SourceDialogText = text
				model.Ask.SourceDialogError = ""
				if err != nil {
					model.Ask.SourceDialogError = err.Error()
				}
				return model
			})
		},
		closeAskSource: func() {
			updateModel(func(model app.AppModel) app.AppModel {
				model.Ask.SourceDialogOpen = false
				model.Ask.SourceDialogTitle = ""
				model.Ask.SourceDialogText = ""
				model.Ask.SourceDialogError = ""
				return model
			})
		},
		cancel: func() {
			if runner.Cancel() {
				updateModel(func(model app.AppModel) app.AppModel {
					model.LastError = "cancel requested"
					return model
				})
			}
		},
	}

	return rootView(model, cbs, tuiApp.Quit)
}

func rootView(model app.AppModel, cbs callbacks, quit func()) tui.Node {
	return tui.ViewWith(
		tui.ViewProps{
			Style: tui.Style{Direction: tui.Column, FlexGrow: 1},
			OnKey: func(ev tui.KeyEvent) bool {
				if ev.Key == tui.KeyCtrlC {
					if model.Busy {
						cbs.cancel()
					} else {
						quit()
					}
					return true
				}
				if ev.Key == tui.KeyEsc {
					if model.Busy {
						cbs.cancel()
					} else {
						cbs.setScreen(app.ScreenAsk)
					}
					return true
				}
				if ev.Key != tui.KeyRune {
					return false
				}
				switch ev.Rune {
				case 'q':
					if !model.Busy {
						quit()
					}
					return true
				case '1':
					cbs.setScreen(app.ScreenAsk)
					return true
				case '2':
					cbs.setScreen(app.ScreenStatus)
					return true
				case '3':
					cbs.setScreen(app.ScreenIndex)
					return true
				case '4':
					cbs.setScreen(app.ScreenSearch)
					return true
				case '5':
					cbs.setScreen(app.ScreenCompact)
					return true
				}
				return false
			},
			Focusable: true,
			AutoFocus: true,
		},
		header(model),
		nav(model, cbs),
		content(model, cbs),
		eventLog(model),
	)
}

func content(model app.AppModel, cbs callbacks) tui.Node {
	switch model.ActiveScreen {
	case app.ScreenIndex:
		return indexScreen(model, cbs)
	case app.ScreenSearch:
		return searchScreen(model, cbs)
	case app.ScreenAsk:
		return askScreen(model, cbs)
	case app.ScreenCompact:
		return compactScreen(model, cbs)
	default:
		return statusScreen(model, cbs)
	}
}

func section(children ...tui.Node) tui.Node {
	return tui.View(tui.Style{FlexGrow: 1, Padding: tui.All(1), Gap: 1, Direction: tui.Column}, children...)
}

func button(label string, disabled bool, onPress func()) tui.Node {
	return tui.Button(tui.ButtonProps{
		Label:    label,
		Disabled: disabled,
		OnPress:  onPress,
		Style: tui.Style{
			Border: tui.BorderAll,
			Height: tui.Cells(3),
			Width:  tui.Cells(len(label) + 4),
		},
		FocusedStyle: tui.Style{Bold: true, Foreground: tui.ANSIColor(3)},
	})
}

func input(value string, placeholder string, onChange func(string)) tui.Node {
	return tui.TextInput(tui.TextInputProps{
		Value:        value,
		Placeholder:  placeholder,
		OnChange:     onChange,
		Style:        tui.Style{Border: tui.BorderBottom, Width: tui.Cells(24)},
		FocusedStyle: tui.Style{Foreground: tui.ANSIColor(3)},
	})
}

func errorText(value string) tui.Node {
	if value == "" {
		return tui.Text("")
	}
	return tui.Text(value, tui.WithTextStyle(tui.Style{Foreground: tui.ANSIColor(1)}))
}

func busyLabel(model app.AppModel) string {
	if !model.Busy {
		return ""
	}
	return fmt.Sprintf("running %s", model.BusyKind)
}
