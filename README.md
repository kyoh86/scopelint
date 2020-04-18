# scopelint

[![Go Report Card](https://goreportcard.com/badge/github.com/kyoh86/scopelint)](https://goreportcard.com/report/github.com/kyoh86/scopelint)
[![CircleCI](https://img.shields.io/circleci/project/github/kyoh86/scopelint.svg)](https://circleci.com/gh/kyoh86/scopelint)
[![Coverage Status](https://img.shields.io/codecov/c/github/kyoh86/scopelint.svg)](https://codecov.io/gh/kyoh86/scopelint)

**scopelint** checks for unpinned variables in go programs.

## OBSOLETED

Use [looppointer](https://github.com/kyoh86/looppointer) or [exportloopref](https://github.com/kyoh86/exportloopref) instead.

I you want to find lints as nervous as possible (with some false-positives),
you should use [looppointer](https://github.com/kyoh86/looppointer).

If you want to find lints as accurately as possible (with some lints ignored),
you should use [exportloopref](https://github.com/kyoh86/exportloopref).

## What's this?

Sample problem code from: https://github.com/kyoh86/scopelint/blob/master/example/readme.go

```
6  values := []string{"a", "b", "c"}
7  var funcs []func()
8  for _, val := range values {
9  	funcs = append(funcs, func() {
10 		fmt.Println(val)
11 	})
12 }
13 for _, f := range funcs {
14 	f()
15 }
16 /*output:
17 c
18 c
19 c
20 */
21 var copies []*string
22 for _, val := range values {
23 	copies = append(copies, &val)
24 }
25 /*(in copies)
26 &"c"
27 &"c"
28 &"c"
29 */
```

In Go, the `val` variable in the above loops is actually a single variable.
So in many case (like the above), using it makes for us annoying bugs.

You can find them with `scopelint`, and fix it.

```
$ scopelint ./example/readme.go
example/readme.go:10:16: Using the variable on range scope "val" in function literal
example/readme.go:23:28: Using a reference for the variable on range scope "val"
Found 2 lint problems; failing.
```

(Fixed sample):

```go
values := []string{"a", "b", "c"}
var funcs []func()
for _, val := range values {
  val := val // pin!
	funcs = append(funcs, func() {
		fmt.Println(val)
	})
}
for _, f := range funcs {
	f()
}
var copies []*string
for _, val := range values {
  val := val // pin!
	copies = append(copies, &val)
}
```

## Install

go get -u github.com/kyoh86/scopelint

## Use

Give the package paths of interest as arguments:

```
scopelint github.com/kyoh86/scopelint/example
```

To check all packages recursively in the current directory:

```
scopelint ./...
```

And also, scopelint supports the following options:

* The `--set-exit-status` flag makes it to set exit status to 1 if any problem variables are found (if you DO NOT it, set --no-set-exit-status)
* The `--vendor` flag enables checking in the `vendor` directories (if you DO NOT it, set `--no-vendor` flag)
* The `--test` flag enables checking in the `*_test.go` files" (if you DO NOT it, set `--no-test` flag)

### Nolint

To ignore issues from use an option comment like //scopelint:ignore.
For example, if it's used inline (not from the beginning of the line), it ignores issues only for the line.

```go
var copies []*string
for _, val := range values {
	copies = append(copies, &val) //scopelint:ignore
}
```

To ignore issues for the block of code put it on the beginning of a line before the block:

```go
var copies []*string
//scopelint:ignore
for _, val := range values {
	copies = append(copies, &val)
}
```

Also, you can exclude all issues in a file by:

```go
//scopelint:ignore
package pkg
```

You may add a comment explaining or justifying why //scopelint:ignore is being used on the same line as the flag itself:

```go
//scopelint:ignore // any comment
```

### Use with gometalinter

scopelint can be used with [gometalinter](https://github.com/alecthomas/gometalinter) in `--linter` flag.

`gometalinter --disable-all --linter 'scope:scopelint {path}:^(?P<path>.*?\.go):(?P<line>\d+):(?P<col>\d+):\s*(?P<message>.*)$'`

## Exit Codes

scopelint returns 1 if any problems were found in the checked files.
It returns 2 if there were any other failures.

## Known Issues

- False positive for for table tests [#4](https://github.com/kyoh86/scopelint/issues/4)
    - I won't fix it because scopelint supports ignore them all. [=> nolint](#Nolint)

## TODO

- Write tests
- License (Some codes copied from [golint](https://github.com/golang/lint))
