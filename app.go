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
	rrule "github.com/teambition/rrule-go"
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
	Id          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Time        EventTime `json:"time"`
	Weekday     Weekday   `json:"weekday"`
}

func (a *App) ImportEvents(path string) []CalendarEvent {
	fmt.Printf("Importing events from %s\n", path)

	// Read the file
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}
	defer file.Close()

	fmt.Printf("File opened\n")

	// Parse the calendar
	cal, err := ics.ParseCalendar(file)
	if err != nil {
		fmt.Printf("Error parsing calendar: %v\n", err)
		return nil
	}

	fmt.Printf("Calendar parsed\n")

	// Get current time and calculate the start of the current week (Sunday)
	now := time.Now()
	// Calculate days since last Sunday (0 if today is Sunday)
	daysSinceSunday := int(now.Weekday())
	weekStart := now.AddDate(0, 0, -daysSinceSunday)
	// Set time to start of day
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	weekEnd := weekStart.AddDate(0, 0, 7)

	fmt.Printf("Week start: %v\n", weekStart)
	fmt.Printf("Week end: %v\n", weekEnd)

	var events []CalendarEvent

	// Process each event
	for _, event := range cal.Events() {
		// Get event start time
		startTime, err := event.GetStartAt()
		if err != nil {
			fmt.Printf("Error getting event start time: %v\n", err)
			continue
		}

		// Skip events not in current week
		if startTime.Before(weekStart) || startTime.After(weekEnd) {
			continue
		}

		// Get event end time
		endTime, err := event.GetEndAt()
		if err != nil {
			fmt.Printf("Error getting event end time: %v\n", err)
			continue
		}

		fmt.Printf("Event start time: %v\n", startTime)
		fmt.Printf("Event end time: %v\n", endTime)

		// Get event title and description
		title := ""
		if prop := event.GetProperty(ics.ComponentPropertySummary); prop != nil {
			title = prop.Value
		}

		description := ""
		if prop := event.GetProperty(ics.ComponentPropertyDescription); prop != nil {
			description = prop.Value
		}

		fmt.Printf("Event title: %s\n", title)
		fmt.Printf("Event description: %s\n", description)

		// Calculate weekday (0 = Sunday, 6 = Saturday)
		weekday := Weekday(startTime.Weekday())

		fmt.Printf("Event weekday: %d\n", weekday)

		// Create CalendarEvent
		calendarEvent := CalendarEvent{
			Id:          event.Id(),
			Title:       title,
			Description: description,
			Time: EventTime{
				Start: startTime.Format(time.RFC3339),
				End:   endTime.Format(time.RFC3339),
			},
			Weekday: weekday,
		}

		fmt.Printf("CalendarEvent: %+v\n", calendarEvent)

		events = append(events, calendarEvent)
	}

	fmt.Printf("Imported %d events\n", len(events))

	return events
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
	daysSinceSunday := int(now.Weekday())
	weekStart := now.AddDate(0, 0, -daysSinceSunday)
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())

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
		return "SU"
	}
}
