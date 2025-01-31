import csv
import ogr
import os
import sys


DATA_DIR = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), 'data')


def open_data_file(filename, mode='r'):
    return open(os.path.join(DATA_DIR, filename), mode)


def import_crime_data(in_file, out_file=None):
    """
    Load Portland crime data from a CSV file ``in_file``, convert coordinate
    data for each crime from NAD83 to WGS84, then write to a file named
    {in_file}_wgs84.csv.

	TODO: This writes columns named "X coordinate" which is actually longitude
	and "Y Coordinate" which is actually latitude.
    """
    skipped = 0
    if out_file is None:
        out_file = '{}_wgs84.csv'.format(in_file.split('.csv')[0])

    # State Plane Coordinate System (Oregon North - EPSG:2269, alt: EPSG:2913).
    nad83 = ogr.osr.SpatialReference()
    nad83.ImportFromEPSG(2269)

    # Latitude/longitude (WGS84 - EPSG:4326)
    wgs84 = ogr.osr.SpatialReference()
    wgs84.ImportFromEPSG(4326)

    transformation = ogr.osr.CoordinateTransformation(nad83, wgs84)

    r = csv.reader(open_data_file(in_file))
    w = csv.writer(open_data_file(out_file, 'w'))

    for i, row in enumerate(r):
        if i == 0:
            w.writerow(row)
            continue

        x, y = float(row[8]) if row[8] else 0, float(row[9] if row[9] else 0)
        if x and y:
            try:
                coord = transformation.TransformPoint(x, y)
            except TypeError:
                skipped += 1
            else:
                # The index order here (1, 0) is intended (longitude, latitude).
                row[8] = coord[1]
                row[9] = coord[0]
                w.writerow(row)
    return skipped


if __name__ == '__main__':
    in_file = sys.argv[1]
    try:
        out_file = sys.argv[2]
    except IndexError:
        out_file = None

    skipped = import_crime_data(in_file, out_file)

    print "Import complete."

    if skipped:
        print '{} records skipped due to missing or invalid ' \
            'coordinates.'.format(skipped)

