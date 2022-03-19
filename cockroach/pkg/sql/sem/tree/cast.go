// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree

import (
	"strings"
	"unicode/utf8"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/errors"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/oidext"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sessiondata"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/types"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util"
	"github.com/lib/pq/oid"
)

// CastContext represents the contexts in which a cast can be performed. There
// are three types of cast contexts: explicit, assignment, and implicit. Not all
// casts can be performed in all contexts. See the description of each context
// below for more details.
//
// The concept of cast contexts is taken directly from Postgres's cast behavior.
// More information can be found in the Postgres documentation on type
// conversion: https://www.postgresql.org/docs/current/typeconv.html
type CastContext uint8

const (
	_ CastContext = iota
	// CastContextExplicit is a cast performed explicitly with the syntax
	// CAST(x AS T) or x::T.
	CastContextExplicit
	// CastContextAssignment is a cast implicitly performed during an INSERT,
	// UPSERT, or UPDATE statement.
	CastContextAssignment
	// CastContextImplicit is a cast performed implicitly. For example, the DATE
	// below is implicitly cast to a TIMESTAMPTZ so that the values can be
	// compared.
	//
	//   SELECT '2021-01-10'::DATE < now()
	//
	CastContextImplicit
)

// String returns the representation of CastContext as a string.
func (cc CastContext) String() string {
	switch cc {
	case CastContextExplicit:
		return "explicit"
	case CastContextAssignment:
		return "assignment"
	case CastContextImplicit:
		return "implicit"
	default:
		return "invalid"
	}
}

// contextOrigin indicates the source of information for a cast's maximum
// context (see cast.maxContext below). It is only used to annotate entries in
// castMap and to perform assertions on cast entries in the init function. It
// has no effect on the behavior of a cast.
type contextOrigin uint8

const (
	_ contextOrigin = iota
	// contextOriginPgCast specifies that a cast's maximum context is based on
	// information in Postgres's pg_cast table.
	contextOriginPgCast
	// contextOriginAutomaticIOConversion specifies that a cast's maximum
	// context is not included in Postgres's pg_cast table. In Postgres's
	// internals, these casts are evaluated by each data type's input and output
	// functions.
	//
	// Automatic casts can only convert to or from string types [1]. Conversions
	// to string types are assignment casts and conversions from string types
	// are explicit casts [2]. These rules are asserted in the init function.
	//
	// [1] https://www.postgresql.org/docs/13/catalog-pg-cast.html#CATALOG-PG-CAST
	// [2] https://www.postgresql.org/docs/13/sql-createcast.html#SQL-CREATECAST-NOTES
	contextOriginAutomaticIOConversion
	// contextOriginLegacyConversion is used for casts that are not supported by
	// Postgres, but are supported by CockroachDB and continue to be supported
	// for backwards compatibility.
	contextOriginLegacyConversion
)

// cast includes details about a cast from one OID to another.
// TODO(mgartner, otan): Move PerformCast logic to this struct.
type cast struct {
	// maxContext is the maximum context in which the cast is allowed. A cast
	// can only be performed in a context that is at or below the specified
	// maximum context.
	//
	// CastContextExplicit casts can only be performed in an explicit context.
	//
	// CastContextAssignment casts can be performed in an explicit context or in
	// an assignment context in an INSERT, UPSERT, or UPDATE statement.
	//
	// CastContextImplicit casts can be performed in any context.
	maxContext CastContext
	// origin is the source of truth for the cast's context. It is used to
	// annotate entries in castMap and to perform assertions on cast entries in
	// the init function. It has no effect on the behavior of a cast.
	origin contextOrigin
	// volatility indicates whether the result of the cast is dependent only on
	// the source value, or dependent on outside factors (such as parameter
	// variables or table contents).
	volatility Volatility
	// volatilityHint is an optional string for VolatilityStable casts. When
	// set, it is used as an error hint suggesting a possible workaround when
	// stable casts are not allowed.
	volatilityHint string
	// intervalStyleAffected is true if the cast is a stable cast when
	// SemaContext.IntervalStyleEnabled is true, and an immutable cast
	// otherwise.
	intervalStyleAffected bool
	// dateStyleAffected is true if the cast is a stable cast when
	// SemaContext.DateStyleEnabled is true, and an immutable cast otherwise.
	dateStyleAffected bool
}

