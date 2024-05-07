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
	"math"
	"slices"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/xy"
)

type edge struct {
	Index int
	Val   float64
}

func cutPolygon(poly *geom.Polygon, fixWindingArr ...bool) (geom.T, error) {
	fixWinding := true
	if len(fixWindingArr) > 0 {
		fixWinding = fixWindingArr[0]
	}

	polygons, err := fixPolygonToList(poly, fixWinding)
	if err != nil {
		return nil, err
	}

	if len(polygons) == 1 {
		polygon := polygons[0]
		if xy.IsRingCounterClockwise(polygon.Layout(), polygon.FlatCoords()) {
			return polygon, nil
		} else {
			return geom.NewPolygon(polygon.Layout()).MustSetCoords(
				[][]geom.Coord{
					{{-180, 90}, {-180, -90}, {180, -90}, {180, 90}},
					polygon.LinearRing(0).Coords(),
				},
			), nil
		}
	}

	// more than one polygon was returned which means we should return a
	// multipolygon

	multiPolygon := geom.NewMultiPolygon(poly.Layout())
	for _, polygon := range polygons {
		multiPolygon.Push(polygon)
	}

	return multiPolygon, nil
}

func fixPolygonToList(poly *geom.Polygon, shouldFixWinding bool) ([]*geom.Polygon, error) {
	if poly.Layout() != geom.XY && poly.Layout() != geom.XYZ {
		return nil, ErrUnsupportedLayout
	}

	var (
		polygons  []*geom.Polygon
		interiors = make([][]geom.Coord, 0)
	)

	numCoords := 2
	if poly.Layout() == geom.XYZ {
		numCoords = 3
	}

	exterior := normalize(poly.LinearRing(0).Coords())
	segments := segment(exterior)

	if len(segments) == 0 {
		if shouldFixWinding {
			correctlyWoundPolygon := fixWinding(poly)
			polygons = append(polygons, correctlyWoundPolygon)
		} else {
			polygons = append(polygons, poly)
		}

		return polygons, nil
	} else {
		for idx := range poly.NumLinearRings() - 1 {
			interior := poly.LinearRing(idx + 1)
			interiorSegments := segment(interior.Coords())
			if len(interiorSegments) > 0 {
				if shouldFixWinding {
					flatCoords := make([]float64, 0, len(interior.Coords())*numCoords)

					// unwrap coordinates
					for _, coord := range interior.Coords() {
						flatCoords = append(flatCoords, math.Mod(coord[0], 360))
						flatCoords = append(flatCoords, coord[1])
						if numCoords == 3 {
							flatCoords = append(flatCoords, coord[2])
						}
					}

					// if the interior ring is counter-clockwise, make it clockwise
					if xy.IsRingCounterClockwise(poly.Layout(), flatCoords) {
						coords := make([]geom.Coord, 0, len(interior.Coords()))
						for idx := range len(interior.Coords()) {
							switch poly.Layout() {
							case geom.XY:
								x := idx * 2
								y := x + 1
								coords = append(coords, geom.Coord{flatCoords[x], flatCoords[y]})
							case geom.XYZ:
								x := idx * 3
								y := x + 1
								z := y + 1
								coords = append(coords, geom.Coord{flatCoords[x], flatCoords[y], flatCoords[z]})
							default:
								return nil, ErrUnsupportedLayout
							}
						}

						slices.Reverse(coords)
						interiorSegments = segment(coords)
					}
				}

				segments = append(segments, interiorSegments...)
			} else {
				interiors = append(interiors, interior.Coords())
			}
		}
	}

	segments = extendOverPoles(segments, shouldFixWinding)
	polygons = buildPolygons(poly.Layout(), segments)

	// add interiors to the correct polygons
	for _, polygon := range polygons {
		remaining := make([][]geom.Coord, 0, len(interiors))
		for _, interior := range interiors {
			interiorPolygon := geom.NewPolygon(poly.Layout()).MustSetCoords([][]geom.Coord{interior})
			if Contains(interiorPolygon, polygon) {
				polygon.Push(geom.NewLinearRing(polygon.Layout()).MustSetCoords(interior))
			} else {
				remaining = append(remaining, interior)
			}
		}

		interiors = remaining
	}

	return polygons, nil
}

