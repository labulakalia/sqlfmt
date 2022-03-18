// Code generated by execgen; DO NOT EDIT.
// Copyright 2020 The Cockroach Authors.
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexecproj

import (
	"github.com/labulakalia/sqlfmt/cockroach/pkg/col/coldata"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colconv"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexec/colexeccmp"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexecerror"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/colexecop"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/execinfra"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree"
)

type defaultCmpProjOp struct {
	projOpBase

	adapter             colexeccmp.ComparisonExprAdapter
	toDatumConverter    *colconv.VecToDatumConverter
	datumToVecConverter func(tree.Datum) interface{}
}

var _ colexecop.Operator = &defaultCmpProjOp{}
var _ execinfra.Releasable = &defaultCmpProjOp{}

func (d *defaultCmpProjOp) Next() coldata.Batch {
	batch := d.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	sel := batch.Selection()
	output := batch.ColVec(d.outputIdx)
	d.allocator.PerformOperation([]coldata.Vec{output}, func() {
		d.toDatumConverter.ConvertBatchAndDeselect(batch)
		leftColumn := d.toDatumConverter.GetDatumColumn(d.col1Idx)
		rightColumn := d.toDatumConverter.GetDatumColumn(d.col2Idx)
		_ = leftColumn[n-1]
		_ = rightColumn[n-1]
		if sel != nil {
			_ = sel[n-1]
		}
		for i := 0; i < n; i++ {
			// Note that we performed a conversion with deselection, so there
			// is no need to check whether sel is non-nil.
			//gcassert:bce
			res, err := d.adapter.Eval(leftColumn[i], rightColumn[i])
			if err != nil {
				colexecerror.ExpectedError(err)
			}
			rowIdx := i
			if sel != nil {
				rowIdx = sel[i]
			}
			// Convert the datum into a physical type and write it out.
			// TODO(yuzefovich): this code block is repeated in several places.
			// Refactor it.
			if res == tree.DNull {
				output.Nulls().SetNull(rowIdx)
			} else {
				converted := d.datumToVecConverter(res)
				coldata.SetValueAt(output, converted, rowIdx)
			}
		}
	})
	return batch
}

func (d *defaultCmpProjOp) Release() {
	d.toDatumConverter.Release()
}

type defaultCmpRConstProjOp struct {
	projConstOpBase
	constArg tree.Datum

	adapter             colexeccmp.ComparisonExprAdapter
	toDatumConverter    *colconv.VecToDatumConverter
	datumToVecConverter func(tree.Datum) interface{}
}

var _ colexecop.Operator = &defaultCmpRConstProjOp{}
var _ execinfra.Releasable = &defaultCmpRConstProjOp{}

func (d *defaultCmpRConstProjOp) Next() coldata.Batch {
	batch := d.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	sel := batch.Selection()
	output := batch.ColVec(d.outputIdx)
	d.allocator.PerformOperation([]coldata.Vec{output}, func() {
		d.toDatumConverter.ConvertBatchAndDeselect(batch)
		nonConstColumn := d.toDatumConverter.GetDatumColumn(d.colIdx)
		_ = nonConstColumn[n-1]
		if sel != nil {
			_ = sel[n-1]
		}
		for i := 0; i < n; i++ {
			// Note that we performed a conversion with deselection, so there
			// is no need to check whether sel is non-nil.
			//gcassert:bce
			res, err := d.adapter.Eval(nonConstColumn[i], d.constArg)
			if err != nil {
				colexecerror.ExpectedError(err)
			}
			rowIdx := i
			if sel != nil {
				rowIdx = sel[i]
			}
			// Convert the datum into a physical type and write it out.
			// TODO(yuzefovich): this code block is repeated in several places.
			// Refactor it.
			if res == tree.DNull {
				output.Nulls().SetNull(rowIdx)
			} else {
				converted := d.datumToVecConverter(res)
				coldata.SetValueAt(output, converted, rowIdx)
			}
		}
	})
	return batch
}

func (d *defaultCmpRConstProjOp) Release() {
	d.toDatumConverter.Release()
}
