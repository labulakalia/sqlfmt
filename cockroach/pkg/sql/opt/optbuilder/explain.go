// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package optbuilder

import (
	"sqlfmt/cockroach/pkg/server/telemetry"
	"sqlfmt/cockroach/pkg/sql/catalog/colinfo"
	"sqlfmt/cockroach/pkg/sql/opt/memo"
	"sqlfmt/cockroach/pkg/sql/pgwire/pgcode"
	"sqlfmt/cockroach/pkg/sql/pgwire/pgerror"
	"sqlfmt/cockroach/pkg/sql/sem/tree"
	"sqlfmt/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/errors"
)

func (b *Builder) buildExplain(explain *tree.Explain, inScope *scope) (outScope *scope) {
	if _, ok := explain.Statement.(*tree.Execute); ok {
		panic(pgerror.New(
			pgcode.FeatureNotSupported, "EXPLAIN EXECUTE is not supported; use EXPLAIN ANALYZE",
		))
	}

	stmtScope := b.buildStmtAtRoot(explain.Statement, nil /* desiredTypes */)

	outScope = inScope.push()

	switch explain.Mode {
	case tree.ExplainPlan:
		telemetry.Inc(sqltelemetry.ExplainPlanUseCounter)

	case tree.ExplainDistSQL:
		telemetry.Inc(sqltelemetry.ExplainDistSQLUseCounter)

	case tree.ExplainOpt:
		if explain.Flags[tree.ExplainFlagVerbose] {
			telemetry.Inc(sqltelemetry.ExplainOptVerboseUseCounter)
		} else {
			telemetry.Inc(sqltelemetry.ExplainOptUseCounter)
		}

	case tree.ExplainVec:
		telemetry.Inc(sqltelemetry.ExplainVecUseCounter)
	case tree.ExplainDDL:
		if explain.Flags[tree.ExplainFlagDeps] {
			telemetry.Inc(sqltelemetry.ExplainDDLDeps)
		} else {
			telemetry.Inc(sqltelemetry.ExplainDDLStages)
		}

	case tree.ExplainGist:
		telemetry.Inc(sqltelemetry.ExplainGist)

	default:
		panic(errors.Errorf("EXPLAIN mode %s not supported", explain.Mode))
	}
	b.synthesizeResultColumns(outScope, colinfo.ExplainPlanColumns)

	input := stmtScope.expr
	private := memo.ExplainPrivate{
		Options:  explain.ExplainOptions,
		ColList:  colsToColList(outScope.cols),
		Props:    stmtScope.makePhysicalProps(),
		StmtType: explain.Statement.StatementReturnType(),
	}
	outScope.expr = b.factory.ConstructExplain(input, &private)
	return outScope
}
