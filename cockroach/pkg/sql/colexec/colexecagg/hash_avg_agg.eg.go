// Code generated by execgen; DO NOT EDIT.
// Copyright 2018 The Cockroach Authors.
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexecagg

import (
	"unsafe"

	"github.com/cockroachdb/apd/v3"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/col/coldata"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexecerror"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colmem"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/types"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/duration"
	"github.com/cockroachdb/errors"
)

// Workaround for bazel auto-generated code. goimports does not automatically
// pick up the right packages when run within the bazel sandbox.
var (
	_ tree.AggType
	_ apd.Context
	_ duration.Duration
)

func newAvgHashAggAlloc(
	allocator *colmem.Allocator, t *types.T, allocSize int64,
) (aggregateFuncAlloc, error) {
	allocBase := aggAllocBase{allocator: allocator, allocSize: allocSize}
	switch t.Family() {
	case types.IntFamily:
		switch t.Width() {
		case 16:
			return &avgInt16HashAggAlloc{aggAllocBase: allocBase}, nil
		case 32:
			return &avgInt32HashAggAlloc{aggAllocBase: allocBase}, nil
		case -1:
		default:
			return &avgInt64HashAggAlloc{aggAllocBase: allocBase}, nil
		}
	case types.DecimalFamily:
		switch t.Width() {
		case -1:
		default:
			return &avgDecimalHashAggAlloc{aggAllocBase: allocBase}, nil
		}
	case types.FloatFamily:
		switch t.Width() {
		case -1:
		default:
			return &avgFloat64HashAggAlloc{aggAllocBase: allocBase}, nil
		}
	case types.IntervalFamily:
		switch t.Width() {
		case -1:
		default:
			return &avgIntervalHashAggAlloc{aggAllocBase: allocBase}, nil
		}
	}
	return nil, errors.Errorf("unsupported avg agg type %s", t.Name())
}

type avgInt16HashAgg struct {
	unorderedAggregateFuncBase
	// curSum keeps track of the sum of elements belonging to the current group,
	// so we can index into the slice once per group, instead of on each
	// iteration.
	curSum apd.Decimal
	// curCount keeps track of the number of non-null elements that we've seen
	// belonging to the current group.
	curCount int64
}

var _ AggregateFunc = &avgInt16HashAgg{}

func (a *avgInt16HashAgg) Compute(
	vecs []coldata.Vec, inputIdxs []uint32, startIdx, endIdx int, sel []int,
) {
	oldCurSumSize := a.curSum.Size()
	vec := vecs[inputIdxs[0]]
	col, nulls := vec.Int16(), vec.Nulls()
	a.allocator.PerformOperation([]coldata.Vec{a.vec}, func() {
		{
			sel = sel[startIdx:endIdx]
			if nulls.MaybeHasNulls() {
				for _, i := range sel {

					var isNull bool
					isNull = nulls.NullAt(i)
					if !isNull {
						v := col.Get(i)

						{

							var tmpDec apd.Decimal //gcassert:noescape
							tmpDec.SetInt64(int64(v))
							if _, err := tree.ExactCtx.Add(&a.curSum, &a.curSum, &tmpDec); err != nil {
								colexecerror.ExpectedError(err)
							}
						}

						a.curCount++
					}
				}
			} else {
				for _, i := range sel {

					var isNull bool
					isNull = false
					if !isNull {
						v := col.Get(i)

						{

							var tmpDec apd.Decimal //gcassert:noescape
							tmpDec.SetInt64(int64(v))
							if _, err := tree.ExactCtx.Add(&a.curSum, &a.curSum, &tmpDec); err != nil {
								colexecerror.ExpectedError(err)
							}
						}

						a.curCount++
					}
				}
			}
		}
	},
	)
	newCurSumSize := a.curSum.Size()
	if newCurSumSize != oldCurSumSize {
		a.allocator.AdjustMemoryUsage(int64(newCurSumSize - oldCurSumSize))
	}
}

func (a *avgInt16HashAgg) Flush(outputIdx int) {
	// The aggregation is finished. Flush the last value. If we haven't found
	// any non-nulls for this group so far, the output for this group should be
	// NULL.
	col := a.vec.Decimal()
	if a.curCount == 0 {
		a.nulls.SetNull(outputIdx)
	} else {

		col[outputIdx].SetInt64(a.curCount)
		if _, err := tree.DecimalCtx.Quo(&col[outputIdx], &a.curSum, &col[outputIdx]); err != nil {
			colexecerror.InternalError(err)
		}
	}
}

