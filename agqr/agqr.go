package agqr

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/yosida95/timetable"
	"github.com/yosida95/timetable/internal"
	"golang.org/x/net/html/atom"
)

const (
	TIMETABLE_URL = "http://www.agqr.jp/timetable/streaming.php"
)

func parseProgram(col *goquery.Selection, weekday time.Weekday, prev *timetable.Program, now time.Time) (*timetable.Program, error) {
	h, m, err := internal.ParseTimeColonSeparated(col.Find("div.time").Text())
	if err != nil {
		return nil, err
	}
	if 4 >= h {
		weekday = (weekday + 1) % 7
	}
	nextDuration := internal.WaitTimeToNextOA(now, h, m, weekday)

	title := col.Find("div.title-p")
	url, _ := title.Find("a").Attr("href")

	prog := &timetable.Program{
		Title:  strings.TrimSpace(title.Text()),
		NextOA: now.Add(nextDuration),

		URL: strings.TrimSpace(url),
	}

	col.Find("div.rp").Children().Each(func(_ int, s *goquery.Selection) {
		switch s.Nodes[0].DataAtom {
		case atom.A:
			href, _ := s.Attr("href")
			href = strings.TrimSpace(href)
			if strings.HasPrefix(href, "mailto:") {
				prog.MailAddr = href[7:]
			}
		}
	})

	val, ok := col.Attr("class")
	switch {
	case !ok:
		prog.IsRepeat = true
	case strings.Contains(val, "bg-l"):
		prog.IsLive = true
	case strings.Contains(val, "bg-f"):
		prog.IsFirst = true
	}

	if prev != nil && prev.NextOA.Before(prog.NextOA) {
		prog.Prev = prev
		prev.Duration = prog.NextOA.Sub(prev.NextOA)
		prev.Next = prog
	} else if prev != nil {
		prev.Duration = prog.NextOA.Sub(prev.NextOA.Add(-7 * 24 * time.Hour))
	}

	return prog, nil
}

func parseTimetable(body io.Reader) (root *timetable.Program, err error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return
	}

	now := time.Now().Truncate(time.Minute)
	var (
		table  [14]*timetable.Program
		offset [7]int
	)
	doc.Find("table.timetb-ag tbody tr").Each(func(_ int, row *goquery.Selection) {
		i := 0
		row.Children().EachWithBreak(func(_ int, col *goquery.Selection) bool {
			node := col.Nodes[0]
			switch node.DataAtom {
			case atom.Td:
				for ; offset[i] > 0; i++ {
					if i >= 6 {
						return false
					}
				}

				if rowspanStr, ok := col.Attr("rowspan"); ok {
					offset[i], err = strconv.Atoi(strings.TrimSpace(rowspanStr))
					if err != nil {
						return false
					}
				}

				weekday := time.Weekday((i + 1) % 7)
				firstI := i * 2
				lastI := firstI + 1

				prev := table[lastI]
				table[lastI], err = parseProgram(col, weekday, prev, now)
				if err != nil {
					return false
				}
				if table[firstI] == nil {
					table[firstI] = table[lastI]
				}

				if root == nil {
					root = table[firstI]
				} else if root.NextOA.After(table[lastI].NextOA) {
					root = table[lastI]
				}

				i++
			}
			return true
		})
		for i := 0; i < 7; i++ {
			if offset[i] > 0 {
				offset[i]--
			}
		}
	})

	for i := 1; i < 15; i += 2 {
		prev := table[i]
		next := table[(i+1)%14]

		if prev.NextOA.Before(next.NextOA) {
			prev.Duration = next.NextOA.Sub(prev.NextOA)
			prev.Next = next
			next.Prev = prev
		}
	}

	return
}

func BuildTimetable() (*timetable.Program, error) {
	req, err := http.NewRequest("GET", TIMETABLE_URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", internal.USERAGENT)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	return parseTimetable(resp.Body)
}
