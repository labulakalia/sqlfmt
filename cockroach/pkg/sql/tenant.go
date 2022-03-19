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
	"time"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/clusterversion"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/jobs"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/jobs/jobspb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/keys"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/roachpb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/security"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/settings"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/spanconfig"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog/bootstrap"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog/descpb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog/systemschema"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/rowenc"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sessiondata"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/log"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/protoutil"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

// rejectIfCantCoordinateMultiTenancy returns an error if the current tenant is
// disallowed from coordinating tenant management operations on behalf of a
// multi-tenant cluster. Only the system tenant has permissions to do so.
func rejectIfCantCoordinateMultiTenancy(codec keys.SQLCodec, op string) error {
	// NOTE: even if we got this wrong, the rest of the function would fail for
	// a non-system tenant because they would be missing a system.tenant table.
	if !codec.ForSystemTenant() {
		return pgerror.Newf(pgcode.InsufficientPrivilege,
			"only the system tenant can %s other tenants", op)
	}
	return nil
}

// rejectIfSystemTenant returns an error if the provided tenant ID is the system
// tenant's ID.
func rejectIfSystemTenant(tenID uint64, op string) error {
	if roachpb.IsSystemTenantID(tenID) {
		return pgerror.Newf(pgcode.InvalidParameterValue,
			"cannot %s tenant \"%d\", ID assigned to system tenant", op, tenID)
	}
	return nil
}

// CreateTenantRecord creates a tenant in system.tenants and installs an initial
// span config (in system.span_configurations) for it. It also initializes the
// usage data in system.tenant_usage if info.Usage is set.
func CreateTenantRecord(
	ctx context.Context, execCfg *ExecutorConfig, txn *kv.Txn, info *descpb.TenantInfoWithUsage,
) error {
	const op = "create"
	if err := rejectIfCantCoordinateMultiTenancy(execCfg.Codec, op); err != nil {
		return err
	}
	if err := rejectIfSystemTenant(info.ID, op); err != nil {
		return err
	}

	tenID := info.ID
	active := info.State == descpb.TenantInfo_ACTIVE
	infoBytes, err := protoutil.Marshal(&info.TenantInfo)
	if err != nil {
		return err
	}

	// Insert into the tenant table and detect collisions.
	if num, err := execCfg.InternalExecutor.ExecEx(
		ctx, "create-tenant", txn, sessiondata.NodeUserSessionDataOverride,
		`INSERT INTO system.tenants (id, active, info) VALUES ($1, $2, $3)`,
		tenID, active, infoBytes,
	); err != nil {
		if pgerror.GetPGCode(err) == pgcode.UniqueViolation {
			return pgerror.Newf(pgcode.DuplicateObject, "tenant \"%d\" already exists", tenID)
		}
		return errors.Wrap(err, "inserting new tenant")
	} else if num != 1 {
		log.Fatalf(ctx, "unexpected number of rows affected: %d", num)
	}

	if u := info.Usage; u != nil {
		consumption, err := protoutil.Marshal(&u.Consumption)
		if err != nil {
			return errors.Wrap(err, "marshaling tenant usage data")
		}
		if num, err := execCfg.InternalExecutor.ExecEx(
			ctx, "create-tenant-usage", txn, sessiondata.NodeUserSessionDataOverride,
			`INSERT INTO system.tenant_usage (
			  tenant_id, instance_id, next_instance_id, last_update,
			  ru_burst_limit, ru_refill_rate, ru_current, current_share_sum,
			  total_consumption)
			VALUES (
				$1, 0, 0, now(),
				$2, $3, $4, 0, 
				$5)`,
			tenID,
			u.RUBurstLimit, u.RURefillRate, u.RUCurrent,
			tree.NewDBytes(tree.DBytes(consumption)),
		); err != nil {
			if pgerror.GetPGCode(err) == pgcode.UniqueViolation {
				return pgerror.Newf(pgcode.DuplicateObject, "tenant \"%d\" already has usage data", tenID)
			}
			return errors.Wrap(err, "inserting tenant usage data")
		} else if num != 1 {
			log.Fatalf(ctx, "unexpected number of rows affected: %d", num)
		}
	}

	if !execCfg.Settings.Version.IsActive(ctx, clusterversion.PreSeedTenantSpanConfigs) {
		return nil
	}

	// Install a single key[1] span config at the start of tenant's keyspace;
	// elsewhere this ensures that we split on the tenant boundary. The subset
	// of entries with spans in the tenant keyspace are, henceforth, governed
	// by the tenant's SQL pods. This entry may be replaced with others when the
	// SQL pods reconcile their zone configs for the first time. When destroying
	// the tenant for good, we'll clear out any left over entries as part of the
	// GC-ing the tenant's record.
	//
	// [1]: It doesn't actually matter what span is inserted here as long as it
	//      starts at the tenant prefix and is fully contained within the tenant
	//      keyspace. The span does not need to extend all the way to the
	//      tenant's prefix end because we only look at start keys for split
	//      boundaries. Whatever is inserted will get cleared out by the
	//      tenant's reconciliation process.

	// TODO(irfansharif): What should this initial default be? Could be this
	// static one, could use host's RANGE TENANT or host's RANGE DEFAULT?
	// Does it even matter given it'll disappear as soon as tenant starts
	// reconciling?
	tenantSpanConfig := execCfg.DefaultZoneConfig.AsSpanConfig()
	// Make sure to enable rangefeeds; the tenant will need them on its system
	// tables as soon as it starts up. It's not unsafe/buggy if we didn't do this,
	// -- the tenant's span config reconciliation process would eventually install
	// appropriate (rangefeed.enabled = true) configs for its system tables, at
	// which point subsystems that rely on rangefeeds are able to proceed. All of
	// this can noticeably slow down pod startup, so we just enable things to
	// start with.
	tenantSpanConfig.RangefeedEnabled = true
	// Make it behave like usual system database ranges, for good measure.
	tenantSpanConfig.GCPolicy.IgnoreStrictEnforcement = true

	tenantPrefix := keys.MakeTenantPrefix(roachpb.MakeTenantID(tenID))
	record, err := spanconfig.MakeRecord(spanconfig.MakeTargetFromSpan(roachpb.Span{
		Key:    tenantPrefix,
		EndKey: tenantPrefix.Next(),
	}), tenantSpanConfig)
	if err != nil {
		return err
	}
	toUpsert := []spanconfig.Record{record}
	scKVAccessor := execCfg.SpanConfigKVAccessor.WithTxn(ctx, txn)
	return scKVAccessor.UpdateSpanConfigRecords(
		ctx, nil /* toDelete */, toUpsert,
	)
}

