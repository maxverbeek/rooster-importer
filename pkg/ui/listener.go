package ui

import (
	"fmt"
	"rooster-importer/pkg/domain"
	"strings"

	"fyne.io/fyne/v2/dialog"
)

func (ui *AppUI) SubscribeToApp(events <-chan interface{}) {
	ui.events <- domain.GuiAttachedAction()

	for event := range events {
		switch e := event.(type) {
		case error:
			dialog.ShowError(e, ui.mainWindow)

		case domain.Information:
			dialog.ShowInformation("Information", string(e), ui.mainWindow)

		case domain.NewState:
			// update UI state from state
			state := domain.UIState(e)

			ui.uploadLabel.SetText(state.SelectedXlsxFile)

			if state.IsLoggedIn {
				ui.loginButton.Disable()
				ui.logoutButton.Enable()
			} else {
				ui.loginButton.Enable()
				ui.logoutButton.Disable()
			}

			if len(state.AvailableCalendars) > 0 {
				ui.calSelect.SetOptions(state.AvailableCalendars)
				ui.calSelect.Enable()
			} else {
				ui.calSelect.Disable()
			}

			// Build text block for event summary
			convertedCount := len(state.ConvertedEvents)
			warningCount := len(state.WarningEvents)
			freeDayCount := len(state.FreeDays)
			skippedCount := len(state.SkippedDays)
			newEventCount := len(state.EventsNotAlreadyInCalendar)

			previewlines := strings.Builder{}

			if convertedCount+freeDayCount > 0 {
				previewlines.WriteString(fmt.Sprintf("Schedule conversion summary:\nNew Schedule events: %d", convertedCount))

				if warningCount > 0 {
					previewlines.WriteString(fmt.Sprintf(" (%d not sure of time)", warningCount))
				}

				previewlines.WriteString(fmt.Sprintf("\nEvents free: %d\nTotal things processed: %d\n", freeDayCount, convertedCount+freeDayCount))
			}

			// Build a preview of the first and last events to be added to the calendar
			if convertedCount > 0 {
				previewlines.WriteString("\nFirst event: ")
				previewlines.WriteString(state.ConvertedEvents[0].Summary())
			}

			if convertedCount > 1 {
				previewlines.WriteString("\nLast event: ")
				previewlines.WriteString(state.ConvertedEvents[convertedCount-1].Summary())
			}

			previewlines.WriteString("\n")

			previewlines.WriteString(fmt.Sprintf("\nNew events: %d\n", newEventCount))

			if newEventCount > 0 {
				previewlines.WriteString(fmt.Sprintf("First event: %s\n", state.EventsNotAlreadyInCalendar[0].Summary()))

				if newEventCount > 1 {
					previewlines.WriteString(fmt.Sprintf("Last event: %s\n", state.EventsNotAlreadyInCalendar[newEventCount-1].Summary()))
				}

				for _, event := range state.EventsNotAlreadyInCalendar {
					previewlines.WriteString(fmt.Sprintf("%s\n", event.Summary()))
				}
			}

			if warningCount > 0 {
				previewlines.WriteString("\nEvents where time is not explicit:\n")

				for _, warning := range state.WarningEvents {
					previewlines.WriteString(warning.Summary())
					previewlines.WriteString("\n")
				}
			}

			if freeDayCount > 0 || skippedCount > 0 {
				previewlines.WriteString(fmt.Sprintf("\nFree dates (%d weekends not included):\n", skippedCount))

				for i, date := range state.FreeDays {
					previewlines.WriteString(fmt.Sprintf("%s  ", date.Format("02/01")))

					if i%7 == 6 {
						previewlines.WriteString("\n")
					}
				}
			}

			ui.preview.SetText(previewlines.String())

			if state.IsLoggedIn && len(state.EventsNotAlreadyInCalendar) > 0 && state.SelectedCalendarName != "" {
				ui.createEventsButton.SetText(fmt.Sprintf("Create %d events in %s", len(state.EventsNotAlreadyInCalendar), state.SelectedCalendarName))
				ui.createEventsButton.Enable()
			} else {
				ui.createEventsButton.SetText("Create Events")
				ui.createEventsButton.Disable()
			}

		case domain.Progress:
			if ui.progress.Hidden {
				ui.progress.Show()
			}

			ui.progress.SetValue(float64(e.Done) / float64(e.Total))

			if e.Done == e.Total {
				ui.createEventsButton.Enable()
				ui.progress.Hide()
			}

		default:
			panic(fmt.Sprintf("unexpected event to UI (type %T)", e))
		}
	}
}
