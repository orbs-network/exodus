module github.com/orbs-network/exodus

go 1.13

require (
	github.com/lib/pq v1.5.2
	github.com/orbs-network/orbs-client-sdk-go v0.15.0
	github.com/orbs-network/orbs-network-go v1.3.13
	github.com/orbs-network/orbs-spec v0.0.0-20200503073830-babdf6adc845
	github.com/orbs-network/scribe v0.2.3
)

replace github.com/orbs-network/orbs-network-go => ../orbs-network-go
