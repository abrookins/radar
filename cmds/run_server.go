package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
	"runtime"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/abrookins/radar"
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
		log.Fatal(err)
	}
	resp := nearby.Crimes().ToJson()
	w.Write(resp.Bytes())
	defer r.Body.Close()
}

func main() {
	var err error
	flag.Parse()

	// Get the project's directory
	_, curFilename, _, _ := runtime.Caller(0)
	parentDir := path.Dir(path.Dir(curFilename))

	finder, err = radar.NewCrimeFinder(path.Join(parentDir, *filename))
	if err != nil {
		log.Fatal("Could not open data file.", err, path.Join(parentDir, *filename))
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/crimes/near/{lat:[-+]?[0-9]*.?[0-9]+.}/{lng:[-+]?[0-9]*.?[0-9]+.}", handler)
	http.Handle("/", r)

	log.Println("Running server on port", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *port), nil))
}
