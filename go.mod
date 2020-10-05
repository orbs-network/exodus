module github.com/orbs-network/exodus

go 1.13

require (
	github.com/lib/pq v1.5.2
	github.com/orbs-network/crypto-lib-go v1.5.0
	github.com/orbs-network/orbs-client-sdk-go v0.18.0
	github.com/orbs-network/orbs-contract-sdk v1.8.0
	github.com/orbs-network/orbs-network-go v1.3.13
	github.com/orbs-network/orbs-spec v0.0.0-20200715083427-d937ef1ec8ef
	github.com/orbs-network/scribe v0.2.3
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
)

replace github.com/orbs-network/orbs-network-go => ../orbs-network-go
