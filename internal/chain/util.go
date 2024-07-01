package chain

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

type Bech32MAddress struct {
	Address string
	Prefix  string
	Bytes   [20]byte
}

func (a *Bech32MAddress) String() string {
	return a.Address
}

func Has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

func IsValidAddress(address string, checksummed bool) bool {
	if !common.IsHexAddress(address) {
		return false
	}
	return !checksummed || common.HexToAddress(address).Hex() == address
}

func IsBech32M(address string) bool {
	prefix, bytes, err := bech32.Decode(address)
	if err != nil {
		return false
	}
	encodedAddress, err := Bech32MFromBytes(prefix, [20]byte(bytes))
	if err != nil {
		return false
	}
	log.Infof("trying with address: %s, calculated: %s", address, encodedAddress.Address)
	return true
}
func Bech32MFromBytes(prefix string, data [20]byte) (*Bech32MAddress, error) {
	// Convert the data from 8-bit groups to 5-bit
	converted, err := bech32.ConvertBits(data[:], 8, 5, true)
	if err != nil {
		return nil, fmt.Errorf("failed to convert bits from 8-bit groups to 5-bit groups: %v", err)
	}

	// Encode the data as Bech32m
	address, err := bech32.EncodeM(prefix, converted)
	if err != nil {
		return nil, fmt.Errorf("failed to encode address as bech32m: %v", err)
	}

	return &Bech32MAddress{
		Address: address,
		Prefix:  prefix,
		Bytes:   data,
	}, nil
}
