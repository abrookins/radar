package radar

import (
	"fmt"
	"testing"
)

func TestCrimeTypeContainsDoesNotExist(t *testing.T) {
	ct := make(CrimeTypes, 0)
	if ct.Contains("Should not exist") {
		t.Error("It should not contain a string that it does not contain")
	}
}

func TestCrimeTypeContainsExists(t *testing.T) {
	ct := make(CrimeTypes, 0)
	str := "Hello"
	ct = append(ct, &str)
	if(!ct.Contains(str)) {
		t.Error("It should contain a string that it does contain")
	}
}

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
	expectedCrimes := Crimes {
		{int64(1), "1/1/2013", "04:30", "Burglary"},
		{int64(2), "1/2/2013", "04:45", "Robbery"},
	}
	expectedJson := "[{1,\"1/1/2013\",\"04:30\",\"Burglary\"},{2,\"1/2/2013\",\"04:45\",\"Robbery\"}]"
	actualJson := expectedCrimes.ToJson()
	if expectedJson != actualJson.String() {
		t.Error("Crimes JSON string is wrong: ", expectedJson, actualJson.String())
	}
}

func TestLocationHasFields(t *testing.T) {
	expectedPoint := Point{}
	expectedPoint.Coordinates = []float64{20.2, 33.34}
	crimes := make([]*Crime, 0)
	l := &Location{&expectedPoint, crimes}
	address1 := &expectedPoint
	address2 := l.Point
	// Struct equality: Compare the addresses of the pointers
	if (address1 != address2) {
		t.Error("Location.Point is not the expected Point", &address1, &address2)
	}
}

func TestLocationsGetOrCreateFromRowDoesNotExist(t *testing.T) {
	csvRow := CsvRow{"13690824","05/27/2011","08:35:00","Liquor Laws","NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212","ELIOT","PORTLAND PREC NO","590","45.53579735412487","-122.66468312170824"}
	locations := make(Locations, 0)
	location, _ := locations.getOrCreateFromRow(csvRow)
	if location.Point.Coordinates[0] != float64(45.53579735412487) {
		t.Error("Location has the wrong x coordinate", location.Point.Coordinates[0])
	}
	if location.Point.Coordinates[1] != float64(-122.66468312170824) {
		t.Error("Location has the wrong y coordiante", location.Point.Coordinates[1])
	}
	if len(locations) != 1 {
		t.Error("Locations should only have one Location")
	}
}

func TestLocationsGetOrCreateFromRowExists(t *testing.T) {
	csvRow := CsvRow{"13690824","05/27/2011","08:35:00","Liquor Laws","NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212","ELIOT","PORTLAND PREC NO","590","45.53579735412487","-122.66468312170824"}
	locations := make(Locations, 0)
	location, _ := locations.getOrCreateFromRow(csvRow)
	// Call again with data at the same coordinates
	location2, _ := locations.getOrCreateFromRow(csvRow)

	if len(locations) != 1 {
		t.Error("Locations should only have one Location")
	}
	if location != location2 {
		t.Error("Locations should have returned the same Location, not created a second one", location, location2)
	}
}

func TestLocationsGetOrCreateFromRowBadLatitude(t *testing.T) {
	// The latitude is munged so it won't convert to float64
	csvRow := CsvRow{"13690824","05/27/2011","08:35:00","Liquor Laws","NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212","ELIOT","PORTLAND PREC NO","590","45.53579735412487","not-a-float"}
	locations := make(Locations, 0)
	_, err := locations.getOrCreateFromRow(csvRow)

	if err == nil {
		t.Error("Should have returned an error due to bad coordinate")
	}
}

func TestLocationsGetOrCreateFromRowBadLongitude(t *testing.T) {
	// The longitude is munged so it won't convert to float64
	csvRow := CsvRow{"13690824","05/27/2011","08:35:00","Liquor Laws","NE SCHUYLER ST and NE 1ST AVE, PORTLAND, OR 97212","ELIOT","PORTLAND PREC NO","590","not-a-float","-122.66468312170824"}
	locations := make(Locations, 0)
	_, err := locations.getOrCreateFromRow(csvRow)

	if err == nil {
		t.Error("Should have returned an error due to bad coordinate")
	}
}
