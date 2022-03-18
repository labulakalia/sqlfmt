// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/testutils"
)

var expectedA = `# Code generated by prereqs. DO NOT EDIT!

bin/a: a a/a.c a/a.f a/a.go a/cgo.go a/ignore.go a/invalid.go b b/b.go b/vendor/foo.com/bar b/vendor/foo.com/bar/bar.go vendor/foo.com/foo vendor/foo.com/foo/foo.go

a:
a/a.c:
a/a.f:
a/a.go:
a/cgo.go:
a/ignore.go:
a/invalid.go:
b:
b/b.go:
b/vendor/foo.com/bar:
b/vendor/foo.com/bar/bar.go:
vendor/foo.com/foo:
vendor/foo.com/foo/foo.go:
`

var expectedB = `# Code generated by prereqs. DO NOT EDIT!

bin/b: b b/b.go b/vendor/foo.com/bar b/vendor/foo.com/bar/bar.go vendor/foo.com/foo vendor/foo.com/foo/foo.go

b:
b/b.go:
b/vendor/foo.com/bar:
b/vendor/foo.com/bar/bar.go:
vendor/foo.com/foo:
vendor/foo.com/foo/foo.go:
`

var expectedFoo = `# Code generated by prereqs. DO NOT EDIT!

bin/foo: vendor/foo.com/foo vendor/foo.com/foo/foo.go

vendor/foo.com/foo:
vendor/foo.com/foo/foo.go:
`

var expectedSpecialChars = `# Code generated by prereqs. DO NOT EDIT!

bin/specialchars: specialchars specialchars/a\[\]\*\?\~ $$%\#.go

specialchars:
specialchars/a\[\]\*\?\~ $$%\#.go:
`

var expectedTestNoDeps = `# Code generated by prereqs. DO NOT EDIT!

bin/test: 

`

func TestPrereqs(t *testing.T) {
	gopath, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}

	const pkg = "testdata/src/example.com"
	absPkg, err := filepath.Abs(pkg)
	if err != nil {
		t.Fatal(err)
	}
	defer func(dir string, err error) {
		_ = os.Chdir(dir)
	}(os.Getwd())
	if err := os.Chdir(pkg); err != nil {
		t.Fatal(err)
	}

	testutils.RunTrueAndFalse(t, "symlink", func(t *testing.T, symlink bool) {
		if symlink {
			tempDir, cleanup := testutils.TempDir(t)
			defer cleanup()

			link := filepath.Join(tempDir, "link")
			if err := os.Symlink(absPkg, link); err != nil {
				t.Fatal(err)
			}

			// You can't chdir into a symlink. Instead, you chdir into the physical
			// path and set the PWD to the logical path. (At least, this is the hack
			// that most shells employ, and os.Getwd respects it.)
			if err := os.Setenv("PWD", link); err != nil {
				t.Fatal(err)
			}
			if cwd, err := os.Getwd(); err != nil {
				t.Fatal(err)
			} else if cwd != link {
				t.Fatalf("failed to chdir into symlink %s (os.Getwd reported %s)", link, cwd)
			}
		}

		for _, tc := range []struct {
			path        string
			exp         string
			includeTest bool
		}{
			{path: "example.com/a", exp: expectedA},
			{path: "./b", exp: expectedB},
			{path: "./a/../b", exp: expectedB},
			{path: "example.com/a/../b", exp: expectedB},
			{path: "example.com/b", exp: expectedB},
			{path: "./vendor/foo.com/foo", exp: expectedFoo},
			{path: "example.com/vendor/foo.com/foo", exp: expectedFoo},
			{path: "example.com/test", exp: expectedTestNoDeps},
		} {
			t.Run(tc.path, func(t *testing.T) {
				var buf bytes.Buffer
				options := testOptions{
					gopath: gopath,
				}
				if err := run(&buf, tc.path, tc.includeTest, "", options); err != nil {
					t.Fatal(err)
				}
				if e, a := tc.exp, buf.String(); e != a {
					t.Fatalf("expected:\n%s\nactual:\n%s\n", e, a)
				}
			})
		}
	})

	// The test for special character handling is special because creating a
	// filename with these special characters is problematic in some
	// contexts. Notably, we can't create such files on Windows or add them to
	// zips, which in turn prohibits the Cockroach source code from being used in
	// those contexts. To workaround this limitation, we create a filesystem
	// overlay and pass that to the package loader.
	t.Run("example.com/specialchars", func(t *testing.T) {
		srcDir := filepath.Join(gopath, "src")
		pkg := "example.com/specialchars"
		pkgPath := filepath.Join(srcDir, pkg)
		fakeContents := []byte("package specialchars\n")

		overlay := make(map[string][]byte)

		// Add a file with special characters in its name.
		overlay[filepath.Join(pkgPath, "a[]*?~ $%#.go")] = fakeContents
		var buf bytes.Buffer
		options := testOptions{
			gopath:    gopath,
			fsOverlay: overlay,
		}
		if err := run(&buf, pkg, false, "", options); err != nil {
			t.Fatal(err)
		}

		exp := expectedSpecialChars
		if e, a := exp, buf.String(); e != a {
			t.Fatalf("expected:\n%s\nactual:\n%s\n", e, a)
		}
	})
}
