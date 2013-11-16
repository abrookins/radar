# radar: Find Crimes in Portland, Oregon

This is a library (and web service) that finds crime data near a WGS84
coordinate.

# Running

To run the web service, check out this code and build it with `go build`
or install with `go get github.com/abrookins/radar`.

You should receive a `radar` binary. Run that as follows:

	GOMAXPROCS=8 ./radar -p 8081 -f data/crime_incident_data_wgs84.csv

Use whatever value for GOMAXPROCS and the port number that makes sense.

# Loading New Data

The code ships with a version of the City of Portland's crime data from 2011.
You can load new data with the included `scripts/import.py` script. This takes
the path of a CSV file and converts the coordinates from NAD83 to WGS84.

To load a file of CSV data from the City (e.g. `crime_incident_data.csv`), run
this command:

    python import.py crime_incident_data.csv

The output will be a file named `{in_file_name}}_wgs84.csv`.

# Benchmarks

You can run benchmarks with `wrk` (https://github.com/wg/wrk) like this:

    wrk -t12 -c400 -d30s "http://localhost:8081/?lat=45.548&lng=-122.6"

Send the web service a lat/long coordinate in Portland, Oregon as `lat` and
`lng` GET parameters.

Output on my machine:

    Running 30s test @ http://localhost:8081/crimes/near/45.5184/-122.6554
      12 threads and 400 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency   190.92ms   22.74ms 290.56ms   72.67%
        Req/Sec   104.24     18.19   146.00     63.32%
      37487 requests in 30.01s, 2.07GB read
      Socket errors: connect 157, read 143, write 0, timeout 2355
    Requests/sec:   1249.21
    Transfer/sec:     70.53MB
    
Benchmark stats change depending on the location you use. Sometimes it's slower
(1200 reqs/sec) and sometimes faster (3500 reqs/sec).

# License

This code is licensed under the MIT license. See LICENSE for details.

