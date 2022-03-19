// Copyright 2017 The Cockroach Authors.
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
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog/catconstants"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog/descpb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/types"
)

// pgExtension is virtual schema which contains virtual tables and/or views
// which are used by postgres extensions. Postgres extensions typically install
// these tables and views on the public schema, but we instead do it in
// our own defined virtual table / schema.
var pgExtension = virtualSchema{
	name: catconstants.PgExtensionSchemaName,
	tableDefs: map[descpb.ID]virtualSchemaDef{
		catconstants.PgExtensionGeographyColumnsTableID: pgExtensionGeographyColumnsTable,
		catconstants.PgExtensionGeometryColumnsTableID:  pgExtensionGeometryColumnsTable,
	},
	validWithNoDatabaseContext: false,
}

func postgisColumnsTablePopulator(
	matchingFamily types.Family,
) func(context.Context, *planner, catalog.DatabaseDescriptor, func(...tree.Datum) error) error {
	return func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		return forEachTableDesc(
			ctx,
			p,
			dbContext,
			hideVirtual,
			func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor) error {
				if !table.IsPhysicalTable() {
					return nil
				}
				if p.CheckAnyPrivilege(ctx, table) != nil {
					return nil
				}
				for _, col := range table.PublicColumns() {
					if col.GetType().Family() != matchingFamily {
						continue
					}

					var datumNDims tree.Datum

					// PostGIS is weird on this one! It has the following behavior:
					//
					// * For Geometry, it uses the 2D shape type, all uppercase.
					// * For Geography, use the correct OGR case for the shape type.

					if err := addRow(
						tree.NewDString(db.GetName()),
						tree.NewDString(scName),
						tree.NewDString(table.GetName()),
						tree.NewDString(col.GetName()),
						datumNDims,
					); err != nil {
						return err
					}
				}
				return nil
			},
		)
	}
}

var pgExtensionGeographyColumnsTable = virtualSchemaTable{
	comment: `Shows all defined geography columns. Matches PostGIS' geography_columns functionality.`,
	schema: `
CREATE TABLE pg_extension.geography_columns (
	f_table_catalog name,
	f_table_schema name,
	f_table_name name,
	f_geography_column name,
	coord_dimension integer,
	srid integer,
	type text
)`,
	populate: postgisColumnsTablePopulator(types.GeographyFamily),
}

var pgExtensionGeometryColumnsTable = virtualSchemaTable{
	comment: `Shows all defined geometry columns. Matches PostGIS' geometry_columns functionality.`,
	schema: `
CREATE TABLE pg_extension.geometry_columns (
	f_table_catalog name,
	f_table_schema name,
	f_table_name name,
	f_geometry_column name,
	coord_dimension integer,
	srid integer,
	type text
)`,
	populate: postgisColumnsTablePopulator(types.GeometryFamily),
}
