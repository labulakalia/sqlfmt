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
	"sort"

	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/base"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/config"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/config/zonepb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/keys"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/kv"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/server/telemetry"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/settings/cluster"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/catalog/descpb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/privilege"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sqlerrors"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sqltelemetry"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/types"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/errorutil"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/protoutil"
	yaml "gopkg.in/yaml.v2"
)

type optionValue struct {
	inheritValue  bool
	explicitValue tree.TypedExpr
}

type setZoneConfigNode struct {
	zoneSpecifier tree.ZoneSpecifier
	allIndexes    bool
	yamlConfig    tree.TypedExpr
	options       map[tree.Name]optionValue
	setDefault    bool

	run setZoneConfigRun
}

// supportedZoneConfigOptions indicates how to translate SQL variable
// assignments in ALTER CONFIGURE ZONE to assignments to the member
// fields of zonepb.ZoneConfig.
var supportedZoneConfigOptions = map[tree.Name]struct {
	requiredType *types.T
	setter       func(*zonepb.ZoneConfig, tree.Datum)
	checkAllowed func(context.Context, *ExecutorConfig, tree.Datum) error // optional
}{
	"range_min_bytes": {
		requiredType: types.Int,
		setter:       func(c *zonepb.ZoneConfig, d tree.Datum) { c.RangeMinBytes = proto.Int64(int64(tree.MustBeDInt(d))) },
	},
	"range_max_bytes": {
		requiredType: types.Int,
		setter:       func(c *zonepb.ZoneConfig, d tree.Datum) { c.RangeMaxBytes = proto.Int64(int64(tree.MustBeDInt(d))) },
	},
	"global_reads": {
		requiredType: types.Bool,
		setter:       func(c *zonepb.ZoneConfig, d tree.Datum) { c.GlobalReads = proto.Bool(bool(tree.MustBeDBool(d))) },
		checkAllowed: func(ctx context.Context, execCfg *ExecutorConfig, d tree.Datum) error {
			if !tree.MustBeDBool(d) {
				// Always allow the value to be unset.
				return nil
			}
			return base.CheckEnterpriseEnabled(
				execCfg.Settings,
				execCfg.ClusterID(),
				execCfg.Organization(),
				"global_reads",
			)
		},
	},
	"num_replicas": {
		requiredType: types.Int,
		setter:       func(c *zonepb.ZoneConfig, d tree.Datum) { c.NumReplicas = proto.Int32(int32(tree.MustBeDInt(d))) },
	},
	"num_voters": {
		requiredType: types.Int,
		setter:       func(c *zonepb.ZoneConfig, d tree.Datum) { c.NumVoters = proto.Int32(int32(tree.MustBeDInt(d))) },
	},
	"gc.ttlseconds": {
		requiredType: types.Int,
		setter: func(c *zonepb.ZoneConfig, d tree.Datum) {
			c.GC = &zonepb.GCPolicy{TTLSeconds: int32(tree.MustBeDInt(d))}
		},
	},
	"constraints": {
		requiredType: types.String,
		setter: func(c *zonepb.ZoneConfig, d tree.Datum) {
			constraintsList := zonepb.ConstraintsList{
				Constraints: c.Constraints,
				Inherited:   c.InheritedConstraints,
			}
			loadYAML(&constraintsList, string(tree.MustBeDString(d)))
			c.Constraints = constraintsList.Constraints
			c.InheritedConstraints = false
		},
	},
	"voter_constraints": {
		requiredType: types.String,
		setter: func(c *zonepb.ZoneConfig, d tree.Datum) {
			voterConstraintsList := zonepb.ConstraintsList{
				Constraints: c.VoterConstraints,
				Inherited:   c.InheritedVoterConstraints(),
			}
			loadYAML(&voterConstraintsList, string(tree.MustBeDString(d)))
			c.VoterConstraints = voterConstraintsList.Constraints
			c.NullVoterConstraintsIsEmpty = true
		},
	},
	"lease_preferences": {
		requiredType: types.String,
		setter: func(c *zonepb.ZoneConfig, d tree.Datum) {
			loadYAML(&c.LeasePreferences, string(tree.MustBeDString(d)))
			c.InheritedLeasePreferences = false
		},
	},
}

