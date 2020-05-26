package config

import (
	"github.com/orbs-network/crypto-lib-go/crypto/encoding"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
)

type OrbsClientConfig struct {
	Endpoint     string
	VirtualChain uint32

	PublicKey  string
	PrivateKey string

	Contract string
}

func (c OrbsClientConfig) Account() (*orbs.OrbsAccount, error) {
	publicKey, err := encoding.DecodeHex(c.PublicKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := encoding.DecodeHex(c.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &orbs.OrbsAccount{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

func (c OrbsClientConfig) Client() *orbs.OrbsClient {
	return orbs.NewClient(c.Endpoint, c.VirtualChain, codec.NETWORK_TYPE_TEST_NET)
}
