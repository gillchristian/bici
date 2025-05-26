package calendar

import (
	"fmt"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	rrule "github.com/teambition/rrule-go"
)

// Helper to sanitize strings for schedule
func sanitizeString(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "#", "-")
	s = strings.TrimSpace(s)
	return s
}

// parseEvent parses a single calendar event and returns a CalendarEvent
func parseEvent(event *ics.VEvent, weekStart, weekEnd time.Time) ([]CalendarEvent, error) {
	// Exclude all-day events if the constant is set
	if ExcludeAllDayEvents {
		dtstartProp := event.GetProperty(ics.ComponentPropertyDtStart)
		if dtstartProp != nil && dtstartProp.GetValueType() == ics.ValueDataTypeDate {
			return nil, nil
		}
	}

	// Get event start time
	startTime, err := event.GetStartAt()
	if err != nil {
		return nil, fmt.Errorf("error getting event start time: %w", err)
	}

	// Get event end time, default to 1 hour after start if not specified
	endTime, err := event.GetEndAt()
	if err != nil {
		endTime = startTime.Add(time.Hour)
	}

	// Get event title and description, sanitize them
	title := ""
	if prop := event.GetProperty(ics.ComponentPropertySummary); prop != nil {
		title = sanitizeString(prop.Value)
	}
	description := ""
	if prop := event.GetProperty(ics.ComponentPropertyDescription); prop != nil {
		description = sanitizeString(prop.Value)
	}

	// Check for RRULE (recurrence rule)
	rruleProp := event.GetProperty("RRULE")
	if rruleProp != nil {
		return parseRecurringEvent(event, rruleProp, startTime, endTime, title, description, weekStart, weekEnd)
	}

	// Non-recurring event: only add if in this week
	if startTime.Before(weekStart) || !startTime.Before(weekEnd) {
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

	dtstart, _ := event.GetStartAt()
	dtend, _ := event.GetEndAt()
	fmt.Printf("[ImportEvents] Processing event: '%s' | DTSTART: %v | DTEND: %v\n", title, dtstart, dtend)

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

	// Use the timezone from startTime
	loc := startTime.Location()
	if loc != time.UTC {
		// Convert weekStart and weekEnd to the event's timezone
		weekStart = weekStart.In(loc)
		weekEnd = weekEnd.In(loc)
	}

	// Set the start time and don't limit the until date
	opt.Dtstart = startTime
	opt.Until = time.Time{} // Don't limit the until date
	rr, err := rrule.NewRRule(*opt)
	if err != nil {
		return nil, fmt.Errorf("error creating RRule: %w", err)
	}

	// Handle EXDATEs (dates to exclude)
	exdates := map[time.Time]bool{}
	exdateProps := event.GetProperties("EXDATE")
	for _, ex := range exdateProps {
		if ex != nil {
			exdate, err := time.Parse(time.RFC3339, ex.Value)
			if err != nil {
				exdate, err = time.Parse("20060102", ex.Value)
			}
			if err != nil {
				exdate, err = time.Parse("20060102T150405", ex.Value)
			}
			if err == nil {
				exdates[exdate] = true
			}
		}
	}

	// Get all occurrences in this week
	occurrences := rr.Between(weekStart, weekEnd, true)
	fmt.Printf("[parseRecurringEvent] Found %d occurrences for event '%s' between %v and %v (in timezone %v)\n",
		len(occurrences), title, weekStart.Format("2006-01-02 15:04:05 -0700"),
		weekEnd.Format("2006-01-02 15:04:05 -0700"), loc)

	// Use a map to deduplicate events based on their start time and title
	dedupMap := make(map[string]CalendarEvent)
	var events []CalendarEvent

	for _, occ := range occurrences {
		// Skip if the occurrence is outside our week window
		if occ.Before(weekStart) || !occ.Before(weekEnd) {
			continue
		}

		// Skip if this occurrence is in the exdates
		skip := false
		for ex := range exdates {
			if occ.Year() == ex.Year() &&
				occ.Month() == ex.Month() &&
				occ.Day() == ex.Day() &&
				occ.Hour() == ex.Hour() &&
				occ.Minute() == ex.Minute() &&
				occ.Second() == ex.Second() {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Calculate the end time for this occurrence
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

		// Create a unique key for this event based on its start time and title
		key := fmt.Sprintf("%s_%s", title, occ.Format(time.RFC3339))
		if _, exists := dedupMap[key]; !exists {
			dedupMap[key] = calendarEvent
			fmt.Printf("[parseRecurringEvent] Included event: '%s' | Start: %s | End: %s | Timezone: %v\n",
				title, occ.Format(time.RFC3339), occEnd.Format(time.RFC3339), occ.Location())
		}
	}

	// Convert the map values to a slice
	for _, event := range dedupMap {
		events = append(events, event)
	}

	return events, nil
}
