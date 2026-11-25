// Package wallet — 每个 SSH 用户一个独立的 Polygon 钱包。
//
// 派生方式：标准 BIP32 / BIP44 Ethereum 路径
//
//	m/44'/60'/0'/0/{userID}
//
// master seed 来自环境变量 GEOYUAN_WALLET_SEED（64 位 hex，即 32 字节）。
// 只要 seed 不变，同一 userID 永远派生出同一个地址，无需存私钥。
//
// 用户钱包初始没有 MATIC，首次 mint 前会从部署钱包自动小额充值 gas。
//
// ⚠️ 这是托管（custodial）方案：私钥由服务器派生，用户无法自行导出。
// 上主网前必须改为用户自持私钥或智能合约账户。
package wallet

import (
	"context"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"geoyuan.cn/ssh/internal/card"
	"geoyuan.cn/ssh/internal/chain"
)

const (
	// mintGas — 单次 mint 的 gas 上限（实测 ~400-550k，因字符串长度浮动）
	mintGas = 700_000
	// 同一地址 60s 内不重复补 gas
	gasCooldown = 60 * time.Second
)

var curveN = crypto.S256().Params().N

// Manager — 用户钱包 + 链上 Mint 的统一入口
type Manager struct {
	svc    *card.Service
	chain  *chain.Client
	seed   []byte
	depKey *ecdsa.PrivateKey
	depAdr common.Address

	mu     sync.Mutex
	funded map[string]time.Time // 地址 → 上次补 gas 时间
}

// New — 创建 Manager。seedHex 为空表示不启用用户钱包（降级为部署钱包代签）。
func New(svc *card.Service, seedHex string) (*Manager, error) {
	m := &Manager{
		svc:    svc,
		funded: make(map[string]time.Time),
	}
	if svc != nil {
		m.chain = svc.Chain()
		m.depKey = svc.SignerKey()
		if m.depKey != nil {
			m.depAdr = crypto.PubkeyToAddress(m.depKey.PublicKey)
		}
	}
	if seedHex = strings.TrimPrefix(strings.TrimSpace(seedHex), "0x"); seedHex != "" {
		b, err := hex.DecodeString(seedHex)
		if err != nil {
			return nil, fmt.Errorf("wallet seed: %w", err)
		}
		if len(b) < 16 {
			return nil, fmt.Errorf("wallet seed too short (%d bytes), need >= 16", len(b))
		}
		m.seed = b
	}
	return m, nil
}

// Ready — 是否启用用户独立钱包（需要 seed + chain + 合约）
func (m *Manager) Ready() bool {
	return m != nil && len(m.seed) > 0 && m.svc != nil && m.svc.ContractReady()
}

// Service — 底层只读卡片服务（MyCards / NeighborMatch）
func (m *Manager) Service() *card.Service {
	if m == nil {
		return nil
	}
	return m.svc
}

// AddressFor — userID 对应的钱包地址（seed 未配置时返回空串）
func (m *Manager) AddressFor(userID int64) string {
	key, err := m.keyFor(userID)
	if err != nil || key == nil {
		return ""
	}
	return crypto.PubkeyToAddress(key.PublicKey).Hex()
}

// Mint — 用 userID 对应的钱包 mint 一张卡（自动补 gas）
func (m *Manager) Mint(userID int64, p card.MintParams) (*card.MintResult, error) {
	if m == nil || m.svc == nil {
		return nil, fmt.Errorf("card service unavailable")
	}
	key, err := m.keyFor(userID)
	if err != nil {
		return nil, err
	}
	if key == nil {
		// 未启用独立钱包 → 沿用部署钱包代签（兼容旧行为）
		return m.svc.Mint(p)
	}
	addr := crypto.PubkeyToAddress(key.PublicKey)
	if err := m.ensureGas(addr); err != nil {
		return nil, fmt.Errorf("gas top-up: %w", err)
	}
	return m.svc.MintAs(key, p)
}

// List — 把 userID 钱包的卡挂到市场（自动补 gas）
func (m *Manager) List(userID int64, tokenID *big.Int, priceWei *big.Int) (txHash string, block uint64, err error) {
	if m == nil || m.svc == nil {
		return "", 0, fmt.Errorf("card service unavailable")
	}
	key, err := m.keyFor(userID)
	if err != nil {
		return "", 0, err
	}
	if key == nil {
		return "", 0, fmt.Errorf("user wallet not configured")
	}
	addr := crypto.PubkeyToAddress(key.PublicKey)
	if err := m.ensureGas(addr); err != nil {
		return "", 0, fmt.Errorf("gas top-up: %w", err)
	}
	return m.svc.List(key, tokenID, priceWei)
}

