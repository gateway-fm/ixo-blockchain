package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ixofoundation/ixo-cosmos/x/ixo"
	"github.com/tendermint/tendermint/crypto"
)

func DidToAddr(did ixo.Did) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(did)))
}
