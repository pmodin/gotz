package core

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// parseArgs parses the command line arguments and applies them to the given configuration.
func ParseFlags(startConfig Config, appVersion string) (Config, time.Time, bool, error) {
	// Define version flag
	version := flag.Bool("version", false, "print version and exit")
	// Check for any changes
	var changed bool
	// Define configuration flags
	var timezones, symbols, tics, stretch, colorize, hours12, live string
	flag.StringVar(
		&timezones,
		"timezones",
		"",
		"timezones to display, comma-separated (for example: 'America/New_York,Europe/London,Asia/Shanghai' or named 'Office:America/New_York,Home:Europe/London' "+
			" - for TZ names see TZ database name in https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)",
	)
	flag.StringVar(
		&symbols,
		"symbols",
		"",
		"symbols to use for time blocks (one of: "+
			SymbolModeRectangles+", "+
			SymbolModeSunMoon+", "+
			SymbolModeMono+")",
	)
	flag.StringVar(
		&tics,
		"tics",
		"",
		"indicates whether to use local time tics on the time axis (one of: true, false)",
	)
	flag.StringVar(
		&stretch,
		"stretch",
		"",
		"indicates whether to stretch across the terminal width at cost of accuracy (one of: true, false)",
	)
	flag.StringVar(
		&colorize,
		"colorize",
		"",
		"indicates whether to colorize the symbols (one of: true, false)",
	)
	flag.StringVar(
		&hours12,
		"hours12",
		"",
		"indicates whether to use 12-hour clock (one of: true, false)",
	)
	flag.StringVar(
		&live,
		"live",
		"",
		"indicates whether to display time live (quit via 'q' or 'Ctrl+C') (one of: true, false)",
	)

	// Define direct flags
	var requestTime string
	var t time.Time = time.Time{}
	flag.StringVar(
		&requestTime,
		"time",
		"",
		"time to display (e.g. 20:00 or 2000 or 20 or 8pm)",
	)

	// Parse flags
	flag.Parse()

	// Check for version flag
	if *version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	// Handle configuration
	if timezones != "" {
		changed = true
		tzs, err := parseTimezones(timezones)
		if err != nil {
			return startConfig, time.Time{}, changed, err
		}
		startConfig.Timezones = tzs
	}
	if symbols != "" {
		changed = true
		startConfig.Style.Symbols = symbols
		if !checkSymbolMode(startConfig.Style.Symbols) {
			return startConfig, time.Time{}, false, fmt.Errorf("invalid symbol mode: %s", symbols)
		}
	}
	if tics != "" {
		changed = true
		if strings.ToLower(tics) == "true" {
			startConfig.Tics = true
		} else if strings.ToLower(tics) == "false" {
			startConfig.Tics = false
		} else {
			return startConfig, time.Time{}, changed, fmt.Errorf("invalid value for tics: %s", tics)
		}
	}
	if stretch != "" {
		changed = true
		if strings.ToLower(stretch) == "true" {
			startConfig.Stretch = true
		} else if strings.ToLower(stretch) == "false" {
			startConfig.Stretch = false
		} else {
			return startConfig, time.Time{}, changed, fmt.Errorf("invalid value for stretch: %s", stretch)
		}
	}
	if colorize != "" {
		changed = true
		if strings.ToLower(colorize) == "true" {
			startConfig.Style.Colorize = true
		} else if strings.ToLower(colorize) == "false" {
			startConfig.Style.Colorize = false
		} else {
			return startConfig, time.Time{}, changed, fmt.Errorf("invalid value for colorize: %s", colorize)
		}
	}
	if hours12 != "" {
		changed = true
		if strings.ToLower(hours12) == "true" {
			startConfig.Hours12 = true
		} else if strings.ToLower(hours12) == "false" {
			startConfig.Hours12 = false
		} else {
			return startConfig, time.Time{}, changed, fmt.Errorf("invalid value for hours12: %s", hours12)
		}
	}
	if live != "" {
		changed = true
		if strings.ToLower(live) == "true" {
			startConfig.Live = true
		} else if strings.ToLower(live) == "false" {
			startConfig.Live = false
		} else {
			return startConfig, time.Time{}, changed, fmt.Errorf("invalid value for live: %s", live)
		}
	}

	// Handle direct flags
	if requestTime != "" {
		// Parse time
		rTime, err := parseTime(requestTime)
		if err != nil {
			return startConfig, time.Time{}, changed, err
		}
		t = rTime
	}

	// Handle last argument as time, if it starts with a digit
	if flag.NArg() > 0 {
		// Get last argument
		lastArg := flag.Arg(flag.NArg() - 1)
		// If last argument is a time, parse it
		if len(lastArg) > 0 && lastArg[0] >= '0' && lastArg[0] <= '9' {
			// Parse time
			rTime, err := parseTime(lastArg)
			if err != nil {
				return startConfig, time.Time{}, changed, err
			}
			t = rTime
		}
	}

	return startConfig, t, changed, nil
}

// parseTimezones parses a comma-separated list of timezones.
func parseTimezones(timezones string) ([]Location, error) {
	var timezoneList []Location
	for _, timezone := range strings.Split(timezones, ",") {
		// Skip empty timezones
		if timezone == "" {
			continue
		}

		if strings.Contains(timezone, ":") {
			// Handle named timezones
			parts := strings.Split(timezone, ":")
			if len(parts) != 2 {
				return timezoneList, fmt.Errorf("invalid timezone: %s", timezone)
			}
			if !checkTimezoneLocation(parts[1]) {
				return timezoneList, fmt.Errorf("invalid timezone: %s", timezone)
			}
			timezoneList = append(timezoneList, Location{
				Name: parts[0],
				TZ:   parts[1],
			})
		} else {
			// Handle simple timezones
			if !checkTimezoneLocation(timezone) {
				return timezoneList, fmt.Errorf("invalid timezone: %s", timezone)
			}
			timezoneList = append(timezoneList, Location{
				Name: timezone,
				TZ:   timezone,
			})
		}
	}
	return timezoneList, nil
}

// checkTimezoneLocation checks if a timezone name is valid.
func checkTimezoneLocation(timezone string) bool {
	_, err := time.LoadLocation(timezone)
	return err == nil
}

// inputTimeFormat defines accepted time formats.
type inputTimeFormat struct {
	// The format string.
	Format string
	// Indicates whether the input declared a date too.
	Date bool
	// Indicates whether the input declared a timezone too.
	TZInfo bool
}

// parseTime parses a time string.
func parseTime(t string) (time.Time, error) {
	// Try all supported formats
	for _, format := range []inputTimeFormat{
		{"15", false, false},
		{"15:04", false, false},
		{"15:04:05", false, false},
		{"3:04pm", false, false},
		{"3:04:05pm", false, false},
		{"3pm", false, false},
		{"1504", false, false},
		{"150405", false, false},
		{"2006-01-02T15:04:05", true, false},
		{"2006-01-02T15:04:05Z07:00", true, true},
	} {
		if t, err := time.Parse(format.Format, t); err == nil {
			n := time.Now()
			if !format.TZInfo {
				if format.Date {
					t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
				} else {
					t = time.Date(n.Year(), n.Month(), n.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
				}
			}
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time: %s", t)
}
