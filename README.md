# radar: Find Crimes in Portland, Oregon

[![Build Status](https://travis-ci.org/abrookins/radar.png)](https://travis-ci.org/abrookins/radar)

This is a library (and web service) that finds crime data near a WGS84
coordinate.

# Running the Server

To run `radar` as a web service, check out this code and build it with `go
build` or install with `go get github.com/abrookins/radar`.

You should receive a `radar` binary that provides an HTTP server. Run that as
follows:

	GOMAXPROCS=4 ./radar -p 8081 -f data/crime_incident_data_wgs84.csv

Use whatever value for GOMAXPROCS and the port number that makes sense.

# Running Tests

With the package installed, navigate to its source directory in your `GOPATH`
and run the following command:

    got test ./...

This is a special form of `go test` that runs tests in sub-packages.

# Loading New Data

The code ships with a version of the City of Portland's crime data from 2011.
You can load new data with the included `scripts/import.py` script. This takes
the path of a CSV file and converts the coordinates from NAD83 to WGS84.

To load a file of CSV data from the City (e.g. `crime_incident_data.csv`), run
this command:

    python import.py crime_incident_data.csv

The output will be a file named `{in_file_name}}_wgs84.csv`.

# Deploying

You can deploy `radar` to Heroku pretty easily. First create an instance using
a popular Go buildpack:

    heroku create -b https://github.com/kr/heroku-buildpack-go.git

The Procfile in the repo should do the necessaries. Now just push to Heroku!

# The API

There is only one endpoint right now: /crimes/near/{latitude}/{longitude}.

Here is an example of a GET:

    GET http://localhost:8081/crimes/near/45.5184/-122.6554

Response:

    {
        "query": {
            "lng": -122.6554,
            "lat": 45.5184
        },
        "locations": [
            {
                "crimes": [
                    {
                        "date": "04/18/2011",
                        "id": 13667453,
                        "time": "15:16:00",
                        "type": "Liquor Laws"
                    },
                    {
                        "date": "04/29/2011",
                        "id": 13672680,
                        "time": "13:45:00",
                        "type": "Motor Vehicle Theft"
                    },
                    {
                        "date": "09/17/2011",
                        "id": 13760105,
                        "time": "10:56:00",
                        "type": "Larceny"
                    },
                    {
                        "date": "12/15/2011",
                        "id": 13815913,
                        "time": "10:02:00",
                        "type": "Larceny"
                    }
                ],
                "point": {
                    "lng": -122.65769669639069,
                    "lat": 45.51793011872208
                }
            },
            {
                "crimes": [
                    {
                        "date": "02/18/2011",
                        "id": 13633694,
                        "time": "09:24:00",
                        "type": "Vandalism"
                    },
                    {
                        "date": "03/12/2011",
                        "id": 13646262,
                        "time": "00:24:00",
                        "type": "DUII"
                    },
                    {
                        "date": "05/07/2011",
                        "id": 13679925,
                        "time": "13:14:00",
                        "type": "Larceny"
                    },
                    {
                        "date": "07/23/2011",
                        "id": 13726944,
                        "time": "01:10:00",
                        "type": "DUII"
                    },
                    {
                        "date": "08/13/2011",
                        "id": 13736457,
                        "time": "14:30:00",
                        "type": "Motor Vehicle Theft"
                    }
                ],
                "point": {
                    "lat": 45.52479664790835,
                    "lng": -122.64852371711835
                }
            }
        ]
    }

# License

This code is licensed under the MIT license. See LICENSE for details.

