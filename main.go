package main

import (
	"rooster-importer/pkg/domain"
	"rooster-importer/pkg/excelreader"
	"rooster-importer/pkg/ui"
)

func main() {
	application := domain.NewApplication()
	application.SetRooserReader(&excelreader.RoosterReader{})

	gui := ui.CreateAppUI()

	go gui.SubscribeToApp(application.GuiStuff())
	go application.SubscribeToGui(gui.Events())

	gui.ShowAndRun()
}
