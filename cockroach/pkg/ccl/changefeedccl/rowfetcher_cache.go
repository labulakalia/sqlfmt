// Copyright 2018 The Cockroach Authors.
//
// Licensed as a CockroachDB Enterprise file under the Cockroach Community
// License (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//     https://sqlfmt/cockroach/blob/master/licenses/CCL.txt

package changefeedccl

import (
	"context"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/jobs/jobspb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/keys"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/roachpb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog/descpb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog/descs"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog/lease"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/row"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/rowenc"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/cache"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/encoding"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

// rowFetcherCache maintains a cache of single table RowFetchers. Given a key
// with an mvcc timestamp, it retrieves the correct TableDescriptor for that key
// and returns a Fetcher initialized with that table. This Fetcher's
// StartScanFrom can be used to turn that key (or all the keys making up the
// column families of one row) into a row.
type rowFetcherCache struct {
	codec           keys.SQLCodec
	leaseMgr        *lease.Manager
	fetchers        *cache.UnorderedCache
	watchedFamilies map[watchedFamily]struct{}

	collection *descs.Collection
	db         *kv.DB

	a tree.DatumAlloc
}

type cachedFetcher struct {
	tableDesc  catalog.TableDescriptor
	fetcher    row.Fetcher
	familyDesc descpb.ColumnFamilyDescriptor
	skip       bool
}

type watchedFamily struct {
	tableID    descpb.ID
	familyName string
}

var rfCacheConfig = cache.Config{
	Policy: cache.CacheFIFO,
	// TODO: If we find ourselves thrashing here in changefeeds on many tables,
	// we can improve performance by eagerly evicting versions using Resolved notifications.
	// A old version with a timestamp entirely before a notification can be safely evicted.
	ShouldEvict: func(size int, _ interface{}, _ interface{}) bool { return size > 1024 },
}

type idVersion struct {
	id      descpb.ID
	version descpb.DescriptorVersion
	family  descpb.FamilyID
}

func newRowFetcherCache(
	ctx context.Context,
	codec keys.SQLCodec,
	leaseMgr *lease.Manager,
	cf *descs.CollectionFactory,
	db *kv.DB,
	details jobspb.ChangefeedDetails,
) *rowFetcherCache {
	specs := details.TargetSpecifications
	watchedFamilies := make(map[watchedFamily]struct{}, len(specs))
	for _, s := range specs {
		watchedFamilies[watchedFamily{tableID: s.TableID, familyName: s.FamilyName}] = struct{}{}
	}
	return &rowFetcherCache{
		codec:           codec,
		leaseMgr:        leaseMgr,
		collection:      cf.NewCollection(ctx, nil /* TemporarySchemaProvider */),
		db:              db,
		fetchers:        cache.NewUnorderedCache(rfCacheConfig),
		watchedFamilies: watchedFamilies,
	}
}

func (c *rowFetcherCache) TableDescForKey(
	ctx context.Context, key roachpb.Key, ts hlc.Timestamp,
) (catalog.TableDescriptor, descpb.FamilyID, error) {
	var tableDesc catalog.TableDescriptor
	key, err := c.codec.StripTenantPrefix(key)
	if err != nil {
		return nil, descpb.FamilyID(0), err
	}
	remaining, tableID, _, err := rowenc.DecodePartialTableIDIndexID(key)
	if err != nil {
		return nil, descpb.FamilyID(0), err
	}

	familyID, err := keys.DecodeFamilyKey(key)
	if err != nil {
		return nil, descpb.FamilyID(0), err
	}

	family := descpb.FamilyID(familyID)

	// Retrieve the target TableDescriptor from the lease manager. No caching
	// is attempted because the lease manager does its own caching.
	desc, err := c.leaseMgr.Acquire(ctx, ts, tableID)
	if err != nil {
		// Manager can return all kinds of errors during chaos, but based on
		// its usage, none of them should ever be terminal.
		return nil, family, changefeedbase.MarkRetryableError(err)
	}
	tableDesc = desc.Underlying().(catalog.TableDescriptor)
	// Immediately release the lease, since we only need it for the exact
	// timestamp requested.
	desc.Release(ctx)
	if tableDesc.ContainsUserDefinedTypes() {
		// If the table contains user defined types, then use the
		// descs.Collection to retrieve a TableDescriptor with type metadata
		// hydrated. We open a transaction here only because the
		// descs.Collection needs one to get a read timestamp. We do this lookup
		// again behind a conditional to avoid allocating any transaction
		// metadata if the table has user defined types. This can be bypassed
		// once (#53751) is fixed. Once the descs.Collection can take in a read
		// timestamp rather than a whole transaction, we can use the
		// descs.Collection directly here.
		// TODO (SQL Schema): #53751.
		if err := c.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			err := txn.SetFixedTimestamp(ctx, ts)
			if err != nil {
				return err
			}
			tableDesc, err = c.collection.GetImmutableTableByID(ctx, txn, tableID, tree.ObjectLookupFlags{})
			return err
		}); err != nil {
			// Manager can return all kinds of errors during chaos, but based on
			// its usage, none of them should ever be terminal.
			return nil, family, changefeedbase.MarkRetryableError(err)
		}
		// Immediately release the lease, since we only need it for the exact
		// timestamp requested.
		c.collection.ReleaseAll(ctx)
	}

	// Skip over the column data.
	for skippedCols := 0; skippedCols < tableDesc.GetPrimaryIndex().NumKeyColumns(); skippedCols++ {
		l, err := encoding.PeekLength(remaining)
		if err != nil {
			return nil, family, err
		}
		remaining = remaining[l:]
	}

	return tableDesc, family, nil
}

