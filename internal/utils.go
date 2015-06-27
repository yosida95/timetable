package internal

import (
	"strconv"
	"strings"
	"time"
)

func ParseTimeColonSeparated(inp string) (h, m int, err error) {
	tokens := strings.SplitN(strings.TrimSpace(inp), ":", 2)
	if len(tokens) != 2 {
		return
	}

	h, err = strconv.Atoi(tokens[0])
	if err != nil {
		return
	}

	m, err = strconv.Atoi(tokens[1])
	return
}

func WaitTimeToNextOA(now time.Time, h, m int, weekday time.Weekday) time.Duration {
	nextDuration := time.Duration((7+weekday-now.Weekday())%7) * 24 * time.Hour
	if now.Hour() < h || now.Hour() == h && now.Minute() < m {
		nextDuration += time.Duration(h-now.Hour()) * time.Hour
		nextDuration += time.Duration(m-now.Minute()) * time.Minute
	} else {
		if now.Weekday() == weekday {
			nextDuration += 7 * 24 * time.Hour
		}
		nextDuration -= time.Duration(now.Hour()-h) * time.Hour
		nextDuration -= time.Duration(now.Minute()-m) * time.Minute
	}

	return nextDuration
}
