// Code generated by execgen; DO NOT EDIT.

// Copyright 2019 The Cockroach Authors.
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexecproj

import (
	"bytes"
	"regexp"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/col/coldata"
)

type projPrefixBytesBytesConstOp struct {
	projConstOpBase
	constArg []byte
}

func (p projPrefixBytesBytesConstOp) Next() coldata.Batch {
	batch := p.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(p.colIdx)
	var col *coldata.Bytes
	col = vec.Bytes()
	projVec := batch.ColVec(p.outputIdx)
	p.allocator.PerformOperation([]coldata.Vec{projVec}, func() {
		// Capture col to force bounds check to work. See
		// https://github.com/golang/go/issues/39756
		col := col
		if projVec.MaybeHasNulls() {
			// We need to make sure that there are no left over null values in the
			// output vector.
			projVec.Nulls().UnsetNulls()
		}
		projCol := projVec.Bool()
		// Some operators can result in NULL with non-NULL inputs, like the JSON
		// fetch value operator, ->. Therefore, _outNulls is defined to allow
		// updating the output Nulls from within _ASSIGN functions when the result
		// of a projection is Null.
		_outNulls := projVec.Nulls()
		if vec.Nulls().MaybeHasNulls() {
			colNulls := vec.Nulls()
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = bytes.HasPrefix(arg, p.constArg)
					}
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = bytes.HasPrefix(arg, p.constArg)
					}
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
			projVec.SetNulls(_outNulls.Or(*colNulls))
		} else {
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					arg := col.Get(i)
					projCol[i] = bytes.HasPrefix(arg, p.constArg)
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					arg := col.Get(i)
					projCol[i] = bytes.HasPrefix(arg, p.constArg)
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
		}
	})
	return batch
}

type projSuffixBytesBytesConstOp struct {
	projConstOpBase
	constArg []byte
}

func (p projSuffixBytesBytesConstOp) Next() coldata.Batch {
	batch := p.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(p.colIdx)
	var col *coldata.Bytes
	col = vec.Bytes()
	projVec := batch.ColVec(p.outputIdx)
	p.allocator.PerformOperation([]coldata.Vec{projVec}, func() {
		// Capture col to force bounds check to work. See
		// https://github.com/golang/go/issues/39756
		col := col
		if projVec.MaybeHasNulls() {
			// We need to make sure that there are no left over null values in the
			// output vector.
			projVec.Nulls().UnsetNulls()
		}
		projCol := projVec.Bool()
		// Some operators can result in NULL with non-NULL inputs, like the JSON
		// fetch value operator, ->. Therefore, _outNulls is defined to allow
		// updating the output Nulls from within _ASSIGN functions when the result
		// of a projection is Null.
		_outNulls := projVec.Nulls()
		if vec.Nulls().MaybeHasNulls() {
			colNulls := vec.Nulls()
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = bytes.HasSuffix(arg, p.constArg)
					}
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = bytes.HasSuffix(arg, p.constArg)
					}
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
			projVec.SetNulls(_outNulls.Or(*colNulls))
		} else {
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					arg := col.Get(i)
					projCol[i] = bytes.HasSuffix(arg, p.constArg)
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					arg := col.Get(i)
					projCol[i] = bytes.HasSuffix(arg, p.constArg)
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
		}
	})
	return batch
}

type projContainsBytesBytesConstOp struct {
	projConstOpBase
	constArg []byte
}

func (p projContainsBytesBytesConstOp) Next() coldata.Batch {
	batch := p.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(p.colIdx)
	var col *coldata.Bytes
	col = vec.Bytes()
	projVec := batch.ColVec(p.outputIdx)
	p.allocator.PerformOperation([]coldata.Vec{projVec}, func() {
		// Capture col to force bounds check to work. See
		// https://github.com/golang/go/issues/39756
		col := col
		if projVec.MaybeHasNulls() {
			// We need to make sure that there are no left over null values in the
			// output vector.
			projVec.Nulls().UnsetNulls()
		}
		projCol := projVec.Bool()
		// Some operators can result in NULL with non-NULL inputs, like the JSON
		// fetch value operator, ->. Therefore, _outNulls is defined to allow
		// updating the output Nulls from within _ASSIGN functions when the result
		// of a projection is Null.
		_outNulls := projVec.Nulls()
		if vec.Nulls().MaybeHasNulls() {
			colNulls := vec.Nulls()
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = bytes.Contains(arg, p.constArg)
					}
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = bytes.Contains(arg, p.constArg)
					}
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
			projVec.SetNulls(_outNulls.Or(*colNulls))
		} else {
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					arg := col.Get(i)
					projCol[i] = bytes.Contains(arg, p.constArg)
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					arg := col.Get(i)
					projCol[i] = bytes.Contains(arg, p.constArg)
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
		}
	})
	return batch
}

