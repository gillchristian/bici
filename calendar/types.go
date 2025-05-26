package calendar

// EventTime represents the start and end times of a calendar event
type EventTime struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// Weekday represents a day of the week
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

// Source represents the origin of a calendar event
type Source string

const (
	SourceExternal Source = "external"
	SourceInternal Source = "internal"
)

// Color represents the color of a calendar event
type Color string

const (
	ColorAmber   Color = "amber"
	ColorBlue    Color = "blue"
	ColorCyan    Color = "cyan"
	ColorEmerald Color = "emerald"
	ColorGray    Color = "gray"
	ColorIndigo  Color = "indigo"
	ColorOrange  Color = "orange"
	ColorPink    Color = "pink"
	ColorPurple  Color = "purple"
	ColorTeal    Color = "teal"
)

// CalendarEvent represents a calendar event
type CalendarEvent struct {
	Id          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Time        EventTime `json:"time"`
	Weekday     Weekday   `json:"weekday"`
	Color       Color     `json:"color"`
	Source      Source    `json:"source"`
}

// String returns the string representation of a Weekday
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