// fixWinding ensures that the exterior ring of the polygon is wound
// counter-clockwise and all interior rings are wound clockwise
func fixWinding(poly *geom.Polygon) *geom.Polygon {
	fixed := geom.NewPolygon(poly.Layout())
	if !xy.IsRingCounterClockwise(poly.Layout(), poly.LinearRing(0).FlatCoords()) {
		// exterior ring should be wound counter-clockwise
		rewoundCoords := poly.LinearRing(0).Coords()
		slices.Reverse(rewoundCoords)
		fixed.MustSetCoords([][]geom.Coord{rewoundCoords})
	}

	// all interior rings should be wound clockwise
	for idx := range poly.NumLinearRings() - 1 {
		interior := poly.LinearRing(idx + 1)
		if xy.IsRingCounterClockwise(poly.Layout(), interior.FlatCoords()) {
			coords := interior.Coords()
			slices.Reverse(coords)
			ring := geom.NewLinearRing(poly.Layout()).MustSetCoords(coords)
			fixed.Push(ring)
		}
	}

	return fixed
}

func segment(coords []geom.Coord) [][]geom.Coord {
	currSegment := make([]geom.Coord, 0)
	segments := make([][]geom.Coord, 0)

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

func extendOverPoles(segments [][]geom.Coord, shouldFixWinding bool) [][]geom.Coord {
	leftStarts := make([]edge, 0)
	rightStarts := make([]edge, 0)
	leftEnds := make([]edge, 0)
	rightEnds := make([]edge, 0)

	for idx, segment := range segments {
		if segment[0][0] == -180 {
			leftStarts = append(leftStarts, edge{Index: idx, Val: segment[0][1]})
		} else {
			rightStarts = append(rightStarts, edge{Index: idx, Val: segment[0][1]})
		}

		if segment[len(segment)-1][0] == -180 {
			leftEnds = append(leftEnds, edge{Index: idx, Val: segment[len(segment)-1][1]})
		} else {
			rightEnds = append(rightEnds, edge{Index: idx, Val: segment[len(segment)-1][1]})
		}
	}

	slices.SortFunc(leftEnds, cmp)
	slices.SortFunc(leftStarts, cmp)
	slices.SortFunc(rightEnds, cmpReverse)
	slices.SortFunc(rightStarts, cmpReverse)

	isOverNorthPole := false
	isOverSouthPole := false

	// deep copy segments
	originalSegments := make([][]geom.Coord, len(segments))
	for ii, sub := range segments {
		originalSegments[ii] = make([]geom.Coord, len(sub))
		for jj, val := range sub {
			originalSegments[ii][jj] = val.Clone()
		}
	}

	// If there's no segment ends between a start and the pole, extend the
	// segment over the pole.
	if len(leftEnds) > 0 && (len(leftStarts) == 0 || leftEnds[0].Val < leftStarts[0].Val) {
		isOverSouthPole = true
		segments[leftEnds[0].Index] = append(segments[leftEnds[0].Index], geom.Coord{-180, -90}, geom.Coord{180, -90})
	}

	if len(rightEnds) > 0 && (len(rightStarts) == 0 || rightEnds[0].Val > rightStarts[0].Val) {
		isOverNorthPole = true
		segments[rightEnds[0].Index] = append(segments[rightEnds[0].Index], geom.Coord{180, 90}, geom.Coord{-180, 90})
	}

	if shouldFixWinding && isOverNorthPole && isOverSouthPole {
		// If we're over both poles reverse all segments, effectively
		// reversing the winding order.
		for _, segment := range originalSegments {
			slices.Reverse(segment)
		}

		return originalSegments
	}

	return segments
}

func buildPolygons(layout geom.Layout, segments [][]geom.Coord) []*geom.Polygon {
	if len(segments) == 0 {
		return []*geom.Polygon{}
	}

	// pop last segment off list
	segment := segments[len(segments)-1]
	segments = segments[:len(segments)-1]

	segmentEnd := segment[len(segment)-1]
	isRight := segmentEnd[0] == 180

	candidates := make([]edge, 0)
	if isSelfClosing(segment) {
		// Self-closing segments might end up joining up with themselves. They
		// might not, e.g. donuts.
		candidates = append(candidates, edge{Index: -1, Val: segment[0][1]})
	}

	for idx, s0 := range segments {
		// Is the start of s0 on the same side as the end of segment?
		if s0[0][0] == segment[len(segment)-1][0] {
			// If so, check the following:
			// - Is the start of s0 closer to the pole than the end of segment, and
			// - is the end of s0 on the other side, or
			// - is the end of s0 further away from the pole than the start of
			//   segment (e.g. donuts)?

			startCloserToNorthPole := s0[0][1] > segmentEnd[1]
			startCloserToSouthPole := s0[0][1] < segmentEnd[1]

			s0End := s0[len(s0)-1]
			endFurtherFromNorthPole := s0End[1] < segment[0][1]
			endFurtherFromSouthPole := s0End[1] > segment[0][1]

			if (isRight && startCloserToNorthPole && (!isSelfClosing(s0) || endFurtherFromNorthPole)) ||
				(!isRight && startCloserToSouthPole && (!isSelfClosing(s0) || endFurtherFromSouthPole)) {
				candidates = append(candidates, edge{Index: idx, Val: s0[0][1]})
			}
		}
	}

	// Sort the candidates so the closest point is first in the list.
	slices.SortFunc(candidates, cmp)
	if !isRight {
		slices.Reverse(candidates)
	}

	index := 0
	if len(candidates) > 0 {
		index = candidates[0].Index
	} else {
		index = -1
	}

	if index > -1 {
		// Join the segments, then re-add them to the list and recurse.
		segment = append(segment, segments[index]...)
		segments = slices.Delete(segments, index, index+1)
		segments = append(segments, segment)

		return buildPolygons(layout, segments)
	} else {
		// This segment should be self-joining, so just build the rest of the
		// polygons without it.
		polygons := buildPolygons(layout, segments)

		// If every point is the same, then we don't need it in the output
		// set of polygons. This happens if, e.g., one corner of an input
		// polygon is on the antimeridian.
		allEqual := true
		for _, pt := range segment {
			allEqual = allEqual && (pt.Equal(layout, segment[0]))
		}

		if !allEqual {
			// if the last element does not equal the first of the polygon
			// close the polygon
			first := segment[0]
			last := segment[len(segment)-1]

			if !first.Equal(layout, last) {
				segment = append(segment, first.Clone())
			}

			polygon := geom.NewPolygon(layout).MustSetCoords([][]geom.Coord{segment})
			polygons = append(polygons, polygon)
		}

		return polygons
	}
}

func isSelfClosing(segment []geom.Coord) bool {
	segmentEnd := segment[len(segment)-1]
	isRight := segmentEnd[0] == 180
	return segment[0][0] == segmentEnd[0] &&
		((isRight && segment[0][1] > segmentEnd[1]) ||
			(!isRight && segment[0][1] < segmentEnd[1]))
}

func normalize(coords []geom.Coord) []geom.Coord {
	// make a copy of the original coordinates
	original := make([]geom.Coord, len(coords))
	for idx, v := range coords {
		original[idx] = v.Clone()
	}

	allAreOnAntiMeridian := true
	// Ensure all longitudes are between -180 and 180, and that tiny floating
	// point differences are ignored
	tol := 1e-08
	for idx, point := range coords {
		switch {
		case math.Abs(point[0]-180.0) <= tol:
			wrappedPrevIdx := int(math.Mod(float64(idx-1), float64(len(coords))))
			if math.Abs(point[1]) != 90 && math.Abs(coords[wrappedPrevIdx][0]+180) <= tol {
				coords[idx] = geom.Coord{-180.0, point[1]}
			} else {
				coords[idx] = geom.Coord{180.0, point[1]}
			}
		case math.Abs(point[0]+180) <= tol:
			wrappedPrevIdx := int(math.Mod(float64(idx-1), float64(len(coords))))
			if math.Abs(point[1]) != 90 && math.Abs(coords[wrappedPrevIdx][0]-180) <= tol {
				coords[idx] = geom.Coord{180.0, point[1]}
			} else {
				coords[idx] = geom.Coord{-180.0, point[1]}
			}
		default:
			coords[idx] = geom.Coord{math.Mod(point[0]+180.0, 360.0) - 180.0, point[1]}
			allAreOnAntiMeridian = false
		}
	}

	if allAreOnAntiMeridian {
		return original
	}

	return coords
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func cmp(a edge, b edge) int {
	if a.Val < b.Val {
		return -1
	} else if a.Val > b.Val {
		return 1
	}

	return 0
}

func cmpReverse(a edge, b edge) int {
	if a.Val > b.Val {
		return -1
	} else if a.Val < b.Val {
		return 1
	}

	return 0
}
