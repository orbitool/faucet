package faucet

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/patrickmn/go-cache"
	"log"
	"math/big"
	"net/http"
	"time"
)

type Faucet struct {
	Config       *Config
	address      common.Address
	privateKey   *ecdsa.PrivateKey
	signer       types.Signer
	eth          *ethclient.Client
	cache        *cache.Cache
	displayValue string
}

func New(cfg *Config) (*Faucet, error) {
	cfg = useDefaults(cfg)

	keyBytes, err := hexutil.Decode(cfg.PrivateKey)
	if err != nil {
		fmt.Println("Failed to decode private KEY")
		return nil, err
	}

	privateKey, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		fmt.Println("Invalid private KEY")
		return nil, err
	}

	client, err := ethclient.Dial(cfg.Provider)
	if err != nil {
		return nil, err
	}

	if cfg.ChainID == nil {
		cfg.ChainID, err = client.ChainID(context.Background())
		if err != nil {
			return nil, err
		}
	}

	return &Faucet{
		Config:       cfg,
		address:      crypto.PubkeyToAddress(privateKey.PublicKey),
		privateKey:   privateKey,
		signer:       types.NewEIP155Signer(cfg.ChainID),
		eth:          client,
		cache:        cache.New(cfg.Delay, cfg.CacheCleanupDuration),
		displayValue: ToEth(cfg.Amount, 18).String(),
	}, nil
}

func (f *Faucet) Address() string {
	return f.address.String()
}

func (f *Faucet) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// 1. Get address and ip
	ip := parseIp(r.RemoteAddr)
	address, err := toEthAddress(r.URL.Query().Get(f.Config.AddressQueryKey))
	if err != nil {
		http.Error(w, "invalid ethereum address", http.StatusBadRequest)
		return
	}

	// 2. Check if account can claim according to the rate limit
	if err := f.canClaim(address.String(), ip); err != nil {
		http.Error(w, err.Error(), http.StatusTooManyRequests)
		return
	}

	// 3. Send the ethereum to the recipient
	receipt, err := f.send(address)
	if err != nil {
		log.Printf("[ERR] Failed to claim -> %v", err)
		http.Error(w, "failed to claim", http.StatusInternalServerError)
		return
	}

	// 4. Mark as complete
	log.Printf("Faucet: Successful claim: [%s] %s -> %s, txhash: %s", ip, f.displayValue, address, receipt.Hash)
	f.hasClaimed(address.String(), ip)

	// 5. Return the receipt as the http response
	json.NewEncoder(w).Encode(receipt)
}

type Receipt struct {
	Hash      string   `json:"hash,omitempty"`
	Address   string   `json:"address,omitempty"`
	Value     *big.Int `json:"value,omitempty"`
	Timestamp int64    `json:"timestamp,omitempty"`
}

func (f *Faucet) send(to common.Address) (*Receipt, error) {
	ctx, _ := context.WithTimeout(context.Background(), f.Config.RPCTimeout)

	nonce, err := f.eth.PendingNonceAt(ctx, f.address)
	if err != nil {
		return nil, err
	}

	gasLimit := uint64(21000)
	gasPrice, err := f.eth.SuggestGasPrice(ctx)
	if err != nil {
		gasPrice = f.Config.FallbackGasPrice
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    f.Config.Amount,
		Gas:      gasLimit,
		GasPrice: gasPrice,
	})

	signedTx, err := types.SignTx(tx, f.signer, f.privateKey)
	if err != nil {
		return nil, err
	}

	if err := f.eth.SendTransaction(ctx, signedTx); err != nil {
		return nil, err
	}

	return &Receipt{Hash: signedTx.Hash().String(), Value: tx.Value(), Timestamp: time.Now().Unix(), Address: to.String()}, nil
}

func (f *Faucet) canClaim(address, ip string) error {
	if _, exists := f.cache.Get(address); exists {
		return errors.New("this address has already claimed recently")
	}

	if _, exists := f.cache.Get(ip); exists {
		return errors.New("this ip has already claimed recently")
	}

	return nil
}

func (f *Faucet) hasClaimed(address, ip string) {
	f.cache.Set(address, byte(0x1), f.Config.Delay)
	f.cache.Set(ip, byte(0x1), f.Config.Delay)
}