// GetTenantRecord retrieves a tenant in system.tenants.
func GetTenantRecord(
	ctx context.Context, execCfg *ExecutorConfig, txn *kv.Txn, tenID uint64,
) (*descpb.TenantInfo, error) {
	row, err := execCfg.InternalExecutor.QueryRowEx(
		ctx, "activate-tenant", txn, sessiondata.NodeUserSessionDataOverride,
		`SELECT info FROM system.tenants WHERE id = $1`, tenID,
	)
	if err != nil {
		return nil, err
	} else if row == nil {
		return nil, pgerror.Newf(pgcode.UndefinedObject, "tenant \"%d\" does not exist", tenID)
	}

	info := &descpb.TenantInfo{}
	infoBytes := []byte(tree.MustBeDBytes(row[0]))
	if err := protoutil.Unmarshal(infoBytes, info); err != nil {
		return nil, err
	}
	return info, nil
}

// updateTenantRecord updates a tenant in system.tenants.
func updateTenantRecord(
	ctx context.Context, execCfg *ExecutorConfig, txn *kv.Txn, info *descpb.TenantInfo,
) error {
	tenID := info.ID
	active := info.State == descpb.TenantInfo_ACTIVE
	infoBytes, err := protoutil.Marshal(info)
	if err != nil {
		return err
	}

	if num, err := execCfg.InternalExecutor.ExecEx(
		ctx, "activate-tenant", txn, sessiondata.NodeUserSessionDataOverride,
		`UPDATE system.tenants SET active = $2, info = $3 WHERE id = $1`,
		tenID, active, infoBytes,
	); err != nil {
		return errors.Wrap(err, "activating tenant")
	} else if num != 1 {
		log.Fatalf(ctx, "unexpected number of rows affected: %d", num)
	}
	return nil
}


