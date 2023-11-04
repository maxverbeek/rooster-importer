package domain

import (
	"context"
	"errors"
	"fmt"
	"io"
	"rooster-importer/pkg/calendar"
	"rooster-importer/pkg/excelreader"
	"time"
)

type Action func(*Application)

func SelectedXlsxFileAction(file io.ReadCloser, filename string, username string) Action {
	return func(a *Application) {
		a.xlsxfile = file
		a.uistate.SelectedXlsxFile = filename

		a.uistate.ConvertedEvents = []*ScheduleEvent{}
		a.uistate.WarningEvents = []*ScheduleEvent{}
		a.uistate.FreeDays = []time.Time{}

		a.guistuff <- NewState(a.uistate)

		if username == "" {
			a.guistuff <- errors.New("vul eerst een naam in")
			return
		}

		entries, err := excelreader.FindScheduleEntries(file, username)

		if err != nil {
			a.guistuff <- err
			return
		}

		events := []*ScheduleEvent{}
		free := []time.Time{}
		warnings := []*ScheduleEvent{}

		for _, entry := range entries {
			event, conversion := NewScheduleEvent(entry.Shift, entry.Date)

			if conversion == ConversionVrij {
				free = append(free, entry.Date)
			} else {
				events = append(events, event)
				if conversion == ConversionDefaulted {
					warnings = append(warnings, event)
				}
			}
		}

		a.eventsForCalendar = events
		a.uistate.ConvertedEvents = events
		a.uistate.WarningEvents = warnings
		a.uistate.FreeDays = free

		a.guistuff <- NewState(a.uistate)
	}
}

func SelectCalendarAction(calendarName string) Action {
	return func(a *Application) {
		a.selectedCalendarName = calendarName
		a.uistate.SelectedCalendarName = calendarName
		a.guistuff <- NewState(a.uistate)
	}
}

func ClickedCalendarLoginAction() Action {
	return func(a *Application) {
		fmt.Println("Login Action")

		calendars, err := listCalendars()

		if err != nil {
			a.guistuff <- err
		} else {
			a.uistate.IsLoggedIn = true
			a.uistate.AvailableCalendars = make([]string, len(calendars))

			fmt.Println(calendars)

			for i, cal := range calendars {
				a.uistate.AvailableCalendars[i] = cal.Name
			}
			a.guistuff <- NewState(a.uistate)
		}
	}
}

func ClickedCalendarLogoutAction() Action {
	return func(a *Application) {
		err := calendar.LogOut()

		if err != nil {
			a.guistuff <- err
		}

		a.uistate.IsLoggedIn = false
		a.uistate.AvailableCalendars = []string{}
		a.guistuff <- NewState(a.uistate)
	}
}

func listCalendars() ([]calendar.CalendarItem, error) {
	client, err := calendar.LogIn()

	if err != nil {
		return nil, fmt.Errorf("cannot login to google calendar: %w", err)
	}

	calendars, err := client.ListCalendars(context.TODO())

	if err != nil {
		return nil, fmt.Errorf("cannot list calendars: %w", err)
	}

	return calendars, nil
}

func GuiAttachedAction() Action {
	return func(a *Application) {
		// When the GUI attaches, determine if user is logged into google cal
		a.uistate.IsLoggedIn = calendar.IsLoggedIn()

		// If a user is already logged in, fetch calendars and show those too
		if a.uistate.IsLoggedIn {
			calendars, err := listCalendars()

			if err != nil {
				a.guistuff <- err
			} else {
				a.uistate.AvailableCalendars = make([]string, len(calendars))

				for i, cal := range calendars {
					a.uistate.AvailableCalendars[i] = cal.Name
				}
			}
		}
		a.guistuff <- NewState(a.uistate)
	}
}

func scheduleToCalendarEvent(sched *ScheduleEvent) calendar.CalendarEvent {
	return calendar.CalendarEvent{
		Title: sched.ScheduleType,
		Start: sched.Start,
		End:   sched.End,
	}
}

func ImportEntriesToCalendar() Action {
	return func(a *Application) {

		client, err := calendar.LogIn()

		if err != nil {
			a.guistuff <- fmt.Errorf("cannot log into google calendar: %w", err)
			return
		}

		ctx := context.Background()

		calendarId, err := client.FindCalendarIdByName(ctx, a.selectedCalendarName)

		if err != nil {
			a.guistuff <- fmt.Errorf("cannot find calendar ID for calendar %s: %w", a.selectedCalendarName, err)
			return
		}

		errors := []error{}

		for i, event := range a.eventsForCalendar {
			calEvent := scheduleToCalendarEvent(event)

			_, err := client.CreateEvent(ctx, calendarId, &calEvent)

			if err != nil {
				fmt.Printf("error in calendar event create: %s\n", err.Error())
				errors = append(errors, err)
			}

			a.guistuff <- Progress{Done: i + 1, Total: len(a.eventsForCalendar)}
		}

		if len(errors) != 0 {
			a.guistuff <- fmt.Errorf("%d errors occured: 1st error: %w", len(errors), errors[0])
			return
		}

		a.guistuff <- Information(fmt.Sprintf("Successfully imported %d events", len(a.eventsForCalendar)))
	}
}
