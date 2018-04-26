package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/yosida95/timetable"
	"github.com/yosida95/timetable/agqr"
	"github.com/yosida95/timetable/qrr"
)

var (
	fname    string
	cronTmpl string
)

func init() {
	flag.StringVar(&fname, "input", "", "query.json")
	flag.StringVar(&cronTmpl, "cron", "{{.MailAddr}}", "in text/template format")
	flag.Parse()
}

func main() {
	var decoder *json.Decoder
	if fname == "" {
		decoder = json.NewDecoder(os.Stdin)
	} else {
		fh, err := os.OpenFile(fname, os.O_RDONLY, 0)
		if err != nil {
			log.Fatal(err)
		}
		defer fh.Close()

		decoder = json.NewDecoder(fh)
	}

	query := make(Query, 0)
	err := decoder.Decode(&query)
	if err != nil {
		log.Fatal(err)
	}
	filter := query.Filter()

	for _, fun := range []func() (*timetable.Program, error){
		agqr.BuildTimetable,
		qrr.BuildTimetable,
	} {
		prog, err := fun()
		if err != nil {
			log.Fatal(err)
		}
		for prog != nil {
			if filter.Match(prog) {
				fmt.Println(prog.Cron(cronTmpl))
			}
			prog = prog.Next
		}
	}
}
