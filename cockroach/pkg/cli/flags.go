// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"net"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/base"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/cli/cliflags"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/roachpb"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/security"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/server"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/envutil"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/netutil/addr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// special global variables used by flag variable definitions below.
// These do not correspond directly to the configuration parameters
// used as input by the CLI commands (these are defined in context
// structs in context.go). Instead, they are used at the *end* of
// command-line parsing to override the defaults in the context
// structs.
//
// Corollaries:
// - it would be a programming error to access these variables directly
//   outside of this file (flags.go)
// - the underlying context parameters must receive defaults in
//   initCLIDefaults() even when they are otherwise overridden by the
//   flags logic, because some tests to not use the flag logic at all.
var serverListenPort, serverSocketDir string
var serverAdvertiseAddr, serverAdvertisePort string
var serverSQLAddr, serverSQLPort string
var serverSQLAdvertiseAddr, serverSQLAdvertisePort string
var serverHTTPAddr, serverHTTPPort string
var localityAdvertiseHosts localityList
var startBackground bool
var storeSpecs base.StoreSpecList

// initPreFlagsDefaults initializes the values of the global variables
// defined above.
func initPreFlagsDefaults() {
	serverListenPort = base.DefaultPort
	serverSocketDir = ""
	serverAdvertiseAddr = ""
	serverAdvertisePort = ""

	serverSQLAddr = ""
	serverSQLPort = ""
	serverSQLAdvertiseAddr = ""
	serverSQLAdvertisePort = ""

	serverHTTPAddr = ""
	serverHTTPPort = base.DefaultHTTPPort

	localityAdvertiseHosts = localityList{}

	startBackground = false

	storeSpecs = base.StoreSpecList{}
}

// AddPersistentPreRunE add 'fn' as a persistent pre-run function to 'cmd'.
// If the command has an existing pre-run function, it is saved and will be called
// at the beginning of 'fn'.
// This allows an arbitrary number of pre-run functions with ordering based
// on the order in which AddPersistentPreRunE is called (usually package init order).
func AddPersistentPreRunE(cmd *cobra.Command, fn func(*cobra.Command, []string) error) {
	// Save any existing hooks.
	wrapped := cmd.PersistentPreRunE

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Run the previous hook if it exists.
		if wrapped != nil {
			if err := wrapped(cmd, args); err != nil {
				return err
			}
		}

		// Now we can call the new function.
		return fn(cmd, args)
	}
}

// stringFlag creates a string flag and registers it with the FlagSet.
// The default value is taken from the variable pointed to by valPtr.
// See context.go to initialize defaults.
func stringFlag(f *pflag.FlagSet, valPtr *string, flagInfo cliflags.FlagInfo) {
	f.StringVarP(valPtr, flagInfo.Name, flagInfo.Shorthand, *valPtr, flagInfo.Usage())
	registerEnvVarDefault(f, flagInfo)
}

// intFlag creates an int flag and registers it with the FlagSet.
// The default value is taken from the variable pointed to by valPtr.
// See context.go to initialize defaults.
func intFlag(f *pflag.FlagSet, valPtr *int, flagInfo cliflags.FlagInfo) {
	f.IntVarP(valPtr, flagInfo.Name, flagInfo.Shorthand, *valPtr, flagInfo.Usage())
	registerEnvVarDefault(f, flagInfo)
}

// boolFlag creates a bool flag and registers it with the FlagSet.
// The default value is taken from the variable pointed to by valPtr.
// See context.go to initialize defaults.
func boolFlag(f *pflag.FlagSet, valPtr *bool, flagInfo cliflags.FlagInfo) {
	f.BoolVarP(valPtr, flagInfo.Name, flagInfo.Shorthand, *valPtr, flagInfo.Usage())
	registerEnvVarDefault(f, flagInfo)
}

// durationFlag creates a duration flag and registers it with the FlagSet.
// The default value is taken from the variable pointed to by valPtr.
// See context.go to initialize defaults.
func durationFlag(f *pflag.FlagSet, valPtr *time.Duration, flagInfo cliflags.FlagInfo) {
	f.DurationVarP(valPtr, flagInfo.Name, flagInfo.Shorthand, *valPtr, flagInfo.Usage())
	registerEnvVarDefault(f, flagInfo)
}

// varFlag creates a custom-variable flag and registers it with the FlagSet.
// The default value is taken from the value's current value.
// See context.go to initialize defaults.
func varFlag(f *pflag.FlagSet, value pflag.Value, flagInfo cliflags.FlagInfo) {
	f.VarP(value, flagInfo.Name, flagInfo.Shorthand, flagInfo.Usage())
	registerEnvVarDefault(f, flagInfo)
}

