// Copyright 2022 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

import (
	"fmt"
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/roachpb"
	"github.com/stretchr/testify/require"
)

func TestRewriteKeyToTenantPrefix(t *testing.T) {
	for _, tc := range []struct {
		oldTenant, newTenant uint64
		suffix               string
	}{
		{1, 1, "abc"},
		{1, 2, "acc"},
		{2, 1, "abc"},
		{1, 1 << 15, "abc"},
		{1 << 15, 1, "abc"},
		{2, 1 << 15, "abc"},
		{1 << 15, 2, "abc"},
		{1 << 15, 1 << 15, "abc"},
		{1 << 13, 1 << 15, "abc"},
		{1 << 15, 1 << 13, "abc"},
	} {
		t.Run(fmt.Sprintf("%d to %d %s", tc.oldTenant, tc.newTenant, tc.suffix), func(t *testing.T) {
			old := append(MakeSQLCodec(roachpb.MakeTenantID(tc.oldTenant)).TablePrefix(5), tc.suffix...)
			new := MakeTenantPrefix(roachpb.MakeTenantID(tc.newTenant))
			got, err := rewriteKeyToTenantPrefix(old, new)
			require.NoError(t, err)

			expect := append(MakeSQLCodec(roachpb.MakeTenantID(tc.newTenant)).TablePrefix(5), tc.suffix...)
			require.Equal(t, expect, got)
		})
	}
}
