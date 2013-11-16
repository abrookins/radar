package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	// Uncomment to profile
	//_ "net/http/pprof"

	"github.com/gorilla/mux"

	"github.com/abrookins/radar/crimes"
)

var finder radar.CrimeFinder
var port = flag.Int("p", 8081, "port number")
var filename = flag.String("f", "", "data filename")

func handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// I trust that the regex gave us float-worthy values.
	lat, _ := strconv.ParseFloat(vars["lat"], 64)
	lng, _ := strconv.ParseFloat(vars["lng"], 64)

	query := radar.Point{}
	query.Coordinates = []float64{lat, lng}
	nearby, err := finder.FindNear(query)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Fatal(err)
		return
	}
	resp := nearby.Crimes().ToJson()
	w.Write(resp.Bytes())
	defer r.Body.Close()
}

func main() {
	var err error
	flag.Parse()

	finder, err = radar.NewCrimeFinder(*filename)
	if err != nil {
		log.Fatal("Could not open data file.", err, *filename)
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/crimes/near/{lat:[-+]?[0-9]*.?[0-9]+.}/{lng:[-+]?[0-9]*.?[0-9]+.}", handler)
	http.Handle("/", r)

	log.Println("Running server on port", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *port), nil))
}
