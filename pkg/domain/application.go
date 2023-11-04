package domain

import (
	"io"
	"time"
)

type Application struct {
	xlsxfile             io.ReadCloser
	selectedCalendarName string
	uistate              UIState
	eventsForCalendar    []*ScheduleEvent

	guistuff chan interface{}
}

type UIState struct {
	SelectedXlsxFile     string
	IsLoggedIn           bool
	SelectedCalendarName string
	AvailableCalendars   []string
	ImportButtonEnabled  bool

	ConvertedEvents []*ScheduleEvent
	WarningEvents   []*ScheduleEvent
	FreeDays        []time.Time
	SkippedDays     []time.Time
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
