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

// Add this constant to control all-day event exclusion
const ExcludeAllDayEvents = true

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
		// Exclude all-day events if the constant is set
		if ExcludeAllDayEvents {
			dtstartProp := event.GetProperty(ics.ComponentPropertyDtStart)
			if dtstartProp != nil && dtstartProp.GetValueType() == ics.ValueDataTypeDate {
				fmt.Printf("Skipping all-day event: %s (UID: %s)\n", event.GetProperty(ics.ComponentPropertySummary).Value, event.Id())
				continue
			}
		}

		// Get event start time
		startTime, err := event.GetStartAt()
		if err != nil {
			fmt.Printf("Error getting event start time: %v\n", err)
			continue
		}

		// Get event end time
		endTime, err := event.GetEndAt()
		if err != nil {
			fmt.Printf("Error getting event end time: %v\n", err)
			continue
		}

		// Get event title and description
		title := ""
		if prop := event.GetProperty(ics.ComponentPropertySummary); prop != nil {
			title = prop.Value
		}

		description := ""
		if prop := event.GetProperty(ics.ComponentPropertyDescription); prop != nil {
			description = prop.Value
		}

		// Check for RRULE (recurrence rule)
		rruleProp := event.GetProperty("RRULE")
		if rruleProp != nil {
			// Parse RRULE using rrule-go, but anchor it to the event's DTSTART
			rruleStr := rruleProp.Value
			opt, err := rrule.StrToROption(rruleStr)
			if err != nil {
				fmt.Printf("Error parsing RRULE: %v\n", err)
				continue
			}
			opt.Dtstart = startTime
			rr, err := rrule.NewRRule(*opt)
			if err != nil {
				fmt.Printf("Error creating RRule: %v\n", err)
				continue
			}

			// Handle EXDATEs (dates to exclude)
			exdates := map[time.Time]bool{}
			exdateProps := event.GetProperties("EXDATE")
			fmt.Printf("  EXDATEs for event '%s' (UID: %s):\n", title, event.Id())
			for _, ex := range exdateProps {
				if ex != nil {
					// Try parsing as RFC3339, then as 20060102 (DATE only), then as 20060102T150405 (iCalendar local time)
					exdate, err := time.Parse(time.RFC3339, ex.Value)
					if err != nil {
						exdate, err = time.Parse("20060102", ex.Value)
					}
					if err != nil {
						exdate, err = time.Parse("20060102T150405", ex.Value)
					}
					if err == nil {
						exdates[exdate] = true
						fmt.Printf("    Parsed EXDATE: %v\n", exdate)
					} else {
						fmt.Printf("    Failed to parse EXDATE: %s\n", ex.Value)
					}
				}
			}

			// Get all occurrences in this week
			occurrences := rr.Between(weekStart, weekEnd, false)
			fmt.Printf("  Occurrences returned by rr.Between: %d\n", len(occurrences))
			for i, occ := range occurrences {
				fmt.Printf("    Occurrence %d: %v\n", i, occ)
			}
			for _, occ := range occurrences {
				// Ensure occ is within weekStart (inclusive) and weekEnd (exclusive)
				if occ.Before(weekStart) || !occ.Before(weekEnd) {
					fmt.Printf("    Skipping occurrence %v: outside week range\n", occ)
					continue
				}
				// Skip if in EXDATE
				skip := false
				for ex := range exdates {
					if occ.Year() == ex.Year() &&
						occ.Month() == ex.Month() &&
						occ.Day() == ex.Day() &&
						occ.Hour() == ex.Hour() &&
						occ.Minute() == ex.Minute() &&
						occ.Second() == ex.Second() {
						skip = true
						fmt.Printf("    Skipping occurrence %v: in EXDATE (event '%s', UID: %s)\n", occ, title, event.Id())
						break
					}
				}
				if skip {
					continue
				}
				fmt.Printf("    Including occurrence %v for event '%s' (UID: %s)\n", occ, title, event.Id())
				// Calculate end time for this occurrence
				occEnd := occ.Add(endTime.Sub(startTime))
				weekday := Weekday(occ.Weekday())
				calendarEvent := CalendarEvent{
					Id:          event.Id(),
					Title:       title,
					Description: description,
					Time: EventTime{
						Start: occ.Format(time.RFC3339),
						End:   occEnd.Format(time.RFC3339),
					},
					Weekday: weekday,
				}
				events = append(events, calendarEvent)
			}
			continue // skip normal single-instance logic
		}

		// Non-recurring event: only add if in this week
		if startTime.Before(weekStart) || startTime.After(weekEnd) {
			continue
		}

		weekday := Weekday(startTime.Weekday())
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
