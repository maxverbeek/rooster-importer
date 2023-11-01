package main

import (
	"rooster-importer/pkg/domain"
	"rooster-importer/pkg/ui"
)

func main() {
	application := domain.NewApplication()

	gui := ui.CreateAppUI()

	go gui.SubscribeToApp(application.GuiStuff())
	go application.SubscribeToGui(gui.Events())

	gui.ShowAndRun()
}
