# scopelint

[![Go Report Card](https://goreportcard.com/badge/github.com/kyoh86/scopelint)](https://goreportcard.com/report/github.com/kyoh86/scopelint)
[![CircleCI](https://img.shields.io/circleci/project/github/kyoh86/scopelint.svg)](https://circleci.com/gh/kyoh86/scopelint)
[![Coverage Status](https://img.shields.io/codecov/c/github/kyoh86/scopelint.svg)](https://codecov.io/gh/kyoh86/scopelint)

**scopelint** checks for unpinned variables in go programs.

## What's this?

```go
values := []string{"a", "b", "c"}
var funcs []func()
for _, val := range values {
	funcs = append(funcs, func() {
		fmt.Println(val)
	})
}
for _, f := range funcs {
	f()
}
/*output:
c
c
c
*/
```

```golang
var copies []*string
for _, val := range values {
	copies = append(copies, &val)
}
/*(in copies)
&"c"
&"c"
&"c"
*/
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


### Use with gometalinter

scopelint can be used with [gometalinter](https://github.com/alecthomas/gometalinter) in `--linter` flag.

`gometalinter --disable-all --linter 'scope:scopelint {path}:^(?P<path>.*?\.go):(?P<line>\d+):(?P<col>\d+):\s*(?P<message>.*)$'`

## Exit Codes

scopelint returns 1 if any problems were found in the checked files.
It returns 2 if there were any other failures.

## TODO

- Write tests
- License (Some codes copied from [golint](https://github.com/golang/lint))
