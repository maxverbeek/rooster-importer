package domain

import (
	"io"
)

type Application struct {
	xlsxfile           io.ReadCloser
	selectedCalendarId string
	uistate            UIState

	guistuff chan interface{}
}

type UIState struct {
	SelectedXlsxFile     string
	IsLoggedIn           bool
	SelectedCalendarName string
	AvailableCalendars   []string
	ImportButtonEnabled  bool
}

func NewApplication() *Application {
	return &Application{
		guistuff: make(chan interface{}),
	}
}

func (a *Application) SubscribeToGui(events <-chan Action) {
	for action := range events {
		action(a)
	}
}

func (a *Application) GuiStuff() <-chan interface{} {
	return a.guistuff
}
