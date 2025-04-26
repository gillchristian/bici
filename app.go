package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/google/uuid"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type EventTime struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

type CalendarEvent struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Time        EventTime `json:"time"`
	Weekday     Weekday   `json:"weekday"`
}

func (a *App) ExportEvents(events []CalendarEvent) (bool, string) {
	// return false, "Not implemented"

	// Create file in Downloads directory
	filePath := "~/Downloads/calendar.ics"

	// Expand the tilde to the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		msg := fmt.Sprintf("Error getting home directory: %v\n", err)
		fmt.Print(msg)
		return false, msg
	}

	if strings.HasPrefix(filePath, "~/") {
		filePath = filepath.Join(homeDir, filePath[2:])
	}

	// Create a new calendar
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)

	// Get current time and calculate the start of the current week (Sunday)
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))

	// Add each event to the calendar
	for _, event := range events {
		fmt.Printf("Processing event: Title=%s, Description=%s, Time=%+v, Weekday=%d\n",
			event.Title, event.Description, event.Time, event.Weekday)

		// Create a unique ID for the event using UUID
		eventID := fmt.Sprintf("%s@bici", uuid.New().String())

		// Parse the start and end times from ISO strings
		startTime, err := time.Parse(time.RFC3339, event.Time.Start)
		if err != nil {
			msg := fmt.Sprintf("Error parsing start time: %v\n", err)
			fmt.Print(msg)
			return false, msg
		}

		endTime, err := time.Parse(time.RFC3339, event.Time.End)
		if err != nil {
			msg := fmt.Sprintf("Error parsing end time: %v\n", err)
			fmt.Print(msg)
			return false, msg
		}

		// Calculate the event date by adding the weekday offset to the week start
		eventDate := weekStart.AddDate(0, 0, int(event.Weekday))

		// Combine the date with the time
		eventStart := time.Date(
			eventDate.Year(),
			eventDate.Month(),
			eventDate.Day(),
			startTime.Hour(),
			startTime.Minute(),
			0, 0, time.Local,
		)

		eventEnd := time.Date(
			eventDate.Year(),
			eventDate.Month(),
			eventDate.Day(),
			endTime.Hour(),
			endTime.Minute(),
			0, 0, time.Local,
		)

		// Create the event
		calEvent := cal.AddEvent(eventID)
		calEvent.SetCreatedTime(time.Now())
		calEvent.SetDtStampTime(time.Now())
		calEvent.SetModifiedAt(time.Now())
		calEvent.SetStartAt(eventStart)
		calEvent.SetEndAt(eventEnd)
		calEvent.SetSummary(event.Title)
		calEvent.SetDescription(event.Description)
	}

	// Create or truncate the file
	file, err := os.Create(filePath)
	if err != nil {
		msg := fmt.Sprintf("Error creating file: %v\n", err)
		fmt.Print(msg)
		return false, msg
	}
	defer file.Close()

	// Write the calendar to the file
	err = cal.SerializeTo(file)
	if err != nil {
		msg := fmt.Sprintf("Error writing calendar to file: %v\n", err)
		fmt.Print(msg)
		return false, msg
	}

	return true, ""
}

// Add String method to Weekday type
func (w Weekday) String() string {
	switch w {
	case Sunday:
		return "SU"
	case Monday:
		return "MO"
	case Tuesday:
		return "TU"
	case Wednesday:
		return "WE"
	case Thursday:
		return "TH"
	case Friday:
		return "FR"
	case Saturday:
		return "SA"
	default:
		return "MO"
	}
}
