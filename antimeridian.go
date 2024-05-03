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

	"github.com/twpayne/go-geom"
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
	return poly, nil
}

func cutMultiPolygon(poly *geom.MultiPolygon) (*geom.MultiPolygon, error) {
	return poly, nil
}
