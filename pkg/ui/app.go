package ui

import (
	"io"
	"rooster-importer/pkg/domain"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type AppUI struct {
	mainWindow  fyne.Window
	uploadLabel *widget.Label
	nameEntry   *widget.Entry
	calSelect   *widget.Select

	loginButton  *widget.Button
	logoutButton *widget.Button

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

	explainerLabel := widget.NewLabel("Vul je naam (1e kolom van je Excel bestand) in, upload je rooster.xlsx hier, en dit ding vult je Google Calendar in")

	uploadBox := ui.createUploadBox()
	googleCalendarBox := ui.createGoogleCalendarBox()

	ui.mainWindow.SetContent(container.NewVBox(
		explainerLabel,
		uploadBox,
		googleCalendarBox,
	))

	return ui
}

func (u *AppUI) createUploadBox() *fyne.Container {
	button := widget.NewButton("Selecteer rooster", u.clickUploadButton)
	u.uploadLabel = widget.NewLabel(NO_FILE_SELECTED)
	u.nameEntry = widget.NewEntry()
	uploader := container.NewHBox(button, u.uploadLabel)

	namelabel := widget.NewLabel("Naam")
	nameform := container.New(layout.NewFormLayout(), namelabel, u.nameEntry)

	uploadBox := container.NewVBox(nameform, uploader)

	return container.NewPadded(uploadBox)
}

func (u *AppUI) createGoogleCalendarBox() *fyne.Container {
	label := widget.NewLabel("Google Calendar stuff")
	u.loginButton = widget.NewButton("Log in", func() {
		u.events <- domain.ClickedCalendarLoginAction()
	})
	u.loginButton.Disable()

	u.logoutButton = widget.NewButton("Log out", func() {
		u.events <- domain.ClickedCalendarLogoutAction()
	})
	u.logoutButton.Disable()

	buttonBox := container.NewHBox(u.loginButton, u.logoutButton)

	u.calSelect = widget.NewSelect([]string{}, func(s string) {
		u.events <- domain.SelectCalendarAction(s)
	})

	u.calSelect.Disable()

	return container.NewPadded(container.NewVBox(label, buttonBox, u.calSelect))
}

func (u *AppUI) Events() <-chan domain.Action {
	return u.events
}

func (u *AppUI) clickUploadButton() {
	fileOpen := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if uc != nil {
			u.events <- domain.SelectedXlsxFileAction(uc, uc.URI().Path())
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
