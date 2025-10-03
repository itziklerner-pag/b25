module github.com/yourusername/b25/tests/integration

go 1.21

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/lib/pq v1.10.9
	github.com/nats-io/nats.go v1.31.0
	github.com/stretchr/testify v1.8.4
	github.com/yourusername/b25/tests/testutil v0.0.0
)

replace github.com/yourusername/b25/tests/testutil => ../testutil

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/nats-io/nkeys v0.4.6 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
