package radar

import (
	"fmt"
	"math"
	"testing"

	"github.com/unit3/kdtree"
)

// Radius of the earth (M)
var EARTH_RADIUS = 3959.0

// GreatCircleDistance calculates the Haversine distance between two points.
// Used to verify distance-based search.
// https://github.com/kellydunn/golang-geo/blob/master/point.go
func (p *Point) GreatCircleDistance(p2 *Point) float64 {
	dLat := (p2.Lat - p.Lat) * (math.Pi / 180.0)
	dLon := (p2.Lng - p.Lng) * (math.Pi / 180.0)

	lat1 := p.Lat * (math.Pi / 180.0)
	lat2 := p2.Lat * (math.Pi / 180.0)

	a1 := math.Sin(dLat/2) * math.Sin(dLat/2)
	a2 := math.Sin(dLon/2) * math.Sin(dLon/2) * math.Cos(lat1) * math.Cos(lat2)

	a := a1 + a2

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EARTH_RADIUS * c
}

// CrimeType tests

func TestCrimeTypeContainsDoesNotExist(t *testing.T) {
	ct := make(CrimeTypes, 0)
	if ct.Contains("Should not exist") {
		t.Error("It should not contain a string that it does not contain")
	}
}

func TestCrimeTypeContainsExists(t *testing.T) {
	ct := make(CrimeTypes, 0)
	str := "Hello"
	ct = append(ct, str)
	if !ct.Contains(str) {
		t.Error("It should contain a string that it does contain")
	}
}

// Crime tests

func TestCrimeHasFields(t *testing.T) {
	expectedId := int64(1)
	expectedDate := "1/1/2013"
	expectedTime := "04:30"
	expectedType := "Burglary"
	c := &Crime{expectedId, expectedDate, expectedTime, expectedType}

	if expectedId != c.Id {
		t.Error("It should have an ID")
	}
	if expectedDate != c.Date {
		t.Error("It should have a Date")
	}
	if expectedTime != c.Time {
		t.Error("It should have a Time")
	}
	if expectedType != c.Type {
		t.Error("It should have a Type")
	}
}

func TestCrimesString(t *testing.T) {
	expectedId := int64(1)
	expectedDate := "1/1/2013"
	expectedTime := "04:30"
	expectedType := "Burglary"
	c := &Crime{expectedId, expectedDate, expectedTime, expectedType}

	expectedString := "(1, 1/1/2013, 04:30, Burglary)"
	actual := fmt.Sprintf("%v", c)

	if expectedString != actual {
		t.Error("Crime did not convert to the right string: ", actual)
	}
}

func TestSearchResultToJson(t *testing.T) {
	crimes := Crimes{
		{int64(1), "1/1/2013", "04:30", "Burglary"},
		{int64(2), "1/2/2013", "04:45", "Robbery"},
	}
	crimePoint := Point{45.1, -122.3}
	location := CrimeLocation{
		&crimePoint,
		crimes,
	}
	queryPoint := Point{45.1, -122.3}
	node := kdtree.Node{}
	node.Coordinates = Coordinates{crimePoint.Lat, crimePoint.Lng}
	searchResult := SearchResult{
		&queryPoint,
		[]*CrimeLocation{&location},
	}
	expectedJson := `{"query":{"lat":45.1,"lng":-122.3},"locations":[{"point":{"lat":45.1,"lng":-122.3},"crimes":[{"id":1,"date":"1/1/2013","time":"04:30","type":"Burglary"},{"id":2,"date":"1/2/2013","time":"04:45","type":"Robbery"}]}]}`
	actualJson, err := searchResult.ToJson()
	jsonString := string(actualJson[:])
	if err != nil {
		t.Error("ToJson returned an error: ", err)
	}
	if expectedJson != jsonString {
		t.Error("Crimes JSON string is wrong. Expected: ", expectedJson, "Actual: ", jsonString)
	}
}

// CrimeLocation tests

func TestCrimeLocationHasFields(t *testing.T) {
	expectedPoint := Point{20.2, 33.34}
	crimes := make([]*Crime, 0)
	l := &CrimeLocation{&expectedPoint, crimes}
	address1 := &expectedPoint
	address2 := l.Point
	// Struct equality: Compare two pointers
	if address1 != address2 {
		t.Error("CrimeLocation.Point is not the expected Point", &address1, &address2)
	}
}

func TestLocationLookupGetOrCreateFromRowDoesNotExist(t *testing.T) {
	csvRow := CsvRow{"13690824", "05/27/2011", "08:35:00", "Liquor Laws", "NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212", "ELIOT", "PORTLAND PREC NO", "590", "45.53579735412487", "-122.66468312170824"}
	locations := make(LocationLookup, 0)
	location, _ := locations.getOrCreateFromCsvRow(csvRow)
	if location.Point.Lat != float64(45.53579735412487) {
		t.Error("CrimeLocation has the wrong latitude", location.Point.Lat)
	}
	if location.Point.Lng != float64(-122.66468312170824) {
		t.Error("CrimeLocation has the wrong longitude", location.Point.Lng)
	}
	if len(locations) != 1 {
		t.Error("LocationLookup should only have one CrimeLocation")
	}
}

