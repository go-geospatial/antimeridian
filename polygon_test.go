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
	func(inFile, outFile string, fixWinding bool) {
		if outFile == "" {
			outFile = inFile
		}

		inp, err := os.ReadFile(fmt.Sprintf("test_data/input/%s.json", inFile))
		Expect(err).To(BeNil())

		out, err := os.ReadFile(fmt.Sprintf("test_data/output/%s.json", outFile))
		Expect(err).To(BeNil())

		var inGeom geom.T
		err = geojson.Unmarshal(inp, &inGeom)
		Expect(err).To(BeNil())

		var outGeom geom.T
		err = geojson.Unmarshal(out, &outGeom)
		Expect(err).To(BeNil())

		result, err := antimeridian.Cut(inGeom, fixWinding)
		Expect(err).To(BeNil())

		resCoords := result.FlatCoords()
		outCoords := outGeom.FlatCoords()

		Expect(resCoords).To(HaveLen(len(outCoords)))

		for idx := range resCoords {
			Expect(resCoords[idx]).To(BeNumerically("~", outCoords[idx], .0000001), fmt.Sprintf("idx %d\nres = %+v\nexp = %+v", idx, resCoords, outCoords))
		}
	},
	Entry("almost touching 180", "almost-180", "almost-180", true),
	Entry("fix winding both poles", "both-poles", "both-poles-fix-winding", true),
	Entry("both poles", "both-poles", "both-poles", false),
	Entry("complex split", "complex-split", "complex-split", true),
	Entry("crossing latitude", "crossing-latitude", "crossing-latitude", true),
	Entry("cw only", "cw-only", "cw-only", true),
	Entry("cw split", "cw-split", "cw-split", true),
	Entry("extra crossing", "extra-crossing", "extra-crossing", true),
	Entry("latitude band", "latitude-band", "latitude-band", true),
	Entry("north pole", "north-pole", "north-pole", true),
	Entry("one ccw hole", "one-ccw-hole", "one-ccw-hole", true),
	Entry("one hole", "one-hole", "one-hole", true),
	Entry("over 180", "over-180", "over-180", true),
	Entry("overlap", "overlap", "overlap", true),
	Entry("point on antimeridian", "point-on-antimeridian", "point-on-antimeridian", true),
	Entry("simple with ccw hole", "simple-with-ccw-hole", "simple-with-ccw-hole", true),
	Entry("simple", "simple", "simple", true),
	Entry("south pole", "south-pole", "south-pole", true),
	Entry("split", "split", "split", true),
	Entry("two holes", "two-holes", "two-holes", true),
)
