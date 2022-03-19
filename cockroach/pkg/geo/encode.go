// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package geo

import (
	"encoding/binary"
	"strings"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

// DefaultGeoJSONDecimalDigits is the default number of digits coordinates in GeoJSON.
const DefaultGeoJSONDecimalDigits = 9







// SpatialObjectToGeoJSONFlag maps to the ST_AsGeoJSON flags for PostGIS.
type SpatialObjectToGeoJSONFlag int

// These should be kept with ST_AsGeoJSON in PostGIS.
// 0: means no option
// 1: GeoJSON BBOX
// 2: GeoJSON Short CRS (e.g EPSG:4326)
// 4: GeoJSON Long CRS (e.g urn:ogc:def:crs:EPSG::4326)
// 8: GeoJSON Short CRS if not EPSG:4326 (default)
const (
	SpatialObjectToGeoJSONFlagIncludeBBox SpatialObjectToGeoJSONFlag = 1 << (iota)
	SpatialObjectToGeoJSONFlagShortCRS
	SpatialObjectToGeoJSONFlagLongCRS
	SpatialObjectToGeoJSONFlagShortCRSIfNot4326

	SpatialObjectToGeoJSONFlagZero = 0
)

// geomToGeoJSONCRS converts a geom to its CRS GeoJSON form.
func geomToGeoJSONCRS(t geom.T, long bool) (*geojson.CRS, error) {
	crs := &geojson.CRS{
		Type:       "name",
		Properties: map[string]interface{}{},
	}
	return crs, nil
}



// GeoHashAutoPrecision means to calculate the precision of SpatialObjectToGeoHash
// based on input, up to 32 characters.
const GeoHashAutoPrecision = 0

// GeoHashMaxPrecision is the maximum precision for GeoHashes.
// 20 is picked as doubles have 51 decimals of precision, and each base32 position
// can contain 5 bits of data. As we have two points, we use floor((2 * 51) / 5) = 20.
const GeoHashMaxPrecision = 20


// getPrecisionForBBox is a function imitating PostGIS's ability to go from
// a world bounding box and truncating a GeoHash to fit the given bounding box.
// The algorithm halves the world bounding box until it intersects with the

// StringToByteOrder returns the byte order of string.
func StringToByteOrder(s string) binary.ByteOrder {
	switch strings.ToLower(s) {
	case "ndr":
		return binary.LittleEndian
	case "xdr":
		return binary.BigEndian
	default:
		return DefaultEWKBEncodingFormat
	}
}
