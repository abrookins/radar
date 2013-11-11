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
	dLat := (p2.Coordinates[1] - p.Coordinates[1]) * (math.Pi / 180.0)
	dLon := (p2.Coordinates[0] - p.Coordinates[0]) * (math.Pi / 180.0)

	lat1 := p.Coordinates[1] * (math.Pi / 180.0)
	lat2 := p2.Coordinates[1] * (math.Pi / 180.0)

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

func TestCrimesToJson(t *testing.T) {
	expectedCrimes := Crimes{
		{int64(1), "1/1/2013", "04:30", "Burglary"},
		{int64(2), "1/2/2013", "04:45", "Robbery"},
	}
	expectedJson := `[{1,"1/1/2013","04:30","Burglary"},{2,"1/2/2013","04:45","Robbery"}]`
	actualJson := expectedCrimes.ToJson()
	if expectedJson != actualJson.String() {
		t.Error("Crimes JSON string is wrong. Expected: ", expectedJson, "Actual: ", actualJson.String())
	}
}

// CrimeLocation tests

func TestCrimeLocationHasFields(t *testing.T) {
	expectedPoint := Point{}
	expectedPoint.Coordinates = []float64{20.2, 33.34}
	crimes := make([]*Crime, 0)
	l := &CrimeLocation{&expectedPoint, crimes}
	address1 := &expectedPoint
	address2 := l.Point
	// Struct equality: Compare two pointers
	if address1 != address2 {
		t.Error("CrimeLocation.Point is not the expected Point", &address1, &address2)
	}
}

func TestCrimeLocationsGetOrCreateFromRowDoesNotExist(t *testing.T) {
	csvRow := CsvRow{"13690824", "05/27/2011", "08:35:00", "Liquor Laws", "NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212", "ELIOT", "PORTLAND PREC NO", "590", "45.53579735412487", "-122.66468312170824"}
	locations := make(CrimeLocations, 0)
	location, _ := locations.getOrCreateFromCsvRow(csvRow)
	if location.Point.Coordinates[0] != float64(45.53579735412487) {
		t.Error("CrimeLocation has the wrong x coordinate", location.Point.Coordinates[0])
	}
	if location.Point.Coordinates[1] != float64(-122.66468312170824) {
		t.Error("CrimeLocation has the wrong y coordiante", location.Point.Coordinates[1])
	}
	if len(locations) != 1 {
		t.Error("CrimeLocations should only have one CrimeLocation")
	}
}

func TestCrimeLocationsGetOrCreateFromRowExists(t *testing.T) {
	csvRow := CsvRow{"13690824", "05/27/2011", "08:35:00", "Liquor Laws", "NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212", "ELIOT", "PORTLAND PREC NO", "590", "45.53579735412487", "-122.66468312170824"}
	locations := make(CrimeLocations, 0)
	location, _ := locations.getOrCreateFromCsvRow(csvRow)
	// Call again with data at the same coordinates
	location2, _ := locations.getOrCreateFromCsvRow(csvRow)

	if len(locations) != 1 {
		t.Error("CrimeLocations should only have one CrimeLocation")
	}
	if location != location2 {
		t.Error("CrimeLocations should have returned the same CrimeLocation, not created a second one", location, location2)
	}
}

func TestCrimeLocationsGetOrCreateFromRowBadLatitude(t *testing.T) {
	// The latitude is munged so it won't convert to float64
	csvRow := CsvRow{"13690824", "05/27/2011", "08:35:00", "Liquor Laws", "NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212", "ELIOT", "PORTLAND PREC NO", "590", "45.53579735412487", "not-a-float"}
	locations := make(CrimeLocations, 0)
	_, err := locations.getOrCreateFromCsvRow(csvRow)

	if err == nil {
		t.Error("Should have returned an error due to bad coordinate")
	}
}

func TestCrimeLocationsGetOrCreateFromRowBadLongitude(t *testing.T) {
	// The longitude is munged so it won't convert to float64
	csvRow := CsvRow{"13690824", "05/27/2011", "08:35:00", "Liquor Laws", "NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212", "ELIOT", "PORTLAND PREC NO", "590", "not-a-float", "-122.66468312170824"}
	locations := make(CrimeLocations, 0)
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
	// 1-length slice just to test that we set the value
	locations := make(CrimeLocations)
	csvRow := CsvRow{"13690824", "05/27/2011", "08:35:00", "Liquor Laws", "NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212", "ELIOT", "PORTLAND PREC NO", "590", "45.53579735412487", "-122.66468312170824"}
	loc, _ := locations.getOrCreateFromCsvRow(csvRow)
	coordKey := loc.GetCoordinateString()
	finder.CrimeLocations = locations
	nodes := make([]*kdtree.Node, 0)
	tree := kdtree.BuildTree(nodes)
	finder.Tree = tree

	if len(finder.CrimeTypes) != 1 {
		t.Error("CrimeFinder.CrimeTypes value is wrong")
	}
	if loc != finder.CrimeLocations[coordKey] {
		t.Error("CrimeFinder.CrimeLocations value is wrong")
	}
	if finder.Tree != tree {
		t.Error("CrimeFinder.Tree value is wrong")
	}
}

func TestCrimeFinderNewCrimeFinder(t *testing.T) {
	finder, err := NewCrimeFinder("data/test.csv")
	if err != nil {
		t.Error("Error creating CrimeFinder: ", err)
	}
	if len(finder.CrimeLocations) != 224 {
		t.Error("Wrong number of CrimeLocations: ", len(finder.CrimeLocations))
	}
}

func TestCrimeFinderAll(t *testing.T) {
	finder, err := NewCrimeFinder("data/test.csv")
	if err != nil {
		t.Error("Error creating CrimeFinder: ", err)
	}
	all := finder.All()
	allCrimes := all.Crimes()

	expectedLocations := 224
	expectedCrimes := 2321

	numCrimeLocations := len(all.CrimeLocations)
	numCrimes := len(allCrimes)

	if expectedLocations != numCrimeLocations {
		t.Error("Wrong number of CrimeLocations: ", numCrimeLocations)
	}
	if expectedCrimes != numCrimes {
		t.Error("Wrong number of Crimes: ", numCrimes)
	}
}

func TestCrimeFinderFindNear(t *testing.T) {
	finder, _ := NewCrimeFinder("data/test.csv")
	point := Point{}
	point.Coordinates = []float64{45.53435699129174, -122.66469510763777}
	result, _ := finder.FindNear(point)

	expectedLocations := 14
	numLocations := len(result.CrimeLocations)

	if expectedLocations != numLocations {
		t.Error("Wrong number of CrimeLocations: ", numLocations)
	}

	// Verify that no distance is more than 0.5 miles
	for _, p := range result.CrimeLocations {
		distance := p.Point.GreatCircleDistance(&point)
		if distance > 0.5 {
			t.Error("FindNear returned a CrimeLocation more than half a mile away")
		}
	}
}