// zoneOptionKeys contains the keys from suportedZoneConfigOptions in
// deterministic order. Needed to make the event log output
// deterministic.
var zoneOptionKeys = func() []string {
	l := make([]string, 0, len(supportedZoneConfigOptions))
	for k := range supportedZoneConfigOptions {
		l = append(l, string(k))
	}
	sort.Strings(l)
	return l
}()

func loadYAML(dst interface{}, yamlString string) {
	if err := yaml.UnmarshalStrict([]byte(yamlString), dst); err != nil {
		panic(err)
	}
}

func (p *planner) SetZoneConfig(ctx context.Context, n *tree.SetZoneConfig) (planNode, error) {
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"CONFIGURE ZONE",
	); err != nil {
		return nil, err
	}

	if !p.ExecCfg().Codec.ForSystemTenant() &&
		!secondaryTenantZoneConfigsEnabled.Get(&p.ExecCfg().Settings.SV) {
		// Return an unimplemented error here instead of referencing the cluster
		// setting here as zone configurations for secondary tenants are intended to
		// be hidden.
		return nil, errorutil.UnsupportedWithMultiTenancy(MultitenancyZoneCfgIssueNo)
	}

	if err := checkPrivilegeForSetZoneConfig(ctx, p, n.ZoneSpecifier); err != nil {
		return nil, err
	}

	if err := p.CheckZoneConfigChangePermittedForMultiRegion(
		ctx,
		n.ZoneSpecifier,
		n.Options,
	); err != nil {
		return nil, err
	}

	var yamlConfig tree.TypedExpr

	if n.YAMLConfig != nil {
		// We have a CONFIGURE ZONE = <expr> assignment.
		// This can be either a literal NULL (deletion), or a string containing YAML.
		// We also support byte arrays for backward compatibility with
		// previous versions of CockroachDB.

		var err error
		yamlConfig, err = p.analyzeExpr(
			ctx, n.YAMLConfig, nil, tree.IndexedVarHelper{}, types.String, false /*requireType*/, "configure zone")
		if err != nil {
			return nil, err
		}

		switch typ := yamlConfig.ResolvedType(); typ.Family() {
		case types.UnknownFamily:
			// Unknown occurs if the user entered a literal NULL. That's OK and will mean deletion.
		case types.StringFamily:
		case types.BytesFamily:
		default:
			return nil, pgerror.Newf(pgcode.InvalidParameterValue,
				"zone config must be of type string or bytes, not %s", typ)
		}
	}

	var options map[tree.Name]optionValue
	if n.Options != nil {
		// We have a CONFIGURE ZONE USING ... assignment.
		// Here we are constrained by the supported ZoneConfig fields,
		// as described by supportedZoneConfigOptions above.

		options = make(map[tree.Name]optionValue)
		for _, opt := range n.Options {
			if _, alreadyExists := options[opt.Key]; alreadyExists {
				return nil, pgerror.Newf(pgcode.InvalidParameterValue,
					"duplicate zone config parameter: %q", tree.ErrString(&opt.Key))
			}
			req, ok := supportedZoneConfigOptions[opt.Key]
			if !ok {
				return nil, pgerror.Newf(pgcode.InvalidParameterValue,
					"unsupported zone config parameter: %q", tree.ErrString(&opt.Key))
			}
			telemetry.Inc(
				sqltelemetry.SchemaSetZoneConfigCounter(
					n.ZoneSpecifier.TelemetryName(),
					string(opt.Key),
				),
			)
			if opt.Value == nil {
				options[opt.Key] = optionValue{inheritValue: true, explicitValue: nil}
				continue
			}
			valExpr, err := p.analyzeExpr(
				ctx, opt.Value, nil, tree.IndexedVarHelper{}, req.requiredType, true /*requireType*/, string(opt.Key))
			if err != nil {
				return nil, err
			}
			options[opt.Key] = optionValue{inheritValue: false, explicitValue: valExpr}
		}
	}

	return &setZoneConfigNode{
		zoneSpecifier: n.ZoneSpecifier,
		allIndexes:    n.AllIndexes,
		yamlConfig:    yamlConfig,
		options:       options,
		setDefault:    n.SetDefault,
	}, nil
}

