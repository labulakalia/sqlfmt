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
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util"
	"github.com/pierrre/geohash"
)



const sridPrefix = "SRID="
const sridPrefixLen = len(sridPrefix)

type defaultSRIDOverwriteSetting bool

const (
	// DefaultSRIDShouldOverwrite implies the parsing function should overwrite
	// the SRID with the defaultSRID.
	DefaultSRIDShouldOverwrite defaultSRIDOverwriteSetting = true
	// DefaultSRIDIsHint implies that the default SRID is only a hint
	// and if the SRID is provided by the given EWKT/EWKB, it should be
	// used instead.
	DefaultSRIDIsHint defaultSRIDOverwriteSetting = false
)


// hasPrefixIgnoreCase returns whether a given str begins with a prefix, ignoring case.
// It assumes that the string and prefix contains only ASCII bytes.
func hasPrefixIgnoreCase(str string, prefix string) bool {
	if len(str) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if util.ToLowerSingleByte(str[i]) != util.ToLowerSingleByte(prefix[i]) {
			return false
		}
	}
	return true
}

// ParseGeometryPointFromGeoHash converts a GeoHash to a Geometry Point




func parseGeoHash(g string, precision int) (geohash.Box, error) {
	if len(g) == 0 {
		return geohash.Box{}, pgerror.Newf(pgcode.InvalidParameterValue, "length of GeoHash must be greater than 0")
	}

	// In PostGIS the parsing is case-insensitive.
	g = strings.ToLower(g)

	// If precision is more than the length of the geohash
	// or if precision is less than 0 then set
	// precision equal to length of geohash.
	if precision > len(g) || precision < 0 {
		precision = len(g)
	}
	box, err := geohash.Decode(g[:precision])
	if err != nil {
		return geohash.Box{}, pgerror.Wrap(err, pgcode.InvalidParameterValue, "invalid GeoHash")
	}
	return box, nil
}

// GeometryToEncodedPolyline turns the provided geometry and precision into a Polyline ASCII
func GeometryToEncodedPolyline(g Geometry, p int) (string, error) {
	gt, err := g.AsGeomT()
	if err != nil {
		return "", errors.Wrap(err, "error parsing input geometry")
	}
	if gt.SRID() != 4326 {
		return "", pgerror.Newf(pgcode.InvalidParameterValue, "only SRID 4326 is supported")
	}

	return encodePolylinePoints(gt.FlatCoords(), p), nil
}


