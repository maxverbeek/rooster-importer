package ui

import (
	"fmt"
	"rooster-importer/pkg/domain"

	"fyne.io/fyne/v2/dialog"
)

func (ui *AppUI) SubscribeToApp(events <-chan interface{}) {
	for event := range events {
		switch e := event.(type) {
		case error:
			dialog.ShowError(e, ui.mainWindow)

		case domain.Information:
			dialog.ShowInformation("Information", string(e), ui.mainWindow)

		default:
			panic(fmt.Sprintf("unexpected event to UI (type %T)", e))
		}
	}
}