func checkPrivilegeForSetZoneConfig(ctx context.Context, p *planner, zs tree.ZoneSpecifier) error {
	// For system ranges, the system database, or system tables, the user must be
	// an admin. Otherwise we require CREATE privileges on the database or table
	// in question.
	if zs.NamedZone != "" {
		return p.RequireAdminRole(ctx, "alter system ranges")
	}
	if zs.Database != "" {
		if zs.Database == "system" {
			return p.RequireAdminRole(ctx, "alter the system database")
		}
		dbDesc, err := p.Descriptors().GetImmutableDatabaseByName(ctx, p.txn,
			string(zs.Database), tree.DatabaseLookupFlags{Required: true})
		if err != nil {
			return err
		}
		dbCreatePrivilegeErr := p.CheckPrivilege(ctx, dbDesc, privilege.CREATE)
		dbZoneConfigPrivilegeErr := p.CheckPrivilege(ctx, dbDesc, privilege.ZONECONFIG)

		// Can set ZoneConfig if user has either CREATE privilege or ZONECONFIG privilege at the Database level
		if dbZoneConfigPrivilegeErr == nil || dbCreatePrivilegeErr == nil {
			return nil
		}

		return pgerror.Newf(pgcode.InsufficientPrivilege,
			"user %s does not have %s or %s privilege on %s %s",
			p.SessionData().User(), privilege.ZONECONFIG, privilege.CREATE, dbDesc.DescriptorType(), dbDesc.GetName())
	}
	tableDesc, err := p.resolveTableForZone(ctx, &zs)
	if err != nil {
		if zs.TargetsIndex() && zs.TableOrIndex.Table.ObjectName == "" {
			err = errors.WithHint(err, "try specifying the index as <tablename>@<indexname>")
		}
		return err
	}
	if tableDesc.GetParentID() == keys.SystemDatabaseID {
		return p.RequireAdminRole(ctx, "alter system tables")
	}

	// Can set ZoneConfig if user has either CREATE privilege or ZONECONFIG privilege at the Table level
	tableCreatePrivilegeErr := p.CheckPrivilege(ctx, tableDesc, privilege.CREATE)
	tableZoneConfigPrivilegeErr := p.CheckPrivilege(ctx, tableDesc, privilege.ZONECONFIG)

	if tableCreatePrivilegeErr == nil || tableZoneConfigPrivilegeErr == nil {
		return nil
	}

	return pgerror.Newf(pgcode.InsufficientPrivilege,
		"user %s does not have %s or %s privilege on %s %s",
		p.SessionData().User(), privilege.ZONECONFIG, privilege.CREATE, tableDesc.DescriptorType(), tableDesc.GetName())
}

// setZoneConfigRun contains the run-time state of setZoneConfigNode during local execution.
type setZoneConfigRun struct {
	numAffected int
}

// ReadingOwnWrites implements the planNodeReadingOwnWrites interface.
// This is because CONFIGURE ZONE performs multiple KV operations on descriptors
// and expects to see its own writes.
func (n *setZoneConfigNode) ReadingOwnWrites()            {}
func (n *setZoneConfigNode) Next(runParams) (bool, error) { return false, nil }
func (n *setZoneConfigNode) Values() tree.Datums          { return nil }
func (*setZoneConfigNode) Close(context.Context)          {}

func (n *setZoneConfigNode) FastPathResults() (int, bool) { return n.run.numAffected, true }

// Check that there are not duplicated values for a particular
// constraint. For example, constraints [+region=us-east1,+region=us-east2]
// will be rejected. Additionally, invalid constraints such as
// [+region=us-east1, -region=us-east1] will also be rejected.
func validateNoRepeatKeysInZone(zone *zonepb.ZoneConfig) error {
	if err := validateNoRepeatKeysInConjunction(zone.Constraints); err != nil {
		return err
	}
	return validateNoRepeatKeysInConjunction(zone.VoterConstraints)
}