func (a *avgInt16HashAgg) Reset() {
	a.curSum = zeroDecimalValue
	a.curCount = 0
}

type avgInt16HashAggAlloc struct {
	aggAllocBase
	aggFuncs []avgInt16HashAgg
}

var _ aggregateFuncAlloc = &avgInt16HashAggAlloc{}

const sizeOfAvgInt16HashAgg = int64(unsafe.Sizeof(avgInt16HashAgg{}))
const avgInt16HashAggSliceOverhead = int64(unsafe.Sizeof([]avgInt16HashAgg{}))

func (a *avgInt16HashAggAlloc) newAggFunc() AggregateFunc {
	if len(a.aggFuncs) == 0 {
		a.allocator.AdjustMemoryUsage(avgInt16HashAggSliceOverhead + sizeOfAvgInt16HashAgg*a.allocSize)
		a.aggFuncs = make([]avgInt16HashAgg, a.allocSize)
	}
	f := &a.aggFuncs[0]
	f.allocator = a.allocator
	a.aggFuncs = a.aggFuncs[1:]
	return f
}

type avgInt32HashAgg struct {
	unorderedAggregateFuncBase
	// curSum keeps track of the sum of elements belonging to the current group,
	// so we can index into the slice once per group, instead of on each
	// iteration.
	curSum apd.Decimal
	// curCount keeps track of the number of non-null elements that we've seen
	// belonging to the current group.
	curCount int64
}

var _ AggregateFunc = &avgInt32HashAgg{}

func (a *avgInt32HashAgg) Compute(
	vecs []coldata.Vec, inputIdxs []uint32, startIdx, endIdx int, sel []int,
) {
	oldCurSumSize := a.curSum.Size()
	vec := vecs[inputIdxs[0]]
	col, nulls := vec.Int32(), vec.Nulls()
	a.allocator.PerformOperation([]coldata.Vec{a.vec}, func() {
		{
			sel = sel[startIdx:endIdx]
			if nulls.MaybeHasNulls() {
				for _, i := range sel {

					var isNull bool
					isNull = nulls.NullAt(i)
					if !isNull {
						v := col.Get(i)

						{

							var tmpDec apd.Decimal //gcassert:noescape
							tmpDec.SetInt64(int64(v))
							if _, err := tree.ExactCtx.Add(&a.curSum, &a.curSum, &tmpDec); err != nil {
								colexecerror.ExpectedError(err)
							}
						}

						a.curCount++
					}
				}
			} else {
				for _, i := range sel {

					var isNull bool
					isNull = false
					if !isNull {
						v := col.Get(i)

						{

							var tmpDec apd.Decimal //gcassert:noescape
							tmpDec.SetInt64(int64(v))
							if _, err := tree.ExactCtx.Add(&a.curSum, &a.curSum, &tmpDec); err != nil {
								colexecerror.ExpectedError(err)
							}
						}

						a.curCount++
					}
				}
			}
		}
	},
	)
	newCurSumSize := a.curSum.Size()
	if newCurSumSize != oldCurSumSize {
		a.allocator.AdjustMemoryUsage(int64(newCurSumSize - oldCurSumSize))
	}
}

func (a *avgInt32HashAgg) Flush(outputIdx int) {
	// The aggregation is finished. Flush the last value. If we haven't found
	// any non-nulls for this group so far, the output for this group should be
	// NULL.
	col := a.vec.Decimal()
	if a.curCount == 0 {
		a.nulls.SetNull(outputIdx)
	} else {

		col[outputIdx].SetInt64(a.curCount)
		if _, err := tree.DecimalCtx.Quo(&col[outputIdx], &a.curSum, &col[outputIdx]); err != nil {
			colexecerror.InternalError(err)
		}
	}
}

func (a *avgInt32HashAgg) Reset() {
	a.curSum = zeroDecimalValue
	a.curCount = 0
}

type avgInt32HashAggAlloc struct {
	aggAllocBase
	aggFuncs []avgInt32HashAgg
}

var _ aggregateFuncAlloc = &avgInt32HashAggAlloc{}

const sizeOfAvgInt32HashAgg = int64(unsafe.Sizeof(avgInt32HashAgg{}))
const avgInt32HashAggSliceOverhead = int64(unsafe.Sizeof([]avgInt32HashAgg{}))

func (a *avgInt32HashAggAlloc) newAggFunc() AggregateFunc {
	if len(a.aggFuncs) == 0 {
		a.allocator.AdjustMemoryUsage(avgInt32HashAggSliceOverhead + sizeOfAvgInt32HashAgg*a.allocSize)
		a.aggFuncs = make([]avgInt32HashAgg, a.allocSize)
	}
	f := &a.aggFuncs[0]
	f.allocator = a.allocator
	a.aggFuncs = a.aggFuncs[1:]
	return f
}

