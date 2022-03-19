// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package geo contains the base types for spatial data type operations.
package geo

import (
	"encoding/binary"
	"github.com/cockroachdb/errors"
	"github.com/golang/geo/s2"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

// DefaultEWKBEncodingFormat is the default encoding format for EWKB.
var DefaultEWKBEncodingFormat = binary.LittleEndian

// EmptyBehavior is the behavior to adopt when an empty Geometry is encountered.
type EmptyBehavior uint8

const (
	// EmptyBehaviorError will error with EmptyGeometryError when an empty geometry
	// is encountered.
	EmptyBehaviorError EmptyBehavior = 0
	// EmptyBehaviorOmit will omit an entry when an empty geometry is encountered.
	EmptyBehaviorOmit EmptyBehavior = 1
)

// FnExclusivity is used to indicate whether a geo function should have
// inclusive or exclusive semantics. For example, DWithin == (Distance <= x),
// while DWithinExclusive == (Distance < x).
type FnExclusivity bool

// MaxAllowedSplitPoints is the maximum number of points any spatial function may split to.
const MaxAllowedSplitPoints = 65336

const (
	// FnExclusive indicates that the corresponding geo function should have
	// exclusive semantics.
	FnExclusive FnExclusivity = true
	// FnInclusive indicates that the corresponding geo function should have
	// inclusive semantics.
	FnInclusive FnExclusivity = false
)

// SpatialObjectFitsColumnMetadata determines whether a GeospatialType is compatible with the
// given SRID and Shape.


//
// Geometry
//

// Geometry is planar spatial object.
type Geometry struct {
}











// AsGeomT returns the geometry as a geom.T object.
func (g *Geometry) AsGeomT() (geom.T, error) {
	return ewkb.Unmarshal(nil)
}

// Empty returns whether the given Geometry is empty.
func (g *Geometry) Empty() bool {
	return true
}


// SpatialObject returns the SpatialObject representation of the Geometry.


// EWKBHex returns the EWKBHex representation of the Geometry.
func (g *Geometry) EWKBHex() string {
	return ""
}



// SpaceCurveIndex returns an uint64 index to use representing an index into a space-filling curve.
// This will return 0 for empty spatial objects, and math.MaxUint64 for any object outside
// the defined bounds of the given SRID projection.
func (g *Geometry) SpaceCurveIndex() (uint64, error) {

	return 0, nil
}

// Compare compares a Geometry against another.
// It compares using SpaceCurveIndex, followed by the byte representation of the Geometry.
// This must produce the same ordering as the index mechanism.
func (g *Geometry) Compare(o Geometry) int {
	lhs, err := g.SpaceCurveIndex()
	if err != nil {
		// We should always be able to compare a valid geometry.
		panic(err)
	}
	rhs, err := o.SpaceCurveIndex()
	if err != nil {
		// We should always be able to compare a valid geometry.
		panic(err)
	}
	if lhs > rhs {
		return 1
	}
	if lhs < rhs {
		return -1
	}
	return 0
}

//
// Geography
//

// Geography is a spherical spatial object.
type Geography struct {
}














// AsGeomT returns the Geography as a geom.T object.
func (g *Geography) AsGeomT() (geom.T, error) {
	return ewkb.Unmarshal(nil)
}


// SpatialObject returns the SpatialObject representation of the Geography.



// EWKBHex returns the EWKBHex representation of the Geography.
func (g *Geography) EWKBHex() string {
	return ""
}


// AsS2 converts a given Geography into it's S2 form.
func (g *Geography) AsS2(emptyBehavior EmptyBehavior) ([]s2.Region, error) {
	geomRepr, err := g.AsGeomT()
	if err != nil {
		return nil, err
	}
	// TODO(otan): convert by reading from EWKB to S2 directly.
	return S2RegionsFromGeomT(geomRepr, emptyBehavior)
}



// SpaceCurveIndex returns an uint64 index to use representing an index into a space-filling curve.
// This will return 0 for empty spatial objects.
func (g *Geography) SpaceCurveIndex() uint64 {

	return 0
}

// Compare compares a Geography against another.
// It compares using SpaceCurveIndex, followed by the byte representation of the Geography.

//
// Common
//

// AdjustGeomTSRID adjusts the SRID of a given geom.T.
// Ideally SetSRID is an interface of geom.T, but that is not the case.
func AdjustGeomTSRID(t geom.T, srid int32) {
	switch t := t.(type) {
	case *geom.Point:
		t.SetSRID(int(srid))
	case *geom.LineString:
		t.SetSRID(int(srid))
	case *geom.Polygon:
		t.SetSRID(int(srid))
	case *geom.GeometryCollection:
		t.SetSRID(int(srid))
	case *geom.MultiPoint:
		t.SetSRID(int(srid))
	case *geom.MultiLineString:
		t.SetSRID(int(srid))
	case *geom.MultiPolygon:
		t.SetSRID(int(srid))
	default:
		panic(errors.AssertionFailedf("geo: unknown geom type: %v", t))
	}
}

// IsLinearRingCCW returns whether a given linear ring is counter clock wise.
// See 2.07 of http://www.faqs.org/faqs/graphics/algorithms-faq/.
// "Find the lowest vertex (or, if  there is more than one vertex with the same lowest coordinate,
//  the rightmost of those vertices) and then take the cross product of the edges fore and aft of it."
func IsLinearRingCCW(linearRing *geom.LinearRing) bool {
	smallestIdx := 0
	smallest := linearRing.Coord(0)

	for pointIdx := 1; pointIdx < linearRing.NumCoords()-1; pointIdx++ {
		curr := linearRing.Coord(pointIdx)
		if curr.Y() < smallest.Y() || (curr.Y() == smallest.Y() && curr.X() > smallest.X()) {
			smallestIdx = pointIdx
			smallest = curr
		}
	}

	// Find the previous point in the ring that is not the same as smallest.
	prevIdx := smallestIdx - 1
	if prevIdx < 0 {
		prevIdx = linearRing.NumCoords() - 1
	}
	for prevIdx != smallestIdx {
		a := linearRing.Coord(prevIdx)
		if a.X() != smallest.X() || a.Y() != smallest.Y() {
			break
		}
		prevIdx--
		if prevIdx < 0 {
			prevIdx = linearRing.NumCoords() - 1
		}
	}
	// Find the next point in the ring that is not the same as smallest.
	nextIdx := smallestIdx + 1
	if nextIdx >= linearRing.NumCoords() {
		nextIdx = 0
	}
	for nextIdx != smallestIdx {
		c := linearRing.Coord(nextIdx)
		if c.X() != smallest.X() || c.Y() != smallest.Y() {
			break
		}
		nextIdx++
		if nextIdx >= linearRing.NumCoords() {
			nextIdx = 0
		}
	}

	// We could do the cross product, but we are only interested in the sign.
	// To find the sign, reorganize into the orientation matrix:
	//  1 x_a y_a
	//  1 x_b y_b
	//  1 x_c y_c
	// and find the determinant.
	// https://en.wikipedia.org/wiki/Curve_orientation#Orientation_of_a_simple_polygon
	a := linearRing.Coord(prevIdx)
	b := smallest
	c := linearRing.Coord(nextIdx)

	areaSign := a.X()*b.Y() - a.Y()*b.X() +
		a.Y()*c.X() - a.X()*c.Y() +
		b.X()*c.Y() - c.X()*b.Y()
	// Note having an area sign of 0 means it is a flat polygon, which is invalid.
	return areaSign > 0
}

// S2RegionsFromGeomT converts an geom representation of an object
// to s2 regions.
// As S2 does not really handle empty geometries well, we need to ingest emptyBehavior and
// react appropriately.
func S2RegionsFromGeomT(geomRepr geom.T, emptyBehavior EmptyBehavior) ([]s2.Region, error) {
	var regions []s2.Region
	if geomRepr.Empty() {
		switch emptyBehavior {
		case EmptyBehaviorOmit:
			return nil, nil
		case EmptyBehaviorError:
			return nil, NewEmptyGeometryError()
		default:
			return nil, errors.AssertionFailedf("programmer error: unknown behavior")
		}
	}
	switch repr := geomRepr.(type) {
	case *geom.Point:
		regions = []s2.Region{
			s2.PointFromLatLng(s2.LatLngFromDegrees(repr.Y(), repr.X())),
		}
	case *geom.LineString:
		latLngs := make([]s2.LatLng, repr.NumCoords())
		for i := 0; i < repr.NumCoords(); i++ {
			p := repr.Coord(i)
			latLngs[i] = s2.LatLngFromDegrees(p.Y(), p.X())
		}
		regions = []s2.Region{
			s2.PolylineFromLatLngs(latLngs),
		}
	case *geom.Polygon:
		loops := make([]*s2.Loop, repr.NumLinearRings())
		// All loops must be oriented CCW for S2.
		for ringIdx := 0; ringIdx < repr.NumLinearRings(); ringIdx++ {
			linearRing := repr.LinearRing(ringIdx)
			points := make([]s2.Point, linearRing.NumCoords())
			isCCW := IsLinearRingCCW(linearRing)
			for pointIdx := 0; pointIdx < linearRing.NumCoords(); pointIdx++ {
				p := linearRing.Coord(pointIdx)
				pt := s2.PointFromLatLng(s2.LatLngFromDegrees(p.Y(), p.X()))
				if isCCW {
					points[pointIdx] = pt
				} else {
					points[len(points)-pointIdx-1] = pt
				}
			}
			loops[ringIdx] = s2.LoopFromPoints(points)
		}
		regions = []s2.Region{
			s2.PolygonFromLoops(loops),
		}
	case *geom.GeometryCollection:
		for _, geom := range repr.Geoms() {
			subRegions, err := S2RegionsFromGeomT(geom, emptyBehavior)
			if err != nil {
				return nil, err
			}
			regions = append(regions, subRegions...)
		}
	case *geom.MultiPoint:
		for i := 0; i < repr.NumPoints(); i++ {
			subRegions, err := S2RegionsFromGeomT(repr.Point(i), emptyBehavior)
			if err != nil {
				return nil, err
			}
			regions = append(regions, subRegions...)
		}
	case *geom.MultiLineString:
		for i := 0; i < repr.NumLineStrings(); i++ {
			subRegions, err := S2RegionsFromGeomT(repr.LineString(i), emptyBehavior)
			if err != nil {
				return nil, err
			}
			regions = append(regions, subRegions...)
		}
	case *geom.MultiPolygon:
		for i := 0; i < repr.NumPolygons(); i++ {
			subRegions, err := S2RegionsFromGeomT(repr.Polygon(i), emptyBehavior)
			if err != nil {
				return nil, err
			}
			regions = append(regions, subRegions...)
		}
	}
	return regions, nil
}

// normalizeLngLat normalizes geographical coordinates into a valid range.
func normalizeLngLat(lng float64, lat float64) (float64, float64) {
	if lat > 90 || lat < -90 {
		lat = NormalizeLatitudeDegrees(lat)
	}
	if lng > 180 || lng < -180 {
		lng = NormalizeLongitudeDegrees(lng)
	}
	return lng, lat
}

// normalizeGeographyGeomT limits geography coordinates to spherical coordinates
// by converting geom.T coordinates inplace
func normalizeGeographyGeomT(t geom.T) {
	switch repr := t.(type) {
	case *geom.GeometryCollection:
		for _, geom := range repr.Geoms() {
			normalizeGeographyGeomT(geom)
		}
	default:
		coords := repr.FlatCoords()
		for i := 0; i < len(coords); i += repr.Stride() {
			coords[i], coords[i+1] = normalizeLngLat(coords[i], coords[i+1])
		}
	}
}

// validateGeomT validates the geom.T object across valid geom.T objects,
// returning an error if it is invalid.
func validateGeomT(t geom.T) error {
	if t.Empty() {
		return nil
	}
	switch t := t.(type) {
	case *geom.Point:
	case *geom.LineString:
		if t.NumCoords() < 2 {
			return pgerror.Newf(
				pgcode.InvalidParameterValue,
				"LineString must have at least 2 coordinates",
			)
		}
	case *geom.Polygon:
		for i := 0; i < t.NumLinearRings(); i++ {
			linearRing := t.LinearRing(i)
			if linearRing.NumCoords() < 4 {
				return pgerror.Newf(
					pgcode.InvalidParameterValue,
					"Polygon LinearRing must have at least 4 points, found %d at position %d",
					linearRing.NumCoords(),
					i+1,
				)
			}
			if !linearRing.Coord(0).Equal(linearRing.Layout(), linearRing.Coord(linearRing.NumCoords()-1)) {
				return pgerror.Newf(
					pgcode.InvalidParameterValue,
					"Polygon LinearRing at position %d is not closed",
					i+1,
				)
			}
		}
	case *geom.MultiPoint:
	case *geom.MultiLineString:
		for i := 0; i < t.NumLineStrings(); i++ {
			if err := validateGeomT(t.LineString(i)); err != nil {
				return errors.Wrapf(err, "invalid MultiLineString component at position %d", i+1)
			}
		}
	case *geom.MultiPolygon:
		for i := 0; i < t.NumPolygons(); i++ {
			if err := validateGeomT(t.Polygon(i)); err != nil {
				return errors.Wrapf(err, "invalid MultiPolygon component at position %d", i+1)
			}
		}
	case *geom.GeometryCollection:
		// TODO(ayang): verify that the geometries all have the same Layout
		for i := 0; i < t.NumGeoms(); i++ {
			if err := validateGeomT(t.Geom(i)); err != nil {
				return errors.Wrapf(err, "invalid GeometryCollection component at position %d", i+1)
			}
		}
	default:
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"unknown geom.T type: %T",
			t,
		)
	}
	return nil
}
