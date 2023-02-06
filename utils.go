package faucet

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"math/big"
	"net"
)

func toEthAddress(address string) (common.Address, error) {
	if !common.IsHexAddress(address) {
		return common.Address{}, errors.New("invalid address")
	}
	return common.HexToAddress(address), nil
}

func parseIp(ip string) string {
	remoteIP, _, err := net.SplitHostPort(ip)
	if err != nil {
		remoteIP = ip
	}
	return remoteIP
}

func ToWei(v float64, decimals int) *big.Int {
	amount := decimal.NewFromFloat(v)
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}

func toEth(value *big.Int, decimals int) decimal.Decimal {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	num, _ := decimal.NewFromString(value.String())
	result := num.Div(mul)

	return result
}
