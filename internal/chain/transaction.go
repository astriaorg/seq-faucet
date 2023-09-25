package chain

import (
	"context"
	"crypto/ed25519"
	"math/big"

	client "github.com/astriaorg/go-sequencer-client/client"
	sqproto "github.com/astriaorg/go-sequencer-client/proto"

	"github.com/ethereum/go-ethereum/common"
)

type TxBuilder interface {
	Sender() common.Address
	Transfer(ctx context.Context, to string, value *big.Int) (common.Hash, error)
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

func (b *TxBuild) Transfer(ctx context.Context, to string, value *big.Int) (common.Hash, error) {
	nonce, err := b.sequencerClient.GetNonce(ctx, b.fromAddress)
	if err != nil {
		panic(err)
	}

	toAddress := common.HexToAddress(to)
	unsignedTx := &sqproto.UnsignedTransaction{
		Nonce: nonce,
		Actions: []*sqproto.Action{
			{
				Value: &sqproto.Action_TransferAction{
					TransferAction: &sqproto.TransferAction{
						To: toAddress.Bytes(),
						Amount: &sqproto.Uint128{
							Lo: value.Uint64(),
							Hi: value.Rsh(value, 64).Uint64(),
						},
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

	// FIXME - is this correct conversion here?
	// FIXME - is this the correct hash?
	return common.HexToHash(string(result.Hash)), err
}