type avgInt64HashAgg struct {
	unorderedAggregateFuncBase
	// curSum keeps track of the sum of elements belonging to the current group,
	// so we can index into the slice once per group, instead of on each
	// iteration.
	curSum apd.Decimal
	// curCount keeps track of the number of non-null elements that we've seen
	// belonging to the current group.
	curCount int64
}

var _ AggregateFunc = &avgInt64HashAgg{}

func (a *avgInt64HashAgg) Compute(
	vecs []coldata.Vec, inputIdxs []uint32, startIdx, endIdx int, sel []int,
) {
	oldCurSumSize := a.curSum.Size()
	vec := vecs[inputIdxs[0]]
	col, nulls := vec.Int64(), vec.Nulls()
	a.allocator.PerformOperation([]coldata.Vec{a.vec}, func() {
		{
			sel = sel[startIdx:endIdx]
			if nulls.MaybeHasNulls() {
				for _, i := range sel {

					var isNull bool
					isNull = nulls.NullAt(i)
					if !isNull {
						v := col.Get(i)

						{

							var tmpDec apd.Decimal //gcassert:noescape
							tmpDec.SetInt64(int64(v))
							if _, err := tree.ExactCtx.Add(&a.curSum, &a.curSum, &tmpDec); err != nil {
								colexecerror.ExpectedError(err)
							}
						}

						a.curCount++
					}
				}
			} else {
				for _, i := range sel {

					var isNull bool
					isNull = false
					if !isNull {
						v := col.Get(i)

						{

							var tmpDec apd.Decimal //gcassert:noescape
							tmpDec.SetInt64(int64(v))
							if _, err := tree.ExactCtx.Add(&a.curSum, &a.curSum, &tmpDec); err != nil {
								colexecerror.ExpectedError(err)
							}
						}

						a.curCount++
					}
				}
			}
		}
	},
	)
	newCurSumSize := a.curSum.Size()
	if newCurSumSize != oldCurSumSize {
		a.allocator.AdjustMemoryUsage(int64(newCurSumSize - oldCurSumSize))
	}
}

func (a *avgInt64HashAgg) Flush(outputIdx int) {
	// The aggregation is finished. Flush the last value. If we haven't found
	// any non-nulls for this group so far, the output for this group should be
	// NULL.
	col := a.vec.Decimal()
	if a.curCount == 0 {
		a.nulls.SetNull(outputIdx)
	} else {

		col[outputIdx].SetInt64(a.curCount)
		if _, err := tree.DecimalCtx.Quo(&col[outputIdx], &a.curSum, &col[outputIdx]); err != nil {
			colexecerror.InternalError(err)
		}
	}
}

func (a *avgInt64HashAgg) Reset() {
	a.curSum = zeroDecimalValue
	a.curCount = 0
}

type avgInt64HashAggAlloc struct {
	aggAllocBase
	aggFuncs []avgInt64HashAgg
}

var _ aggregateFuncAlloc = &avgInt64HashAggAlloc{}

const sizeOfAvgInt64HashAgg = int64(unsafe.Sizeof(avgInt64HashAgg{}))
const avgInt64HashAggSliceOverhead = int64(unsafe.Sizeof([]avgInt64HashAgg{}))

func (a *avgInt64HashAggAlloc) newAggFunc() AggregateFunc {
	if len(a.aggFuncs) == 0 {
		a.allocator.AdjustMemoryUsage(avgInt64HashAggSliceOverhead + sizeOfAvgInt64HashAgg*a.allocSize)
		a.aggFuncs = make([]avgInt64HashAgg, a.allocSize)
	}
	f := &a.aggFuncs[0]
	f.allocator = a.allocator
	a.aggFuncs = a.aggFuncs[1:]
	return f
}

type avgDecimalHashAgg struct {
	unorderedAggregateFuncBase
	// curSum keeps track of the sum of elements belonging to the current group,
	// so we can index into the slice once per group, instead of on each
	// iteration.
	curSum apd.Decimal
	// curCount keeps track of the number of non-null elements that we've seen
	// belonging to the current group.
	curCount int64
}

var _ AggregateFunc = &avgDecimalHashAgg{}

