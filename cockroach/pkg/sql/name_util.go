// Copyright 2021 The Cockroach Authors.
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

	"sqlfmt/cockroach/pkg/clusterversion"
	"sqlfmt/cockroach/pkg/kv"
	"sqlfmt/cockroach/pkg/sql/catalog"
	"sqlfmt/cockroach/pkg/sql/catalog/catalogkeys"
	"sqlfmt/cockroach/pkg/sql/catalog/descpb"
	"sqlfmt/cockroach/pkg/util/log"
)

func (p *planner) dropNamespaceEntry(
	ctx context.Context, b *kv.Batch, desc catalog.MutableDescriptor,
) {
	// Delete current namespace entry.
	deleteNamespaceEntryAndMaybeAddDrainingName(ctx, b, p, desc, desc)
}

func (p *planner) renameNamespaceEntry(
	ctx context.Context, b *kv.Batch, oldNameKey catalog.NameKey, desc catalog.MutableDescriptor,
) {
	// Delete old namespace entry.
	deleteNamespaceEntryAndMaybeAddDrainingName(ctx, b, p, oldNameKey, desc)

	// Write new namespace entry.
	marshalledKey := catalogkeys.EncodeNameKey(p.ExecCfg().Codec, desc)
	if p.extendedEvalCtx.Tracing.KVTracingEnabled() {
		log.VEventf(ctx, 2, "CPut %s -> %d", marshalledKey, desc.GetID())
	}
	b.CPut(marshalledKey, desc.GetID(), nil)
}

func deleteNamespaceEntryAndMaybeAddDrainingName(
	ctx context.Context,
	b *kv.Batch,
	p *planner,
	nameKeyToDelete catalog.NameKey,
	desc catalog.MutableDescriptor,
) {
	if !p.execCfg.Settings.Version.IsActive(ctx, clusterversion.AvoidDrainingNames) {
		desc.AddDrainingName(descpb.NameInfo{
			ParentID:       nameKeyToDelete.GetParentID(),
			ParentSchemaID: nameKeyToDelete.GetParentSchemaID(),
			Name:           nameKeyToDelete.GetName(),
		})
		return
	}
	marshalledKey := catalogkeys.EncodeNameKey(p.ExecCfg().Codec, nameKeyToDelete)
	if p.extendedEvalCtx.Tracing.KVTracingEnabled() {
		log.VEventf(ctx, 2, "Del %s", marshalledKey)
	}
	b.Del(marshalledKey)
}
