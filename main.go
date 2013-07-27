package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"code.google.com/p/biogo.kdtree"
)


func getKey(x float64, y float64) string {
	return  fmt.Sprintf("%s,%s", x, y)
}

type Crime struct {
	Id int64
	Date string
	Time string
	Type string
	Point kdtree.Point
}

func (c Crime) String() string {
	return fmt.Sprintf("(%v, %v, %v, %v, %v)", c.Id, c.Date, c.Time, c.Type, c.Point)
}

type RawCrimes [][]string

type Location struct {
	Point kdtree.Point
	Crimes []*Crime
}

type Locations map[string]*Location

func (locs Locations) AllPoints() kdtree.Points {
	points := make(kdtree.Points, 0)
	for _, loc := range locs {
		points = append(points, loc.Point)
	}
	return points
}

func (locs Locations) AllCrimes() []*Crime {
	crimes := make([]*Crime, 0)
	for _, loc := range locs {
		for _, crime := range loc.Crimes {
			crimes = append(crimes, crime)
		}
	}
	return crimes
}

func (locs Locations) Near(query kdtree.Point) Locations {
	tree := kdtree.New(locs.AllPoints(), false)
	keeper := kdtree.NewNKeeper(25)
	tree.NearestSet(keeper, query)
	locations := make(Locations)

	for i := 0; i < keeper.Len(); i++ {
		n := keeper.Pop(); if n == nil {
			// A nil node marks the end of the heap.
			break
		}
		node := n.(kdtree.ComparableDist)
		point := node.Comparable.(kdtree.Point)
		key := getKey(point[0], point[1])
		location, ok := locs[key]
		if (ok) {
			locations[key] = location
		}
	}
	return locations
}


func isFloat(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}
	return true
}

func readCrimes(filename string) (RawCrimes, error) {
	f, err := os.Open(filename); if err != nil {
		return nil, err
	}

	reader := csv.NewReader(f)
	reader.TrailingComma = true
	rows, err := reader.ReadAll(); if err != nil {
		return nil, err
	}

	filteredRows := make(RawCrimes, 0)
	for _, row := range rows {
		if row[8] == "" || row[9] == "" {
			continue
		}
		if !isFloat(row[8]) || !isFloat(row[9]) {
			continue
		}
		filteredRows = append(filteredRows, row)
	}

	return filteredRows, nil
}

func floatForCol(col int, row []string) (float64, error) {
	val := row[col]
	id := row[0]
	f, err := strconv.ParseFloat(val, 64); if err != nil {
		log.Printf("Could not parse column %v of row %v. Bad value: %v", col, id, val)
		return 0, err
	}
	return f, nil
}

func rawCoords(row []string) ([]float64, error) {
	coords := make([]float64, 2)
	lat, err := floatForCol(8, row); if err != nil {
		return nil, err
	}
	coords[0] = lat
	lng, err := floatForCol(9, row); if err != nil {
		return nil, err
	}
	coords[1] = lng
	return coords, nil
}

func parseCrimes(rows RawCrimes) (Locations, error) {
	locations := map[string]*Location{}

	for _, row := range rows {
		coords, err := rawCoords(row); if err != nil {
			return nil, err
		}
		key := getKey(coords[0], coords[1])
		location, ok := locations[key]
		var point kdtree.Point
		if ok {
			point = location.Point
		} else {
			point = kdtree.Point{coords[0], coords[1]}
			location = &Location{point, make([]*Crime, 0)}
			locations[key] = location
		}
		id, err := strconv.ParseInt(row[0], 0, 64); if err != nil {
			fmt.Printf("Could not convert ID to int: %v\n", row[0])
		}
		location.Crimes = append(location.Crimes, &Crime{id, row[1], row[2], row[3], point})
	}

	return locations, nil
}

var locations Locations

func getLocations() (Locations, error) {
	if locations != nil {
		return locations, nil
	}
	rows, err := readCrimes(os.Args[1]); if err != nil {
		return nil, err
	}
	locations, err = parseCrimes(rows); if err != nil {
		return nil, err
	}
	return locations, nil
}

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
	lat, err := strconv.ParseFloat(latRaw[0], 64); if err != nil {
		log.Fatal(err)
		return
	}
	lng, err := strconv.ParseFloat(latRaw[0], 64); if err != nil {
		log.Fatal(err)
		return
	}
	locations, err := getLocations(); if err != nil {
		log.Fatal(err)
		return
	}
	query := kdtree.Point{lat, lng}
	nearby := locations.Near(query)
	resp, err := json.Marshal(nearby.AllCrimes()); if err != nil {
		log.Fatal(err)
	} else {
		w.Write(resp)
	}
	defer r.Body.Close()
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Running server on port", os.Args[2])
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", os.Args[2]), nil))
}
