// scopelint checks for unpinned variables in go programs.
package main

import (
	"errors"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/kyoh86/scopelint/scopelint"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var params struct {
	argCount      int
	arguments     arguments
	setExitStatus bool
	vendor        bool
	test          bool
}

var problems int

func main() {
	app := kingpin.New("scopelint", "Checks for unpinned variables in go programs")
	app.Author("kyoh86").Version("0.1.0")
	app.VersionFlag.Short('v')

	app.Flag("set-exit-status", "Set exit status to 1 if any problem variables are found").Default("true").BoolVar(&params.setExitStatus)
	app.Flag("vendor", "Search lints in the `vendor` directories").Default("true").BoolVar(&params.vendor)
	app.Flag("test", "Search lints in the `*_test.go` files").Default("true").BoolVar(&params.test)
	arg := app.Arg("packages", "Set target packages")
	arg.CounterVar(&params.argCount)
	arg.SetValue(&params.arguments)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	for _, dir := range params.arguments.directories {
		lintImportedPackage(build.ImportDir(dir, 0))
	}
	for _, pkgname := range importPaths(params.arguments.packages) {
		lintImportedPackage(build.Import(pkgname, ".", 0))
	}
	lintFiles(params.arguments.files...)

	if params.setExitStatus && problems > 0 {
		fmt.Fprintf(os.Stderr, "Found %d lint problems; failing.\n", problems)
		os.Exit(1)
	}
}

