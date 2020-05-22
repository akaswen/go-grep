go-grep
an exploration of go with a grep-like command line tool

pre-requisites
Installation of go

build instructions
run go build .

usage
simple add the go-grep binary to runpath and then run with go-grep "expr"

Possible flags:

-I - allows automatically ignoring .git folder and everything in .gitignore. Can also specify additional files to ignore with = e.g. go-grep something_to_grep -I=file/to/ignore
