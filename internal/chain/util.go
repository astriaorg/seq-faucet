package chain

import (
	"github.com/ethereum/go-ethereum/common"
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
