package domain

import (
	"io"
	"time"
)

type Application struct {
	xlsxfile             io.ReadCloser
	selectedCalendarName string
	selectedCalendarId   string
	uistate              UIState
	eventsForCalendar    []*ScheduleEvent
	eventsInCalendar     []*ScheduleEvent
	newEventsForCalendar []*ScheduleEvent

	guistuff chan interface{}
}

type UIState struct {
	SelectedXlsxFile     string
	IsLoggedIn           bool
	SelectedCalendarName string
	AvailableCalendars   []string
	ImportButtonEnabled  bool

	ConvertedEvents            []*ScheduleEvent
	WarningEvents              []*ScheduleEvent
	EventsNotAlreadyInCalendar []*ScheduleEvent
	FreeDays                   []time.Time
	SkippedDays                []time.Time
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

func (a *Application) DeduplicateEvents() {
	if len(a.eventsForCalendar) == 0 {
		a.newEventsForCalendar = make([]*ScheduleEvent, 0)
		a.uistate.EventsNotAlreadyInCalendar = a.newEventsForCalendar

		a.guistuff <- NewState(a.uistate)
		return
	}

	if len(a.eventsInCalendar) == 0 {
		a.newEventsForCalendar = a.eventsForCalendar
		a.uistate.EventsNotAlreadyInCalendar = a.newEventsForCalendar

		a.guistuff <- NewState(a.uistate)
		return
	}

	existing := make(map[ScheduleEvent]bool)

	for _, incalEvent := range a.eventsInCalendar {
		existing[*incalEvent] = true
	}

	a.newEventsForCalendar = []*ScheduleEvent{}

	for _, newEvent := range a.eventsForCalendar {
		if _, exists := existing[*newEvent]; !exists {
			a.newEventsForCalendar = append(a.newEventsForCalendar, newEvent)
		}
	}

	a.uistate.EventsNotAlreadyInCalendar = a.newEventsForCalendar
}
