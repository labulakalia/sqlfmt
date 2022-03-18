// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package testcat_test

import (
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/opt/testutils/opttester"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/opt/testutils/testcat"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/datadriven"
)

func TestCatalog(t *testing.T) {
	defer leaktest.AfterTest(t)()

	datadriven.Walk(t, testutils.TestDataPath(t), func(t *testing.T, path string) {
		catalog := testcat.New()

		datadriven.RunTest(t, path, func(t *testing.T, d *datadriven.TestData) string {
			tester := opttester.New(catalog, d.Input)
			return tester.RunCommand(t, d)
		})
	})
}
