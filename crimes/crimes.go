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

// One half mile of latitude in the WGS84 coordinate system.
const HALF_MILE_LAT = 0.00714

// One half mile of longitude in the WGS84 coordinate system.
const HALF_MILE_LNG = 0.00724

// A Point represents a latitude and longitude coordinate pair.
type Point struct {
	Lat float64
	Lng float64
}

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

// A location in the City's data with a coordinate at which crimes occurred.
type CrimeLocation struct {
	Point  *Point
	Crimes []*Crime
}

// This will help us find the CrimeLocation that a kd-tree node refers to.
type LocationLookup map[string]*CrimeLocation

// getOrCreateFromCsvRow gets an existing CrimeLocation for the coordinate
// stored in "row", or creates a CrimeLocation for that coordinate if one does
// not exist.
func (locs LocationLookup) getOrCreateFromCsvRow(row CsvRow) (*CrimeLocation, error) {
	var location *CrimeLocation
	var pointExists bool
	coords, err := floatCoordsFromRow(row)
	if err != nil {
		return nil, err
	}
	key := GetCoordinateKey(coords[0], coords[1])
	location, pointExists = locs[key]
	if !pointExists {
		point := Point{coords[0], coords[1]}
		location = &CrimeLocation{&point, make([]*Crime, 0)}
		locs[key] = location
	}
	return location, nil
}

// The result of a search for crimes near a location.
type SearchResult struct {
	Query     *Point
	Locations []*CrimeLocation
}

// Points returns all of the coordinates of a SearchResult's LocationLookup.
func (r SearchResult) Points() Points {
	points := make(Points, len(r.Locations))
	for _, loc := range r.Locations {
		points = append(points, loc.Point)
	}
	return points
}

// Crimes returns all of the Crimes in a SearchResult.
func (r SearchResult) Crimes() Crimes {
	crimes := make(Crimes, 0)
	for _, loc := range r.Locations {
		for _, crime := range loc.Crimes {
			crimes = append(crimes, crime)
		}
	}
	return crimes
}

// ToJson returns a SearchResult marshalled to JSON bytes.
// XXX: This is terrible but gained several hundred requests/sec over json.Marshall.
func (r SearchResult) ToJson() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf(`{"query":{"lat":%v,"lng":%v},"locations":[`, r.Query.Lat, r.Query.Lng))
	totalLocations := len(r.Locations)

	for x, location := range r.Locations {
		total := len(location.Crimes)
		buf.WriteString(fmt.Sprintf(`{"point":{"lat":%v,"lng":%v},`, location.Point.Lat, location.Point.Lng))
		buf.WriteString(`"crimes":[`)
		line := `{"id":%v,"date":"%v","time":"%v","type":"%v"}`
		for i, crime := range location.Crimes {
			isLast := i == total-1
			buf.WriteString(fmt.Sprintf(line, crime.Id, crime.Date, crime.Time, crime.Type))
			if (total > 1) && !isLast {
				buf.WriteString(",")
			}
		}
		buf.WriteString("]}")
		isLast :=  x == totalLocations-1
		if (totalLocations > 1) && !isLast {
			buf.WriteString(",")
		}	
	}
	buf.WriteString("]}")
	return buf.Bytes(), nil
}

// An object that can find crimes near a WGS84 coordinate.
type CrimeFinder struct {
	LocationLookup LocationLookup
	CrimeTypes     CrimeTypes
	Tree           *kdtree.Tree
}

// Locations returned a slice of all the CrimeLocations in this CrimeFinder
func (finder *CrimeFinder) Locations() []*CrimeLocation {
	locations := make([]*CrimeLocation, 0)
	for _, location := range finder.LocationLookup {
		locations = append(locations, location)
	}
	return locations
}

// FindNear returns a SearchResult containing LocationLookup within a half-mile of ``query``
func (finder *CrimeFinder) FindNear(query Point) (SearchResult, error) {
	nearby := SearchResult{}
	nearby.Query = &query
	nearby.Locations = make([]*CrimeLocation, 0)
	ranges := map[int]kdtree.Range{
		0: {query.Lat - HALF_MILE_LAT, query.Lat + HALF_MILE_LAT},
		1: {query.Lng - HALF_MILE_LNG, query.Lng + HALF_MILE_LNG}}
	results, err := finder.Tree.FindRange(ranges)
	if err != nil {
		return nearby, err
	}
	for i := 0; i < len(results); i++ {
		node := results[i]
		// If we have a record for this coordinate, add it to ``nearby``.
		key := GetCoordinateKey(node.Coordinates[0], node.Coordinates[1])
		location, exists := finder.LocationLookup[key]
		if exists {
			nearby.Locations = append(nearby.Locations, location)
		}
	}
	return nearby, nil
}

// All returns a SearchResult containing all LocationLookup in the CrimeFinder.
func (finder *CrimeFinder) All() SearchResult {
	all := SearchResult{}
	all.Locations = finder.Locations()
	return all
}

// loadFromCsv hydrates a CrimeFinder from CSV data.
func (finder *CrimeFinder) loadFromCsv(rows CsvRows) error {
	locations := make(LocationLookup)
	numCrimes := 0
	for _, row := range rows {
		location, err := locations.getOrCreateFromCsvRow(row)
		if err != nil {
			continue
		}
		// Parse the "id" column as an int64
		id, err := strconv.ParseInt(row[0], 0, 64)
		if err != nil {
			continue
		}
		crimeType := string(row[3])
		if !finder.CrimeTypes.Contains(crimeType) {
			finder.CrimeTypes = append(finder.CrimeTypes, crimeType)
		}
		location.Crimes = append(location.Crimes, &Crime{id, row[1], row[2], crimeType})
		numCrimes += 1
	}
	log.Printf("Loaded %v crimes and %v locations", numCrimes, len(locations))
	finder.LocationLookup = locations
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
	nodes := make([]*kdtree.Node, 0)
	for _, location := range finder.LocationLookup {
		node := kdtree.Node{}
		node.Coordinates = Coordinates{location.Point.Lat, location.Point.Lng}
		nodes = append(nodes, &node)
	}
	finder.Tree = kdtree.BuildTree(nodes)
	return finder, nil
}

// GetCoordinateKey returns a pair of float64 coordinates as strings.
func GetCoordinateKey(x float64, y float64) string {
	return fmt.Sprintf("%v,%v", x, y)
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