// stringSliceFlag creates a string slice flag and registers it with the FlagSet.
// The default value is taken from the value's current value.
// See context.go to initialize defaults.
func stringSliceFlag(f *pflag.FlagSet, valPtr *[]string, flagInfo cliflags.FlagInfo) {
	f.StringSliceVar(valPtr, flagInfo.Name, *valPtr, flagInfo.Usage())
	registerEnvVarDefault(f, flagInfo)
}

// aliasStrVar wraps a string configuration option and is meant
// to be used in addition to / next to another flag that targets the
// same option. It does not implement "default values" so that the
// main flag can perform the default logic.
type aliasStrVar struct{ p *string }

// String implements the pflag.Value interface.
func (a aliasStrVar) String() string { return "" }

// Set implements the pflag.Value interface.
func (a aliasStrVar) Set(v string) error {
	if v != "" {
		*a.p = v
	}
	return nil
}

// Type implements the pflag.Value interface.
func (a aliasStrVar) Type() string { return "string" }

// addrSetter wraps a address/port configuration option pair and
// enables setting them both with a single command-line flag.
type addrSetter struct {
	addr *string
	port *string
}

// String implements the pflag.Value interface.
func (a addrSetter) String() string {
	return net.JoinHostPort(*a.addr, *a.port)
}

// Type implements the pflag.Value interface.
func (a addrSetter) Type() string { return "<addr/host>[:<port>]" }

// Set implements the pflag.Value interface.
func (a addrSetter) Set(v string) error {
	addr, port, err := addr.SplitHostPort(v, *a.port)
	if err != nil {
		return err
	}
	*a.addr = addr
	*a.port = port
	return nil
}

// clusterNameSetter wraps the cluster name variable
// and verifies its format during configuration.
type clusterNameSetter struct {
	clusterName *string
}

// String implements the pflag.Value interface.
func (a clusterNameSetter) String() string { return *a.clusterName }

// Type implements the pflag.Value interface.
func (a clusterNameSetter) Type() string { return "<identifier>" }

// Set implements the pflag.Value interface.
func (a clusterNameSetter) Set(v string) error {
	if v == "" {
		return errors.New("cluster name cannot be empty")
	}
	if len(v) > maxClusterNameLength {
		return errors.Newf(`cluster name can contain at most %d characters`, maxClusterNameLength)
	}
	if !clusterNameRe.MatchString(v) {
		return errClusterNameInvalidFormat
	}
	*a.clusterName = v
	return nil
}

var errClusterNameInvalidFormat = errors.New(`cluster name must contain only letters, numbers or the "-" and "." characters`)

// clusterNameRe matches valid cluster names.
// For example, "a", "a123" and "a-b" are OK,
// but "0123", "a-" and "123a" are not OK.
var clusterNameRe = regexp.MustCompile(`^[a-zA-Z](?:[-a-zA-Z0-9]*[a-zA-Z0-9]|)$`)

const maxClusterNameLength = 256

type keyTypeFilter int8

const (
	showAll keyTypeFilter = iota
	showValues
	showIntents
	showTxns
)

// String implements the pflag.Value interface.
func (f *keyTypeFilter) String() string {
	switch *f {
	case showValues:
		return "values"
	case showIntents:
		return "intents"
	case showTxns:
		return "txns"
	}
	return "all"
}

// Type implements the pflag.Value interface.
func (f *keyTypeFilter) Type() string { return "<key type>" }

// Set implements the pflag.Value interface.
func (f *keyTypeFilter) Set(v string) error {
	switch v {
	case "values":
		*f = showValues
	case "intents":
		*f = showIntents
	case "txns":
		*f = showTxns
	default:
		return errors.Newf("invalid key filter type '%s'", v)
	}
	return nil
}

const backgroundEnvVar = "COCKROACH_BACKGROUND_RESTART"

// flagSetForCmd is a replacement for cmd.Flag() that properly merges
// persistent and local flags, until the upstream bug
// https://github.com/spf13/cobra/issues/961 has been fixed.
func flagSetForCmd(cmd *cobra.Command) *pflag.FlagSet {
	_ = cmd.LocalFlags() // force merge persistent+local flags
	return cmd.Flags()
}

type tenantIDWrapper struct {
	tenID *roachpb.TenantID
}

func (w *tenantIDWrapper) String() string {
	return w.tenID.String()
}
func (w *tenantIDWrapper) Set(s string) error {
	tenID, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, "invalid tenant ID")
	}
	if tenID == 0 {
		return errors.New("invalid tenant ID")
	}
	*w.tenID = roachpb.MakeTenantID(tenID)
	return nil
}

func (w *tenantIDWrapper) Type() string {
	return "number"
}

