package calendar_test

import (
	"context"
	"rooster-importer/pkg/calendar"
	"testing"
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
