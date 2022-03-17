// Copyright 2021 The Cockroach Authors.
//
// Licensed as a CockroachDB Enterprise file under the Cockroach Community
// License (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//     https://sqlfmt/cockroach/blob/master/licenses/CCL.txt

package changefeedccl

import (
	"context"

	"sqlfmt/cockroach/pkg/base"
	"sqlfmt/cockroach/pkg/config"
	"sqlfmt/cockroach/pkg/kv"
	"sqlfmt/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"sqlfmt/cockroach/pkg/roachpb"
	"sqlfmt/cockroach/pkg/rpc"
	"sqlfmt/cockroach/pkg/server/status"
	"sqlfmt/cockroach/pkg/settings/cluster"
	"sqlfmt/cockroach/pkg/storage"
	"sqlfmt/cockroach/pkg/testutils/serverutils"
	"sqlfmt/cockroach/pkg/util/hlc"
	"sqlfmt/cockroach/pkg/util/stop"
	"sqlfmt/cockroach/pkg/util/uuid"
)

// testServerShim is a kludge to get a few more tests working in
// tenant-mode.
//
// Currently, our TestFeedFactory has a Server() method that returns a
// TestServerInterface. The TestTenantInterface returned by
// StartTenant isn't a TestServerInterface.
//
// TODO(ssd): Clean this up. Perhaps we can add a SQLServer() method
// to TestFeedFactory that returns just the bits that are shared.
type testServerShim struct {
	serverutils.TestTenantInterface
	kvServer serverutils.TestServerInterface
}

const unsupportedShimMethod = `
This TestServerInterface method is not supported for tenants. Either disable this test on tenants by using the
feedOptionNoTenants option or add an appropriate implementation for this method to testServerShim.
`

var _ serverutils.TestServerInterface = (*testServerShim)(nil)

func (t *testServerShim) ServingSQLAddr() string {
	return t.SQLAddr()
}

func (t *testServerShim) Stopper() *stop.Stopper                { panic(unsupportedShimMethod) }
func (t *testServerShim) Start(context.Context) error           { panic(unsupportedShimMethod) }
func (t *testServerShim) Node() interface{}                     { panic(unsupportedShimMethod) }
func (t *testServerShim) NodeID() roachpb.NodeID                { panic(unsupportedShimMethod) }
func (t *testServerShim) ClusterID() uuid.UUID                  { panic(unsupportedShimMethod) }
func (t *testServerShim) ServingRPCAddr() string                { panic(unsupportedShimMethod) }
func (t *testServerShim) RPCAddr() string                       { panic(unsupportedShimMethod) }
func (t *testServerShim) DB() *kv.DB                            { panic(unsupportedShimMethod) }
func (t *testServerShim) RPCContext() *rpc.Context              { panic(unsupportedShimMethod) }
func (t *testServerShim) LeaseManager() interface{}             { panic(unsupportedShimMethod) }
func (t *testServerShim) InternalExecutor() interface{}         { panic(unsupportedShimMethod) }
func (t *testServerShim) ExecutorConfig() interface{}           { panic(unsupportedShimMethod) }
func (t *testServerShim) TracerI() interface{}                  { panic(unsupportedShimMethod) }
func (t *testServerShim) GossipI() interface{}                  { panic(unsupportedShimMethod) }
func (t *testServerShim) RangeFeedFactory() interface{}         { panic(unsupportedShimMethod) }
func (t *testServerShim) Clock() *hlc.Clock                     { panic(unsupportedShimMethod) }
func (t *testServerShim) DistSenderI() interface{}              { panic(unsupportedShimMethod) }
func (t *testServerShim) MigrationServer() interface{}          { panic(unsupportedShimMethod) }
func (t *testServerShim) SQLServer() interface{}                { panic(unsupportedShimMethod) }
func (t *testServerShim) SQLLivenessProvider() interface{}      { panic(unsupportedShimMethod) }
func (t *testServerShim) StartupMigrationsManager() interface{} { panic(unsupportedShimMethod) }
func (t *testServerShim) NodeLiveness() interface{}             { panic(unsupportedShimMethod) }
func (t *testServerShim) HeartbeatNodeLiveness() error          { panic(unsupportedShimMethod) }
func (t *testServerShim) NodeDialer() interface{}               { panic(unsupportedShimMethod) }
func (t *testServerShim) SetDistSQLSpanResolver(spanResolver interface{}) {
	panic(unsupportedShimMethod)
}
func (t *testServerShim) MustGetSQLCounter(name string) int64        { panic(unsupportedShimMethod) }
func (t *testServerShim) MustGetSQLNetworkCounter(name string) int64 { panic(unsupportedShimMethod) }
func (t *testServerShim) WriteSummaries() error                      { panic(unsupportedShimMethod) }
func (t *testServerShim) GetFirstStoreID() roachpb.StoreID           { panic(unsupportedShimMethod) }
func (t *testServerShim) GetStores() interface{}                     { panic(unsupportedShimMethod) }
func (t *testServerShim) ClusterSettings() *cluster.Settings         { panic(unsupportedShimMethod) }
func (t *testServerShim) Decommission(
	ctx context.Context, targetStatus livenesspb.MembershipStatus, nodeIDs []roachpb.NodeID,
) error {
	panic(unsupportedShimMethod)
}
func (t *testServerShim) SplitRange(
	splitKey roachpb.Key,
) (left roachpb.RangeDescriptor, right roachpb.RangeDescriptor, err error) {
	panic(unsupportedShimMethod)
}
func (t *testServerShim) MergeRanges(
	leftKey roachpb.Key,
) (merged roachpb.RangeDescriptor, err error) {
	panic(unsupportedShimMethod)
}
func (t *testServerShim) ExpectedInitialRangeCount() (int, error) { panic(unsupportedShimMethod) }
func (t *testServerShim) ForceTableGC(
	ctx context.Context, database, table string, timestamp hlc.Timestamp,
) error {
	panic(unsupportedShimMethod)
}
func (t *testServerShim) UpdateChecker() interface{} { panic(unsupportedShimMethod) }
func (t *testServerShim) StartTenant(
	ctx context.Context, params base.TestTenantArgs,
) (serverutils.TestTenantInterface, error) {
	panic(unsupportedShimMethod)
}
func (t *testServerShim) ScratchRange() (roachpb.Key, error)       { panic(unsupportedShimMethod) }
func (t *testServerShim) Engines() []storage.Engine                { panic(unsupportedShimMethod) }
func (t *testServerShim) MetricsRecorder() *status.MetricsRecorder { panic(unsupportedShimMethod) }
func (t *testServerShim) CollectionFactory() interface{}           { panic(unsupportedShimMethod) }
func (t *testServerShim) SystemTableIDResolver() interface{}       { panic(unsupportedShimMethod) }
func (t *testServerShim) SpanConfigKVSubscriber() interface{}      { panic(unsupportedShimMethod) }
func (t *testServerShim) SystemConfigProvider() config.SystemConfigProvider {
	panic(unsupportedShimMethod)
}
