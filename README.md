# scopelint

**scopelint** checks for unpinned variables in go programs.

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

The `--set-exit-status` flag makes it to set exit status to 1 if any problem variables are found (if you DO NOT it, set --no-set-exit-status)
The `--vendor` flag enables checking in the `vendor` directories (if you DO NOT it, set `--no-vendor` flag)
The `--test` flag enables checking in the `*_test.go` files" (if you DO NOT it, set `--no-test` flag)

## Exit Codes

scopelint returns 1 if any problems were found in the checked files.
It returns 2 if there were any other failures.

## TODO

- Write tests
- License (Some codes copied from [golint](https://github.com/golang/lint))