func (a *avgDecimalHashAgg) Compute(
	vecs []coldata.Vec, inputIdxs []uint32, startIdx, endIdx int, sel []int,
) {
	oldCurSumSize := a.curSum.Size()
	vec := vecs[inputIdxs[0]]
	col, nulls := vec.Decimal(), vec.Nulls()
	a.allocator.PerformOperation([]coldata.Vec{a.vec}, func() {
		{
			sel = sel[startIdx:endIdx]
			if nulls.MaybeHasNulls() {
				for _, i := range sel {

					var isNull bool
					isNull = nulls.NullAt(i)
					if !isNull {
						v := col.Get(i)

						{

							_, err := tree.ExactCtx.Add(&a.curSum, &a.curSum, &v)
							if err != nil {
								colexecerror.ExpectedError(err)
							}
						}

						a.curCount++
					}
				}
			} else {
				for _, i := range sel {

					var isNull bool
					isNull = false
					if !isNull {
						v := col.Get(i)

						{

							_, err := tree.ExactCtx.Add(&a.curSum, &a.curSum, &v)
							if err != nil {
								colexecerror.ExpectedError(err)
							}
						}

						a.curCount++
					}
				}
			}
		}
	},
	)
	newCurSumSize := a.curSum.Size()
	if newCurSumSize != oldCurSumSize {
		a.allocator.AdjustMemoryUsage(int64(newCurSumSize - oldCurSumSize))
	}
}

func (a *avgDecimalHashAgg) Flush(outputIdx int) {
	// The aggregation is finished. Flush the last value. If we haven't found
	// any non-nulls for this group so far, the output for this group should be
	// NULL.
	col := a.vec.Decimal()
	if a.curCount == 0 {
		a.nulls.SetNull(outputIdx)
	} else {

		col[outputIdx].SetInt64(a.curCount)
		if _, err := tree.DecimalCtx.Quo(&col[outputIdx], &a.curSum, &col[outputIdx]); err != nil {
			colexecerror.InternalError(err)
		}
	}
}

func (a *avgDecimalHashAgg) Reset() {
	a.curSum = zeroDecimalValue
	a.curCount = 0
}

type avgDecimalHashAggAlloc struct {
	aggAllocBase
	aggFuncs []avgDecimalHashAgg
}

var _ aggregateFuncAlloc = &avgDecimalHashAggAlloc{}

const sizeOfAvgDecimalHashAgg = int64(unsafe.Sizeof(avgDecimalHashAgg{}))
const avgDecimalHashAggSliceOverhead = int64(unsafe.Sizeof([]avgDecimalHashAgg{}))

func (a *avgDecimalHashAggAlloc) newAggFunc() AggregateFunc {
	if len(a.aggFuncs) == 0 {
		a.allocator.AdjustMemoryUsage(avgDecimalHashAggSliceOverhead + sizeOfAvgDecimalHashAgg*a.allocSize)
		a.aggFuncs = make([]avgDecimalHashAgg, a.allocSize)
	}
	f := &a.aggFuncs[0]
	f.allocator = a.allocator
	a.aggFuncs = a.aggFuncs[1:]
	return f
}

type avgFloat64HashAgg struct {
	unorderedAggregateFuncBase
	// curSum keeps track of the sum of elements belonging to the current group,
	// so we can index into the slice once per group, instead of on each
	// iteration.
	curSum float64
	// curCount keeps track of the number of non-null elements that we've seen
	// belonging to the current group.
	curCount int64
}

var _ AggregateFunc = &avgFloat64HashAgg{}

func (a *avgFloat64HashAgg) Compute(
	vecs []coldata.Vec, inputIdxs []uint32, startIdx, endIdx int, sel []int,
) {
	var oldCurSumSize uintptr
	vec := vecs[inputIdxs[0]]
	col, nulls := vec.Float64(), vec.Nulls()
	a.allocator.PerformOperation([]coldata.Vec{a.vec}, func() {
		{
			sel = sel[startIdx:endIdx]
			if nulls.MaybeHasNulls() {
				for _, i := range sel {

					var isNull bool
					isNull = nulls.NullAt(i)
					if !isNull {
						v := col.Get(i)

						{

							a.curSum = float64(a.curSum) + float64(v)
						}

						a.curCount++
					}
				}
			} else {
				for _, i := range sel {

					var isNull bool
					isNull = false
					if !isNull {
						v := col.Get(i)

						{

							a.curSum = float64(a.curSum) + float64(v)
						}

						a.curCount++
					}
				}
			}
		}
	},
	)
	var newCurSumSize uintptr
	if newCurSumSize != oldCurSumSize {
		a.allocator.AdjustMemoryUsage(int64(newCurSumSize - oldCurSumSize))
	}
}