type projRegexpBytesBytesConstOp struct {
	projConstOpBase
	constArg *regexp.Regexp
}

func (p projRegexpBytesBytesConstOp) Next() coldata.Batch {
	batch := p.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(p.colIdx)
	var col *coldata.Bytes
	col = vec.Bytes()
	projVec := batch.ColVec(p.outputIdx)
	p.allocator.PerformOperation([]coldata.Vec{projVec}, func() {
		// Capture col to force bounds check to work. See
		// https://github.com/golang/go/issues/39756
		col := col
		if projVec.MaybeHasNulls() {
			// We need to make sure that there are no left over null values in the
			// output vector.
			projVec.Nulls().UnsetNulls()
		}
		projCol := projVec.Bool()
		// Some operators can result in NULL with non-NULL inputs, like the JSON
		// fetch value operator, ->. Therefore, _outNulls is defined to allow
		// updating the output Nulls from within _ASSIGN functions when the result
		// of a projection is Null.
		_outNulls := projVec.Nulls()
		if vec.Nulls().MaybeHasNulls() {
			colNulls := vec.Nulls()
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = p.constArg.Match(arg)
					}
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = p.constArg.Match(arg)
					}
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
			projVec.SetNulls(_outNulls.Or(*colNulls))
		} else {
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					arg := col.Get(i)
					projCol[i] = p.constArg.Match(arg)
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					arg := col.Get(i)
					projCol[i] = p.constArg.Match(arg)
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
		}
	})
	return batch
}

type projNotPrefixBytesBytesConstOp struct {
	projConstOpBase
	constArg []byte
}

func (p projNotPrefixBytesBytesConstOp) Next() coldata.Batch {
	batch := p.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(p.colIdx)
	var col *coldata.Bytes
	col = vec.Bytes()
	projVec := batch.ColVec(p.outputIdx)
	p.allocator.PerformOperation([]coldata.Vec{projVec}, func() {
		// Capture col to force bounds check to work. See
		// https://github.com/golang/go/issues/39756
		col := col
		if projVec.MaybeHasNulls() {
			// We need to make sure that there are no left over null values in the
			// output vector.
			projVec.Nulls().UnsetNulls()
		}
		projCol := projVec.Bool()
		// Some operators can result in NULL with non-NULL inputs, like the JSON
		// fetch value operator, ->. Therefore, _outNulls is defined to allow
		// updating the output Nulls from within _ASSIGN functions when the result
		// of a projection is Null.
		_outNulls := projVec.Nulls()
		if vec.Nulls().MaybeHasNulls() {
			colNulls := vec.Nulls()
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = !bytes.HasPrefix(arg, p.constArg)
					}
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = !bytes.HasPrefix(arg, p.constArg)
					}
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
			projVec.SetNulls(_outNulls.Or(*colNulls))
		} else {
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					arg := col.Get(i)
					projCol[i] = !bytes.HasPrefix(arg, p.constArg)
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					arg := col.Get(i)
					projCol[i] = !bytes.HasPrefix(arg, p.constArg)
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
		}
	})
	return batch
}

type projNotSuffixBytesBytesConstOp struct {
	projConstOpBase
	constArg []byte
}

func (p projNotSuffixBytesBytesConstOp) Next() coldata.Batch {
	batch := p.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(p.colIdx)
	var col *coldata.Bytes
	col = vec.Bytes()
	projVec := batch.ColVec(p.outputIdx)
	p.allocator.PerformOperation([]coldata.Vec{projVec}, func() {
		// Capture col to force bounds check to work. See
		// https://github.com/golang/go/issues/39756
		col := col
		if projVec.MaybeHasNulls() {
			// We need to make sure that there are no left over null values in the
			// output vector.
			projVec.Nulls().UnsetNulls()
		}
		projCol := projVec.Bool()
		// Some operators can result in NULL with non-NULL inputs, like the JSON
		// fetch value operator, ->. Therefore, _outNulls is defined to allow
		// updating the output Nulls from within _ASSIGN functions when the result
		// of a projection is Null.
		_outNulls := projVec.Nulls()
		if vec.Nulls().MaybeHasNulls() {
			colNulls := vec.Nulls()
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = !bytes.HasSuffix(arg, p.constArg)
					}
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = !bytes.HasSuffix(arg, p.constArg)
					}
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
			projVec.SetNulls(_outNulls.Or(*colNulls))
		} else {
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					arg := col.Get(i)
					projCol[i] = !bytes.HasSuffix(arg, p.constArg)
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					arg := col.Get(i)
					projCol[i] = !bytes.HasSuffix(arg, p.constArg)
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
		}
	})
	return batch
}