// clearTenant deletes the tenant's data.
func clearTenant(ctx context.Context, execCfg *ExecutorConfig, info *descpb.TenantInfo) error {
	// Confirm tenant is ready to be cleared.
	if info.State != descpb.TenantInfo_DROP {
		return errors.Errorf("tenant %d is not in state DROP", info.ID)
	}

	log.Infof(ctx, "clearing data for tenant %d", info.ID)

	prefix := keys.MakeTenantPrefix(roachpb.MakeTenantID(info.ID))
	prefixEnd := prefix.PrefixEnd()

	log.VEventf(ctx, 2, "ClearRange %s - %s", prefix, prefixEnd)
	// ClearRange cannot be run in a transaction, so create a non-transactional
	// batch to send the request.
	b := &kv.Batch{}
	b.AddRawRequest(&roachpb.ClearRangeRequest{
		RequestHeader: roachpb.RequestHeader{Key: prefix, EndKey: prefixEnd},
	})

	return errors.Wrapf(execCfg.DB.Run(ctx, b), "clearing tenant %d data", info.ID)
}

// DestroyTenant implements the tree.TenantOperator interface.
func (p *planner) DestroyTenant(ctx context.Context, tenID uint64, synchronous bool) error {
	const op = "destroy"
	if err := rejectIfCantCoordinateMultiTenancy(p.execCfg.Codec, op); err != nil {
		return err
	}
	if err := rejectIfSystemTenant(tenID, op); err != nil {
		return err
	}

	// Retrieve the tenant's info.
	info, err := GetTenantRecord(ctx, p.execCfg, p.txn, tenID)
	if err != nil {
		return errors.Wrap(err, "destroying tenant")
	}

	if info.State == descpb.TenantInfo_DROP {
		return errors.Errorf("tenant %d is already in state DROP", tenID)
	}

	// Mark the tenant as dropping.
	info.State = descpb.TenantInfo_DROP
	if err := updateTenantRecord(ctx, p.execCfg, p.txn, info); err != nil {
		return errors.Wrap(err, "destroying tenant")
	}

	jobID, err := gcTenantJob(ctx, p.execCfg, p.txn, p.User(), tenID, synchronous)
	if err != nil {
		return errors.Wrap(err, "scheduling gc job")
	}
	if synchronous {
		p.extendedEvalCtx.Jobs.add(jobID)
	}
	return nil
}

// GCTenantSync clears the tenant's data and removes its record.
func GCTenantSync(ctx context.Context, execCfg *ExecutorConfig, info *descpb.TenantInfo) error {
	const op = "gc"
	if err := rejectIfCantCoordinateMultiTenancy(execCfg.Codec, op); err != nil {
		return err
	}
	if err := rejectIfSystemTenant(info.ID, op); err != nil {
		return err
	}

	if err := clearTenant(ctx, execCfg, info); err != nil {
		return errors.Wrap(err, "clear tenant")
	}

	err := execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		if num, err := execCfg.InternalExecutor.ExecEx(
			ctx, "delete-tenant", txn, sessiondata.NodeUserSessionDataOverride,
			`DELETE FROM system.tenants WHERE id = $1`, info.ID,
		); err != nil {
			return errors.Wrapf(err, "deleting tenant %d", info.ID)
		} else if num != 1 {
			log.Fatalf(ctx, "unexpected number of rows affected: %d", num)
		}

		if _, err := execCfg.InternalExecutor.ExecEx(
			ctx, "delete-tenant-usage", txn, sessiondata.NodeUserSessionDataOverride,
			`DELETE FROM system.tenant_usage WHERE tenant_id = $1`, info.ID,
		); err != nil {
			return errors.Wrapf(err, "deleting tenant %d usage", info.ID)
		}

		if execCfg.Settings.Version.IsActive(ctx, clusterversion.TenantSettingsTable) {
			if _, err := execCfg.InternalExecutor.ExecEx(
				ctx, "delete-tenant-settings", txn, sessiondata.NodeUserSessionDataOverride,
				`DELETE FROM system.tenant_settings WHERE tenant_id = $1`, info.ID,
			); err != nil {
				return errors.Wrapf(err, "deleting tenant %d settings", info.ID)
			}
		}

		if !execCfg.Settings.Version.IsActive(ctx, clusterversion.PreSeedTenantSpanConfigs) {
			return nil
		}

		// Clear out all span config records left over by the tenant.
		tenID := roachpb.MakeTenantID(info.ID)
		tenantPrefix := keys.MakeTenantPrefix(tenID)
		tenantSpan := roachpb.Span{
			Key:    tenantPrefix,
			EndKey: tenantPrefix.PrefixEnd(),
		}

		systemTarget, err := spanconfig.MakeTenantKeyspaceTarget(tenID, tenID)
		if err != nil {
			return err
		}
		scKVAccessor := execCfg.SpanConfigKVAccessor.WithTxn(ctx, txn)
		records, err := scKVAccessor.GetSpanConfigRecords(
			ctx, []spanconfig.Target{
				spanconfig.MakeTargetFromSpan(tenantSpan),
				spanconfig.MakeTargetFromSystemTarget(systemTarget),
			},
		)
		if err != nil {
			return err
		}

		toDelete := make([]spanconfig.Target, len(records))
		for i, record := range records {
			toDelete[i] = record.GetTarget()
		}
		return scKVAccessor.UpdateSpanConfigRecords(ctx, toDelete, nil /* toUpsert */)
	})
	return errors.Wrapf(err, "deleting tenant %d record", info.ID)
}

