// Package scopelint privides a linter for scopes of variable in `for {}`.
package scopelint

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"sort"
	"strings"
)

// A Linter lints Go source code.
type Linter struct{}

// Lint lints src.
func (l *Linter) Lint(filename string, src []byte) ([]Problem, error) {
	return l.LintFiles(map[string][]byte{filename: src})
}

// LintFiles lints a set of files of a single package.
// The argument is a map of filename to source.
func (l *Linter) LintFiles(files map[string][]byte) ([]Problem, error) {
	if len(files) == 0 {
		return nil, nil
	}

	pkg := &Package{
		FileSet: token.NewFileSet(),
		Files:   make(map[string]*File),
	}

	var pkgName string
	for filename, src := range files {
		astFile, err := parser.ParseFile(pkg.FileSet, filename, src, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		if pkgName == "" {
			pkgName = astFile.Name.Name
		} else if strings.TrimSuffix(astFile.Name.Name, "_test") != strings.TrimSuffix(pkgName, "_test") {
			return nil, fmt.Errorf("%s is in package %s, not %s", filename, astFile.Name.Name, pkgName)
		}
		pkg.Files[filename] = &File{
			Package:    pkg,
			ASTFile:    astFile,
			FileSet:    pkg.FileSet,
			Source:     src,
			Filename:   filename,
			CommentMap: ast.NewCommentMap(pkg.FileSet, astFile, astFile.Comments),
		}
	}
	return pkg.lint(), nil
}

// Package represents a package being linted.
type Package struct {
	FileSet *token.FileSet
	Files   map[string]*File

	TypesPackage *types.Package
	TypesInfo    *types.Info

	Problems []Problem
}

func (p *Package) lint() []Problem {
	for _, f := range p.Files {
		f.lint()
	}

	sort.Sort(problemsByPosition(p.Problems))

	return p.Problems
}

// File represents a File being linted.
type File struct {
	Package    *Package
	ASTFile    *ast.File
	FileSet    *token.FileSet
	Source     []byte
	Filename   string
	CommentMap ast.CommentMap
}

func (f *File) lint() {
	ast.Walk(&Node{
		File:          *f,
		DangerObjects: map[*ast.Object]int{},
		UnsafeObjects: map[*ast.Object]int{},
		SkipFuncs:     map[*ast.FuncLit]int{},
	}, f.ASTFile)
}

// Node represents a Node being linted.
type Node struct {
	File
	DangerObjects map[*ast.Object]int
	UnsafeObjects map[*ast.Object]int
	SkipFuncs     map[*ast.FuncLit]int
	Ignore        bool
}

// Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
func (n *Node) Visit(node ast.Node) ast.Visitor {
	next := *n
	if node == nil {
		return &next
	}
CGS_LOOP:
	for _, cg := range n.File.CommentMap[node] {
		for _, com := range cg.List {
			if hasOptionComment(com.Text, "ignore") {
				next.Ignore = true
				break CGS_LOOP
			}
		}
	}
	switch typedNode := node.(type) {
	case *ast.ForStmt:
		switch init := typedNode.Init.(type) {
		case *ast.AssignStmt:
			for _, lh := range init.Lhs {
				switch tlh := lh.(type) {
				case *ast.Ident:
					n.UnsafeObjects[tlh.Obj] = 0
				}
			}
		}

	case *ast.RangeStmt:
		// Memory variables declarated in range statement
		switch k := typedNode.Key.(type) {
		case *ast.Ident:
			n.UnsafeObjects[k.Obj] = 0
		}
		switch v := typedNode.Value.(type) {
		case *ast.Ident:
			n.UnsafeObjects[v.Obj] = 0
		}

	case *ast.UnaryExpr:
		if typedNode.Op == token.AND {
			switch ident := typedNode.X.(type) {
			case *ast.Ident:
				if _, unsafe := n.UnsafeObjects[ident.Obj]; unsafe {
					ref := ""
					n.errorf(ident, 1, n.Ignore, link(ref), category("range-scope"), "Using a reference for the variable on range scope %q", ident.Name)
				}
			}
		}

	case *ast.Ident:
		if _, obj := n.DangerObjects[typedNode.Obj]; obj {
			// It is the naked variable in scope of range statement.
			ref := ""
			n.errorf(node, 1, n.Ignore, link(ref), category("range-scope"), "Using the variable on range scope %q in function literal", typedNode.Name)
			break
		}

	case *ast.CallExpr:
		// Ignore func literals that'll be called immediately.
		switch funcLit := typedNode.Fun.(type) {
		case *ast.FuncLit:
			n.SkipFuncs[funcLit] = 0
		}

	case *ast.FuncLit:
		if _, skip := n.SkipFuncs[typedNode]; !skip {
			dangers := map[*ast.Object]int{}
			for d := range n.DangerObjects {
				dangers[d] = 0
			}
			for u := range n.UnsafeObjects {
				dangers[u] = 0
				n.UnsafeObjects[u]++
			}
			next.DangerObjects = dangers
			return &next
		}

	case *ast.ReturnStmt:
		unsafe := map[*ast.Object]int{}
		for u := range n.UnsafeObjects {
			if n.UnsafeObjects[u] == 0 {
				continue
			}
			unsafe[u] = n.UnsafeObjects[u]
		}
		next.UnsafeObjects = unsafe
		return &next
	}
	return &next
}

type link string
type category string

// The variadic arguments may start with link and category types,
// and must end with a format string and any arguments.
// It returns the new Problem.
func (f *File) errorf(n ast.Node, confidence float64, ignore bool, args ...interface{}) *Problem {
	pos := f.FileSet.Position(n.Pos())
	if pos.Filename == "" {
		pos.Filename = f.Filename
	}
	return f.Package.errorfAt(pos, confidence, ignore, args...)
}

func (p *Package) errorfAt(pos token.Position, confidence float64, ignore bool, args ...interface{}) *Problem {
	problem := Problem{
		Position:   pos,
		Confidence: confidence,
		Ignored:    ignore,
	}
	if pos.Filename != "" {
		// The file might not exist in our mapping if a //line directive was encountered.
		if f, ok := p.Files[pos.Filename]; ok {
			problem.LineText = srcLine(f.Source, pos)
		}
	}

argLoop:
	for len(args) > 1 { // always leave at least the format string in args
		switch v := args[0].(type) {
		case link:
			problem.Link = string(v)
		case category:
			problem.Category = string(v)
		default:
			break argLoop
		}
		args = args[1:]
	}

	problem.Text = fmt.Sprintf(args[0].(string), args[1:]...)

	p.Problems = append(p.Problems, problem)
	return &p.Problems[len(p.Problems)-1]
}

// srcLine returns the complete line at p, including the terminating newline.
func srcLine(src []byte, p token.Position) string {
	// Run to end of line in both directions if not at line start/end.
	lo, hi := p.Offset, p.Offset+1
	for lo > 0 && src[lo-1] != '\n' {
		lo--
	}
	for hi < len(src) && src[hi-1] != '\n' {
		hi++
	}
	return string(src[lo:hi])
}
