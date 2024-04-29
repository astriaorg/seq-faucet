package chain

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"math/big"

	primproto "buf.build/gen/go/astria/primitives/protocolbuffers/go/astria/primitive/v1"
	txproto "buf.build/gen/go/astria/protocol-apis/protocolbuffers/go/astria/protocol/transactions/v1alpha1"
	"github.com/astriaorg/go-sequencer-client/client"
	"github.com/cometbft/cometbft/libs/bytes"

	"github.com/ethereum/go-ethereum/common"
)

const (
	// DefaultSequencerNetworkId is the default network id for the sequencer.
	DefaultSequencerNetworkId = "astria-dusk-5"
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

	amount, err := convertToUint128(value)
	if err != nil {
		panic(err)
	}

	toBytes, err := hex.DecodeString(to)
	if err != nil {
		panic(err)
	}
	toAddr := &primproto.Address{
		Inner: toBytes,
	}

	unsignedTx := &txproto.UnsignedTransaction{
		Params: &txproto.TransactionParams{
			ChainId: DefaultSequencerNetworkId,
			Nonce:   nonce,
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
	bigInt := new(big.Int)

	// check if the number is negative or overflows Uint128
	if bigInt.Sign() < 0 {
		return nil, fmt.Errorf("negative number not allowed")
	} else if bigInt.BitLen() > 128 {
		return nil, fmt.Errorf("value overflows Uint128")
	}

	// split the big.Int into two uint64s
	// convert the big.Int to uint64, which will drop the higher 64 bits
	lo := bigInt.Uint64()
	// shift the big.Int to the right by 64 bits and convert to uint64
	hi := bigInt.Rsh(bigInt, 64).Uint64()
	uint128 := &primproto.Uint128{
		Lo: lo,
		Hi: hi,
	}

	return uint128, nil
}