// ErrUnwatchedFamily is a sentinel error that indicates this part of the row
// is not being watched and does not need to be decoded.
var ErrUnwatchedFamily = errors.New("watched table but unwatched family")

func (c *rowFetcherCache) RowFetcherForColumnFamily(
	tableDesc catalog.TableDescriptor, family descpb.FamilyID,
) (*row.Fetcher, error) {
	idVer := idVersion{id: tableDesc.GetID(), version: tableDesc.GetVersion(), family: family}
	if v, ok := c.fetchers.Get(idVer); ok {
		f := v.(*cachedFetcher)
		if f.skip {
			return &f.fetcher, ErrUnwatchedFamily
		}
		// Ensure that all user defined types are up to date with the cached
		// version and the desired version to use the cache. It is safe to use
		// UserDefinedTypeColsHaveSameVersion if we have a hit because we are
		// guaranteed that the tables have the same version. Additionally, these
		// fetchers are always initialized with a single tabledesc.Immutable.
		// TODO (zinger): Only check types used in the relevant family.
		if catalog.UserDefinedTypeColsHaveSameVersion(tableDesc, f.tableDesc) {
			return &f.fetcher, nil
		}
	}

	familyDesc, err := tableDesc.FindFamilyByID(family)
	if err != nil {
		return nil, err
	}

	f := &cachedFetcher{
		tableDesc:  tableDesc,
		familyDesc: *familyDesc,
	}
	rf := &f.fetcher

	_, wholeTableWatched := c.watchedFamilies[watchedFamily{tableID: tableDesc.GetID()}]
	if !wholeTableWatched {
		_, familyWatched := c.watchedFamilies[watchedFamily{tableID: tableDesc.GetID(), familyName: familyDesc.Name}]
		if !familyWatched {
			f.skip = true
			return rf, ErrUnwatchedFamily
		}
	}

	var spec descpb.IndexFetchSpec

	// TODO (zinger): Make fetchColumnIDs only the family and the primary key.
	// This seems to cause an error without further work but would be more efficient.
	if err := rowenc.InitIndexFetchSpec(
		&spec, c.codec, tableDesc, tableDesc.GetPrimaryIndex(), tableDesc.PublicColumnIDs(),
	); err != nil {
		return nil, err
	}

	if err := rf.Init(
		context.TODO(),
		false, /* reverse */
		descpb.ScanLockingStrength_FOR_NONE,
		descpb.ScanLockingWaitPolicy_BLOCK,
		0, /* lockTimeout */
		&c.a,
		nil, /* memMonitor */
		&spec,
	); err != nil {
		return nil, err
	}

	// Necessary because virtual columns are not populated.
	// TODO(radu): should we stop requesting those columns from the fetcher?
	rf.IgnoreUnexpectedNulls = true

	c.fetchers.Add(idVer, f)
	return rf, nil
}
