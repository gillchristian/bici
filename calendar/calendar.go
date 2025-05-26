package calendar

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	ics "github.com/arran4/golang-ical"
)

// Add this constant to control all-day event exclusion
const ExcludeAllDayEvents = true

// ImportEvents reads a calendar file and returns a list of events
func ImportEvents(path string) ([]CalendarEvent, error) {
	// Open the calendar file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening calendar file: %w", err)
	}
	defer file.Close()

	// Parse the calendar
	cal, err := ics.ParseCalendar(file)
	if err != nil {
		return nil, fmt.Errorf("error parsing calendar file: %w", err)
	}

	// Get the current week's start and end
	now := time.Now()
	weekStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	for weekStart.Weekday() != time.Sunday {
		weekStart = weekStart.AddDate(0, 0, -1)
	}
	weekEnd := weekStart.AddDate(0, 0, 7)

	fmt.Printf("[ImportEvents] Now: %v\n", now)
	fmt.Printf("[ImportEvents] weekStart (Sunday): %v\n", weekStart)
	fmt.Printf("[ImportEvents] weekEnd (next Sunday): %v\n", weekEnd)

	var events []CalendarEvent
	for _, event := range cal.Events() {
		parsedEvents, err := parseEvent(event, weekStart, weekEnd)
		if err != nil {
			return nil, fmt.Errorf("error parsing event: %w", err)
		}
		events = append(events, parsedEvents...)
	}

	fmt.Printf("[ImportEvents] Total events included: %d\n", len(events))
	return events, nil
}

// ExportEvents exports a list of events to a calendar file
func ExportEvents(events []CalendarEvent) error {
	// Create a new calendar
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)

	// Add each event to the calendar
	for _, event := range events {
		// Parse the start and end times
		startTime, err := time.Parse(time.RFC3339, event.Time.Start)
		if err != nil {
			return fmt.Errorf("error parsing start time: %w", err)
		}
		endTime, err := time.Parse(time.RFC3339, event.Time.End)
		if err != nil {
			return fmt.Errorf("error parsing end time: %w", err)
		}

		// Create a new event
		calEvent := cal.AddEvent(event.Id)
		calEvent.SetCreatedTime(time.Now())
		calEvent.SetDtStampTime(time.Now())
		calEvent.SetStartAt(startTime)
		calEvent.SetEndAt(endTime)
		calEvent.SetSummary(event.Title)
		calEvent.SetDescription(event.Description)
	}

	// Get the user's Downloads directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}
	downloadsDir := filepath.Join(homeDir, "Downloads")

	// Create the calendar file
	calFile := filepath.Join(downloadsDir, "calendar.ics")
	f, err := os.Create(calFile)
	if err != nil {
		return fmt.Errorf("error creating calendar file: %w", err)
	}
	defer f.Close()

	// Write the calendar to the file
	if err := cal.SerializeTo(f); err != nil {
		return fmt.Errorf("error writing calendar file: %w", err)
	}

	return nil
}
