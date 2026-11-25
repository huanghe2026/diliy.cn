// Package chain — Polygon RPC 交互（ethclient）
//
// 用法：
//
//	client := chain.New(chain.Env())
//	gasPrice, _ := client.GasPrice(ctx)
//	balance, _ := client.BalanceAt(ctx, addr, nil)
package chain

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Config — Polygon 连接配置
type Config struct {
	RPCURL       string // Polygon 节点 (Alchemy/Infura/公共)
	ChainID      int64  // Polygon mainnet=137, Amoy testnet=80002
	ContractAddr string // GeoyuanCard 合约地址
	PrivateKey   string // 签名用私钥 (hex 无 0x)
}

// 默认 Polygon Amoy testnet（开发期用 testnet 免费）
func DefaultConfig() Config {
	return Config{
		RPCURL:       envOr("GEOYUAN_RPC_URL", "https://rpc-amoy.polygon.technology"),
		ChainID:      80002, // Amoy
		ContractAddr: envOr("GEOYUAN_CONTRACT", ""),
		PrivateKey:   strings.TrimPrefix(envOr("GEOYUAN_PRIVATE_KEY", ""), "0x"),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Client — Polygon 客户端（每次调用独立派生超时 ctx）
type Client struct {
	cfg Config
	eth *ethclient.Client
}

// New — 创建客户端（连接失败返回 error，不 panic）
func New(cfg Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	eth, err := ethclient.DialContext(ctx, cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("polygon rpc dial %s: %w", cfg.RPCURL, err)
	}
	return &Client{cfg: cfg, eth: eth}, nil
}

// Eth — 暴露底层 ethclient，给 abigen 绑定用（实现 bind.ContractBackend）
func (c *Client) Eth() *ethclient.Client { return c.eth }

// Close — 关连接
func (c *Client) Close() { c.eth.Close() }

// IsUp — 探活
func (c *Client) IsUp() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := c.eth.ChainID(ctx)
	return err == nil
}

// GasPrice — 当前 gas price（带 1 次重试）
func (c *Client) GasPrice() (*big.Int, error) {
	return withRetry(func() (*big.Int, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return c.eth.SuggestGasPrice(ctx)
	})
}

// Balance — 地址 MATIC 余额 (wei → MATIC)
func (c *Client) Balance(addrHex string) (float64, error) {
	wei, err := c.BalanceWei(addrHex)
	if err != nil {
		return 0, err
	}
	matic := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
	f, _ := matic.Float64()
	return f, nil
}

// BalanceWei — 地址余额（wei 原值，避免浮点误差）
func (c *Client) BalanceWei(addrHex string) (*big.Int, error) {
	addr := common.HexToAddress(addrHex)
	return withRetry(func() (*big.Int, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return c.eth.BalanceAt(ctx, addr, nil)
	})
}

// Nonce — 下一个交易 nonce（pending）
func (c *Client) Nonce(addrHex string) (uint64, error) {
	addr := common.HexToAddress(addrHex)
	type res struct {
		n   uint64
		err error
	}
	r, err := withRetry(func() (res, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		n, err := c.eth.PendingNonceAt(ctx, addr)
		return res{n: n, err: err}, err
	})
	return r.n, err
}

// ContractAddr — 返回合约地址
func (c *Client) ContractAddr() common.Address {
	return common.HexToAddress(c.cfg.ContractAddr)
}

// RPCURL — 返回 RPC 地址（用于日志）
func (c *Client) RPCURL() string { return c.cfg.RPCURL }

// ChainID — 返回链 ID
func (c *Client) ChainID() int64 { return c.cfg.ChainID }

// ─── 辅助 ────────────────────────────────────────────

// withRetry — 简单 1 次重试包装（RPC 偶发抖动）
func withRetry[T any](fn func() (T, error)) (T, error) {
	v, err := fn()
	if err == nil {
		return v, nil
	}
	time.Sleep(300 * time.Millisecond)
	return fn()
}

// HexToBytes — "0xabc" → []byte{0xab, 0xc}
func HexToBytes(s string) ([]byte, error) {
	return hex.DecodeString(strings.TrimPrefix(s, "0x"))
}
