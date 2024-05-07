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

func cutMultiPolygon(multiPoly *geom.MultiPolygon, fixWindingArr ...bool) (*geom.MultiPolygon, error) {
	fixWinding := true
	if len(fixWindingArr) > 0 {
		fixWinding = fixWindingArr[0]
	}

	multiPolygon := geom.NewMultiPolygon(multiPoly.Layout())

	for idx := range multiPoly.NumPolygons() {
		poly := multiPoly.Polygon(idx)
		fixedPolys, err := fixPolygonToList(poly, fixWinding)
		if err != nil {
			return nil, err
		}

		for _, polygon := range fixedPolys {
			multiPolygon.Push(polygon)
		}
	}

	return multiPolygon, nil
}
