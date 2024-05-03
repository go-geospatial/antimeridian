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

package antimeridian_test

import (
	"fmt"
	"os"

	"github.com/go-geospatial/antimeridian"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

var _ = DescribeTable("Various Polygons",
	func(name string) {
		inp, err := os.ReadFile(fmt.Sprintf("test_data/input/%s.json", name))
		Expect(err).To(BeNil())

		out, err := os.ReadFile(fmt.Sprintf("test_data/output/%s.json", name))
		Expect(err).To(BeNil())

		var inGeom geom.T
		err = geojson.Unmarshal(inp, &inGeom)
		Expect(err).To(BeNil())

		var outGeom geom.T
		err = geojson.Unmarshal(out, &outGeom)
		Expect(err).To(BeNil())

		result, err := antimeridian.Cut(inGeom)
		Expect(err).To(BeNil())
		Expect(result).To(Equal(outGeom))
	},
	Entry("almost touching 180", "almost-180"),
)
