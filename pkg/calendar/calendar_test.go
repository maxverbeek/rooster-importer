package calendar_test

import (
	"context"
	"rooster-importer/pkg/calendar"
	"testing"
	"time"
)

func TestLogout(t *testing.T) {
	calendar.LogOut()
}

func TestLogin(t *testing.T) {
	_, err := calendar.LogIn()

	if err != nil {
		t.Fatal(err)
	}

}

func TestList(t *testing.T) {
	client, err := calendar.LogIn()

	if err != nil {
		t.Fatal(err)
	}

	list, err := client.ListCalendars(context.TODO())

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("calendar IDs: %+v", list)
}

func TestListEvents(t *testing.T) {
	client, err := calendar.LogIn()

	if err != nil {
		t.Fatal(err)
	}

	calendars, err := client.ListCalendars(context.TODO())

	if err != nil {
		t.Fatal(err)
	}

	var id string

	for _, cal := range calendars {
		if cal.Name == "Nereas Werk" {
			id = cal.Id
		}
	}

	if id == "" {
		t.Fatal("Nereas werk not found")
	}

	events, err := client.ListEvents(context.TODO(), id)

	if err != nil {
		t.Fatal(err)
	}

	for _, e := range events {
		t.Logf("% 15s: (allday=%t) start: %s, end: %s", e.Title, e.AllDay, e.Start.Format(time.RFC3339), e.End.Format(time.RFC3339))
	}
}
