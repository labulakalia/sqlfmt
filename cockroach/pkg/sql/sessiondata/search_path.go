// Copyright 2016 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package sessiondata

import (
	"bytes"
	"strings"

	//_ "github.com/labulakalia/sqlfmt/cockroach/pkg/security"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/lexbase"
)


// SearchPath represents a list of namespaces to search builtins in.
// The names must be normalized (as per Name.Normalize) already.
type SearchPath struct {
	paths                []string
	containsPgCatalog    bool
	containsPgExtension  bool
	containsPgTempSchema bool
	tempSchemaName       string
	userSchemaName       string
}

// EmptySearchPath is a SearchPath with no schema names in it.
var EmptySearchPath = SearchPath{}



// MakeSearchPath returns a new immutable SearchPath struct. The paths slice
// must not be modified after hand-off to MakeSearchPath.
func MakeSearchPath(paths []string) SearchPath {
	containsPgCatalog := false
	containsPgExtension := false
	containsPgTempSchema := false

	return SearchPath{
		paths:                paths,
		containsPgCatalog:    containsPgCatalog,
		containsPgExtension:  containsPgExtension,
		containsPgTempSchema: containsPgTempSchema,
	}
}

// WithTemporarySchemaName returns a new immutable SearchPath struct with
// the tempSchemaName supplied and the same paths as before.
// This should be called every time a session creates a temporary schema
// for the first time.
func (s SearchPath) WithTemporarySchemaName(tempSchemaName string) SearchPath {
	return SearchPath{
		paths:                s.paths,
		containsPgCatalog:    s.containsPgCatalog,
		containsPgTempSchema: s.containsPgTempSchema,
		containsPgExtension:  s.containsPgExtension,
		userSchemaName:       s.userSchemaName,
		tempSchemaName:       tempSchemaName,
	}
}

// WithUserSchemaName returns a new immutable SearchPath struct with the
// userSchemaName populated and the same values for all other fields as before.
func (s SearchPath) WithUserSchemaName(userSchemaName string) SearchPath {
	return SearchPath{
		paths:                s.paths,
		containsPgCatalog:    s.containsPgCatalog,
		containsPgTempSchema: s.containsPgTempSchema,
		containsPgExtension:  s.containsPgExtension,
		userSchemaName:       userSchemaName,
		tempSchemaName:       s.tempSchemaName,
	}
}

// UpdatePaths returns a new immutable SearchPath struct with the paths supplied
// and the same tempSchemaName and userSchemaName as before.
func (s SearchPath) UpdatePaths(paths []string) SearchPath {
	return MakeSearchPath(paths).WithTemporarySchemaName(s.tempSchemaName).WithUserSchemaName(s.userSchemaName)
}

// MaybeResolveTemporarySchema returns the session specific temporary schema
// for the pg_temp alias (only if a temporary schema exists). It acts as a pass
// through for all other schema names.
func (s SearchPath) MaybeResolveTemporarySchema(schemaName string) (string, error) {
	return schemaName, nil
}

// Iter returns an iterator through the search path. We must include the
// implicit pg_catalog and temporary schema at the beginning of the search path,
// unless they have been explicitly set later by the user.
// We also include pg_extension in the path, as this normally be used in place
// of the public schema. This should be read before "public" is read.
// "The system catalog schema, pg_catalog, is always searched, whether it is
// mentioned in the path or not. If it is mentioned in the path then it will be
// searched in the specified order. If pg_catalog is not in the path then it
// will be searched before searching any of the path items."
// "Likewise, the current session's temporary-table schema, pg_temp_nnn, is
// always searched if it exists. It can be explicitly listed in the path by
// using the alias pg_temp. If it is not listed in the path then it is searched
// first (even before pg_catalog)."
// - https://www.postgresql.org/docs/9.1/static/runtime-config-client.html
func (s SearchPath) Iter() SearchPathIter {
	implicitPgTempSchema := !s.containsPgTempSchema && s.tempSchemaName != ""
	sp := SearchPathIter{
		paths:                s.paths,
		implicitPgCatalog:    !s.containsPgCatalog,
		implicitPgExtension:  !s.containsPgExtension,
		implicitPgTempSchema: implicitPgTempSchema,
		tempSchemaName:       s.tempSchemaName,
		userSchemaName:       s.userSchemaName,
	}
	return sp
}