func validateNoRepeatKeysInConjunction(conjunctions []zonepb.ConstraintsConjunction) error {
	for _, constraints := range conjunctions {
		// Because we expect to have a small number of constraints, a nested
		// loop is probably better than allocating a map.
		for i, curr := range constraints.Constraints {
			for _, other := range constraints.Constraints[i+1:] {
				// We don't want to enter the other validation logic if both of the constraints
				// are attributes, due to the keys being the same for attributes.
				if curr.Key == "" && other.Key == "" {
					if curr.Value == other.Value {
						return pgerror.Newf(pgcode.CheckViolation,
							"incompatible zone constraints: %q and %q", curr, other)
					}
				} else {
					if curr.Type == zonepb.Constraint_REQUIRED {
						if other.Type == zonepb.Constraint_REQUIRED && other.Key == curr.Key ||
							other.Type == zonepb.Constraint_PROHIBITED && other.Key == curr.Key && other.Value == curr.Value {
							return pgerror.Newf(pgcode.CheckViolation,
								"incompatible zone constraints: %q and %q", curr, other)
						}
					} else if curr.Type == zonepb.Constraint_PROHIBITED {
						// If we have a -k=v pair, verify that there are not any
						// +k=v pairs in the constraints.
						if other.Type == zonepb.Constraint_REQUIRED && other.Key == curr.Key && other.Value == curr.Value {
							return pgerror.Newf(pgcode.CheckViolation,
								"incompatible zone constraints: %q and %q", curr, other)
						}
					}
				}
			}
		}
	}
	return nil
}

// accumulateUniqueConstraints returns a list of unique constraints in the
// given zone config proto.
func accumulateUniqueConstraints(zone *zonepb.ZoneConfig) []zonepb.Constraint {
	constraints := make([]zonepb.Constraint, 0)
	addToValidate := func(c zonepb.Constraint) {
		var alreadyInList bool
		for _, val := range constraints {
			if c == val {
				alreadyInList = true
				break
			}
		}
		if !alreadyInList {
			constraints = append(constraints, c)
		}
	}
	for _, constraints := range zone.Constraints {
		for _, constraint := range constraints.Constraints {
			addToValidate(constraint)
		}
	}
	for _, constraints := range zone.VoterConstraints {
		for _, constraint := range constraints.Constraints {
			addToValidate(constraint)
		}
	}
	for _, leasePreferences := range zone.LeasePreferences {
		for _, constraint := range leasePreferences.Constraints {
			addToValidate(constraint)
		}
	}
	return constraints
}

// validateZoneAttrsAndLocalities ensures that all constraints/lease preferences
// specified in the new zone config snippet are actually valid, meaning that
// they match at least one node. This protects against user typos causing
// zone configs that silently don't work as intended.

// validateZoneAttrsAndLocalitiesForSystemTenant performs all the constraint/
// lease preferences validation for the system tenant. The system tenant is
// allowed to reference both locality and non-locality attributes as it has
// access to node information via the NodeStatusServer.

// validateZoneLocalitiesForSecondaryTenants performs all the constraint/lease
// preferences validation for secondary tenants. Secondary tenants are only
// allowed to reference locality attributes as they only have access to region
// information via the RegionServer. Even then, they're only allowed to
// reference the "region" and "zone" tiers.
//

// MultitenancyZoneCfgIssueNo points to the multitenancy zone config issue number.
const MultitenancyZoneCfgIssueNo = 49854

type zoneConfigUpdate struct {
	id    descpb.ID
	value []byte
}

func prepareZoneConfigWrites(
	ctx context.Context,
	execCfg *ExecutorConfig,
	targetID descpb.ID,
	table catalog.TableDescriptor,
	zone *zonepb.ZoneConfig,
	hasNewSubzones bool,
) (_ *zoneConfigUpdate, err error) {
	if len(zone.Subzones) > 0 {
		st := execCfg.Settings
		zone.SubzoneSpans, err = GenerateSubzoneSpans(
			st, execCfg.ClusterID(), execCfg.Codec, table, zone.Subzones, hasNewSubzones)
		if err != nil {
			return nil, err
		}
	} else {
		// To keep the Subzone and SubzoneSpan arrays consistent
		zone.SubzoneSpans = nil
	}
	if zone.IsSubzonePlaceholder() && len(zone.Subzones) == 0 {
		return &zoneConfigUpdate{id: targetID}, nil
	}
	buf, err := protoutil.Marshal(zone)
	if err != nil {
		return nil, pgerror.Wrap(err, pgcode.CheckViolation, "could not marshal zone config")
	}
	return &zoneConfigUpdate{id: targetID, value: buf}, nil
}

func writeZoneConfig(
	ctx context.Context,
	txn *kv.Txn,
	targetID descpb.ID,
	table catalog.TableDescriptor,
	zone *zonepb.ZoneConfig,
	execCfg *ExecutorConfig,
	hasNewSubzones bool,
) (numAffected int, err error) {
	update, err := prepareZoneConfigWrites(ctx, execCfg, targetID, table, zone, hasNewSubzones)
	if err != nil {
		return 0, err
	}
	return writeZoneConfigUpdate(ctx, txn, execCfg, update)
}

