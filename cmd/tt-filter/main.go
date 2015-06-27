package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/yosida95/timetable/agqr"
)

var (
	fname string
)

func init() {
	flag.StringVar(&fname, "input", "", "query.json")
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

	prog, err := agqr.BuildTimetable()
	if err != nil {
		log.Fatal(err)
	}

	for prog != nil {
		if filter.Match(prog) {
			log.Println(prog)
		}

		prog = prog.Next
	}
}