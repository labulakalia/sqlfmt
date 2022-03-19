// Copyright 2020 The Cockroach Authors.
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
	"context"
	"fmt"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/scheduledjobs"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/security"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sessiondata"
)

type controlSchedulesNode struct {
	rows    planNode
	command tree.ScheduleCommand
	numRows int
}


// FastPathResults implements the planNodeFastPath interface.
func (n *controlSchedulesNode) FastPathResults() (int, bool) {
	return n.numRows, true
}

// JobSchedulerEnv returns JobSchedulerEnv.
func JobSchedulerEnv(execCfg *ExecutorConfig) scheduledjobs.JobSchedulerEnv {
	return scheduledjobs.ProdJobSchedulerEnv
}
// DeleteSchedule deletes specified schedule.
func DeleteSchedule(
	ctx context.Context, execCfg *ExecutorConfig, txn *kv.Txn, scheduleID int64,
) error {
	env := JobSchedulerEnv(execCfg)
	_, err := execCfg.InternalExecutor.ExecEx(
		ctx,
		"delete-schedule",
		txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		fmt.Sprintf(
			"DELETE FROM %s WHERE schedule_id = $1",
			env.ScheduledJobsTableName(),
		),
		scheduleID,
	)
	return err
}

// Next implements planNode interface.
func (*controlSchedulesNode) Next(runParams) (bool, error) { return false, nil }

// Values implements planNode interface.
func (*controlSchedulesNode) Values() tree.Datums { return nil }

// Close implements planNode interface.
func (n *controlSchedulesNode) Close(ctx context.Context) {
	n.rows.Close(ctx)
}
