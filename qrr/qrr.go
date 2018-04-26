package qrr

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/yosida95/timetable"
	"github.com/yosida95/timetable/internal"
	"golang.org/x/net/html/atom"
)

const TIMETABLE_URL = "http://www.agqr.jp/timetable/radio.php"
const START_OF_DAY = 11

func parseTimetable(body io.Reader) (*timetable.Program, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}

	table := doc.Find("table.timetb-am").First()

	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), START_OF_DAY, 0, 0, 0, today.Location())
	weekdays := make([]*timetable.Program, 7)
	offset := (int(today.Weekday()) + 6) % 7
	for i := 0; i < 7; i++ {
		weekdays[i] = &timetable.Program{
			NextOA: today.AddDate(0, 0, i-offset),
		}
		if i > 0 {
			weekdays[i].Prev = weekdays[i-1]
			weekdays[i-1].Next = weekdays[i]
		}
	}
	root := weekdays[0]

	cursor := make([]int, 7)
	indices := make([]int, 0, 7)
	table.Find("tbody tr").EachWithBreak(func(_ int, row *goquery.Selection) bool {
		min := cursor[0]
		indices = indices[:0]
		indices = append(indices, 0)
		for i := 1; i < 7; i++ {
			if x := cursor[i]; x <= min {
				if x < min {
					min = x
					indices = indices[:0]
				}
				indices = append(indices, i)
			}
		}

		var hour int
		var oclock bool
		row.Children().EachWithBreak(func(j int, row *goquery.Selection) bool {
			if j > 7 {
				err = errors.New("excess columns")
				return false
			}
			switch row.Get(0).DataAtom {
			case atom.Th:
				if j > 0 {
					err = errors.New("unexpected <th>")
					return false
				}

				hour, err = strconv.Atoi(row.Text())
				if err != nil {
					return false
				}
				oclock = true
			case atom.Td:
				var index int
				if oclock {
					index = indices[j-1]
				} else {
					index = indices[j]
				}

				var span int
				span, err = strconv.Atoi(row.AttrOr("rowspan", "1"))
				if err != nil {
					return false
				}
				span *= 30
				cursor[index] += span

				prev := weekdays[index]
				var prog *timetable.Program
				if prev.Duration == 0 {
					prog = prev
				} else {
					prog = &timetable.Program{
						NextOA: prev.NextOA.Add(prev.Duration),
						Prev:   prev,
					}
					if prev.Next != nil {
						prev.Next.Prev = prog
						prog.Next = prev.Next
					}
					prev.Next = prog
					weekdays[index] = prog
				}
				prog.Title = strings.TrimFunc(row.Find("div.title-p").Text(), unicode.IsSpace)
				prog.Duration = time.Duration(span) * time.Minute
				prog.IsFirst = true
				row.Find("a").Each(func(_ int, a *goquery.Selection) {
					href, ok := a.Attr("href")
					if !ok {
						return
					}
					if strings.HasPrefix(href, "mailto:") {
						prog.MailAddr = href[7:]
					} else {
						prog.URL = href
					}
				})

				if oclock && (prog.NextOA.Hour() != hour || prog.NextOA.Minute() != 0) {
					err = errors.New("out of sync")
					return false
				}
			default:
				err = errors.New("unexpected tag")
				return false
			}
			return true
		})
		if err != nil {
			return false
		}
		return true
	})
	if err != nil {
		return nil, err
	}

	return root, nil
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
