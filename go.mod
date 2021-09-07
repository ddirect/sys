module github.com/ddirect/sys

go 1.16

replace github.com/ddirect/filetest => ../filetest

replace github.com/ddirect/check => ../check

replace github.com/ddirect/format => ../format

replace github.com/ddirect/xrand => ../xrand

require (
	github.com/ddirect/check v0.0.0-00010101000000-000000000000
	github.com/ddirect/filetest v0.0.0-00010101000000-000000000000
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
)
