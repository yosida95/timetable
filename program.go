package timetable

import (
	"time"
)

type Program struct {
	Title    string
	NextOA   time.Time
	Duration time.Duration

	URL      string
	MailAddr string

	Prev *Program
	Next *Program
}
