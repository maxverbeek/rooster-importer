package domain

import (
	"context"
	"fmt"
	"io"
	"rooster-importer/pkg/calendar"
	"rooster-importer/pkg/excelreader"
)

type Action func(*Application)

func SelectedXlsxFileAction(file io.ReadCloser, name string) Action {
	return func(a *Application) {
		a.xlsxfile = file
		a.uistate.SelectedXlsxFile = name
		_, err := excelreader.HandleSelectedFile(file, "Nerea")

		if err != nil {
			a.guistuff <- err
		}
	}
}

func SelectCalendarAction(calendarId string) Action {
	return func(a *Application) {
		a.selectedCalendarId = calendarId
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
