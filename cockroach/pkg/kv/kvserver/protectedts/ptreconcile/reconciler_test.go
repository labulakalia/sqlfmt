// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package ptreconcile_test

import (
	"context"
	"testing"
	"time"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/base"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/keys"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv/kvserver/protectedts/ptreconcile"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/roachpb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/settings/cluster"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils/testcluster"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/leaktest"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/syncutil"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

func TestReconciler(t *testing.T) {
	defer leaktest.AfterTest(t)()

	ctx := context.Background()
	testutils.RunTrueAndFalse(t, "reconciler", func(t *testing.T, withDeprecatedSpans bool) {
		tc := testcluster.StartTestCluster(t, 1, base.TestClusterArgs{
			ServerArgs: base.TestServerArgs{
				Knobs: base.TestingKnobs{
					ProtectedTS: &protectedts.TestingKnobs{DisableProtectedTimestampForMultiTenant: withDeprecatedSpans},
				},
			},
		})
		defer tc.Stopper().Stop(ctx)

		// Now I want to create some artifacts that should get reconciled away and
		// then make sure that they do and others which should not do not.
		s0 := tc.Server(0)
		ptp := s0.ExecutorConfig().(sql.ExecutorConfig).ProtectedTimestampProvider

		settings := cluster.MakeTestingClusterSettings()
		const testTaskType = "foo"
		var state = struct {
			mu       syncutil.Mutex
			toRemove map[string]struct{}
		}{}
		state.toRemove = map[string]struct{}{}
		r := ptreconcile.New(settings, s0.DB(), ptp,
			ptreconcile.StatusFuncs{
				testTaskType: func(
					ctx context.Context, txn *kv.Txn, meta []byte,
				) (shouldRemove bool, err error) {
					state.mu.Lock()
					defer state.mu.Unlock()
					_, shouldRemove = state.toRemove[string(meta)]
					return shouldRemove, nil
				},
			})
		require.NoError(t, r.StartReconciler(ctx, s0.Stopper()))
		recMeta := "a"
		rec1 := ptpb.Record{
			ID:        uuid.MakeV4().GetBytes(),
			Timestamp: s0.Clock().Now(),
			Mode:      ptpb.PROTECT_AFTER,
			MetaType:  testTaskType,
			Meta:      []byte(recMeta),
		}
		if withDeprecatedSpans {
			rec1.DeprecatedSpans = []roachpb.Span{{Key: keys.MinKey, EndKey: keys.MaxKey}}
		} else {
			rec1.Target = ptpb.MakeClusterTarget()
		}
		require.NoError(t, s0.DB().Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			return ptp.Protect(ctx, txn, &rec1)
		}))

		t.Run("update settings", func(t *testing.T) {
			ptreconcile.ReconcileInterval.Override(ctx, &settings.SV, time.Millisecond)
			testutils.SucceedsSoon(t, func() error {
				require.Equal(t, int64(0), r.Metrics().RecordsRemoved.Count())
				require.Equal(t, int64(0), r.Metrics().ReconciliationErrors.Count())
				if processed := r.Metrics().RecordsProcessed.Count(); processed < 1 {
					return errors.Errorf("expected processed to be at least 1, got %d", processed)
				}
				return nil
			})
		})
		t.Run("reconcile", func(t *testing.T) {
			state.mu.Lock()
			state.toRemove[recMeta] = struct{}{}
			state.mu.Unlock()

			ptreconcile.ReconcileInterval.Override(ctx, &settings.SV, time.Millisecond)
			testutils.SucceedsSoon(t, func() error {
				require.Equal(t, int64(0), r.Metrics().ReconciliationErrors.Count())
				if removed := r.Metrics().RecordsRemoved.Count(); removed != 1 {
					return errors.Errorf("expected processed to be 1, got %d", removed)
				}
				return nil
			})
			require.Regexp(t, protectedts.ErrNotExists, s0.DB().Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
				_, err := ptp.GetRecord(ctx, txn, rec1.ID.GetUUID())
				return err
			}))
		})
	})
}