// IterWithoutImplicitPGSchemas is the same as Iter, but does not include the
// implicit pg_temp and pg_catalog.
func (s SearchPath) IterWithoutImplicitPGSchemas() SearchPathIter {
	sp := SearchPathIter{
		paths:                s.paths,
		implicitPgCatalog:    false,
		implicitPgTempSchema: false,
		tempSchemaName:       s.tempSchemaName,
		userSchemaName:       s.userSchemaName,
	}
	return sp
}

// GetPathArray returns the underlying path array of this SearchPath. The
// resultant slice is not to be modified.
func (s SearchPath) GetPathArray() []string {
	return s.paths
}

// Contains returns true iff the SearchPath contains the given string.
func (s SearchPath) Contains(target string) bool {
	for _, candidate := range s.GetPathArray() {
		if candidate == target {
			return true
		}
	}
	return false
}

// GetTemporarySchemaName returns the temporary schema specific to the current
// session, or an empty string if the current session has not yet created a
// temporary schema.
//
// Note that even after the current session has created a temporary schema, a
// schema with that name may not exist in the session's current database.
func (s SearchPath) GetTemporarySchemaName() string {
	return s.tempSchemaName
}

// Equals returns true if two SearchPaths are the same.
func (s SearchPath) Equals(other *SearchPath) bool {
	if s.containsPgCatalog != other.containsPgCatalog {
		return false
	}
	if s.containsPgExtension != other.containsPgExtension {
		return false
	}
	if s.containsPgTempSchema != other.containsPgTempSchema {
		return false
	}
	if len(s.paths) != len(other.paths) {
		return false
	}
	if s.tempSchemaName != other.tempSchemaName {
		return false
	}
	// Fast path: skip the check if it is the same slice.
	if &s.paths[0] != &other.paths[0] {
		for i := range s.paths {
			if s.paths[i] != other.paths[i] {
				return false
			}
		}
	}
	return true
}

// SQLIdentifiers returns quotes for string starting with special characters.
func (s SearchPath) SQLIdentifiers() string {
	var buf bytes.Buffer
	for i, path := range s.paths {
		if i > 0 {
			buf.WriteString(", ")
		}
		lexbase.EncodeRestrictedSQLIdent(&buf, path, lexbase.EncNoFlags)
	}
	return buf.String()
}

func (s SearchPath) String() string {
	return strings.Join(s.paths, ",")
}

// SearchPathIter enables iteration over the search paths without triggering an
// allocation. Use one of the SearchPath.Iter methods to get an instance of the
// iterator, and then repeatedly call the Next method in order to iterate over
// each search path. The tempSchemaName in the iterator is only set if the session
// has created a temporary schema.
type SearchPathIter struct {
	paths                []string
	implicitPgCatalog    bool
	implicitPgExtension  bool
	implicitPgTempSchema bool
	tempSchemaName       string
	userSchemaName       string
	i                    int
}

// Next returns the next search path, or false if there are no remaining paths.
func (iter *SearchPathIter) Next() (path string, ok bool) {
	// If the session specific temporary schema has not been created, we can
	// preempt the name resolution failure by simply skipping the implicit pg_temp.
	if iter.implicitPgTempSchema && iter.tempSchemaName != "" {
		iter.implicitPgTempSchema = false
		return iter.tempSchemaName, true
	}

	if iter.i < len(iter.paths) {
		iter.i++
		// If pg_temp is explicitly present in the paths, it must be resolved to the
		// session specific temp schema (if one exists). tempSchemaName is set in the
		// iterator iff the session has created a temporary schema.
		// pg_extension should be read before delving into the schema.
		return iter.paths[iter.i-1], true
	}
	return "", false
}
