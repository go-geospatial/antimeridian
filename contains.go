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

import "github.com/twpayne/go-geom"

// Contains checks if polygon is contained by the polygon containedBy
func Contains(polygon *geom.Polygon, containedBy *geom.Polygon) bool {
	if polygon.NumLinearRings() == 0 || containedBy.NumLinearRings() == 0 {
		return false
	}

	// polygon is contained if all points fall within containedBy
	testPolygonExterior := polygon.LinearRing(0)
	containmentRing := containedBy.LinearRing(0)
	within := true
	for _, pt := range testPolygonExterior.Coords() {
		within = within && ContainsPoint(pt, containmentRing)
	}

	if within {
		// the point is within the exterior of the polygon
		// check if it falls within any holes in the interior
		for idx := range containedBy.NumLinearRings() - 1 {
			hole := containedBy.LinearRing(idx + 1)
			inHole := true
			for _, pt := range testPolygonExterior.Coords() {
				inHole = inHole && ContainsPoint(pt, hole)
			}

			if inHole {
				return false
			}
		}
	}

	return within
}

// ContainsPoint checks if the point pt is within the linear ring
func ContainsPoint(pt geom.Coord, ring *geom.LinearRing) bool {
	coords := ring.Coords()
	if len(coords) < 3 {
		return false
	}

	in := rayIntersectsSegment(pt, coords[len(coords)-1], coords[0])
	for ii := 1; ii < len(coords); ii++ {
		if rayIntersectsSegment(pt, coords[ii-1], coords[ii]) {
			in = !in
		}
	}

	return in
}

func rayIntersectsSegment(p, a, b geom.Coord) bool {
	return (a[1] > p[1]) != (b[1] > p[1]) &&
		p[0] < (b[0]-a[0])*(p[1]-a[1])/(b[1]-a[1])+a[0]
}
