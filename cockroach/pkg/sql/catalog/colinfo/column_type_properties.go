// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colinfo

import (
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/types"
)

// CheckDatumTypeFitsColumnType verifies that a given scalar value
// type is valid to be stored in a column of the given column type.
//
// For the purpose of this analysis, column type aliases are not
// considered to be different (eg. TEXT and VARCHAR will fit the same
// scalar type String).
//
// This is used by the UPDATE, INSERT and UPSERT code.
func CheckDatumTypeFitsColumnType(col catalog.Column, typ *types.T) error {
	if typ.Family() == types.UnknownFamily {
		return nil
	}
	if !typ.Equivalent(col.GetType()) {
		return pgerror.Newf(pgcode.DatatypeMismatch,
			"value type %s doesn't match type %s of column %q",
			typ.String(), col.GetType().String(), tree.ErrNameString(col.GetName()))
	}
	return nil
}

// CanHaveCompositeKeyEncoding returns true if key columns of the given kind can
// have a composite encoding. For such types, it can be decided on a
// case-by-base basis whether a given Datum requires the composite encoding.
//
// As an example of a composite encoding, collated string key columns are
// encoded partly as a key and partly as a value. The key part is the collation
// key, so that different strings that collate equal cannot both be used as
// keys. The value part is the usual UTF-8 encoding of the string, stored so
// that it can be recovered later for inspection/display.
func CanHaveCompositeKeyEncoding(typ *types.T) bool {
	switch typ.Family() {
	case types.FloatFamily,
		types.DecimalFamily,
		types.CollatedStringFamily:
		return true
	case types.ArrayFamily:
		return CanHaveCompositeKeyEncoding(typ.ArrayContents())
	case types.TupleFamily:
		for _, t := range typ.TupleContents() {
			if CanHaveCompositeKeyEncoding(t) {
				return true
			}
		}
		return false
	case types.BoolFamily,
		types.IntFamily,
		types.DateFamily,
		types.TimestampFamily,
		types.IntervalFamily,
		types.StringFamily,
		types.BytesFamily,
		types.TimestampTZFamily,
		types.OidFamily,
		types.UuidFamily,
		types.INetFamily,
		types.TimeFamily,
		types.JsonFamily,
		types.TimeTZFamily,
		types.BitFamily,
		types.GeometryFamily,
		types.GeographyFamily,
		types.EnumFamily,
		types.Box2DFamily:
		return false
	case types.UnknownFamily,
		types.AnyFamily:
		fallthrough
	default:
		return true
	}
}
