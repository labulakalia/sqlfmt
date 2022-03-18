// Copyright 2016 The Cockroach Authors.
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
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/build"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/cli/clierrorplus"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/cli/cliflags"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/cli/clisqlexec"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/settings"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/settings/cluster"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/startupmigrations"
	"github.com/cockroachdb/errors/oserror"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var manPath string

var genManCmd = &cobra.Command{
	Use:   "man",
	Short: "generate man pages for CockroachDB",
	Long: `This command generates man pages for CockroachDB.

By default, this places man pages into the "man/man1" directory under the
current directory. Use "--path=PATH" to override the output directory. For
example, to install man pages globally on many Unix-like systems,
use "--path=/usr/local/share/man/man1".
`,
	Args: cobra.NoArgs,
	RunE: clierrorplus.MaybeDecorateError(runGenManCmd),
}

func runGenManCmd(cmd *cobra.Command, args []string) error {
	info := build.GetInfo()
	header := &doc.GenManHeader{
		Section: "1",
		Manual:  "CockroachDB Manual",
		Source:  fmt.Sprintf("CockroachDB %s", info.Tag),
	}

	if !strings.HasSuffix(manPath, string(os.PathSeparator)) {
		manPath += string(os.PathSeparator)
	}

	if _, err := os.Stat(manPath); err != nil {
		if oserror.IsNotExist(err) {
			if err := os.MkdirAll(manPath, 0755); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if err := doc.GenManTree(cmd.Root(), header, manPath); err != nil {
		return err
	}

	// TODO(cdo): The man page generated by the cobra package doesn't include a list of commands, so
	// one has to notice the "See Also" section at the bottom of the page to know which commands
	// are supported. I'd like to make this better somehow.

	fmt.Println("Generated CockroachDB man pages in", manPath)
	return nil
}

var autoCompletePath string

var genAutocompleteCmd = &cobra.Command{
	Use:   "autocomplete [shell]",
	Short: "generate autocompletion script for CockroachDB",
	Long: `Generate autocompletion script for CockroachDB.

If no arguments are passed, or if 'bash' is passed, a bash completion file is
written to ./cockroach.bash. If 'fish' is passed, a fish completion file
is written to ./cockroach.fish. If 'zsh' is passed, a zsh completion file is written
to ./_cockroach. Use "--out=/path/to/file" to override the output file location.

Note that for the generated file to work on OS X with bash, you'll need to install
Homebrew's bash-completion package (or an equivalent) and follow the post-install
instructions.
`,
	Args:      cobra.OnlyValidArgs,
	ValidArgs: []string{"bash", "zsh", "fish"},
	RunE:      clierrorplus.MaybeDecorateError(runGenAutocompleteCmd),
}

func runGenAutocompleteCmd(cmd *cobra.Command, args []string) error {
	var shell string
	if len(args) > 0 {
		shell = args[0]
	} else {
		shell = "bash"
	}

	var err error
	switch shell {
	case "bash":
		if autoCompletePath == "" {
			autoCompletePath = "cockroach.bash"
		}
		err = cmd.Root().GenBashCompletionFile(autoCompletePath)
	case "fish":
		if autoCompletePath == "" {
			autoCompletePath = "cockroach.fish"
		}
		err = cmd.Root().GenFishCompletionFile(autoCompletePath, true /* include description */)
	case "zsh":
		if autoCompletePath == "" {
			autoCompletePath = "_cockroach"
		}
		err = cmd.Root().GenZshCompletionFile(autoCompletePath)
	}
	if err != nil {
		return err
	}

	fmt.Printf("Generated %s completion file: %s\n", shell, autoCompletePath)
	return nil
}

var aesSize int
var overwriteKey bool

var genEncryptionKeyCmd = &cobra.Command{
	Use:   "encryption-key <key-file>",
	Short: "generate store key for encryption at rest",
	Long: `Generate store key for encryption at rest.

Generates a key suitable for use as a store key for Encryption At Rest.
The resulting key file will be 32 bytes (random key ID) + key_size in bytes.
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		encryptionKeyPath := args[0]

		// Check encryptionKeySize is suitable for the encryption algorithm.
		if aesSize != 128 && aesSize != 192 && aesSize != 256 {
			return fmt.Errorf("store key size should be 128, 192, or 256 bits, got %d", aesSize)
		}

		// 32 bytes are reserved for key ID.
		kSize := aesSize/8 + 32
		b := make([]byte, kSize)
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("failed to create key with size %d bytes", kSize)
		}

		// Write key to the file with owner read/write permission.
		openMode := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		if !overwriteKey {
			openMode |= os.O_EXCL
		}

		f, err := os.OpenFile(encryptionKeyPath, openMode, 0600)
		if err != nil {
			return err
		}
		n, err := f.Write(b)
		if err == nil && n < len(b) {
			err = io.ErrShortWrite
		}
		if err1 := f.Close(); err == nil {
			err = err1
		}

		if err != nil {
			return err
		}

		fmt.Printf("successfully created AES-%d key: %s\n", aesSize, encryptionKeyPath)
		return nil
	},
}

var includeReservedSettings bool
var excludeSystemSettings bool

var genSettingsListCmd = &cobra.Command{
	Use:   "settings-list",
	Short: "output a list of available cluster settings",
	Long: `
Output the list of cluster settings known to this binary.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		wrapCode := func(s string) string {
			if sqlExecCtx.TableDisplayFormat == clisqlexec.TableDisplayRawHTML {
				return fmt.Sprintf("<code>%s</code>", s)
			}
			return s
		}

		// Fill a Values struct with the defaults.
		s := cluster.MakeTestingClusterSettings()
		settings.NewUpdater(&s.SV).ResetRemaining(context.Background())

		var rows [][]string
		for _, name := range settings.Keys(settings.ForSystemTenant) {
			setting, ok := settings.Lookup(name, settings.LookupForLocalAccess, settings.ForSystemTenant)
			if !ok {
				panic(fmt.Sprintf("could not find setting %q", name))
			}

			if excludeSystemSettings && setting.Class() == settings.SystemOnly {
				continue
			}

			if setting.Visibility() != settings.Public {
				// We don't document non-public settings at this time.
				continue
			}

			typ, ok := settings.ReadableTypes[setting.Typ()]
			if !ok {
				panic(fmt.Sprintf("unknown setting type %q", setting.Typ()))
			}
			var defaultVal string
			if sm, ok := setting.(*settings.VersionSetting); ok {
				defaultVal = sm.SettingsListDefault()
			} else {
				defaultVal = setting.String(&s.SV)
				if override, ok := startupmigrations.SettingsDefaultOverrides[name]; ok {
					defaultVal = override
				}
			}
			row := []string{wrapCode(name), typ, wrapCode(defaultVal), setting.Description()}
			rows = append(rows, row)
		}

		sliceIter := clisqlexec.NewRowSliceIter(rows, "dddd")
		cols := []string{"Setting", "Type", "Default", "Description"}
		return sqlExecCtx.PrintQueryOutput(os.Stdout, stderr, cols, sliceIter)
	},
}

var genCmd = &cobra.Command{
	Use:   "gen [command]",
	Short: "generate auxiliary files",
	Long:  "Generate manpages, example shell settings, example databases, etc.",
	RunE:  UsageAndErr,
}

var genCmds = []*cobra.Command{
	genManCmd,
	genAutocompleteCmd,
	genExamplesCmd,
	genHAProxyCmd,
	genSettingsListCmd,
	genEncryptionKeyCmd,
}

func init() {
	genManCmd.PersistentFlags().StringVar(&manPath, "path", "man/man1",
		"path where man pages will be outputted")
	genAutocompleteCmd.PersistentFlags().StringVar(&autoCompletePath, "out", "",
		"path to generated autocomplete file")
	genHAProxyCmd.PersistentFlags().StringVar(&haProxyPath, "out", "haproxy.cfg",
		"path to generated haproxy configuration file")
	varFlag(genHAProxyCmd.Flags(), &haProxyLocality, cliflags.Locality)
	genEncryptionKeyCmd.PersistentFlags().IntVarP(&aesSize, "size", "s", 128,
		"AES key size for encryption at rest (one of: 128, 192, 256)")
	genEncryptionKeyCmd.PersistentFlags().BoolVar(&overwriteKey, "overwrite", false,
		"Overwrite key if it exists")
	genSettingsListCmd.PersistentFlags().BoolVar(&includeReservedSettings, "include-reserved", false,
		"include undocumented 'reserved' settings")
	genSettingsListCmd.PersistentFlags().BoolVar(&excludeSystemSettings, "without-system-only", false,
		"do not list settings only applicable to system tenant")

	genCmd.AddCommand(genCmds...)
}