// gcTenantJob clears the tenant's data and removes its record using a GC job.
func gcTenantJob(
	ctx context.Context,
	execCfg *ExecutorConfig,
	txn *kv.Txn,
	user security.SQLUsername,
	tenID uint64,
	synchronous bool,
) (jobspb.JobID, error) {
	// Queue a GC job that will delete the tenant data and finally remove the
	// row from `system.tenants`.
	gcDetails := jobspb.SchemaChangeGCDetails{}
	gcDetails.Tenant = &jobspb.SchemaChangeGCDetails_DroppedTenant{
		ID:       tenID,
		DropTime: timeutil.Now().UnixNano(),
	}
	progress := jobspb.SchemaChangeGCProgress{}
	if synchronous {
		progress.Tenant = &jobspb.SchemaChangeGCProgress_TenantProgress{
			Status: jobspb.SchemaChangeGCProgress_DELETING,
		}
	}
	gcJobRecord := jobs.Record{
		Description:   fmt.Sprintf("GC for tenant %d", tenID),
		Username:      user,
		Details:       gcDetails,
		Progress:      progress,
		NonCancelable: true,
	}
	jobID := execCfg.JobRegistry.MakeJobID()
	if _, err := execCfg.JobRegistry.CreateJobWithTxn(
		ctx, gcJobRecord, jobID, txn,
	); err != nil {
		return 0, err
	}
	return jobID, nil
}

// GCTenant implements the tree.TenantOperator interface.
func (p *planner) GCTenant(ctx context.Context, tenID uint64) error {
	// TODO(jeffswenson): Delete internal_crdb.gc_tenant after the DestroyTenant
	// changes are deployed to all Cockroach Cloud serverless hosts.
	if !p.ExtendedEvalContext().TxnImplicit {
		return errors.Errorf("gc_tenant cannot be used inside a transaction")
	}
	var info *descpb.TenantInfo
	if txnErr := p.ExecCfg().DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		var err error
		info, err = GetTenantRecord(ctx, p.execCfg, p.txn, tenID)
		return err
	}); txnErr != nil {
		return errors.Wrapf(txnErr, "retrieving tenant %d", tenID)
	}

	// Confirm tenant is ready to be cleared.
	if info.State != descpb.TenantInfo_DROP {
		return errors.Errorf("tenant %d is not in state DROP", info.ID)
	}

	_, err := gcTenantJob(
		ctx, p.ExecCfg(), p.Txn(), p.User(), tenID, false, /* synchronous */
	)
	return err
}

// UpdateTenantResourceLimits implements the tree.TenantOperator interface.
func (p *planner) UpdateTenantResourceLimits(
	ctx context.Context,
	tenantID uint64,
	availableRU float64,
	refillRate float64,
	maxBurstRU float64,
	asOf time.Time,
	asOfConsumedRequestUnits float64,
) error {
	const op = "update-resource-limits"
	if err := rejectIfCantCoordinateMultiTenancy(p.execCfg.Codec, op); err != nil {
		return err
	}
	if err := rejectIfSystemTenant(tenantID, op); err != nil {
		return err
	}
	return p.ExecCfg().TenantUsageServer.ReconfigureTokenBucket(
		ctx, p.Txn(), roachpb.MakeTenantID(tenantID),
		availableRU, refillRate, maxBurstRU, asOf, asOfConsumedRequestUnits,
	)
}

// TestingUpdateTenantRecord is a public wrapper around updateTenantRecord
// intended for testing purposes.
func TestingUpdateTenantRecord(
	ctx context.Context, execCfg *ExecutorConfig, txn *kv.Txn, info *descpb.TenantInfo,
) error {
	return updateTenantRecord(ctx, execCfg, txn, info)
}
