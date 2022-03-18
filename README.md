# sqlfmt

Experimental SQL formatter with width-aware output

Based on http://homepages.inf.ed.ac.uk/wadler/papers/prettier/prettier.pdf.

Requires having the master branch of https://github.com/cockroachdb/cockroach checked out at `$GOPATH/src/github.com/cockroachdb/cockroach`, and running `make` in that directory.

> make project as module

# Used
```go
go get github.com/labulakalia/sqlfmt
```