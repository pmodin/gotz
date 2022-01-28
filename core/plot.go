package core

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/term"
)

type Timeslot struct {
	Time time.Time
}

func (c Config) PlotTime() error {
	// Set hours to plot
	hours := 24
	// Get terminal width
	width := GetTerminalWidth()
	// Get current time
	t := time.Now()
	// Determine time slot basics
	timeSlots := make([]Timeslot, width)
	nowSlot := width / 2
	slotMinutes := hours * 60 / width
	offsetMinutes := slotMinutes * width / 2
	// Print header
	fmt.Println(strings.Repeat(" ", nowSlot-4) + "now v " + t.Format("15:04:05"))
	// Prepare slots
	for i := 0; i < width; i++ {
		// Get time of slot
		slotTime := t.Add(time.Duration(i*slotMinutes-offsetMinutes) * time.Minute)
		// Store timeslot info
		timeSlots[i] = Timeslot{
			Time: slotTime,
		}
	}
	// Prepare timezones to plot
	timezones := make([]*time.Location, len(c.Timezones)+1)
	descriptions := make([]string, len(c.Timezones)+1)
	timezones[0] = time.Local
	descriptions[0] = "Local"
	for i, tz := range c.Timezones {
		// Get timezone
		loc, err := time.LoadLocation(tz.TZ)
		if err != nil {
			return fmt.Errorf("error loading timezone %s: %s", tz.TZ, err)
		}
		// Store timezone
		timezones[i+1] = loc
		descriptions[i+1] = tz.Name
	}

	// Plot all timezones
	for i := range timezones {
		// Print header
		fmt.Printf("%s: %s (%s)\n", descriptions[i], FormatTime(t.In(timezones[i])), FormatDay(t.In(timezones[i])))
		for j := 0; j < width; j++ {
			// Convert to tz time
			tzTime := timeSlots[j].Time.In(timezones[i])
			// Get symbol of slot
			symbol := GetHourSymbol(tzTime.Hour())
			if j == nowSlot {
				symbol = "|"
			}
			fmt.Print(symbol)
		}
		fmt.Println()
	}

	return nil
}

func FormatTime(t time.Time) string {
	return t.Format("15:04:05")
}

func FormatDay(t time.Time) string {
	return t.Format("Mon 02 Jan 2006")
}

// GetTerminalWidth returns the width of the terminal.
func GetTerminalWidth() int {
	width, _, err := term.GetSize(0)
	if err != nil {
		return 80
	}
	return width
}

// GetHourSymbol returns a symbol representing the hour in a day.
func GetHourSymbol(hour int) string {
	switch {
	case hour >= 6 && hour < 12:
		return "☀"
	case hour >= 12 && hour < 18:
		return "☼"
	default:
		return "☾"
	}
}
