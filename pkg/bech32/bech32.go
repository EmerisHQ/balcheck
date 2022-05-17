package bech32

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func HexDecode(addr string) (string, error) {
	_, data, err := bech32.DecodeAndConvert(addr)
	if err != nil {
		return "", err
	}

	h := hex.EncodeToString(data)
	return h, nil
}

func HexEncode(hrp, hexAddr string) (string, error) {
	data, err := hex.DecodeString(hexAddr)
	if err != nil {
		return "", err
	}

	res, err := bech32.ConvertAndEncode(hrp, data)
	if err != nil {
		return "", err
	}

	return res, nil
}