// processEnvVarDefaults injects the current value of flag-related
// environment variables into the initial value of the settings linked
// to the flags, during initialization and before the command line is
// actually parsed. For example, it will inject the value of
// $COCKROACH_URL into the urlParser object linked to the --url flag.
func processEnvVarDefaults(cmd *cobra.Command) error {
	fl := flagSetForCmd(cmd)

	var retErr error
	fl.VisitAll(func(f *pflag.Flag) {
		envv, ok := f.Annotations[envValueAnnotationKey]
		if !ok || len(envv) < 2 {
			// No env var associated. Nothing to do.
			return
		}
		varName, value := envv[0], envv[1]
		if err := fl.Set(f.Name, value); err != nil {
			retErr = errors.CombineErrors(retErr,
				errors.Wrapf(err, "setting --%s from %s", f.Name, varName))
		}
	})
	return retErr
}

const (
	// envValueAnnotationKey is the map key used in pflag.Flag instances
	// to associate flags with a possible default value set by an
	// env var.
	envValueAnnotationKey = "envvalue"
)

// registerEnvVarDefault registers a deferred initialization of a flag
// from an environment variable.
// The caller is responsible for ensuring that the flagInfo has been
// defined in the FlagSet already.
func registerEnvVarDefault(f *pflag.FlagSet, flagInfo cliflags.FlagInfo) {
	if flagInfo.EnvVar == "" {
		return
	}

	value, set := envutil.EnvString(flagInfo.EnvVar, 2)
	if !set {
		// Env var is not set. Nothing to do.
		return
	}

	if err := f.SetAnnotation(flagInfo.Name, envValueAnnotationKey, []string{flagInfo.EnvVar, value}); err != nil {
		// This should never happen: an error is only returned if the flag
		// name was not defined yet.
		panic(err)
	}
}

// extraServerFlagInit configures the server.Config based on the command-line flags.
// It is only called when the command being ran is one of the start commands.
func extraServerFlagInit(cmd *cobra.Command) error {
	if err := security.SetCertPrincipalMap(startCtx.serverCertPrincipalMap); err != nil {
		return err
	}
	serverCfg.User = security.NodeUserName()
	serverCfg.Insecure = startCtx.serverInsecure
	serverCfg.SSLCertsDir = startCtx.serverSSLCertsDir

	// Construct the main RPC listen address.
	serverCfg.Addr = net.JoinHostPort(startCtx.serverListenAddr, serverListenPort)

	fs := flagSetForCmd(cmd)

	// Helper for .Changed that is nil-aware as not all of the `cmd`s may have
	// all of the flags.
	changed := func(set *pflag.FlagSet, name string) bool {
		f := set.Lookup(name)
		return f != nil && f.Changed
	}

	// Construct the socket name, if requested. The flags may not be defined for
	// `cmd` so be cognizant of that.
	//
	// If --socket-dir is set, then we'll use that.
	// There are two cases:
	// 1. --socket-dir is set and is empty; in this case the user is telling us
	//    "disable the socket".
	// 2. is set and non-empty. Then it should be used as specified.
	if changed(fs, cliflags.SocketDir.Name) {
		if serverSocketDir == "" {
			serverCfg.SocketFile = ""
		} else {
			serverCfg.SocketFile = filepath.Join(serverSocketDir, ".s.PGSQL."+serverListenPort)
		}
	}

	// Fill in the defaults for --advertise-addr.
	if serverAdvertiseAddr == "" {
		serverAdvertiseAddr = startCtx.serverListenAddr
	}
	if serverAdvertisePort == "" {
		serverAdvertisePort = serverListenPort
	}
	serverCfg.AdvertiseAddr = net.JoinHostPort(serverAdvertiseAddr, serverAdvertisePort)

	// Fill in the defaults for --sql-addr.
	if serverSQLAddr == "" {
		serverSQLAddr = startCtx.serverListenAddr
	}
	if serverSQLPort == "" {
		serverSQLPort = serverListenPort
	}
	serverCfg.SQLAddr = net.JoinHostPort(serverSQLAddr, serverSQLPort)
	serverCfg.SplitListenSQL = fs.Lookup(cliflags.ListenSQLAddr.Name).Changed

	// Fill in the defaults for --advertise-sql-addr, if the flag exists on `cmd`.
	advSpecified := changed(fs, cliflags.AdvertiseAddr.Name) ||
		changed(fs, cliflags.AdvertiseHost.Name)
	if serverSQLAdvertiseAddr == "" {
		if advSpecified {
			serverSQLAdvertiseAddr = serverAdvertiseAddr
		} else {
			serverSQLAdvertiseAddr = serverSQLAddr
		}
	}
	if serverSQLAdvertisePort == "" {
		if advSpecified && !serverCfg.SplitListenSQL {
			serverSQLAdvertisePort = serverAdvertisePort
		} else {
			serverSQLAdvertisePort = serverSQLPort
		}
	}
	serverCfg.SQLAdvertiseAddr = net.JoinHostPort(serverSQLAdvertiseAddr, serverSQLAdvertisePort)

	// Fill in the defaults for --http-addr.
	if serverHTTPAddr == "" {
		serverHTTPAddr = startCtx.serverListenAddr
	}
	if startCtx.unencryptedLocalhostHTTP {
		// If --unencrypted-localhost-http was specified, we want to
		// override whatever was specified or derived from other flags for
		// the host part of --http-addr.
		//
		// Before we do so, we'll check whether the user explicitly
		// specified something contradictory, and tell them that's no
		// good.
		if (changed(fs, cliflags.ListenHTTPAddr.Name) ||
			changed(fs, cliflags.ListenHTTPAddrAlias.Name)) &&
			(serverHTTPAddr != "" && serverHTTPAddr != "localhost") {
			return errors.WithHintf(
				errors.Newf("--unencrypted-localhost-http is incompatible with --http-addr=%s:%s",
					serverHTTPAddr, serverHTTPPort),
				`When --unencrypted-localhost-http is specified, use --http-addr=:%s or omit --http-addr entirely.`, serverHTTPPort)
		}

		// Now do the override proper.
		serverHTTPAddr = "localhost"
		// We then also tell the server to disable TLS for the HTTP
		// listener.
		serverCfg.DisableTLSForHTTP = true
	}
	serverCfg.HTTPAddr = net.JoinHostPort(serverHTTPAddr, serverHTTPPort)

	// Fill the advertise port into the locality advertise addresses.
	for i, a := range localityAdvertiseHosts {
		host, port, err := addr.SplitHostPort(a.Address.AddressField, serverAdvertisePort)
		if err != nil {
			return err
		}
		localityAdvertiseHosts[i].Address.AddressField = net.JoinHostPort(host, port)
	}
	serverCfg.LocalityAddresses = localityAdvertiseHosts

	return nil
}

