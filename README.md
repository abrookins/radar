# radar: Find Crimes in Portland, Oregon

This is a library (and web service) that finds crime data near a WGS84
coordinate.

# Running

To run the web service, check out this code and build it with `go build`.

You should receive a `run_server` binary. Run that as follows:

	GOMAXPROCS=8 ./run_server -p 8081 -f data/crime_incident_data_wgs84.csv

Use whatever value for GOMAXPROCS and the port number that makes sense.

# Loading New Data

The code ships with a version of the City of Portland's crime data from 2011.
You can load new data with the included `scripts/import.py` script. This takes
the path of a CSV file and converts the coordinates from NAD83 to WGS84.

To load a file of CSV data from the City (`e.g. crime_incident_data.csv`), run
this command:

	python import.py crime_incident_data.csv

The output will be a file named `{in_file_name}}_wgs84.csv`.

# Benchmarks

You can bencharmk with `wrk` (https://github.com/wg/wrk) like this:

	wrk -t12 -c400 -d30s "http://localhost:8081/?lat=45.548&lng=-122.6"

Just give it the port you are running `radar` on and a lat/long coordinate in
Portland.

Output on my machine:

	Running 30s test @ http://localhost:8081/?lat=45.548&lng=-122.6
	  12 threads and 400 connections
	  Thread Stats   Avg      Stdev     Max   +/- Stdev
		Latency    76.52ms   13.56ms 174.38ms   70.77%
		Req/Sec   255.16    156.90   483.00     43.80%
	  93565 requests in 30.01s, 1.63GB read
	  Socket errors: connect 157, read 120, write 0, timeout 2254
	Requests/sec:   3118.16
	Transfer/sec:     55.53MB

# License

This code is licensed under the MIT license. See LICENSE for details.