type projNotContainsBytesBytesConstOp struct {
	projConstOpBase
	constArg []byte
}

func (p projNotContainsBytesBytesConstOp) Next() coldata.Batch {
	batch := p.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(p.colIdx)
	var col *coldata.Bytes
	col = vec.Bytes()
	projVec := batch.ColVec(p.outputIdx)
	p.allocator.PerformOperation([]coldata.Vec{projVec}, func() {
		// Capture col to force bounds check to work. See
		// https://github.com/golang/go/issues/39756
		col := col
		if projVec.MaybeHasNulls() {
			// We need to make sure that there are no left over null values in the
			// output vector.
			projVec.Nulls().UnsetNulls()
		}
		projCol := projVec.Bool()
		// Some operators can result in NULL with non-NULL inputs, like the JSON
		// fetch value operator, ->. Therefore, _outNulls is defined to allow
		// updating the output Nulls from within _ASSIGN functions when the result
		// of a projection is Null.
		_outNulls := projVec.Nulls()
		if vec.Nulls().MaybeHasNulls() {
			colNulls := vec.Nulls()
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = !bytes.Contains(arg, p.constArg)
					}
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = !bytes.Contains(arg, p.constArg)
					}
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
			projVec.SetNulls(_outNulls.Or(*colNulls))
		} else {
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					arg := col.Get(i)
					projCol[i] = !bytes.Contains(arg, p.constArg)
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					arg := col.Get(i)
					projCol[i] = !bytes.Contains(arg, p.constArg)
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
		}
	})
	return batch
}

type projNotRegexpBytesBytesConstOp struct {
	projConstOpBase
	constArg *regexp.Regexp
}

func (p projNotRegexpBytesBytesConstOp) Next() coldata.Batch {
	batch := p.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(p.colIdx)
	var col *coldata.Bytes
	col = vec.Bytes()
	projVec := batch.ColVec(p.outputIdx)
	p.allocator.PerformOperation([]coldata.Vec{projVec}, func() {
		// Capture col to force bounds check to work. See
		// https://github.com/golang/go/issues/39756
		col := col
		if projVec.MaybeHasNulls() {
			// We need to make sure that there are no left over null values in the
			// output vector.
			projVec.Nulls().UnsetNulls()
		}
		projCol := projVec.Bool()
		// Some operators can result in NULL with non-NULL inputs, like the JSON
		// fetch value operator, ->. Therefore, _outNulls is defined to allow
		// updating the output Nulls from within _ASSIGN functions when the result
		// of a projection is Null.
		_outNulls := projVec.Nulls()
		if vec.Nulls().MaybeHasNulls() {
			colNulls := vec.Nulls()
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = !p.constArg.Match(arg)
					}
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					if !colNulls.NullAt(i) {
						// We only want to perform the projection operation if the value is not null.
						arg := col.Get(i)
						projCol[i] = !p.constArg.Match(arg)
					}
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
			projVec.SetNulls(_outNulls.Or(*colNulls))
		} else {
			if sel := batch.Selection(); sel != nil {
				sel = sel[:n]
				for _, i := range sel {
					arg := col.Get(i)
					projCol[i] = !p.constArg.Match(arg)
				}
			} else {
				_ = projCol.Get(n - 1)
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					arg := col.Get(i)
					projCol[i] = !p.constArg.Match(arg)
				}
			}
			// _outNulls has been updated from within the _ASSIGN function to include
			// any NULLs that resulted from the projection.
			// If $hasNulls is true, union _outNulls with the set of input Nulls.
			// If $hasNulls is false, then there are no input Nulls. _outNulls is
			// projVec.Nulls() so there is no need to call projVec.SetNulls().
		}
	})
	return batch
}
