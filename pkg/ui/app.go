package ui

import (
	"io"
	"rooster-importer/pkg/domain"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type AppUI struct {
	mainWindow  fyne.Window
	uploadLabel *widget.Label

	events chan domain.Action
}

type XlsxHandler interface {
	HandleSelectedFile(io.ReadCloser)
}

const NO_FILE_SELECTED = "(geen bestand geselecteerd)"

func CreateAppUI() *AppUI {
	ui := &AppUI{}
	ui.events = make(chan domain.Action, 4)

	a := app.New()
	ui.mainWindow = a.NewWindow("Fix je rooster naar Google Calendar")

	explainerLabel := widget.NewLabel("Upload je rooster.xlsx hier, en dit ding vult je Google Calendar in")

	uploadBox := ui.createUploadBox()

	ui.mainWindow.SetContent(container.NewVBox(
		explainerLabel,
		uploadBox,
	))

	return ui
}

func (u *AppUI) createUploadBox() *fyne.Container {
	button := widget.NewButton("Selecteer rooster", u.clickUploadButton)
	u.uploadLabel = widget.NewLabel(NO_FILE_SELECTED)
	box := container.NewHBox(button, u.uploadLabel)

	return box
}

func (u *AppUI) Events() <-chan domain.Action {
	return u.events
}

func (u *AppUI) clickUploadButton() {
	fileOpen := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if uc != nil {
			u.uploadLabel.SetText(uc.URI().Path())
			u.events <- domain.SelectedXlsxFileAction(uc)
		} else {
			u.uploadLabel.SetText(NO_FILE_SELECTED)
		}
	}, u.mainWindow)

	fileOpen.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
	fileOpen.Show()
}

func (u *AppUI) ShowAndRun() {
	u.mainWindow.ShowAndRun()
	close(u.events)
}
