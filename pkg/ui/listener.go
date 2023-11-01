package ui

import (
	"fmt"
	"rooster-importer/pkg/domain"

	"fyne.io/fyne/v2/dialog"
)

func (ui *AppUI) SubscribeToApp(events <-chan interface{}) {
	ui.events <- domain.GuiAttachedAction()

	for event := range events {
		switch e := event.(type) {
		case error:
			dialog.ShowError(e, ui.mainWindow)

		case domain.Information:
			dialog.ShowInformation("Information", string(e), ui.mainWindow)

		case domain.NewState:
			// update UI state from state
			state := domain.UIState(e)

			ui.uploadLabel.SetText(state.SelectedXlsxFile)

			if state.IsLoggedIn {
				ui.loginButton.Disable()
				ui.logoutButton.Enable()
			} else {
				ui.loginButton.Enable()
				ui.logoutButton.Disable()
			}

			if len(state.AvailableCalendars) > 0 {
				ui.calSelect.SetOptions(state.AvailableCalendars)
				ui.calSelect.Enable()
			} else {
				ui.calSelect.Disable()
			}

		default:
			panic(fmt.Sprintf("unexpected event to UI (type %T)", e))
		}
	}
}
