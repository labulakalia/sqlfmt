// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package sql

import (
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog/colinfo"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/opt"
)

type opaqueMetadata struct {
	info    string
	plan    planNode
	columns colinfo.ResultColumns
}

var _ opt.OpaqueMetadata = &opaqueMetadata{}

func (o *opaqueMetadata) ImplementsOpaqueMetadata()      {}
func (o *opaqueMetadata) String() string                 { return o.info }
func (o *opaqueMetadata) Columns() colinfo.ResultColumns { return o.columns }
