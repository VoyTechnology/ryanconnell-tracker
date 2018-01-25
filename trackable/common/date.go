package common

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Date struct {
	Day   int
	Month int
	Year  int
}

type NullDate struct {
	Date  Date
	Valid bool
}

var months = []string{
	"january", "february", "march", "april", "may", "june", "july",
	"august", "september", "october", "november", "december",
}

var dateSuffixes = []string{
	"th", "st", "nd", "rd", "th", "th", "th", "th", "th", "th",
}

func (d *Date) String() string {
	return fmt.Sprintf("%02d-%02d-%4d", d.Day, d.Month, d.Year)
}

func (d *Date) FancyString() string {
	var suffix string
	if d.Day >= 10 && d.Day <= 20 {
		suffix = "th"
	} else {
		suffix = dateSuffixes[d.Day%10]
	}
	return fmt.Sprintf("%s %d%s, %4d", strings.Title(months[d.Month-1]), d.Day, suffix, d.Year)
}

func (d *Date) isEmpty() bool {
	return d.Day == 0 && d.Month == 0 && d.Year == 0
}

func (d *Date) ToTime() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
}

func (d *Date) CompareTo(date *Date) int {
	if date == nil {
		return 1
	}
	if d == nil {
		return -1
	}

	if d.Year == date.Year {
		if d.Month == date.Month {
			if d.Day == date.Day {
				return 0
			}
			if d.Day > date.Day {
				return 1
			}
		}
		if d.Month > date.Month {
			return 1
		}
	}
	if d.Year > date.Year {
		return 1
	}
	return -1
}

func (nd *NullDate) Scan(value interface{}) error {
	if val, ok := value.([]uint8); ok {
		err := nd.fromStr(string(val))
		if err != nil {
			return err
		}
		nd.Valid = true
	} else {
		return fmt.Errorf("Unsupported type given for NullDate: %T", value)
	}
	return nil
}

func (nd NullDate) Value() (driver.Value, error) {
	if !nd.Valid {
		return nil, nil
	}
	return fmt.Sprintf("%04d-%02d-%02d", nd.Date.Year, nd.Date.Month, nd.Date.Day), nil
}

func (nd *NullDate) fromStr(str string) error {
	values := strings.Split(str, "-")
	var err error
	if nd.Date.Year, err = StringToInt(values[0]); err != nil {
		return err
	}
	if nd.Date.Month, err = StringToInt(values[1]); err != nil {
		return err
	}
	if nd.Date.Day, err = StringToInt(values[2]); err != nil {
		return err
	}
	return nil
}

func IsDate(str string) bool {
	for _, month := range months {
		if strings.Contains(strings.ToLower(str), month) {
			return true
		}
	}
	return false
}

func ToDate(str string) (*Date, error) {
	date := &Date{}
	regex, err := regexp.Compile(`([a-zA-Z]+)[^0-9]+([0-9]+)[^0-9]+([0-9]+)`)
	matches := regex.FindStringSubmatch(str)
	if len(matches) >= 3 {
		if date.Day, err = StringToInt(matches[2]); err != nil {
			return nil, err
		}
		if date.Month = matchMonth(matches[1]); date.Month == 0 {
			return nil, fmt.Errorf("Invalid month: %s", matches[1])
		}
		if date.Year, err = StringToInt(matches[3]); err != nil {
			return nil, err
		}
	} else {
		return date, fmt.Errorf("Unable to match regexp against str")
	}
	return date, nil
}

func matchMonth(str string) int {
	for i, month := range months {
		if strings.ToLower(str) == month {
			return i + 1
		}
	}
	return 0
}