// Buy — 用 userID 钱包买下一张挂售卡（自动补 gas + 自动带 MATIC）
func (m *Manager) Buy(userID int64, tokenID *big.Int, valueWei *big.Int) (txHash string, block uint64, err error) {
	if m == nil || m.svc == nil {
		return "", 0, fmt.Errorf("card service unavailable")
	}
	key, err := m.keyFor(userID)
	if err != nil {
		return "", 0, err
	}
	if key == nil {
		return "", 0, fmt.Errorf("user wallet not configured")
	}
	addr := crypto.PubkeyToAddress(key.PublicKey)
	// 买需要 gas + 卡的价格本身
	if err := m.ensureGas(addr); err != nil {
		return "", 0, fmt.Errorf("gas top-up: %w", err)
	}
	return m.svc.Buy(key, tokenID, valueWei)
}

// ─── gas ────────────────────────────────────────────

func (m *Manager) ensureGas(addr common.Address) error {
	if m.chain == nil {
		return nil // 无链客户端，交给 mint 自己报错
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	wei, err := m.chain.BalanceWei(addr.Hex())
	if err != nil {
		return fmt.Errorf("read balance: %w", err)
	}
	gasPrice, err := m.chain.GasPrice()
	if err != nil {
		return fmt.Errorf("gas price: %w", err)
	}
	need := new(big.Int).Mul(gasPrice, big.NewInt(mintGas)) // 一次 mint 所需
	if wei.Cmp(need) >= 0 {
		return nil
	}
	if m.depKey == nil {
		return fmt.Errorf("user wallet %s has insufficient MATIC for gas and no gas signer is configured (CARD_CHAIN_KEY missing)", addr.Hex())
	}

	// 冷却，避免并发请求重复转账
	h := addr.Hex()
	m.mu.Lock()
	if t, ok := m.funded[h]; ok && time.Since(t) < gasCooldown {
		m.mu.Unlock()
		return nil
	}
	m.funded[h] = time.Now()
	m.mu.Unlock()

	// 补到「两次 mint 的 gas」水位，避免每 mint 一次都要转账
	topUp := new(big.Int).Sub(new(big.Int).Mul(need, big.NewInt(2)), wei)

	nonce, err := m.chain.Nonce(m.depAdr.Hex())
	if err != nil {
		return fmt.Errorf("deployer nonce: %w", err)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      21000,
		To:       &addr,
		Value:    topUp,
	})
	signed, err := types.SignTx(tx, types.NewLondonSigner(big.NewInt(m.chain.ChainID())), m.depKey)
	if err != nil {
		return fmt.Errorf("sign top-up: %w", err)
	}
	if err := m.chain.Eth().SendTransaction(ctx, signed); err != nil {
		return fmt.Errorf("send top-up: %w", err)
	}
	if _, err := bind.WaitMined(ctx, m.chain.Eth(), signed); err != nil {
		return fmt.Errorf("wait top-up: %w", err)
	}
	return nil
}

// ─── BIP32 派生 ──────────────────────────────────────

func (m *Manager) keyFor(userID int64) (*ecdsa.PrivateKey, error) {
	if len(m.seed) == 0 {
		return nil, nil
	}
	if userID <= 0 {
		return nil, fmt.Errorf("invalid userID %d for wallet derivation", userID)
	}
	return derive(m.seed, uint32(userID))
}

type extKey struct {
	key   []byte // 32B private key
	chain []byte // 32B chain code
}

// derive — m/44'/60'/0'/0/{index}
func derive(seed []byte, index uint32) (*ecdsa.PrivateKey, error) {
	k := masterFromSeed(seed)
	path := []uint32{
		44 + 0x80000000, // 44'
		60 + 0x80000000, // 60'
		0 + 0x80000000,  // 0'
		0,               // 0
		index,           // account index = userID
	}
	for _, i := range path {
		var err error
		k, err = k.child(i)
		if err != nil {
			return nil, err
		}
	}
	return crypto.ToECDSA(k.key)
}

func masterFromSeed(seed []byte) extKey {
	mac := hmac.New(sha512.New, []byte("Bitcoin seed"))
	mac.Write(seed)
	I := mac.Sum(nil)
	return extKey{key: I[:32], chain: I[32:]}
}

// child — BIP32 CKDpriv
func (k extKey) child(i uint32) (extKey, error) {
	var data []byte
	if i >= 0x80000000 { // hardened
		data = make([]byte, 0, 37)
		data = append(data, 0x00)
		data = append(data, k.key...)
	} else {
		priv, err := crypto.ToECDSA(k.key)
		if err != nil {
			return extKey{}, fmt.Errorf("ckd parent key: %w", err)
		}
		data = append(data, crypto.FromECDSAPub(&priv.PublicKey)...)
	}
	var idx [4]byte
	binary.BigEndian.PutUint32(idx[:], i)
	data = append(data, idx[:]...)

	mac := hmac.New(sha512.New, k.chain)
	mac.Write(data)
	I := mac.Sum(nil)

	il := new(big.Int).SetBytes(I[:32])
	il.Add(il, new(big.Int).SetBytes(k.key))
	il.Mod(il, curveN)
	if il.Sign() == 0 {
		return extKey{}, fmt.Errorf("derived zero key at index %d", i)
	}
	childKey := make([]byte, 32)
	il.FillBytes(childKey)
	return extKey{key: childKey, chain: I[32:]}, nil
}