func TestLocationLookupGetOrCreateFromRowExists(t *testing.T) {
	csvRow := CsvRow{"13690824", "05/27/2011", "08:35:00", "Liquor Laws", "NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212", "ELIOT", "PORTLAND PREC NO", "590", "45.53579735412487", "-122.66468312170824"}
	locations := make(LocationLookup, 0)
	location, _ := locations.getOrCreateFromCsvRow(csvRow)
	// Call again with data at the same coordinates
	location2, _ := locations.getOrCreateFromCsvRow(csvRow)

	if len(locations) != 1 {
		t.Error("LocationLookup should only have one CrimeLocation")
	}
	if location != location2 {
		t.Error("LocationLookup should have returned the same CrimeLocation, not created a second one", location, location2)
	}
}

func TestLocationLookupGetOrCreateFromRowBadLatitude(t *testing.T) {
	// The latitude is munged so it won't convert to float64
	csvRow := CsvRow{"13690824", "05/27/2011", "08:35:00", "Liquor Laws", "NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212", "ELIOT", "PORTLAND PREC NO", "590", "45.53579735412487", "not-a-float"}
	locations := make(LocationLookup, 0)
	_, err := locations.getOrCreateFromCsvRow(csvRow)

	if err == nil {
		t.Error("Should have returned an error due to bad coordinate")
	}
}

func TestLocationLookupGetOrCreateFromRowBadLongitude(t *testing.T) {
	// The longitude is munged so it won't convert to float64
	csvRow := CsvRow{"13690824", "05/27/2011", "08:35:00", "Liquor Laws", "NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212", "ELIOT", "PORTLAND PREC NO", "590", "not-a-float", "-122.66468312170824"}
	locations := make(LocationLookup, 0)
	_, err := locations.getOrCreateFromCsvRow(csvRow)

	if err == nil {
		t.Error("Should have returned an error due to bad coordinate")
	}
}

// CrimeFinder tests

func TestCrimeFinderFields(t *testing.T) {
	finder := CrimeFinder{}
	// 1-length slice just to test that we set the value
	crimeTypes := make(CrimeTypes, 1)
	finder.CrimeTypes = crimeTypes
	locations := make(LocationLookup)

	csvRow := CsvRow{"13690824", "05/27/2011", "08:35:00", "Liquor Laws", "NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212", "ELIOT", "PORTLAND PREC NO", "590", "45.53579735412487", "-122.66468312170824"}
	_, err := locations.getOrCreateFromCsvRow(csvRow)
	if err != nil {
		t.Error("Could not create CrimeLocation: ", err)
	}
	finder.LocationLookup = locations
	nodes := make([]*kdtree.Node, 0)
	for _, location := range finder.LocationLookup {
		node := kdtree.Node{}
		node.Coordinates = Coordinates{location.Point.Lat, location.Point.Lng}
		nodes = append(nodes, &node)
	}
	tree := kdtree.BuildTree(nodes)
	finder.Tree = tree

	if len(finder.CrimeTypes) != 1 {
		t.Error("CrimeFinder.CrimeTypes value is wrong")
	}
	if finder.Tree != tree {
		t.Error("CrimeFinder.Tree value is wrong")
	}
}

func TestCrimeFinderNewCrimeFinder(t *testing.T) {
	finder, err := NewCrimeFinder("../data/test.csv")
	if err != nil {
		t.Error("Error creating CrimeFinder: ", err)
	}
	if len(finder.LocationLookup) != 224 {
		t.Error("Wrong number of LocationLookup: ", len(finder.LocationLookup))
	}
}

func TestCrimeFinderAll(t *testing.T) {
	finder, err := NewCrimeFinder("../data/test.csv")
	if err != nil {
		t.Error("Error creating CrimeFinder: ", err)
	}
	all := finder.All()

	expectedLocations := 224
	numLocations := len(all.Locations)

	if expectedLocations != numLocations {
		t.Error("Wrong number of Locations in the LocationLookup table: ", numLocations)
	}
}

func TestCrimeFinderLocations(t *testing.T) {
	finder, err := NewCrimeFinder("../data/test.csv")
	if err != nil {
		t.Error("Error creating CrimeFinder: ", err)
	}
	locations := finder.Locations()

	expectedLocations := 224
	numLocations := len(locations)

	if expectedLocations != numLocations {
		t.Error("Wrong number of Locations in the returned by finder.Locations(): ", numLocations)
	}
}

func TestCrimeFinderFindNear(t *testing.T) {
	finder, _ := NewCrimeFinder("../data/test.csv")
	point := Point{45.53435699129174, -122.66469510763777}
	result, _ := finder.FindNear(point)

	expectedLocations := 14
	numLocations := len(result.Locations)

	if expectedLocations != numLocations {
		t.Error("FindNear returned the number of LocationLookup: ", numLocations)
	}

	if *result.Query != point {
		t.Error("FindNear result had the wrong query", result.Query)
	}

	// Verify that no distance is more than 0.5 miles
	for _, p := range result.Locations {
		distance := p.Point.GreatCircleDistance(&point)
		if distance > 0.5 {
			t.Error("FindNear returned a CrimeLocation more than half a mile away")
		}
	}
}

// A regression test to make sure we find locations near a known-good location.
func TestCrimeFinderFindNearRegression(t *testing.T) {
	finder, _ := NewCrimeFinder("../data/crime_incident_data_wgs84.csv")
	point := Point{45.5184, -122.6554}
	result, _ := finder.FindNear(point)

	expectedLocations := 247
	numLocations := len(result.Locations)

	if expectedLocations != numLocations {
		t.Error("Wrong number of Locations: ", numLocations)
	}
}

func TestGetCoordinateKey(t *testing.T) {
	x := 45.1
	y := -122.1
	key := GetCoordinateKey(x, y)
	if key != "45.1,-122.1" {
		t.Error("Coordinate key is wrong: ", key)
	}
}
