package faucet

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"math/big"
	"net"
	"net/http"
	"strings"
)

func toEthAddress(address string) (common.Address, error) {
	if !common.IsHexAddress(address) {
		return common.Address{}, errors.New("invalid address")
	}
	return common.HexToAddress(address), nil
}

// based on https://github.com/chainflag/eth-faucet/blob/main/internal/server/limiter.go
func getIPFromRequest(proxies int, r *http.Request) string {
	if proxies > 0 {
		forwarded := r.Header.Get("X-Forwarded-For")
		if forwarded != "" {
			ips := strings.Split(forwarded, ",")
			idx := len(ips) - proxies
			if idx < 0 {
				idx = 0
			}
			return strings.TrimSpace(ips[idx])
		}
	}

	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteIP = r.RemoteAddr
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

func ToEth(value *big.Int, decimals int) decimal.Decimal {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	num, _ := decimal.NewFromString(value.String())
	result := num.Div(mul)

	return result
}
