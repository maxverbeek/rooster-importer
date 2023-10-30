package calendar

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
}

func (c *CalendarClient) ListCalendars(ctx context.Context) ([]string, error) {
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(c.client))
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve Calendar client: %w", err)
	}

	list, err := srv.CalendarList.List().Do()

	if err != nil {
		return nil, fmt.Errorf("couldnt list calendars: %w", err)
	}

	calendars := []string{}

	for _, item := range list.Items {
		calendars = append(calendars, item.Id)
	}

	return calendars, nil
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

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	channel := make(chan TokenReceivedMessage, 1)
	go StartWebserverForCallback("0.0.0.0:10321", channel)

	msg := <-channel

	if msg.Error != nil {
		log.Fatal(msg.Error)
	}

	tok, err := config.Exchange(context.TODO(), msg.Code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
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

func LogIn() (*CalendarClient, error) {
	tokenLocation, err := tokenLocation()

	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(credentialContents, calendar.CalendarEventsScope)

	if err != nil {
		return nil, err
	}

	tok, err := tokenFromFile(tokenLocation)

	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenLocation, tok)
	}

	client := config.Client(context.Background(), tok)

	return &CalendarClient{
		client: client,
	}, nil
}

func main() {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}
	fmt.Println("Upcoming events:")
	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("%v (%v)\n", item.Summary, date)
		}
	}
}
