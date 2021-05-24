module myplugin

go 1.16

replace github.com/tliron/prudence => "/Depot/Projects/Three Crickets/prudence-go"

require (
	github.com/dgraph-io/badger/v3 v3.2011.1
	github.com/tliron/kutil v0.1.30
	github.com/tliron/prudence v0.0.0-00010101000000-000000000000
)