// Fill the store paths.
// We have different defaults for server and tenant pod, and we don't want incorrect
// default to show up in flag help. To achieve that we create empty spec in private
// flag copy of spec and then copy this value if it was populated.
// If it isn't populated, default from server config is used for server commands or
// alternative default is generated by PreRun multi-tenant hook.
func extraStoreFlagInit(cmd *cobra.Command) error {
	fs := flagSetForCmd(cmd)
	if fs.Changed(cliflags.Store.Name) {
		serverCfg.Stores = storeSpecs
	}
	return nil
}

func extraClientFlagInit() error {
	// A command can be either a 'cert' command or an actual client command.
	// TODO(knz): Clean this up to not use a global variable for the
	// principal map.
	principalMap := certCtx.certPrincipalMap
	if principalMap == nil {
		principalMap = cliCtx.certPrincipalMap
	}
	if err := security.SetCertPrincipalMap(principalMap); err != nil {
		return err
	}
	serverCfg.Addr = net.JoinHostPort(cliCtx.clientConnHost, cliCtx.clientConnPort)
	serverCfg.AdvertiseAddr = serverCfg.Addr
	serverCfg.SQLAddr = net.JoinHostPort(cliCtx.clientConnHost, cliCtx.clientConnPort)
	serverCfg.SQLAdvertiseAddr = serverCfg.SQLAddr
	if serverHTTPAddr == "" {
		serverHTTPAddr = startCtx.serverListenAddr
	}
	serverCfg.HTTPAddr = net.JoinHostPort(serverHTTPAddr, serverHTTPPort)

	// If CLI/SQL debug mode is requested, override the echo mode here,
	// so that the initial client/server handshake reveals the SQL being
	// sent.
	if sqlConnCtx.DebugMode {
		sqlConnCtx.Echo = true
	}
	return nil
}

func mtStartSQLFlagsInit(cmd *cobra.Command) error {
	// Override default store for mt to use a per tenant store directory.
	fs := flagSetForCmd(cmd)
	if !fs.Changed(cliflags.Store.Name) {
		// We assume that we only need to change top level store as temp dir configs are
		// initialized when start is executed and temp dirs inherit path from first store.
		tenantID := fs.Lookup(cliflags.TenantID.Name).Value.String()
		serverCfg.Stores.Specs[0].Path = server.DefaultSQLNodeStorePathPrefix + tenantID
	}
	return nil
}

// VarFlag is exported for use in package cliccl.
var VarFlag = varFlag
