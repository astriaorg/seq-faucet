package chain

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"math/big"

	primproto "buf.build/gen/go/astria/primitives/protocolbuffers/go/astria/primitive/v1"
	txproto "buf.build/gen/go/astria/protocol-apis/protocolbuffers/go/astria/protocol/transactions/v1alpha1"
	client "github.com/astriaorg/go-sequencer-client/client"
	"github.com/cometbft/cometbft/libs/bytes"

	"github.com/ethereum/go-ethereum/common"
)

type TxBuilder interface {
	Sender() common.Address
	Transfer(ctx context.Context, to string, value *big.Int) (bytes.HexBytes, error)
}

type TxBuild struct {
	sequencerClient  client.Client
	privateKey       *ed25519.PrivateKey
	signer           client.Signer
	fromAddress      common.Address
	sequencerChainId string
}

func NewTxBuilder(provider string, privateKey *ed25519.PrivateKey, chainId string) (TxBuilder, error) {
	sequencerClient, err := client.NewClient(provider)
	if err != nil {
		return nil, err
	}

	signer := client.NewSigner(*privateKey)

	return &TxBuild{
		sequencerClient:  *sequencerClient,
		privateKey:       privateKey,
		signer:           *signer,
		fromAddress:      signer.Address(),
		sequencerChainId: chainId,
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

	amount, err := convertToUint128(value)
	if err != nil {
		panic(err)
	}

	toAddress := common.HexToAddress(to)
	toAddr := &primproto.Address{
		Inner: toAddress.Bytes(),
	}

	unsignedTx := &txproto.UnsignedTransaction{
		Params: &txproto.TransactionParams{
			Nonce:   nonce,
			ChainId: b.sequencerChainId,
		},
		Actions: []*txproto.Action{
			{
				Value: &txproto.Action_TransferAction{
					TransferAction: &txproto.TransferAction{
						To:         toAddr,
						Amount:     amount,
						AssetId:    client.DefaultAstriaAssetID[:],
						FeeAssetId: client.DefaultAstriaAssetID[:],
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

// convertToUint128 converts a string to a Uint128 protobuf
func convertToUint128(numStr *big.Int) (*primproto.Uint128, error) {

	// check if the number is negative or overflows Uint128
	if numStr.Sign() < 0 {
		return nil, fmt.Errorf("negative number not allowed")
	} else if numStr.BitLen() > 128 {
		return nil, fmt.Errorf("value overflows Uint128")
	}

	// split the big.Int into two uint64s
	// convert the big.Int to uint64, which will drop the higher 64 bits
	lo := numStr.Uint64()
	// shift the big.Int to the right by 64 bits and convert to uint64
	hi := numStr.Rsh(numStr, 64).Uint64()
	uint128 := &primproto.Uint128{
		Lo: lo,
		Hi: hi,
	}

	return uint128, nil
}
