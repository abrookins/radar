// Package radar loads crime data from the City of Portland and makes it
// available to search using WGS84 coordinates.

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

// One half mile in the WGS84 coordinate system.
//
// This is a rough estimate of distance derived from the fact that one degree
// of latitude or longitude is 70 miles wide at the equator. This distance is
// constant outside of the equator for latitude, but not for longitude.
const HALF_MILE = 0.00714

// A Point represents a 2d coordinate within a kd-tree.
type Point kdtree.Node

type Points []*Point

type CsvRow []string

type CsvRows []CsvRow

type Coordinates []float64

// We store a slice of all the types of crime in the CSV data. This isn't used
// anywhere in the code but is something we make available to clients.
type CrimeTypes []string

func (types CrimeTypes) Contains(crimeType string) bool {
	for _, t := range types {
		if crimeType == t {
			return true
		}
	}
	return false
}

// Data for a single crime in the City's CSV data (one row).
type Crime struct {
	Id   int64
	Date string
	Time string
	Type string
}

// String formats a string version of a Crime.
func (c Crime) String() string {
	return fmt.Sprintf("(%v, %v, %v, %v)", c.Id, c.Date, c.Time, c.Type)
}

type Crimes []*Crime

// ToJson creates a bytes.Buffer containing a JSON representation of Crimes.
func (cs Crimes) ToJson() *bytes.Buffer {
	buf := new(bytes.Buffer)
	total := len(cs)
	buf.WriteString("[")
	line := `{%v,"%v","%v","%v"}`

	for i, crime := range cs {
		isLast := total > 1 && i == total-1
		buf.WriteString(fmt.Sprintf(line, crime.Id, crime.Date, crime.Time, crime.Type))
		if !isLast {
			buf.WriteString(",")
		}
	}
	buf.WriteString("]")
	return buf
}

// A location in the City's data with a coordinate at which crimes occurred.
type CrimeLocation struct {
	Point  *Point
	Crimes []*Crime
}

// GetCoordinateString returns a CrimeLocation's coordinates formtted as a
// string.
func (loc CrimeLocation) GetCoordinateString() string {
	return fmt.Sprintf("%s,%s", loc.Point.Coordinates[0], loc.Point.Coordinates[1])
}

// A map of coordinate strings (slices can't be map keys) to the CrimeLocation
// at the coordinate.
type CrimeLocations map[string]*CrimeLocation

// getOrCreateFromCsvRow gets an existing CrimeLocation for the coordinate
// stored in "row", or creates a CrimeLocation for that coordinate if one does
// not exist.
func (locs CrimeLocations) getOrCreateFromCsvRow(row CsvRow) (*CrimeLocation, error) {
	var location *CrimeLocation
	var pointExists bool
	coords, err := floatCoordsFromRow(row)
	if err != nil {
		return nil, err
	}
	key := getCoordKey(coords[0], coords[1])
	location, pointExists = locs[key]
	if !pointExists {
		point := Point{}
		point.Coordinates = []float64{coords[0], coords[1]}
		location = &CrimeLocation{&point, make([]*Crime, 0)}
		locs[key] = location
	}
	return location, nil
}

// The result of a search for crimes near a location.
type SearchResult struct {
	CrimeLocations CrimeLocations
}

// Points returns all of the coordinates of a SearchResult's CrimeLocations.
func (r SearchResult) Points() Points {
	points := make(Points, 0)
	for _, loc := range r.CrimeLocations {
		points = append(points, loc.Point)
	}
	return points
}

// Crimes returns all of the Crimes in a SearchResult.
func (r SearchResult) Crimes() Crimes {
	crimes := make(Crimes, 0)
	for _, loc := range r.CrimeLocations {
		for _, crime := range loc.Crimes {
			crimes = append(crimes, crime)
		}
	}
	return crimes
}

// An object that can find crimes near a WGS84 coordinate.
type CrimeFinder struct {
	CrimeLocations CrimeLocations
	CrimeTypes     CrimeTypes
	Tree           *kdtree.Tree
}

// FindNear returns a SearchResult containing CrimeLocations within a half-mile of ``point``
func (t *CrimeFinder) FindNear(query Point) (SearchResult, error) {
	nearby := SearchResult{}
	nearby.CrimeLocations = make(CrimeLocations)
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
		location, ok := t.CrimeLocations[key]
		if ok {
			nearby.CrimeLocations[key] = location
		}
	}
	return nearby, nil
}

// All returns a SearchResult containing all CrimeLocations in the CrimeFinder.
func (t *CrimeFinder) All() SearchResult {
	all := SearchResult{}
	all.CrimeLocations = t.CrimeLocations
	return all
}

// loadFromCsv hydrates a CrimeFinder from CSV data.
func (t *CrimeFinder) loadFromCsv(rows CsvRows) error {
	locations := make(CrimeLocations)
	numCrimes := 0
	for _, row := range rows {
		location, err := locations.getOrCreateFromCsvRow(row)
		if err != nil {
			return err
		}
		// Parse the "id" column as an int64
		id, err := strconv.ParseInt(row[0], 0, 64)
		if err != nil {
			return err
		}
		crimeType := string(row[3])
		if !t.CrimeTypes.Contains(crimeType) {
			t.CrimeTypes = append(t.CrimeTypes, crimeType)
		}
		location.Crimes = append(location.Crimes, &Crime{id, row[1], row[2], crimeType})
		numCrimes += 1
	}
	log.Printf("Loaded %v crimes and %v locations", numCrimes, len(locations))
	t.CrimeLocations = locations
	return nil
}

// NewCrimeFinder creates a new CrimeFinder loaded from CSV data.
func NewCrimeFinder(filename string) (CrimeFinder, error) {
	var err error
	finder := CrimeFinder{}
	rows, err := readCrimes(filename)
	if err != nil {
		return finder, err
	}
	err = finder.loadFromCsv(rows)
	if err != nil {
		return finder, err
	}
	points := finder.All().Points()
	nodes := make([]*kdtree.Node, len(points))
	for i, p := range points {
		n := kdtree.Node(*p)
		nodes[i] = &n
	}
	finder.Tree = kdtree.BuildTree(nodes)
	return finder, nil
}

// getCoordKey returns a pair of float64 coordinates as strings.
func getCoordKey(x float64, y float64) string {
	return fmt.Sprintf("%s,%s", x, y)
}

// isFloat checks if a string is coercible to a float.
func isFloat(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}
	return true
}

// readCrimes reads CSV data from a file identified by filename.
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

// floatForCol tries to coerce a specific column of a CSV file into float64.
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

// floatCoordsFromRow tries to return a coordinate pair for a row of CSV data.
func floatCoordsFromRow(row CsvRow) (Coordinates, error) {
	coords := make(Coordinates, 2)
	latitudeColumn := 8
	longitudeColumn := 9
	lat, err := floatForCol(latitudeColumn, row)
	if err != nil {
		return nil, err
	}
	coords[0] = lat
	lng, err := floatForCol(longitudeColumn, row)
	if err != nil {
		return nil, err
	}
	coords[1] = lng
	return coords, nil
}