func writeZoneConfigUpdate(
	ctx context.Context, txn *kv.Txn, execCfg *ExecutorConfig, update *zoneConfigUpdate,
) (numAffected int, _ error) {
	if update.value == nil {
		return execCfg.InternalExecutor.Exec(ctx, "delete-zone", txn,
			"DELETE FROM system.zones WHERE id = $1", update.id)
	}
	return execCfg.InternalExecutor.Exec(ctx, "update-zone", txn,
		"UPSERT INTO system.zones (id, config) VALUES ($1, $2)", update.id, update.value)
}

// getZoneConfigRaw looks up the zone config with the given ID. Unlike
// getZoneConfig, it does not attempt to ascend the zone config hierarchy. If no
// zone config exists for the given ID, it returns nil.
func getZoneConfigRaw(
	ctx context.Context, txn *kv.Txn, codec keys.SQLCodec, settings *cluster.Settings, id descpb.ID,
) (*zonepb.ZoneConfig, error) {
	kv, err := txn.Get(ctx, config.MakeZoneKey(codec, id))
	if err != nil {
		return nil, err
	}
	if kv.Value == nil {
		return nil, nil
	}
	var zone zonepb.ZoneConfig
	if err := kv.ValueProto(&zone); err != nil {
		return nil, err
	}
	return &zone, nil
}

// getZoneConfigRawBatch looks up the zone config with the given IDs.
// Unlike getZoneConfig, it does not attempt to ascend the zone config hierarchy.
// If no zone config exists for the given ID, the map entry is not provided.
func getZoneConfigRawBatch(
	ctx context.Context,
	txn *kv.Txn,
	codec keys.SQLCodec,
	settings *cluster.Settings,
	ids []descpb.ID,
) (map[descpb.ID]*zonepb.ZoneConfig, error) {
	b := txn.NewBatch()
	for _, id := range ids {
		b.Get(config.MakeZoneKey(codec, id))
	}
	if err := txn.Run(ctx, b); err != nil {
		return nil, err
	}
	ret := make(map[descpb.ID]*zonepb.ZoneConfig, len(b.Results))
	for idx, r := range b.Results {
		if r.Err != nil {
			return nil, r.Err
		}
		var zone zonepb.ZoneConfig
		row := r.Rows[0]
		if row.Value == nil {
			continue
		}
		if err := row.ValueProto(&zone); err != nil {
			return nil, err
		}
		ret[ids[idx]] = &zone
	}
	return ret, nil
}

// RemoveIndexZoneConfigs removes the zone configurations for some
// indexes being dropped. It is a no-op if there is no zone
// configuration, there's no index zone configs to be dropped,
// or it is run on behalf of a tenant.
//
// It operates entirely on the current goroutine and is thus able to
// reuse an existing client.Txn safely.
func RemoveIndexZoneConfigs(
	ctx context.Context,
	txn *kv.Txn,
	execCfg *ExecutorConfig,
	tableDesc catalog.TableDescriptor,
	indexIDs []uint32,
) error {
	if !execCfg.Codec.ForSystemTenant() {
		// Tenants are agnostic to zone configs.
		return nil
	}
	zone, err := getZoneConfigRaw(ctx, txn, execCfg.Codec, execCfg.Settings, tableDesc.GetID())
	if err != nil {
		return err
	}
	// If there are no zone configs, there's nothing to remove.
	if zone == nil {
		return nil
	}

	// Look through all of the subzones and determine if we need to remove any
	// of them. We only want to rewrite the zone config below if there's actual
	// work to be done here.
	zcRewriteNecessary := false
	for _, indexID := range indexIDs {
		for _, s := range zone.Subzones {
			if s.IndexID == indexID {
				// We've found an subzone that matches the given indexID. Delete all of
				// this index's subzones and move on to the next index.
				zone.DeleteIndexSubzones(indexID)
				zcRewriteNecessary = true
				break
			}
		}
	}

	if zcRewriteNecessary {
		// Ignore CCL required error to allow schema change to progress.
		_, err = writeZoneConfig(ctx, txn, tableDesc.GetID(), tableDesc, zone, execCfg, false /* hasNewSubzones */)
		if err != nil && !sqlerrors.IsCCLRequiredError(err) {
			return err
		}
	}

	return nil
}
