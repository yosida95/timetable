package main

import (
	"strings"

	"github.com/yosida95/timetable"
)

type filter interface {
	Match(*timetable.Program) bool
}

type filterFunc func(*timetable.Program) bool

func (fun filterFunc) Match(p *timetable.Program) bool {
	return fun(p)
}

type titleFilter string

func (f titleFilter) Match(p *timetable.Program) bool {
	return strings.Contains(p.Title, string(f))
}

type urlFilter string

func (f urlFilter) Match(p *timetable.Program) bool {
	return p.URL == string(f)
}

type mailAddrFilter string

func (f mailAddrFilter) Match(p *timetable.Program) bool {
	return p.MailAddr == string(f)
}

func isFirstFilter(p *timetable.Program) bool {
	return p.IsFirst
}

func isLiveFilter(p *timetable.Program) bool {
	return p.IsLive
}

func isRepeatFilter(p *timetable.Program) bool {
	return p.IsRepeat
}

type joinedFilter []filter

type AndFilter joinedFilter

func (f AndFilter) Match(prog *timetable.Program) bool {
	for _, cond := range f {
		if !cond.Match(prog) {
			return false
		}
	}

	return true
}

type OrFilter joinedFilter

func (f OrFilter) Match(prog *timetable.Program) bool {
	for _, cond := range f {
		if cond.Match(prog) {
			return true
		}
	}

	return false
}
