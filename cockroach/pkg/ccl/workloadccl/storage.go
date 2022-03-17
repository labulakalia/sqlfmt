// Copyright 2022 The Cockroach Authors.
//
// Licensed as a CockroachDB Enterprise file under the Cockroach Community
// License (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//     https://sqlfmt/cockroach/blob/master/licenses/CCL.txt

package workloadccl

import (
	"context"

	"sqlfmt/cockroach/pkg/base"
	"sqlfmt/cockroach/pkg/cloud"
	// Import all the cloud provider storage we care about.
	_ "sqlfmt/cockroach/pkg/cloud/amazon"
	_ "sqlfmt/cockroach/pkg/cloud/azure"
	_ "sqlfmt/cockroach/pkg/cloud/gcp"
	"sqlfmt/cockroach/pkg/security"
	clustersettings "sqlfmt/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/errors"
)

const storageError = `failed to create google cloud client ` +
	`(You may need to setup the GCS application default credentials: ` +
	`'gcloud auth application-default login --project=cockroach-shared')`

// GetStorage returns a cloud storage implementation
// The caller is responsible for closing it.
func GetStorage(ctx context.Context, cfg FixtureConfig) (cloud.ExternalStorage, error) {
	switch cfg.StorageProvider {
	case "gs", "s3", "azure":
	default:
		return nil, errors.AssertionFailedf("unsupported external storage provider; valid providers are gs, s3, and azure")
	}

	s, err := cloud.ExternalStorageFromURI(ctx, cfg.ObjectPathToURI(),
		base.ExternalIODirConfig{}, clustersettings.MakeClusterSettings(),
		nil, security.SQLUsername{}, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, storageError)
	}
	return s, nil
}
