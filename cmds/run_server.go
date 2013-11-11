package main

// import _ "net/http/pprof"

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
	"runtime"
	"strconv"

	"github.com/abrookins/radar"
)

var finder radar.CrimeFinder
var port = flag.Int("p", 8081, "port number")
var filename = flag.String("f", "", "data filename")

func handler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	latRaw := q["lat"]
	lngRaw := q["lng"]
	if len(latRaw) == 0 {
		log.Print("lng field not received")
		return
	} else if len(lngRaw) == 0 {
		log.Print("lat field not received")
		return
	}
	lat, err := strconv.ParseFloat(latRaw[0], 64)
	if err != nil {
		log.Fatal(err)
		return
	}
	lng, err := strconv.ParseFloat(lngRaw[0], 64)
	if err != nil {
		log.Fatal(err)
		return
	}
	query := radar.Point{}
	query.Coordinates = []float64{lat, lng}
	nearby, err := finder.FindNear(query)
	if err != nil {
		log.Fatal(err)
	}
	resp := nearby.Crimes().ToJson()
	w.Write(resp.Bytes())
	defer r.Body.Close()
}

func main() {
	var err error
	flag.Parse()
	_, curFilename, _, _ := runtime.Caller(0)
	parentDir := path.Dir(path.Dir(curFilename))
	tracker, err = radar.NewCrimeTracker(path.Join(parentDir, *filename))
	if err != nil {
		log.Fatal("Could not open data file.", err, path.Join(parentDir, *filename))
		return
	}

	http.HandleFunc("/", handler)
	fmt.Println("Running server on port", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *port), nil))
}