func lintFiles(filenames ...string) {
	files := make(map[string][]byte)
	for _, filename := range filenames {
		src, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		files[filename] = src
	}

	l := new(scopelint.Linter)
	ps, err := l.LintFiles(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	for _, p := range ps {
		fmt.Printf("%v: %s\n", p.Position, p.Text)
		problems++
	}
}

func lintImportedPackage(pkg *build.Package, err error) {
	if err != nil {
		if _, nogo := err.(*build.NoGoError); nogo {
			// Don't complain if the failure is due to no Go source files.
			return
		}
		fmt.Fprintln(os.Stderr, err)
		return
	}

	var files []string
	files = append(files, pkg.GoFiles...)
	files = append(files, pkg.CgoFiles...)
	if params.test {
		files = append(files, pkg.TestGoFiles...)
		files = append(files, pkg.XTestGoFiles...)
	}
	if pkg.Dir != "." {
		for i, f := range files {
			files[i] = filepath.Join(pkg.Dir, f)
		}
	}
	if !params.vendor {
		var target []string
		for _, f := range files {
			if strings.Contains(f, "/vendor/") {
				continue
			}
			target = append(target, f)
		}
		files = target
	}

	lintFiles(files...)
}

/*

This file holds a direct copy of the import path matching code of
https://github.com/golang/go/blob/master/src/cmd/go/main.go. It can be
replaced when https://golang.org/issue/8768 is resolved.

It has been updated to follow upstream changes in a few ways.

*/

var buildContext = build.Default

var (
	goroot    = filepath.Clean(runtime.GOROOT())
	gorootSrc = filepath.Join(goroot, "src")
)

// importPathsNoDotExpansion returns the import paths to use for the given
// command line, but it does no ... expansion.
func importPathsNoDotExpansion(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	var out []string
	for _, a := range args {
		// Arguments are supposed to be import paths, but
		// as a courtesy to Windows developers, rewrite \ to /
		// in command-line arguments.  Handles .\... and so on.
		if filepath.Separator == '\\' {
			a = strings.Replace(a, `\`, `/`, -1)
		}

		// Put argument in canonical form, but preserve leading ./.
		if strings.HasPrefix(a, "./") {
			a = "./" + path.Clean(a)
			if a == "./." {
				a = "."
			}
		} else {
			a = path.Clean(a)
		}
		if a == allPackage || a == standardPackages {
			out = append(out, allPackages(a)...)
			continue
		}
		out = append(out, a)
	}
	return out
}

// importPaths returns the import paths to use for the given command line.
func importPaths(args []string) []string {
	args = importPathsNoDotExpansion(args)
	var out []string
	for _, a := range args {
		if strings.Contains(a, "...") {
			if build.IsLocalImport(a) {
				out = append(out, allPackagesInFS(a)...)
			} else {
				out = append(out, allPackages(a)...)
			}
			continue
		}
		out = append(out, a)
	}
	return out
}

// matchPattern(pattern)(name) reports whether
// name matches pattern.  Pattern is a limited glob
// pattern in which '...' means 'any string' and there
// is no other special syntax.
func matchPattern(pattern string) func(name string) bool {
	re := regexp.QuoteMeta(pattern)
	re = strings.Replace(re, `\.\.\.`, `.*`, -1)
	// Special case: foo/... matches foo too.
	if strings.HasSuffix(re, `/.*`) {
		re = re[:len(re)-len(`/.*`)] + `(/.*)?`
	}
	reg := regexp.MustCompile(`^` + re + `$`)
	return func(name string) bool {
		return reg.MatchString(name)
	}
}

// hasPathPrefix reports whether the path s begins with the
// elements in prefix.
func hasPathPrefix(s, prefix string) bool {
	switch {
	default:
		return false
	case len(s) == len(prefix):
		return s == prefix
	case len(s) > len(prefix):
		if prefix != "" && prefix[len(prefix)-1] == '/' {
			return strings.HasPrefix(s, prefix)
		}
		return s[len(prefix)] == '/' && s[:len(prefix)] == prefix
	}
}

// treeCanMatchPattern(pattern)(name) reports whether
// name or children of name can possibly match pattern.
// Pattern is the same limited glob accepted by matchPattern.
func treeCanMatchPattern(pattern string) func(name string) bool {
	wildCard := false
	if i := strings.Index(pattern, "..."); i >= 0 {
		wildCard = true
		pattern = pattern[:i]
	}
	return func(name string) bool {
		return len(name) <= len(pattern) && hasPathPrefix(pattern, name) ||
			wildCard && strings.HasPrefix(name, pattern)
	}
}

// allPackages returns all the packages that can be found
// under the $GOPATH directories and $GOROOT matching pattern.
// The pattern is either "all" (all packages), "std" (standard packages)
// or a path including "...".
func allPackages(pattern string) []string {
	pkgs := matchPackages(pattern)
	if len(pkgs) == 0 {
		fmt.Fprintf(os.Stderr, "warning: %q matched no packages\n", pattern)
	}
	return pkgs
}

const (
	standardPackages = "std"
	commandPackages  = "cmd"
	allPackage       = "all"
)

func matchPackages(pattern string) []string {
	match := func(string) bool { return true }
	treeCanMatch := func(string) bool { return true }
	if pattern != allPackage && pattern != standardPackages {
		match = matchPattern(pattern)
		treeCanMatch = treeCanMatchPattern(pattern)
	}

	have := map[string]bool{
		"builtin": true, // ignore pseudo-package that exists only for documentation
	}
	if !buildContext.CgoEnabled {
		have["runtime/cgo"] = true // ignore during walk
	}
	var pkgs []string

	// Commands
	cmd := filepath.Join(goroot, "src/cmd") + string(filepath.Separator)
	filepath.Walk(cmd, func(path string, fi os.FileInfo, err error) error {
		if err != nil || !fi.IsDir() || path == cmd {
			return nil
		}
		name := path[len(cmd):]
		if !treeCanMatch(name) {
			return filepath.SkipDir
		}
		// Commands are all in cmd/, not in subdirectories.
		if strings.Contains(name, string(filepath.Separator)) {
			return filepath.SkipDir
		}

		// We use, e.g., cmd/gofmt as the pseudo import path for gofmt.
		name = "cmd/" + name
		if have[name] {
			return nil
		}
		have[name] = true
		if !match(name) {
			return nil
		}
		_, err = buildContext.ImportDir(path, 0)
		if err != nil {
			if _, noGo := err.(*build.NoGoError); !noGo {
				log.Print(err)
			}
			return nil
		}
		pkgs = append(pkgs, name)
		return nil
	})

	for _, src := range buildContext.SrcDirs() {
		if (pattern == standardPackages || pattern == commandPackages) && src != gorootSrc {
			continue
		}
		src = filepath.Clean(src) + string(filepath.Separator)
		root := src
		if pattern == commandPackages {
			root += commandPackages + string(filepath.Separator)
		}
		filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
			if err != nil || !fi.IsDir() || path == src {
				return nil
			}

			// Avoid .foo, _foo, and testdata directory trees.
			_, elem := filepath.Split(path)
			if strings.HasPrefix(elem, ".") || strings.HasPrefix(elem, "_") || elem == "testdata" {
				return filepath.SkipDir
			}

			name := filepath.ToSlash(path[len(src):])
			if pattern == standardPackages && (strings.Contains(name, ".") || name == commandPackages) {
				// The name "std" is only the standard library.
				// If the name is cmd, it's the root of the command tree.
				return filepath.SkipDir
			}
			if !treeCanMatch(name) {
				return filepath.SkipDir
			}
			if have[name] {
				return nil
			}
			have[name] = true
			if !match(name) {
				return nil
			}
			_, err = buildContext.ImportDir(path, 0)
			if err != nil {
				if _, noGo := err.(*build.NoGoError); noGo {
					return nil
				}
			}
			pkgs = append(pkgs, name)
			return nil
		})
	}
	return pkgs
}

// allPackagesInFS is like allPackages but is passed a pattern
// beginning ./ or ../, meaning it should scan the tree rooted
// at the given directory.  There are ... in the pattern too.
func allPackagesInFS(pattern string) []string {
	pkgs := matchPackagesInFS(pattern)
	if len(pkgs) == 0 {
		fmt.Fprintf(os.Stderr, "warning: %q matched no packages\n", pattern)
	}
	return pkgs
}

func matchPackagesInFS(pattern string) []string {
	// Find directory to begin the scan.
	// Could be smarter but this one optimization
	// is enough for now, since ... is usually at the
	// end of a path.
	i := strings.Index(pattern, "...")
	dir, _ := path.Split(pattern[:i])

	// pattern begins with ./ or ../.
	// path.Clean will discard the ./ but not the ../.
	// We need to preserve the ./ for pattern matching
	// and in the returned import paths.
	prefix := ""
	if strings.HasPrefix(pattern, "./") {
		prefix = "./"
	}
	match := matchPattern(pattern)

	var pkgs []string
	filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil || !fi.IsDir() {
			return nil
		}
		if path == dir {
			// filepath.Walk starts at dir and recurses. For the recursive case,
			// the path is the result of filepath.Join, which calls filepath.Clean.
			// The initial case is not Cleaned, though, so we do this explicitly.
			//
			// This converts a path like "./io/" to "io". Without this step, running
			// "cd $GOROOT/src/pkg; go list ./io/..." would incorrectly skip the io
			// package, because prepending the prefix "./" to the unclean path would
			// result in "././io", and match("././io") returns false.
			path = filepath.Clean(path)
		}

		// Avoid .foo, _foo, and testdata directory trees, but do not avoid "." or "..".
		_, elem := filepath.Split(path)
		dot := strings.HasPrefix(elem, ".") && elem != "." && elem != ".."
		if dot || strings.HasPrefix(elem, "_") || elem == "testdata" {
			return filepath.SkipDir
		}

		name := prefix + filepath.ToSlash(path)
		if !match(name) {
			return nil
		}
		if _, err = build.ImportDir(path, 0); err != nil {
			if _, noGo := err.(*build.NoGoError); !noGo {
				log.Print(err)
			}
			return nil
		}
		pkgs = append(pkgs, name)
		return nil
	})
	return pkgs
}

type arguments struct {
	packages    []string
	files       []string
	directories []string
}

func (a *arguments) String() string {
	return strings.Join(append(append(a.packages, a.files...), a.directories...), ",")
}

func isDirExists(filename string) bool {
	fi, err := os.Stat(filename)
	return err == nil && fi.IsDir()
}

func isFileExists(filename string) bool {
	fi, err := os.Stat(filename)
	return err == nil && !fi.IsDir()
}

func (a *arguments) Set(arg string) error {
	if strings.HasSuffix(arg, "/...") && isDirExists(arg[:len(arg)-len("/...")]) {
		return a.addDirectory(allPackagesInFS(arg)...)
	}
	if isDirExists(arg) {
		return a.addDirectory(arg)
	}
	if isFileExists(arg) {
		return a.addFile(arg)
	}
	return a.addPackage(arg)
}

func (a *arguments) addDirectory(directory ...string) error {
	if len(a.packages) > 0 || len(a.files) > 0 {
		return errors.New("cannot search lints in directories with packages or files")
	}
	for _, d := range directory {
		a.directories = append(a.directories, d)
	}
	return nil
}

func (a *arguments) addFile(file string) error {
	if len(a.packages) > 0 || len(a.directories) > 0 {
		return errors.New("cannot search lints in files with packages or directories")
	}
	a.files = append(a.files, file)
	return nil
}

func (a *arguments) addPackage(p string) error {
	if len(a.files) > 0 || len(a.directories) > 0 {
		return errors.New("cannot search lints in packages with files or directories")
	}
	a.packages = append(a.packages, p)
	return nil
}

func (a *arguments) Get() interface{} {
	switch {
	case a.packages != nil:
		return a.packages
	case a.files != nil:
		return a.files
	case a.directories != nil:
		return a.directories
	}
	return []string{"."}
}

func (a *arguments) IsCumulative() bool {
	return true
}