// castMap defines valid casts. It maps from a source OID to a target OID to a
// cast struct that contains information about the cast. Some possible casts,
// such as casts from the UNKNOWN type and casts from a type to the identical
// type, are not defined in the castMap and are instead codified in ValidCast.
//
// Validation is performed on the map in init().
//
// Entries with a contextOriginPgCast origin were automatically generated by the
// cast_map_gen.sh script. The script outputs some types that we do not support.
// Those types were manually deleted. Entries with
// contextOriginAutomaticIOConversion origin were manually added.
var castMap = map[oid.Oid]map[oid.Oid]cast{
	oid.T_bit: {
		oid.T_bit:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:   {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:   {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:   {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_bool: {
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float4:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float8:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int2:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:    {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_char: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oidext.T_box2d: {
		oidext.T_geometry: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_bpchar: {
		oid.T_bpchar:  {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions from bpchar to other types.
		oid.T_bit:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_box2d: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bytea:    {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_date: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "CHAR to DATE casts depend on session DateStyle; use parse_date(string) instead",
			dateStyleAffected: true,
		},
		oid.T_float4:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_inet:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_interval: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility:            VolatilityImmutable,
			volatilityHint:        "CHAR to INTERVAL casts depend on session IntervalStyle; use parse_interval(string) instead",
			intervalStyleAffected: true,
		},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_record:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_time: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "CHAR to TIME casts depend on session DateStyle; use parse_time(string) instead",
			dateStyleAffected: true,
		},
		oid.T_timestamp: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "CHAR to TIMESTAMP casts are context-dependent because of relative timestamp strings " +
				"like 'now' and session settings such as DateStyle; use parse_timestamp(string) instead.",
			dateStyleAffected: true,
		},
		oid.T_timestamptz: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
		},
		oid.T_timetz: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "CHAR to TIMETZ casts depend on session DateStyle; use parse_timetz(char) instead",
			dateStyleAffected: true,
		},
		oid.T_uuid:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_void:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_bytea: {
		oidext.T_geography: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_uuid:         {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		// Casts from BYTEA to string types are stable, since they depend on
		// the bytea_output session variable.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_char: {
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:    {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_name: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		// Automatic I/O conversions from "char" to other types.
		oid.T_bit:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_box2d: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bytea:    {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_date: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    `"char" to DATE casts depend on session DateStyle; use parse_date(string) instead`,
			dateStyleAffected: true,
		},
		oid.T_float4:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_inet:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_interval: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility:            VolatilityImmutable,
			volatilityHint:        `"char" to INTERVAL casts depend on session IntervalStyle; use parse_interval(string) instead`,
			intervalStyleAffected: true,
		},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_record:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_time: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    `"char" to TIME casts depend on session DateStyle; use parse_time(string) instead`,
			dateStyleAffected: true,
		},
		oid.T_timestamp: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: `"char" to TIMESTAMP casts are context-dependent because of relative timestamp strings ` +
				"like 'now' and session settings such as DateStyle; use parse_timestamp(string) instead.",
			dateStyleAffected: true,
		},
		oid.T_timestamptz: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
		},
		oid.T_timetz: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    `"char" to TIMETZ casts depend on session DateStyle; use parse_timetz(string) instead`,
			dateStyleAffected: true,
		},
		oid.T_uuid:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_void:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_date: {
		oid.T_float4:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float8:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int2:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int8:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_timestamp:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timestamptz: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityStable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility: VolatilityImmutable,
			volatilityHint: "DATE to CHAR casts are dependent on DateStyle; consider " +
				"using to_char(date) instead.",
			dateStyleAffected: true,
		},
		oid.T_char: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility: VolatilityImmutable,
			volatilityHint: `DATE to "char" casts are dependent on DateStyle; consider ` +
				"using to_char(date) instead.",
			dateStyleAffected: true,
		},
		oid.T_name: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility: VolatilityImmutable,
			volatilityHint: "DATE to NAME casts are dependent on DateStyle; consider " +
				"using to_char(date) instead.",
			dateStyleAffected: true,
		},
		oid.T_text: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility: VolatilityImmutable,
			volatilityHint: "DATE to STRING casts are dependent on DateStyle; consider " +
				"using to_char(date) instead.",
			dateStyleAffected: true,
		},
		oid.T_varchar: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility: VolatilityImmutable,
			volatilityHint: "DATE to VARCHAR casts are dependent on DateStyle; consider " +
				"using to_char(date) instead.",
			dateStyleAffected: true,
		},
	},
	oid.T_float4: {
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float8:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_interval: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:  {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		// Casts from FLOAT4 to string types are stable, since they depend on the
		// extra_float_digits session variable.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_float8: {
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float4:   {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_interval: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:  {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		// Casts from FLOAT8 to string types are stable, since they depend on the
		// extra_float_digits session variable.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oidext.T_geography: {
		oid.T_bytea:        {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oidext.T_geometry: {
		oidext.T_box2d:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_bytea:        {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_text:         {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_inet: {
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_char: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_int2: {
		oid.T_bit:          {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_bool:         {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_date:         {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float4:       {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_interval:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regnamespace: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regproc:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regprocedure: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regrole:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regtype:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timestamp:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_timestamptz:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_varbit:       {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_int4: {
		oid.T_bit:          {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_bool:         {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_char:         {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_date:         {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float4:       {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_interval:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regnamespace: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regproc:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regprocedure: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regrole:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regtype:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timestamp:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_timestamptz:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_varbit:       {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_int8: {
		oid.T_bit:          {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_bool:         {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_date:         {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float4:       {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_interval:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regnamespace: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regproc:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regprocedure: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regrole:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regtype:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timestamp:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_timestamptz:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_varbit:       {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_interval: {
		oid.T_float4:   {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float8:   {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int2:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int8:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_interval: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_numeric:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_time:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar: {
			maxContext:            CastContextAssignment,
			origin:                contextOriginAutomaticIOConversion,
			volatility:            VolatilityImmutable,
			volatilityHint:        "INTERVAL to CHAR casts depend on IntervalStyle; consider using to_char(interval)",
			intervalStyleAffected: true,
		},
		oid.T_char: {
			maxContext:            CastContextAssignment,
			origin:                contextOriginAutomaticIOConversion,
			volatility:            VolatilityImmutable,
			volatilityHint:        `INTERVAL to "char" casts depend on IntervalStyle; consider using to_char(interval)`,
			intervalStyleAffected: true,
		},
		oid.T_name: {
			maxContext:            CastContextAssignment,
			origin:                contextOriginAutomaticIOConversion,
			volatility:            VolatilityImmutable,
			volatilityHint:        "INTERVAL to NAME casts depend on IntervalStyle; consider using to_char(interval)",
			intervalStyleAffected: true,
		},
		oid.T_text: {
			maxContext:            CastContextAssignment,
			origin:                contextOriginAutomaticIOConversion,
			volatility:            VolatilityImmutable,
			volatilityHint:        "INTERVAL to STRING casts depend on IntervalStyle; consider using to_char(interval)",
			intervalStyleAffected: true,
		},
		oid.T_varchar: {
			maxContext:            CastContextAssignment,
			origin:                contextOriginAutomaticIOConversion,
			volatility:            VolatilityImmutable,
			volatilityHint:        "INTERVAL to VARCHAR casts depend on IntervalStyle; consider using to_char(interval)",
			intervalStyleAffected: true,
		},
	},
	oid.T_jsonb: {
		oid.T_bool:         {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float4:       {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_name: {
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_char: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		// Automatic I/O conversions from NAME to other types.
		oid.T_bit:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_box2d: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bytea:    {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_date: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "NAME to DATE casts depend on session DateStyle; use parse_date(string) instead",
			dateStyleAffected: true,
		},
		oid.T_float4:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_inet:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_interval: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility:            VolatilityImmutable,
			volatilityHint:        "NAME to INTERVAL casts depend on session IntervalStyle; use parse_interval(string) instead",
			intervalStyleAffected: true,
		},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_record:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_time: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "NAME to TIME casts depend on session DateStyle; use parse_time(string) instead",
			dateStyleAffected: true,
		},
		oid.T_timestamp: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "NAME to TIMESTAMP casts are context-dependent because of relative timestamp strings " +
				"like 'now' and session settings such as DateStyle; use parse_timestamp(string) instead.",
			dateStyleAffected: true,
		},
		oid.T_timestamptz: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
		},
		oid.T_timetz: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "NAME to TIMETZ casts depend on session DateStyle; use parse_timetz(string) instead",
			dateStyleAffected: true,
		},
		oid.T_uuid:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_void:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_numeric: {
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float4:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float8:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_interval: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:  {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_oid: {
		// TODO(mgartner): Casts to INT2 should not be allowed.
		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regnamespace: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regproc:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regprocedure: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regrole:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regtype:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_record: {
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_regclass: {
		// TODO(mgartner): Casts to INT2 should not be allowed.
		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_regnamespace: {
		// TODO(mgartner): Casts to INT2 should not be allowed.
		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_regproc: {
		// TODO(mgartner): Casts to INT2 should not be allowed.
		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regprocedure: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_regprocedure: {
		// TODO(mgartner): Casts to INT2 should not be allowed.
		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regproc:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_regrole: {
		// TODO(mgartner): Casts to INT2 should not be allowed.
		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_regtype: {
		// TODO(mgartner): Casts to INT2 should not be allowed.
		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_text: {
		oid.T_bpchar:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_char:        {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geometry: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_name:        {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityStable},
		// We include a TEXT->TEXT entry to mimic the VARCHAR->VARCHAR entry
		// that is included in the pg_cast table. Postgres doesn't include a
		// TEXT->TEXT entry because it does not allow width-limited TEXT types,
		// so a cast from TEXT->TEXT is always a trivial no-op because the types
		// are always identical (see ValidCast). Because we support
		// width-limited TEXT types with STRING(n), it is possible to have
		// non-identical TEXT types. So, we must include a TEXT->TEXT entry so
		// that casts from STRING(n)->STRING(m) are valid.
		//
		// TODO(#72980): If we use the VARCHAR OID for STRING(n) types rather
		// then the TEXT OID, and we can remove this entry.
		oid.T_text:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions from TEXT to other types.
		oid.T_bit:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_box2d: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bytea:    {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_date: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "STRING to DATE casts depend on session DateStyle; use parse_date(string) instead",
			dateStyleAffected: true,
		},
		oid.T_float4:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_inet:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_interval: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility:            VolatilityImmutable,
			volatilityHint:        "STRING to INTERVAL casts depend on session IntervalStyle; use parse_interval(string) instead",
			intervalStyleAffected: true,
		},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_record:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_time: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "STRING to TIME casts depend on session DateStyle; use parse_time(string) instead",
			dateStyleAffected: true,
		},
		oid.T_timestamp: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "STRING to TIMESTAMP casts are context-dependent because of relative timestamp strings " +
				"like 'now' and session settings such as DateStyle; use parse_timestamp(string) instead.",
			dateStyleAffected: true,
		},
		oid.T_timestamptz: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
		},
		oid.T_timetz: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "STRING to TIMETZ casts depend on session DateStyle; use parse_timetz(string) instead",
			dateStyleAffected: true,
		},
		oid.T_uuid:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_void:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_time: {
		oid.T_interval: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_time:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timetz:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityStable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_timestamp: {
		oid.T_date:        {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float4:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float8:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int2:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int8:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_time:        {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timestamp:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timestamptz: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityStable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility: VolatilityImmutable,
			volatilityHint: "TIMESTAMP to CHAR casts are dependent on DateStyle; consider " +
				"using to_char(timestamp) instead.",
			dateStyleAffected: true,
		},
		oid.T_char: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility: VolatilityImmutable,
			volatilityHint: `TIMESTAMP to "char" casts are dependent on DateStyle; consider ` +
				"using to_char(timestamp) instead.",
			dateStyleAffected: true,
		},
		oid.T_name: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility: VolatilityImmutable,
			volatilityHint: "TIMESTAMP to NAME casts are dependent on DateStyle; consider " +
				"using to_char(timestamp) instead.",
			dateStyleAffected: true,
		},
		oid.T_text: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility: VolatilityImmutable,
			volatilityHint: "TIMESTAMP to STRING casts are dependent on DateStyle; consider " +
				"using to_char(timestamp) instead.",
			dateStyleAffected: true,
		},
		oid.T_varchar: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility: VolatilityImmutable,
			volatilityHint: "TIMESTAMP to VARCHAR casts are dependent on DateStyle; consider " +
				"using to_char(timestamp) instead.",
			dateStyleAffected: true,
		},
	},
	oid.T_timestamptz: {
		oid.T_date:    {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityStable},
		oid.T_float4:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float8:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int2:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int8:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_time:    {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityStable},
		oid.T_timestamp: {
			maxContext:     CastContextAssignment,
			origin:         contextOriginPgCast,
			volatility:     VolatilityStable,
			volatilityHint: "TIMESTAMPTZ to TIMESTAMP casts depend on the current timezone; consider using AT TIME ZONE 'UTC' instead",
		},
		oid.T_timestamptz: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timetz:      {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityStable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "TIMESTAMPTZ to CHAR casts depend on the current timezone; consider " +
				"using to_char(t AT TIME ZONE 'UTC') instead.",
		},
		oid.T_char: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: `TIMESTAMPTZ to "char" casts depend on the current timezone; consider ` +
				"using to_char(t AT TIME ZONE 'UTC') instead.",
		},
		oid.T_name: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "TIMESTAMPTZ to NAME casts depend on the current timezone; consider " +
				"using to_char(t AT TIME ZONE 'UTC') instead.",
		},
		oid.T_text: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "TIMESTAMPTZ to STRING casts depend on the current timezone; consider " +
				"using to_char(t AT TIME ZONE 'UTC') instead.",
		},
		oid.T_varchar: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "TIMESTAMPTZ to VARCHAR casts depend on the current timezone; consider " +
				"using to_char(t AT TIME ZONE 'UTC') instead.",
		},
	},
	oid.T_timetz: {
		oid.T_time:   {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timetz: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_uuid: {
		oid.T_bytea: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_varbit: {
		oid.T_bit:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:   {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:   {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int8:   {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions to string types.
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_varchar: {
		oid.T_bpchar:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_char:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_name:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityStable},
		oid.T_text:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar:  {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		// Automatic I/O conversions from VARCHAR to other types.
		oid.T_bit:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_box2d: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bytea:    {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_date: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "VARCHAR to DATE casts depend on session DateStyle; use parse_date(string) instead",
			dateStyleAffected: true,
		},
		oid.T_float4:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_inet:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_interval: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			// TODO(mgartner): This should be stable.
			volatility:            VolatilityImmutable,
			volatilityHint:        "VARCHAR to INTERVAL casts depend on session IntervalStyle; use parse_interval(string) instead",
			intervalStyleAffected: true,
		},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_record:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_time: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "VARCHAR to TIME casts depend on session DateStyle; use parse_time(string) instead",
			dateStyleAffected: true,
		},
		oid.T_timestamp: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "VARCHAR to TIMESTAMP casts are context-dependent because of relative timestamp strings " +
				"like 'now' and session settings such as DateStyle; use parse_timestamp(string) instead.",
			dateStyleAffected: true,
		},
		oid.T_timestamptz: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
		},
		oid.T_timetz: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "VARCHAR to TIMETZ casts depend on session DateStyle; use parse_timetz(string) instead",
			dateStyleAffected: true,
		},
		oid.T_uuid:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_void:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_void: {
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
}

// init performs sanity checks on castMap.
func init() {
	var stringTypes = [...]oid.Oid{
		oid.T_bpchar,
		oid.T_name,
		oid.T_char,
		oid.T_varchar,
		oid.T_text,
	}
	isStringType := func(o oid.Oid) bool {
		for _, strOid := range stringTypes {
			if o == strOid {
				return true
			}
		}
		return false
	}

	typeName := func(o oid.Oid) string {
		if name, ok := oidext.TypeName(o); ok {
			return name
		}
		panic(errors.AssertionFailedf("no type name for oid %d", o))
	}

	// Assert that there is a cast to and from every string type.
	for _, strType := range stringTypes {
		for otherType := range castMap {
			if strType == otherType {
				continue
			}
			strTypeName := typeName(strType)
			otherTypeName := typeName(otherType)
			if _, from := castMap[strType][otherType]; !from && otherType != oid.T_unknown {
				panic(errors.AssertionFailedf("there must be a cast from %s to %s", strTypeName, otherTypeName))
			}
			if _, to := castMap[otherType][strType]; !to {
				panic(errors.AssertionFailedf("there must be a cast from %s to %s", otherTypeName, strTypeName))
			}
		}
	}

	// Assert that each cast is valid.
	for src, tgts := range castMap {
		for tgt, ent := range tgts {
			srcStr := typeName(src)
			tgtStr := typeName(tgt)

			// Assert that maxContext, method, and origin have been set.
			if ent.maxContext == CastContext(0) {
				panic(errors.AssertionFailedf("cast from %s to %s has no maxContext set", srcStr, tgtStr))
			}
			if ent.origin == contextOrigin(0) {
				panic(errors.AssertionFailedf("cast from %s to %s has no origin set", srcStr, tgtStr))
			}

			// Casts from a type to the same type should be implicit.
			if src == tgt {
				if ent.maxContext != CastContextImplicit {
					panic(errors.AssertionFailedf(
						"cast from %s to %s must be an implicit cast",
						srcStr, tgtStr,
					))
				}
			}

			// Automatic I/O conversions to string types are assignment casts.
			if isStringType(tgt) && ent.origin == contextOriginAutomaticIOConversion &&
				ent.maxContext != CastContextAssignment {
				panic(errors.AssertionFailedf(
					"automatic conversion from %s to %s must be an assignment cast",
					srcStr, tgtStr,
				))
			}

			// Automatic I/O conversions from string types are explicit casts.
			if isStringType(src) && !isStringType(tgt) && ent.origin == contextOriginAutomaticIOConversion &&
				ent.maxContext != CastContextExplicit {
				panic(errors.AssertionFailedf(
					"automatic conversion from %s to %s must be an explicit cast",
					srcStr, tgtStr,
				))
			}
		}
	}
}

// ForEachCast calls fn for every valid cast from a source type to a target
// type.
func ForEachCast(fn func(src, tgt oid.Oid)) {
	for src, tgts := range castMap {
		for tgt := range tgts {
			fn(src, tgt)
		}
	}
}

// ValidCast returns true if a valid cast exists from src to tgt in the given
// context.
func ValidCast(src, tgt *types.T, ctx CastContext) bool {
	srcFamily := src.Family()
	tgtFamily := tgt.Family()

	// If src and tgt are array types, check for a valid cast between their
	// content types.
	if srcFamily == types.ArrayFamily && tgtFamily == types.ArrayFamily {
		return ValidCast(src.ArrayContents(), tgt.ArrayContents(), ctx)
	}

	// If src and tgt are tuple types, check for a valid cast between each
	// corresponding tuple element.
	//
	// Casts from a tuple type to AnyTuple are a no-op so they are always valid.
	// If tgt is AnyTuple, we continue to lookupCast below which contains a
	// special case for these casts.
	if srcFamily == types.TupleFamily && tgtFamily == types.TupleFamily && tgt != types.AnyTuple {
		srcTypes := src.TupleContents()
		tgtTypes := tgt.TupleContents()
		// The tuple types must have the same number of elements.
		if len(srcTypes) != len(tgtTypes) {
			return false
		}
		for i := range srcTypes {
			if ok := ValidCast(srcTypes[i], tgtTypes[i], ctx); !ok {
				return false
			}
		}
		return true
	}

	// If src and tgt are not both array or tuple types, check castMap for a
	// valid cast.
	c, ok := lookupCast(src, tgt, false /* intervalStyleEnabled */, false /* dateStyleEnabled */)
	if ok {
		return c.maxContext >= ctx
	}

	return false
}

// lookupCast returns a cast that describes the cast from src to tgt if it
// exists. If it does not exist, ok=false is returned.
func lookupCast(src, tgt *types.T, intervalStyleEnabled, dateStyleEnabled bool) (cast, bool) {
	srcFamily := src.Family()
	tgtFamily := tgt.Family()
	srcFamily.Name()

	// Unknown is the type given to an expression that statically evaluates
	// to NULL. NULL can be immutably cast to any type in any context.
	if srcFamily == types.UnknownFamily {
		return cast{
			maxContext: CastContextImplicit,
			volatility: VolatilityImmutable,
		}, true
	}

	// Enums have dynamic OIDs, so they can't be populated in castMap. Instead,
	// we dynamically create cast structs for valid enum casts.
	if srcFamily == types.EnumFamily && tgtFamily == types.StringFamily {
		// Casts from enum types to strings are immutable and allowed in
		// assignment contexts.
		return cast{
			maxContext: CastContextAssignment,
			volatility: VolatilityImmutable,
		}, true
	}
	if tgtFamily == types.EnumFamily {
		switch srcFamily {
		case types.StringFamily:
			// Casts from string types to enums are immutable and allowed in
			// explicit contexts.
			return cast{
				maxContext: CastContextExplicit,
				volatility: VolatilityImmutable,
			}, true
		case types.UnknownFamily:
			// Casts from unknown to enums are immutable and allowed in implicit
			// contexts.
			return cast{
				maxContext: CastContextImplicit,
				volatility: VolatilityImmutable,
			}, true
		}
	}

	// Casts from array types to string types are stable and allowed in
	// assignment contexts.
	if srcFamily == types.ArrayFamily && tgtFamily == types.StringFamily {
		return cast{
			maxContext: CastContextAssignment,
			volatility: VolatilityStable,
		}, true
	}

	// Casts from array and tuple types to string types are immutable and
	// allowed in assignment contexts.
	// TODO(mgartner): Tuple to string casts should be stable. They are
	// immutable to avoid backward incompatibility with previous versions, but
	// this is incorrect and can causes corrupt indexes, corrupt tables, and
	// incorrect query results.
	if srcFamily == types.TupleFamily && tgtFamily == types.StringFamily {
		return cast{
			maxContext: CastContextAssignment,
			volatility: VolatilityImmutable,
		}, true
	}

	// Casts from any tuple type to AnyTuple are no-ops, so they are implicit
	// and immutable.
	if srcFamily == types.TupleFamily && tgt == types.AnyTuple {
		return cast{
			maxContext: CastContextImplicit,
			volatility: VolatilityImmutable,
		}, true
	}

	// Casts from string types to array and tuple types are stable and allowed
	// in explicit contexts.
	if srcFamily == types.StringFamily &&
		(tgtFamily == types.ArrayFamily || tgtFamily == types.TupleFamily) {
		return cast{
			maxContext: CastContextExplicit,
			volatility: VolatilityStable,
		}, true
	}

	if tgts, ok := castMap[src.Oid()]; ok {
		if c, ok := tgts[tgt.Oid()]; ok {
			if intervalStyleEnabled && c.intervalStyleAffected ||
				dateStyleEnabled && c.dateStyleAffected {
				c.volatility = VolatilityStable
			}
			return c, true
		}
	}

	// If src and tgt are the same type, the immutable cast is valid in any
	// context. This logic is intentionally after the lookup into castMap so
	// that entries in castMap are preferred.
	if src.Oid() == tgt.Oid() {
		return cast{
			maxContext: CastContextImplicit,
			volatility: VolatilityImmutable,
		}, true
	}

	return cast{}, false
}

// LookupCastVolatility returns the volatility of a valid cast.
func LookupCastVolatility(from, to *types.T, sd *sessiondata.SessionData) (_ Volatility, ok bool) {
	fromFamily := from.Family()
	toFamily := to.Family()
	// Special case for casting between arrays.
	if fromFamily == types.ArrayFamily && toFamily == types.ArrayFamily {
		return LookupCastVolatility(from.ArrayContents(), to.ArrayContents(), sd)
	}
	// Special case for casting between tuples.
	if fromFamily == types.TupleFamily && toFamily == types.TupleFamily {
		fromTypes := from.TupleContents()
		toTypes := to.TupleContents()
		// Handle case where an overload makes a tuple get casted to tuple{}.
		if len(toTypes) == 1 && toTypes[0].Family() == types.AnyFamily {
			return VolatilityStable, true
		}
		if len(fromTypes) != len(toTypes) {
			return 0, false
		}
		maxVolatility := VolatilityLeakProof
		for i := range fromTypes {
			v, lookupOk := LookupCastVolatility(fromTypes[i], toTypes[i], sd)
			if !lookupOk {
				return 0, false
			}
			if v > maxVolatility {
				maxVolatility = v
			}
		}
		return maxVolatility, true
	}

	intervalStyleEnabled := false
	dateStyleEnabled := false
	if sd != nil {
		intervalStyleEnabled = sd.IntervalStyleEnabled
		dateStyleEnabled = sd.DateStyleEnabled
	}

	cast, ok := lookupCast(from, to, intervalStyleEnabled, dateStyleEnabled)
	if !ok {
		return 0, false
	}
	return cast.volatility, true
}

// PerformCast performs a cast from the provided Datum to the specified
// types.T. The original datum is returned if its type is identical
// to the specified type.
func PerformCast(ctx *EvalContext, d Datum, t *types.T) (Datum, error) {
	ret, err := performCastWithoutPrecisionTruncation(ctx, d, t, true /* truncateWidth */)
	if err != nil {
		return nil, err
	}
	return AdjustValueToType(t, ret)
}

// PerformAssignmentCast performs an assignment cast from the provided Datum to
// the specified type. The original datum is returned if its type is identical
// to the specified type.
//
// It is similar to PerformCast, but differs because it errors when a bit-array
// or string values are too wide for the given type, rather than truncating the
// value. The one exception to this is casts to the special "char" type which
// are truncated.
func PerformAssignmentCast(ctx *EvalContext, d Datum, t *types.T) (Datum, error) {
	if !ValidCast(d.ResolvedType(), t, CastContextAssignment) {
		return nil, pgerror.Newf(
			pgcode.CannotCoerce,
			"invalid assignment cast: %s -> %s", d.ResolvedType(), t,
		)
	}
	d, err := performCastWithoutPrecisionTruncation(ctx, d, t, false /* truncateWidth */)
	if err != nil {
		return nil, err
	}
	return AdjustValueToType(t, d)
}

// AdjustValueToType checks that the width (for strings, byte arrays, and bit
// strings) and scale (decimal). and, shape/srid (for geospatial types) fits the
// specified column type.
//
// Additionally, some precision truncation may occur for the specified column type.
//
// In case of decimals, it can truncate fractional digits in the input
// value in order to fit the target column. If the input value fits the target
// column, it is returned unchanged. If the input value can be truncated to fit,
// then a truncated copy is returned. Otherwise, an error is returned.
//
// In the case of time, it can truncate fractional digits of time datums
// to its relevant rounding for the given type definition.
//
// In the case of geospatial types, it will check whether the SRID and Shape in the
// datum matches the type definition.
//
// This method is used by casts and parsing. It is important to note that this
// function will error if the given value is too wide for the given type. For
// explicit casts and parsing, inVal should be truncated before this function is
// called so that an error is not returned. For assignment casts, inVal should
// not be truncated before this function is called, so that an error is
// returned. The one exception for assignment casts is for the special "char"
// type. An assignment cast to "char" does not error and truncates a value if
// the width of the value is wider than a single character. For this exception,
// AdjustValueToType performs the truncation itself.
func AdjustValueToType(typ *types.T, inVal Datum) (outVal Datum, err error) {
	switch typ.Family() {
	case types.StringFamily, types.CollatedStringFamily:
		var sv string
		if v, ok := AsDString(inVal); ok {
			sv = string(v)
		} else if v, ok := inVal.(*DCollatedString); ok {
			sv = v.Contents
		}
		sv = adjustStringValueToType(typ, sv)
		if typ.Width() > 0 && utf8.RuneCountInString(sv) > int(typ.Width()) {
			return nil, pgerror.Newf(pgcode.StringDataRightTruncation,
				"value too long for type %s",
				typ.SQLString())
		}

		if typ.Oid() == oid.T_bpchar || typ.Oid() == oid.T_char {
			if _, ok := AsDString(inVal); ok {
				return NewDString(sv), nil
			} else if _, ok := inVal.(*DCollatedString); ok {
				return NewDCollatedString(sv, typ.Locale(), &CollationEnvironment{})
			}
		}
	case types.IntFamily:
		if v, ok := AsDInt(inVal); ok {
			if typ.Width() == 32 || typ.Width() == 16 {
				// Width is defined in bits.
				width := uint(typ.Width() - 1)

				// We're performing range checks in line with Go's
				// implementation of math.(Max|Min)(16|32) numbers that store
				// the boundaries of the allowed range.
				// NOTE: when updating the code below, make sure to update
				// execgen/cast_gen_util.go as well.
				shifted := v >> width
				if (v >= 0 && shifted > 0) || (v < 0 && shifted < -1) {
					if typ.Width() == 16 {
						return nil, ErrInt2OutOfRange
					}
					return nil, ErrInt4OutOfRange
				}
			}
		}
	case types.BitFamily:
		if v, ok := AsDBitArray(inVal); ok {
			if typ.Width() > 0 {
				bitLen := v.BitLen()
				switch typ.Oid() {
				case oid.T_varbit:
					if bitLen > uint(typ.Width()) {
						return nil, pgerror.Newf(pgcode.StringDataRightTruncation,
							"bit string length %d too large for type %s", bitLen, typ.SQLString())
					}
				default:
					if bitLen != uint(typ.Width()) {
						return nil, pgerror.Newf(pgcode.StringDataLengthMismatch,
							"bit string length %d does not match type %s", bitLen, typ.SQLString())
					}
				}
			}
		}
	case types.DecimalFamily:
		if inDec, ok := inVal.(*DDecimal); ok {
			if inDec.Form != apd.Finite || typ.Precision() == 0 {
				// Non-finite form or unlimited target precision, so no need to limit.
				break
			}
			if int64(typ.Precision()) >= inDec.NumDigits() && typ.Scale() == inDec.Exponent {
				// Precision and scale of target column are sufficient.
				break
			}

			var outDec DDecimal
			outDec.Set(&inDec.Decimal)
			err := LimitDecimalWidth(&outDec.Decimal, int(typ.Precision()), int(typ.Scale()))
			if err != nil {
				return nil, errors.Wrapf(err, "type %s", typ.SQLString())
			}
			return &outDec, nil
		}
	case types.ArrayFamily:
		if inArr, ok := inVal.(*DArray); ok {
			var outArr *DArray
			elementType := typ.ArrayContents()
			for i, inElem := range inArr.Array {
				outElem, err := AdjustValueToType(elementType, inElem)
				if err != nil {
					return nil, err
				}
				if outElem != inElem {
					if outArr == nil {
						outArr = &DArray{}
						*outArr = *inArr
						outArr.Array = make(Datums, len(inArr.Array))
						copy(outArr.Array, inArr.Array[:i])
					}
				}
				if outArr != nil {
					outArr.Array[i] = inElem
				}
			}
			if outArr != nil {
				return outArr, nil
			}
		}
	case types.TimeFamily:
		if in, ok := inVal.(*DTime); ok {
			return in.Round(TimeFamilyPrecisionToRoundDuration(typ.Precision())), nil
		}
	case types.TimestampFamily:
		if in, ok := inVal.(*DTimestamp); ok {
			return in.Round(TimeFamilyPrecisionToRoundDuration(typ.Precision()))
		}
	case types.TimestampTZFamily:
		if in, ok := inVal.(*DTimestampTZ); ok {
			return in.Round(TimeFamilyPrecisionToRoundDuration(typ.Precision()))
		}
	case types.TimeTZFamily:
		if in, ok := inVal.(*DTimeTZ); ok {
			return in.Round(TimeFamilyPrecisionToRoundDuration(typ.Precision())), nil
		}
	case types.IntervalFamily:
		if in, ok := inVal.(*DInterval); ok {
			itm, err := typ.IntervalTypeMetadata()
			if err != nil {
				return nil, err
			}
			return NewDInterval(in.Duration, itm), nil
		}
	case types.GeometryFamily:
	case types.GeographyFamily:

	}
	return inVal, nil
}

// adjustStringToType checks that the width for strings fits the
// specified column type.
func adjustStringValueToType(typ *types.T, sv string) string {
	switch typ.Oid() {
	case oid.T_char:
		// "char" is supposed to truncate long values
		return util.TruncateString(sv, 1)
	case oid.T_bpchar:
		// bpchar types truncate trailing whitespace.
		return strings.TrimRight(sv, " ")
	}
	return sv
}

// formatBitArrayToType formats bit arrays such that they fill the total width
// if too short, or truncate if too long.
func formatBitArrayToType(d *DBitArray, t *types.T) *DBitArray {
	if t.Width() == 0 || d.BitLen() == uint(t.Width()) {
		return d
	}
	a := d.BitArray.Clone()
	switch t.Oid() {
	case oid.T_varbit:
		// VARBITs do not have padding attached, so only truncate.
		if uint(t.Width()) < a.BitLen() {
			a = a.ToWidth(uint(t.Width()))
		}
	default:
		a = a.ToWidth(uint(t.Width()))
	}
	return &DBitArray{a}
}

// performCastWithoutPrecisionTruncation performs the cast, but does not perform
// precision truncation. For example, if d is of type DECIMAL(6, 2) and t is
// DECIMAL(4, 2), d is not truncated to fit into t. However, if truncateWidth is
// true, widths are truncated to match the target type t for some types,
// including the bit and string types. If truncateWidth is false, the input
// datum is not truncated.
//
// In an ideal state, components of AdjustValueToType should be embedded into
// this function, but the code base needs a general refactor of parsing
// and casting logic before this can happen.
// See also: #55094.
func performCastWithoutPrecisionTruncation(
	ctx *EvalContext, d Datum, t *types.T, truncateWidth bool,
) (Datum, error) {
	// No conversion is needed if d is NULL.
	if d == DNull {
		return d, nil
	}

	// If we're casting a DOidWrapper, then we want to cast the wrapped datum.
	// It is also reasonable to lose the old Oid value too.
	// Note that we pass in nil as the first argument since we're not interested
	// in evaluating the placeholders.
	return nil, pgerror.Newf(
		pgcode.CannotCoerce, "invalid cast: %s -> %s", d.ResolvedType(), t)
}

// performIntToOidCast casts the input integer to the OID type given by the
// input types.T.
func performIntToOidCast(ctx *EvalContext, t *types.T, v DInt) (Datum, error) {
	switch t.Oid() {
	case oid.T_oid:
		return &DOid{semanticType: t, DInt: v}, nil
	case oid.T_regtype:
		// Mapping an dOid to a regtype is easy: we have a hardcoded map.
		ret := &DOid{semanticType: t, DInt: v}
		if typ, ok := types.OidToType[oid.Oid(v)]; ok {
			ret.name = typ.PGName()
		} else if types.IsOIDUserDefinedType(oid.Oid(v)) {
			typ, err := ctx.Planner.ResolveTypeByOID(ctx.Context, oid.Oid(v))
			if err != nil {
				return nil, err
			}
			ret.name = typ.PGName()
		} else if v == 0 {
			return wrapAsZeroOid(t), nil
		}
		return ret, nil

	case oid.T_regproc, oid.T_regprocedure:
		// Mapping an dOid to a regproc is easy: we have a hardcoded map.
		name, ok := OidToBuiltinName[oid.Oid(v)]
		ret := &DOid{semanticType: t, DInt: v}
		if !ok {
			if v == 0 {
				return wrapAsZeroOid(t), nil
			}
			return ret, nil
		}
		ret.name = name
		return ret, nil

	default:
		if v == 0 {
			return wrapAsZeroOid(t), nil
		}

		dOid, err := ctx.Planner.ResolveOIDFromOID(ctx.Ctx(), t, NewDOid(v))
		if err != nil {
			dOid = NewDOid(v)
			dOid.semanticType = t
		}
		return dOid, nil
	}
}

func roundDecimalToInt(ctx *EvalContext, d *apd.Decimal) (int64, error) {
	var newD apd.Decimal
	if _, err := DecimalCtx.RoundToIntegralValue(&newD, d); err != nil {
		return 0, err
	}
	i, err := newD.Int64()
	if err != nil {
		return 0, ErrIntOutOfRange
	}
	return i, nil
}