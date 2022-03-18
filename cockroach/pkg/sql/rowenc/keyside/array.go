// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keyside

import (
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/types"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
)

// encodeArrayKey generates an ordered key encoding of an array.
// The encoding format for an array [a, b] is as follows:
// [arrayMarker, enc(a), enc(b), terminator].
// The terminator is guaranteed to be less than all encoded values,
// so two arrays with the same prefix but different lengths will sort
// correctly. The key difference is that NULL values need to be encoded
// differently, because the standard NULL encoding conflicts with the
// terminator byte. This NULL value is chosen to be larger than the
// terminator but less than all existing encoded values.
func encodeArrayKey(b []byte, array *tree.DArray, dir encoding.Direction) ([]byte, error) {
	var err error
	b = encoding.EncodeArrayKeyMarker(b, dir)
	for _, elem := range array.Array {
		if elem == tree.DNull {
			b = encoding.EncodeNullWithinArrayKey(b, dir)
		} else {
			b, err = Encode(b, elem, dir)
			if err != nil {
				return nil, err
			}
		}
	}
	return encoding.EncodeArrayKeyTerminator(b, dir), nil
}

// decodeArrayKey decodes an array key generated by encodeArrayKey.
func decodeArrayKey(
	a *tree.DatumAlloc, t *types.T, buf []byte, dir encoding.Direction,
) (tree.Datum, []byte, error) {
	var err error
	buf, err = encoding.ValidateAndConsumeArrayKeyMarker(buf, dir)
	if err != nil {
		return nil, nil, err
	}

	result := tree.NewDArray(t.ArrayContents())

	for {
		if len(buf) == 0 {
			return nil, nil, errors.AssertionFailedf("invalid array encoding (unterminated)")
		}
		if encoding.IsArrayKeyDone(buf, dir) {
			buf = buf[1:]
			break
		}
		var d tree.Datum
		if encoding.IsNextByteArrayEncodedNull(buf, dir) {
			d = tree.DNull
			buf = buf[1:]
		} else {
			d, buf, err = Decode(a, t.ArrayContents(), buf, dir)
			if err != nil {
				return nil, nil, err
			}
		}
		if err := result.Append(d); err != nil {
			return nil, nil, err
		}
	}
	return result, buf, nil
}
