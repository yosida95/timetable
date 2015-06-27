package timetable

import (
	"bytes"
	"fmt"
	"text/template"
	"time"
)

type Program struct {
	Title    string
	NextOA   time.Time
	Duration time.Duration

	URL      string
	MailAddr string

	IsFirst  bool
	IsRepeat bool
	IsLive   bool

	Prev *Program
	Next *Program
}

func (p *Program) Cron(tmplStr string) string {
	cmd := bytes.NewBuffer(nil)

	tmpl := template.Must(template.New("cron").Parse(tmplStr))
	err := tmpl.Execute(cmd, p)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf(
		"%2d %2d * * %d %s",
		p.NextOA.Minute(), p.NextOA.Hour(), p.NextOA.Weekday(), cmd.String())
}
