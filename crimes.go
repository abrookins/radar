package radar

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/unit3/kdtree"
)

// A HALF_MILE is one half of a mile in the WGS84 coordinate system.
const HALF_MILE = 0.01

// A Point represents a 2d coordinate within a kd-tree.
type Point kdtree.Node

type Points []*Point

type CsvRow []string

type CsvRows []CsvRow

type Coordinates []float64

type CrimeTypes []*string

func (types CrimeTypes) Contains(crimeType string) bool {
	for _, t := range types {
		if crimeType == *t {
			return true
		}
	}
	return false
}

func (types CrimeTypes) GetOrCreate(crimeType string) string {
	if !types.Contains(crimeType) {
		types = append(types, &crimeType)
	}
	return crimeType
}

type Crime struct {
	Id   int64
	Date string
	Time string
	Type string
}

func (c Crime) String() string {
	return fmt.Sprintf("(%v, %v, %v, %v)", c.Id, c.Date, c.Time, c.Type)
}

type Crimes []*Crime

func (cs Crimes) ToJson() bytes.Buffer {
	var buf bytes.Buffer
	buf.WriteString("[")
	total := len(cs)
	for i, crime := range cs {
		var comma string
		if (i == total - 1) {
			comma = ""
		} else {
			comma = ","
		}
		buf.WriteString(fmt.Sprintf("{%v,\"%v\",\"%v\",\"%v\"}%v", crime.Id, crime.Date, crime.Time, crime.Type, comma))
	}
	buf.WriteString("]")
	return buf
}

type Location struct {
	Point  *Point
	Crimes []*Crime
}

type Locations map[string]*Location

func (locs Locations) getOrCreateFromRow(row CsvRow) (*Location, error) {
	var location *Location
	var pointExists bool
	coords, err := rawCoords(row)
	if err != nil {
		return location, err
	}
	key := getCoordKey(coords[0], coords[1])
	location, pointExists = locs[key]
	if !pointExists {
		point := Point{}
		point.Coordinates = []float64{coords[0], coords[1]}
		location = &Location{&point, make([]*Crime, 0)}
		locs[key] = location
	}
	return location, nil
}

type CrimeTracker struct {
	Locations  Locations
	CrimeTypes CrimeTypes
	Tree       *kdtree.Tree
}

func (t CrimeTracker) AllPoints() Points {
	points := make(Points, 0)
	for _, loc := range t.Locations {
		points = append(points, loc.Point)
	}
	return points
}

func (t CrimeTracker) AllCrimes() Crimes {
	crimes := make(Crimes, 0)
	for _, loc := range t.Locations {
		for _, crime := range loc.Crimes {
			crimes = append(crimes, crime)
		}
	}
	return crimes
}

func (t CrimeTracker) Near(query Point) (CrimeTracker, error) {
	nearby := CrimeTracker{}
	nearby.Locations = make(Locations)
	ranges := map[int]kdtree.Range{
		0: {query.Coordinates[0] - HALF_MILE, query.Coordinates[0] + HALF_MILE},
		1: {query.Coordinates[1] - HALF_MILE, query.Coordinates[1] + HALF_MILE}}
	results, err := t.Tree.FindRange(ranges)
	if err != nil {
		return nearby, err
	}
	for i := 0; i < len(results); i++ {
		node := results[i]
		key := getCoordKey(node.Coordinates[0], node.Coordinates[1])
		// If we have a record for this coordinate, add it to ``nearby``.
		location, ok := t.Locations[key]
		if ok {
			nearby.Locations[key] = location
		}
	}
	return nearby, nil
}

func (t CrimeTracker) parseCrimes(rows CsvRows) (Locations, error) {
	locations := make(Locations)
	for _, row := range rows {
		location, err := locations.getOrCreateFromRow(row)
		if err != nil {
			return locations, err
		}
		id, err := strconv.ParseInt(row[0], 0, 64)
		if err != nil {
			return locations, err
		}
		crimeType := t.CrimeTypes.GetOrCreate(string(row[3]))
		location.Crimes = append(location.Crimes, &Crime{id, row[1], row[2], crimeType})
	}
	return locations, nil
}

func NewCrimeTracker(filename string) (CrimeTracker, error) {
	var err error
	tracker := CrimeTracker{}
	rows, err := readCrimes(filename)
	if err != nil {
		return tracker, err
	}
	tracker.Locations, err = tracker.parseCrimes(rows)
	if err != nil {
		return tracker, err
	}
	points := tracker.AllPoints()
	nodes := make([]*kdtree.Node, len(points))
	for i, p := range points {
		n := kdtree.Node(*p)
		nodes[i] = &n
	}
	tracker.Tree = kdtree.BuildTree(nodes)
	tracker.CrimeTypes = make(CrimeTypes, 0)
	return tracker, nil
}

func getCoordKey(x float64, y float64) string {
	return fmt.Sprintf("%s,%s", x, y)
}

func isFloat(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}
	return true
}

func readCrimes(filename string) (CsvRows, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(f)
	reader.TrailingComma = true
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	filteredRows := make(CsvRows, 0)
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

func floatForCol(col int, row CsvRow) (float64, error) {
	val := row[col]
	id := row[0]
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("Could not parse column %v of row %v. Bad value: %v", col, id, val)
		return 0, err
	}
	return f, nil
}

func rawCoords(row CsvRow) (Coordinates, error) {
	coords := make(Coordinates, 2)
	lat, err := floatForCol(8, row)
	if err != nil {
		return nil, err
	}
	coords[0] = lat
	lng, err := floatForCol(9, row)
	if err != nil {
		return nil, err
	}
	coords[1] = lng
	return coords, nil
}
