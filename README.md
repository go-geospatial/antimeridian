# antimeridian

[![PkgGoDev](https://pkg.go.dev/badge/github.com/go-geospatial/antimeridian)](https://pkg.go.dev/github.com/go-geospatial/antimeridian)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-geospatial/antimeridian)](https://goreportcard.com/report/github.com/go-geospatial/antimeridian)

Package `antimeridian` fixes shapes that cross the antimeridian.

![Example Polygon](docs/example-polygon.webp "Example Polygon")

Lines and Polygons that cross the antimeridian are problematic for many web mapping packages and PostGIS. The official [GeoJSON spec](https://datatracker.ietf.org/doc/html/rfc7946#section-3.1.9) recommends features that cross the antimeridian, "SHOULD be represented by cutting it in two such that neither part's representation crosses the antimeridian."

## Usage

```shell
go get github.com/go-geospatial/antimeridian
```

Then:

```go
fixedGeom := antimeridian.Cut(geomCrossingAntiMeridian)
```
