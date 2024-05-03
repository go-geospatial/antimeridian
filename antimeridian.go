// Copyright 2024
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package antimeridian

import (
	"errors"
	"math"
	"slices"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/xy"
)

var (
	ErrUnsupportedType = errors.New("unsupported geometry type")
)

// Cut divides a geometry at the antimeridian and the
// poles. A multi-geometry is returned with the cut
// portions of the original geometry.
func Cut(obj geom.T) (geom.T, error) {
	switch geometry := obj.(type) {
	case *geom.Polygon:
		return cutPolygon(geometry)
	case *geom.MultiPolygon:
		return cutMultiPolygon(geometry)
	default:
		// unsupported type
		return obj, ErrUnsupportedType
	}
}

func cutPolygon(poly *geom.Polygon) (geom.T, error) {

	var polygons []*geom.Polygon
	segments := segment(poly.LinearRing(0).Coords())

	if len(segments) == 0 {
		polygons = append(poly)
	} else {
		interiors := make([][]geom.Coord, 0)
		for idx := range poly.NumLinearRings() - 1 {
			interior := poly.LinearRing(idx+1)
			interiorSegments := segment(interior.Coords())
			if interior_segments:
				if fix_winding:
					unwrapped_linearring = LinearRing(
						list((x % 360, y) for x, y in interior.coords)
					)
					if shapely.is_ccw(unwrapped_linearring):
						FixWindingWarning.warn()
						interior_segments = segment(list(reversed(interior.coords)))
				segments.extend(interior_segments)
			else:
				interiors.append(interior)
		}
	}

	/*
	   segments = extend_over_poles(
	       segments,
	       force_north_pole=force_north_pole,
	       force_south_pole=force_south_pole,
	       fix_winding=fix_winding,
	   )
	   polygons = build_polygons(segments)
	   assert polygons
	   for i, polygon in enumerate(polygons):
	       for j, interior in enumerate(interiors):
	           if polygon.contains(interior):
	               interior = interiors.pop(j)
	               polygon_interiors = list(polygon.interiors)
	               polygon_interiors.append(interior)
	               polygons[i] = Polygon(polygon.exterior, polygon_interiors)
	*/

	if len(polygons) == 1 {
		polygon := polygons[0]
		if xy.IsRingCounterClockwise(polygon.Layout(), polygon.FlatCoords()) {
			return polygon, nil
		} else {
			return geom.NewPolygon(polygon.Layout()).MustSetCoords(
				[][]geom.Coord{{{-180, 90}, {-180, -90}, {180, -90}, {180, 90}}},
			), nil
		}
	}

	return poly, nil
}

func cutMultiPolygon(poly *geom.MultiPolygon) (*geom.MultiPolygon, error) {
	return poly, nil
}

func segment(coords []geom.Coord) [][]geom.Coord {
	currSegment := make([]geom.Coord, 0)
	segments := make([][]geom.Coord, 0)

	// Ensure all longitudes are between -180 and 180, and that tiny floating
	// point differences are ignored
	tol := 1e-08
	for idx, point := range coords {
		switch {
		case math.Abs(point[0]-180.0) <= tol:
			coords[idx] = geom.Coord{180.0, point[1]}
		case math.Abs(point[0]+180) <= tol:
			coords[idx] = geom.Coord{-180.0, point[1]}
		default:
			coords[idx] = geom.Coord{math.Mod(point[0]+180.0, 360.0) - 180.0, point[1]}
		}
	}

	// create segments
	for idx := range len(coords) - 1 {
		start, end := coords[idx], coords[idx+1]
		currSegment = append(currSegment, start)

		switch {
		case (end[0]-start[0] > 180) && (end[0]-start[0] != 360):
			// left
			latitude := crossingLat(start, end)
			currSegment = append(currSegment, geom.Coord{-180.0, latitude})
			segments = append(segments, currSegment)
			currSegment = []geom.Coord{{180.0, latitude}}
		case (start[0]-end[0] > 180) && (start[0]-end[0] != 360):
			// right
			latitude := crossingLat(end, start)
			currSegment = append(currSegment, geom.Coord{180.0, latitude})
			segments = append(segments, currSegment)
			currSegment = []geom.Coord{{-180, latitude}}
		}
	}

	switch {
	case len(segments) == 0:
		// no antimeridian crossings
		return segments
	case slices.Compare[[]float64](coords[len(coords)-1], segments[0][0]) == 0:
		// join polygons
		segments[0] = append(currSegment, segments[0]...)
	default:
		currSegment = append(currSegment, coords[len(coords)-1])
		segments = append(segments, currSegment)
	}

	return segments
}

func crossingLat(start, end geom.Coord) float64 {
	switch {
	case math.Abs(start[0]) == 180.0:
		return start[1]
	case math.Abs(end[0]) == 180.0:
		return end[1]
	}

	latDelta := end[1] - start[1]

	if end[0] > 0 {
		return roundFloat(start[1]+(180.0-start[0])*latDelta/(end[0]+360.0-start[0]), 7)
	} else {
		return roundFloat(start[1]+(start[0]+180.0)*latDelta/(start[0]+360.0-end[0]), 7)
	}
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
