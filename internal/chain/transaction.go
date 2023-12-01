package chain

import (
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"math/big"

	primproto "buf.build/gen/go/astria/astria/protocolbuffers/go/astria/primitive/v1"
	sqproto "buf.build/gen/go/astria/astria/protocolbuffers/go/astria/sequencer/v1alpha1"
	client "github.com/astriaorg/go-sequencer-client/client"
	"github.com/cometbft/cometbft/libs/bytes"

	"github.com/ethereum/go-ethereum/common"
)

type TxBuilder interface {
	Sender() common.Address
	Transfer(ctx context.Context, to string, value *big.Int) (bytes.HexBytes, error)
}

type TxBuild struct {
	sequencerClient client.Client
	privateKey      *ed25519.PrivateKey
	signer          client.Signer
	fromAddress     common.Address
}

func NewTxBuilder(provider string, privateKey *ed25519.PrivateKey) (TxBuilder, error) {
	sequencerClient, err := client.NewClient(provider)
	if err != nil {
		return nil, err
	}

	signer := client.NewSigner(*privateKey)

	return &TxBuild{
		sequencerClient: *sequencerClient,
		privateKey:      privateKey,
		signer:          *signer,
		fromAddress:     signer.Address(),
	}, nil
}

func (b *TxBuild) Sender() common.Address {
	return b.fromAddress
}

func (b *TxBuild) Transfer(ctx context.Context, to string, value *big.Int) (bytes.HexBytes, error) {
	nonce, err := b.sequencerClient.GetNonce(ctx, b.fromAddress)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 16)
	value.FillBytes(buf)

	leastSignificant64 := binary.BigEndian.Uint64(buf[8:])
	mostSignificant64 := binary.BigEndian.Uint64(buf[:8])

	toAddress := common.HexToAddress(to)
	unsignedTx := &sqproto.UnsignedTransaction{
		Nonce: nonce,
		Actions: []*sqproto.Action{
			{
				Value: &sqproto.Action_TransferAction{
					TransferAction: &sqproto.TransferAction{
						To: toAddress.Bytes(),
						Amount: &primproto.Uint128{
							Lo: leastSignificant64,
							Hi: mostSignificant64,
						},
						AssetId: client.DefaultAstriaAssetID[:],
					},
				},
			},
		},
	}

	signedTx, err := b.signer.SignTransaction(unsignedTx)
	if err != nil {
		panic(err)
	}
	result, err := b.sequencerClient.BroadcastTxSync(ctx, signedTx)

	return result.Hash, err
}
