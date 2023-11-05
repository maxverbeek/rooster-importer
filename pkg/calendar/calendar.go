package calendar

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var (
	//go:embed credentials.json
	credentialContents []byte
)

type TokenReceivedMessage struct {
	Code  string
	Error error
}

type CalendarClient struct {
	client *http.Client
	srv    *calendar.Service
}

type CalendarItem struct {
	Id    string
	Name  string
	Color string
}

type CalendarEvent struct {
	Title  string
	Start  time.Time
	End    time.Time
	AllDay bool
}

func (c *CalendarClient) ListCalendars(ctx context.Context) ([]CalendarItem, error) {
	list, err := c.srv.CalendarList.List().Context(ctx).Do()

	if err != nil {
		return nil, fmt.Errorf("couldnt list calendars: %w", err)
	}

	items := []CalendarItem{}

	for _, item := range list.Items {
		items = append(items, CalendarItem{
			Id:    item.Id,
			Name:  item.Summary,
			Color: item.BackgroundColor,
		})
	}
	return items, nil
}

func (c *CalendarClient) ListEvents(ctx context.Context, calendarId string) ([]CalendarEvent, error) {
	googleEvents := []*calendar.Event{}
	err := c.srv.Events.List(calendarId).Context(ctx).Pages(ctx, func(e *calendar.Events) error {
		googleEvents = append(googleEvents, e.Items...)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("couldn't get calendar events: %w", err)
	}

	ourEvents := make([]CalendarEvent, len(googleEvents))

	for i, event := range googleEvents {
		calendarEvent := CalendarEvent{
			Title: event.Summary,
		}
		if event.Start.Date != "" && event.End.Date != "" {

			// All day event
			calendarEvent.AllDay = true

			locationStart, err := time.LoadLocation(event.Start.TimeZone)

			if err != nil {
				return nil, fmt.Errorf("cannot decode timezone: %w", err)
			}

			calendarEvent.Start, err = time.ParseInLocation(time.DateOnly, event.Start.Date, locationStart)

			if err != nil {
				return nil, fmt.Errorf("couldnt parse start time of event %d (%s): %w", i, event.Summary, err)
			}

			locationEnd, err := time.LoadLocation(event.End.TimeZone)

			if err != nil {
				return nil, fmt.Errorf("cannot decode timezone: %w", err)
			}

			calendarEvent.End, err = time.ParseInLocation(time.DateOnly, event.End.Date, locationEnd)

			if err != nil {
				return nil, fmt.Errorf("couldnt parse end time of event %d (%s): %w", i, event.Summary, err)
			}
		} else if event.Start.DateTime != "" && event.End.DateTime != "" {
			// Not all day
			calendarEvent.AllDay = false
			calendarEvent.Start, err = time.Parse(time.RFC3339, event.Start.DateTime)

			if err != nil {
				return nil, fmt.Errorf("couldnt parse start time of event %d (%s): %w", i, event.Summary, err)
			}

			calendarEvent.End, err = time.Parse(time.RFC3339, event.End.DateTime)

			if err != nil {
				return nil, fmt.Errorf("couldnt parse end time of event %d (%s): %w", i, event.Summary, err)
			}
		} else {
			return nil, fmt.Errorf("not sure what type of event %d: %s is", i, event.Summary)
		}

		ourEvents[i] = calendarEvent
	}

	return ourEvents, nil
}

func (c *CalendarClient) CreateEvent(ctx context.Context, calendarId string, event *CalendarEvent) (*calendar.Event, error) {

	googlecalendarevent := &calendar.Event{
		Summary: event.Title,
	}

	if event.AllDay {
		// For all day events, use the Date attribute
		googlecalendarevent.Start = &calendar.EventDateTime{Date: event.Start.Format(time.DateOnly), TimeZone: "Europe/Amsterdam"}
		googlecalendarevent.End = &calendar.EventDateTime{Date: event.End.Format(time.DateOnly), TimeZone: "Europe/Amsterdam"}
	} else {
		googlecalendarevent.Start = &calendar.EventDateTime{DateTime: event.Start.Format(time.RFC3339), TimeZone: "Europe/Amsterdam"}
		googlecalendarevent.End = &calendar.EventDateTime{DateTime: event.End.Format(time.RFC3339), TimeZone: "Europe/Amsterdam"}
	}

	gcalevent, err := c.srv.Events.Insert(calendarId, googlecalendarevent).Context(ctx).Do()

	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return gcalevent, nil
}

func (c *CalendarClient) FindCalendarIdByName(ctx context.Context, calendarName string) (string, error) {
	calendars, err := c.ListCalendars(ctx)

	if err != nil {
		return "", err
	}

	id := ""
	occurrences := 0

	for _, item := range calendars {
		if calendarName == item.Name {
			id = item.Id
			occurrences += 1
		}
	}

	if occurrences == 1 {
		return id, nil
	} else if occurrences == 0 {
		return "", errors.New(fmt.Sprintf("calendar name %s not found", calendarName))
	} else {
		return "", errors.New(fmt.Sprintf("calendar name %s is not unique (%d occurrences)", calendarName, occurrences))
	}
}

func StartWebserverForCallback(addr string, channel chan<- TokenReceivedMessage) {
	handler := http.NewServeMux()

	server := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		var msg TokenReceivedMessage

		if code == "" {
			msg = TokenReceivedMessage{
				Error: errors.New("wrong callback URL, did not have a code parameter"),
			}
		} else {
			msg = TokenReceivedMessage{
				Code: code,
			}
		}

		channel <- msg
		server.Shutdown(context.Background())
		fmt.Printf("Shutdown callback server")
	})

	server.ListenAndServe()

}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	channel := make(chan TokenReceivedMessage, 1)
	go StartWebserverForCallback("0.0.0.0:10321", channel)

	err := browser.OpenURL(authURL)

	if err != nil {
		return nil, err
	}

	msg := <-channel

	if msg.Error != nil {
		return nil, msg.Error
	}

	tok, err := config.Exchange(context.TODO(), msg.Code)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token from web: %w", err)
	}
	fmt.Println("received token")
	return tok, nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("cannot save token file: %s", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func tokenLocation() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	importerdir := fmt.Sprintf("%s/rooster-importer", dir)

	fi, err := os.Stat(importerdir)

	if err != nil {
		os.Mkdir(importerdir, 0775)
	} else if !fi.Mode().IsDir() {
		return "", errors.New(fmt.Sprintf("%s is not a directory", importerdir))
	}

	return fmt.Sprintf("%s/token.json", importerdir), nil
}

func LogOut() error {
	filepath, err := tokenLocation()

	if err != nil {
		return err
	}

	return os.Remove(filepath)
}

func IsLoggedIn() bool {
	path, _ := tokenLocation()
	_, err := os.Stat(path)

	return !os.IsNotExist(err)
}

func LogIn() (*CalendarClient, error) {
	tokenLocation, err := tokenLocation()

	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(credentialContents, calendar.CalendarEventsScope, calendar.CalendarReadonlyScope)

	if err != nil {
		return nil, err
	}

	tok, err := tokenFromFile(tokenLocation)

	if err != nil {
		tok, err = getTokenFromWeb(config)

		if err != nil {
			return nil, fmt.Errorf("failed to get token from web: %w", err)
		}
		saveToken(tokenLocation, tok)
	}

	client := config.Client(context.Background(), tok)

	srv, err := calendar.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve Calendar client: %w", err)
	}

	return &CalendarClient{
		client: client,
		srv:    srv,
	}, nil
}
