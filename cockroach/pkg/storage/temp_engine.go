// Copyright 2017 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package storage

import (
	"context"

	"github.com/cockroachdb/pebble"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv/kvserver/diskmap"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/log"
)



type pebbleTempEngine struct {
	db *pebble.DB
}

// Close implements the diskmap.Factory interface.
func (r *pebbleTempEngine) Close() {
	err := r.db.Close()
	if err != nil {
		log.Fatalf(context.TODO(), "%v", err)
	}
}

// NewSortedDiskMap implements the diskmap.Factory interface.
func (r *pebbleTempEngine) NewSortedDiskMap() diskmap.SortedDiskMap {
	return newPebbleMap(r.db, false /* allowDuplications */)
}

// NewSortedDiskMultiMap implements the diskmap.Factory interface.
func (r *pebbleTempEngine) NewSortedDiskMultiMap() diskmap.SortedDiskMap {
	return newPebbleMap(r.db, true /* allowDuplicates */)
}