func (a *avgFloat64HashAgg) Flush(outputIdx int) {
	// The aggregation is finished. Flush the last value. If we haven't found
	// any non-nulls for this group so far, the output for this group should be
	// NULL.
	col := a.vec.Float64()
	if a.curCount == 0 {
		a.nulls.SetNull(outputIdx)
	} else {
		col[outputIdx] = a.curSum / float64(a.curCount)
	}
}

func (a *avgFloat64HashAgg) Reset() {
	a.curSum = zeroFloat64Value
	a.curCount = 0
}

type avgFloat64HashAggAlloc struct {
	aggAllocBase
	aggFuncs []avgFloat64HashAgg
}

var _ aggregateFuncAlloc = &avgFloat64HashAggAlloc{}

const sizeOfAvgFloat64HashAgg = int64(unsafe.Sizeof(avgFloat64HashAgg{}))
const avgFloat64HashAggSliceOverhead = int64(unsafe.Sizeof([]avgFloat64HashAgg{}))

func (a *avgFloat64HashAggAlloc) newAggFunc() AggregateFunc {
	if len(a.aggFuncs) == 0 {
		a.allocator.AdjustMemoryUsage(avgFloat64HashAggSliceOverhead + sizeOfAvgFloat64HashAgg*a.allocSize)
		a.aggFuncs = make([]avgFloat64HashAgg, a.allocSize)
	}
	f := &a.aggFuncs[0]
	f.allocator = a.allocator
	a.aggFuncs = a.aggFuncs[1:]
	return f
}

type avgIntervalHashAgg struct {
	unorderedAggregateFuncBase
	// curSum keeps track of the sum of elements belonging to the current group,
	// so we can index into the slice once per group, instead of on each
	// iteration.
	curSum duration.Duration
	// curCount keeps track of the number of non-null elements that we've seen
	// belonging to the current group.
	curCount int64
}

var _ AggregateFunc = &avgIntervalHashAgg{}

func (a *avgIntervalHashAgg) Compute(
	vecs []coldata.Vec, inputIdxs []uint32, startIdx, endIdx int, sel []int,
) {
	var oldCurSumSize uintptr
	vec := vecs[inputIdxs[0]]
	col, nulls := vec.Interval(), vec.Nulls()
	a.allocator.PerformOperation([]coldata.Vec{a.vec}, func() {
		{
			sel = sel[startIdx:endIdx]
			if nulls.MaybeHasNulls() {
				for _, i := range sel {

					var isNull bool
					isNull = nulls.NullAt(i)
					if !isNull {
						v := col.Get(i)
						a.curSum = a.curSum.Add(v)
						a.curCount++
					}
				}
			} else {
				for _, i := range sel {

					var isNull bool
					isNull = false
					if !isNull {
						v := col.Get(i)
						a.curSum = a.curSum.Add(v)
						a.curCount++
					}
				}
			}
		}
	},
	)
	var newCurSumSize uintptr
	if newCurSumSize != oldCurSumSize {
		a.allocator.AdjustMemoryUsage(int64(newCurSumSize - oldCurSumSize))
	}
}

func (a *avgIntervalHashAgg) Flush(outputIdx int) {
	// The aggregation is finished. Flush the last value. If we haven't found
	// any non-nulls for this group so far, the output for this group should be
	// NULL.
	col := a.vec.Interval()
	if a.curCount == 0 {
		a.nulls.SetNull(outputIdx)
	} else {
		col[outputIdx] = a.curSum.Div(int64(a.curCount))
	}
}

func (a *avgIntervalHashAgg) Reset() {
	a.curSum = zeroIntervalValue
	a.curCount = 0
}

type avgIntervalHashAggAlloc struct {
	aggAllocBase
	aggFuncs []avgIntervalHashAgg
}

var _ aggregateFuncAlloc = &avgIntervalHashAggAlloc{}

const sizeOfAvgIntervalHashAgg = int64(unsafe.Sizeof(avgIntervalHashAgg{}))
const avgIntervalHashAggSliceOverhead = int64(unsafe.Sizeof([]avgIntervalHashAgg{}))

func (a *avgIntervalHashAggAlloc) newAggFunc() AggregateFunc {
	if len(a.aggFuncs) == 0 {
		a.allocator.AdjustMemoryUsage(avgIntervalHashAggSliceOverhead + sizeOfAvgIntervalHashAgg*a.allocSize)
		a.aggFuncs = make([]avgIntervalHashAgg, a.allocSize)
	}
	f := &a.aggFuncs[0]
	f.allocator = a.allocator
	a.aggFuncs = a.aggFuncs[1:]
	return f
}
