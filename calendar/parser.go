package calendar

import (
	"fmt"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	rrule "github.com/teambition/rrule-go"
)

// parseEvent parses a single calendar event and returns a CalendarEvent
func parseEvent(event *ics.VEvent, weekStart, weekEnd time.Time) ([]CalendarEvent, error) {
	// Exclude all-day events if the constant is set
	if ExcludeAllDayEvents {
		dtstartProp := event.GetProperty(ics.ComponentPropertyDtStart)
		if dtstartProp != nil && dtstartProp.GetValueType() == ics.ValueDataTypeDate {
			fmt.Printf("Skipping all-day event: %s (UID: %s)\n", event.GetProperty(ics.ComponentPropertySummary).Value, event.Id())
			return nil, nil
		}
	}

	// Get event start time
	startTime, err := event.GetStartAt()
	if err != nil {
		return nil, fmt.Errorf("error getting event start time: %w", err)
	}

	// Get event end time
	endTime, err := event.GetEndAt()
	if err != nil {
		return nil, fmt.Errorf("error getting event end time: %w", err)
	}

	// Get event title and description
	title := ""
	if prop := event.GetProperty(ics.ComponentPropertySummary); prop != nil {
		title = prop.Value
		// Sanitize title
		title = strings.ReplaceAll(title, "/", "-")
		title = strings.ReplaceAll(title, "#", "-")
	}

	description := ""
	if prop := event.GetProperty(ics.ComponentPropertyDescription); prop != nil {
		description = prop.Value
		// Sanitize description
		description = strings.ReplaceAll(description, "/", "-")
		description = strings.ReplaceAll(description, "#", "-")
	}

	// Check for RRULE (recurrence rule)
	rruleProp := event.GetProperty("RRULE")
	if rruleProp != nil {
		return parseRecurringEvent(event, rruleProp, startTime, endTime, title, description, weekStart, weekEnd)
	}

	// Non-recurring event: only add if in this week
	if startTime.Before(weekStart) || startTime.After(weekEnd) {
		return nil, nil
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
		Color:   ColorBlue,
		Weekday: weekday,
		Source:  SourceExternal,
	}
	return []CalendarEvent{calendarEvent}, nil
}

// parseRecurringEvent handles the parsing of recurring events
func parseRecurringEvent(event *ics.VEvent, rruleProp *ics.IANAProperty, startTime, endTime time.Time, title, description string, weekStart, weekEnd time.Time) ([]CalendarEvent, error) {
	// Parse RRULE using rrule-go, but anchor it to the event's DTSTART
	rruleStr := rruleProp.Value
	opt, err := rrule.StrToROption(rruleStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing RRULE: %w", err)
	}
	opt.Dtstart = startTime
	rr, err := rrule.NewRRule(*opt)
	if err != nil {
		return nil, fmt.Errorf("error creating RRule: %w", err)
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

	var events []CalendarEvent
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
			Color:   ColorBlue,
			Weekday: weekday,
			Source:  SourceExternal,
		}
		events = append(events, calendarEvent)
	}
	return events, nil
}
