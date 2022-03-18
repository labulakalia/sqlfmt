// Copyright 2021 The Cockroach Authors.
//
// Licensed as a CockroachDB Enterprise file under the Cockroach Community
// License (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//     https://sqlfmt/cockroach/blob/master/licenses/CCL.txt

package streamproducer

import (
	"github.com/labulakalia/sqlfmt/cockroach/pkg/ccl/streamingccl/streampb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/ccl/utilccl"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/streaming"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/hlc"
)

type replicationStreamManagerImpl struct{}

// CompleteStreamIngestion implements ReplicationStreamManager interface.
func (r *replicationStreamManagerImpl) CompleteStreamIngestion(
	evalCtx *tree.EvalContext,
	txn *kv.Txn,
	streamID streaming.StreamID,
	cutoverTimestamp hlc.Timestamp,
) error {
	return completeStreamIngestion(evalCtx, txn, streamID, cutoverTimestamp)
}

// StartReplicationStream implements ReplicationStreamManager interface.
func (r *replicationStreamManagerImpl) StartReplicationStream(
	evalCtx *tree.EvalContext, txn *kv.Txn, tenantID uint64,
) (streaming.StreamID, error) {
	return startReplicationStreamJob(evalCtx, txn, tenantID)
}

// UpdateReplicationStreamProgress implements ReplicationStreamManager interface.
func (r *replicationStreamManagerImpl) UpdateReplicationStreamProgress(
	evalCtx *tree.EvalContext, streamID streaming.StreamID, frontier hlc.Timestamp, txn *kv.Txn,
) (streampb.StreamReplicationStatus, error) {
	return heartbeatReplicationStream(evalCtx, streamID, frontier, txn)
}

// StreamPartition returns a value generator which yields events for the specified partition.
// opaqueSpec contains streampb.PartitionSpec protocol message.
// streamID specifies the streaming job this partition belongs too.
func (r *replicationStreamManagerImpl) StreamPartition(
	evalCtx *tree.EvalContext, streamID streaming.StreamID, opaqueSpec []byte,
) (tree.ValueGenerator, error) {
	return streamPartition(evalCtx, streamID, opaqueSpec)
}

// GetReplicationStreamSpec implements ReplicationStreamManager interface.
func (r *replicationStreamManagerImpl) GetReplicationStreamSpec(
	evalCtx *tree.EvalContext, txn *kv.Txn, streamID streaming.StreamID,
) (*streampb.ReplicationStreamSpec, error) {
	return getReplicationStreamSpec(evalCtx, txn, streamID)
}

func newReplicationStreamManagerWithPrivilegesCheck(
	evalCtx *tree.EvalContext,
) (streaming.ReplicationStreamManager, error) {
	isAdmin, err := evalCtx.SessionAccessor.HasAdminRole(evalCtx.Context)
	if err != nil {
		return nil, err
	}

	if !isAdmin {
		return nil,
			pgerror.New(pgcode.InsufficientPrivilege, "replication restricted to ADMIN role")
	}

	execCfg := evalCtx.Planner.ExecutorConfig().(*sql.ExecutorConfig)
	enterpriseCheckErr := utilccl.CheckEnterpriseEnabled(
		execCfg.Settings, execCfg.ClusterID(), execCfg.Organization(), "REPLICATION")
	if enterpriseCheckErr != nil {
		return nil, pgerror.Wrap(enterpriseCheckErr,
			pgcode.InsufficientPrivilege, "replication requires enterprise license")
	}

	return &replicationStreamManagerImpl{}, nil
}

func init() {
	streaming.GetReplicationStreamManagerHook = newReplicationStreamManagerWithPrivilegesCheck
}
