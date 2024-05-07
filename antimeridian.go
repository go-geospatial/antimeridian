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
	ErrUnsupportedType   = errors.New("unsupported geometry type")
	ErrUnsupportedLayout = errors.New("unsupported geometry layout")
)

// Cut divides a geometry at the antimeridian and the poles. A multi-geometry is
// returned with the cut portions of the original geometry. If no cuts are
// necessary Cut will return the original geometry with the winding normalized.
//
// By default Cut attempts to fix improperly wound geometries; howevever, there
// are instances where the polygon may be correctly wound but antimeridian
// cannot determine this to be so; for example, when the polygon extends over
// both the north and south pole. For these instances, pass fixWinding = false
func Cut(obj geom.T, fixWinding ...bool) (geom.T, error) {
	switch geometry := obj.(type) {
	case *geom.Polygon:
		return cutPolygon(geometry, fixWinding...)
	case *geom.MultiPolygon:
		return cutMultiPolygon(geometry, fixWinding...)
	default:
		// unsupported type
		return obj, ErrUnsupportedType
	}
}
