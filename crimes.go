package radar

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/unit3/kdtree"
)


const HALF_MILE = 0.01

type Point kdtree.Node

type Points []Point

type CsvRows [][]string

type CsvRow []string

type Coordinates []float64

type Crime struct {
	Id    int64
	Date  string
	Time  string
	Type  string
	Point Point
}

type Location struct {
	Point  Point
	Crimes []*Crime
}

type Locations map[string]*Location

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

func (c Crime) String() string {
	return fmt.Sprintf("(%v, %v, %v, %v, %v)", c.Id, c.Date, c.Time, c.Type, c.Point)
}

func (locs Locations) AllPoints() Points {
	points := make(Points, 0)
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

func (locs Locations) Near(query Point) (Locations, error) {
	points := locs.AllPoints()
	nodes := make([]*kdtree.Node, len(points))
	for i, p := range points {
		n := kdtree.Node(p)
		nodes[i] = &n
	}
	tree := kdtree.BuildTree(nodes)
	locations := make(Locations)
	ranges := map[int]kdtree.Range{0: {query.Coordinates[0] - HALF_MILE, query.Coordinates[0] + HALF_MILE},
						           1: {query.Coordinates[1] - HALF_MILE, query.Coordinates[1] + HALF_MILE}}
	fmt.Println(ranges)
	results, err := tree.FindRange(ranges)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(results); i++ {
		node := results[i]
		key := getCoordKey(node.Coordinates[0], node.Coordinates[1])
		location, ok := locs[key]
		if ok {
			locations[key] = location
		}
	}
	return locations, nil
}

func NewLocationManager(filename string) (Locations, error) {
	var locations Locations
	rows, err := readCrimes(filename)
	if err != nil {
		return nil, err
	}
	locations, err = parseCrimes(rows)
	if err != nil {
		return nil, err
	}
	return locations, nil
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

func parseCrimes(rows CsvRows) (Locations, error) {
	locations := map[string]*Location{}

	for _, row := range rows {
		coords, err := rawCoords(row)
		if err != nil {
			return nil, err
		}
		key := getCoordKey(coords[0], coords[1])
		location, ok := locations[key]
		var point Point
		if ok {
			point = location.Point
		} else {
			point = Point{}
			point.Coordinates = []float64{coords[0], coords[1]}
			location = &Location{point, make([]*Crime, 0)}
			locations[key] = location
		}
		id, err := strconv.ParseInt(row[0], 0, 64)
		if err != nil {
			fmt.Printf("Could not convert ID to int: %v\n", row[0])
		}
		location.Crimes = append(location.Crimes, &Crime{id, row[1], row[2], row[3], point})
	}

	return locations, nil
}
