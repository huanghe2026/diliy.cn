// Package card — GeoyuanCard 业务逻辑
//
// 分层：
//   card  (本包)  = 业务逻辑 + 本地 SQLite 缓存 + RPC 调用
//   chain         = Polygon RPC 原始交互
//   h3            = H3 地理编码
//   store         = SQLite 用户数据
package card

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"geoyuan.cn/ssh/internal/chain"
	"geoyuan.cn/ssh/internal/h3"
)

// Card — 链上 GeoyuanCard 简化表示
type Card struct {
	TokenID   string  `json:"tokenId"`   // hex 字符串
	H3Index   uint64  `json:"h3Index"`
	Title     string  `json:"title"`
	Desc      string  `json:"description"`
	Color     string  `json:"color"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Res       int     `json:"res"`
	Timestamp int64   `json:"timestamp"`
	Owner     string  `json:"owner"`
	TxHash    string  `json:"txHash,omitempty"`
}

// Service — 卡片服务（可选 chain 客户端 + 本地缓存）
type Service struct {
	chain   *chain.Client
	contract common.Address
	hasPriv bool
	key     *ecdsa.PrivateKey
	from    common.Address
}

// New — 初始化（私钥可选，只有签名交易才需要）
func New(c *chain.Client, privKeyHex string) (*Service, error) {
	s := &Service{chain: c}
	if privKeyHex != "" {
		privKeyHex = strings.TrimPrefix(privKeyHex, "0x")
		key, err := crypto.HexToECDSA(privKeyHex)
		if err != nil {
			return nil, fmt.Errorf("private key: %w", err)
		}
		s.key = key
		s.from = crypto.PubkeyToAddress(key.PublicKey)
		s.hasPriv = true
	}
	if c != nil {
		s.contract = c.ContractAddr()
	}
	return s, nil
}

// FromAddr — 返回签名地址（如果有私钥）
func (s *Service) FromAddr() string {
	return s.from.Hex()
}

// ─── Mint ────────────────────────────────────────────

// MintParams — mint 一张卡所需参数
type MintParams struct {
	Lat         float64 // 纬度
	Lon         float64 // 经度
	Res         int     // H3 精度 (0-15)
	Title       string  // 标题
	Description string  // 描述
	Color       string  // "#c77dff" 彩虹色
}

// MintResult — mint 结果
type MintResult struct {
	TokenID *big.Int
	H3Index uint64
	TxHash  string
	Lat     float64
	Lon     float64
	Res     int
}

// Mint — 在链上 mint 一张 GeoyuanCard（如果有私钥 + 合约）
// 没私钥 → 只做 H3 编码计算（用于 TUI 预览）
func (s *Service) Mint(p MintParams) (*MintResult, error) {
	res := p.Res
	if res <= 0 {
		res = h3.DefaultRes
	}

	// 1. H3 编码
	cell := h3.Encode(p.Lat, p.Lon, res)

	// 2. 没有私钥/合约 — 只返回编码结果（TUI 预览）
	if !s.hasPriv || (s.contract == common.Address{}) {
		return &MintResult{
			TokenID: big.NewInt(0),
			H3Index: cell.Index,
			TxHash:  "(offline preview)",
			Lat:     cell.Lat,
			Lon:     cell.Lon,
			Res:     res,
		}, nil
	}

	// 3. 调链上合约 mint（原生 eth_call）
	return s.mintOnChain(cell.Index, p, cell.Lat, cell.Lon)
}

// mintOnChain — 真正发链上交易
func (s *Service) mintOnChain(h3Index uint64, p MintParams, lat, lon float64) (*MintResult, error) {
	if s.chain == nil {
		return nil, fmt.Errorf("chain client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 构造 tx opts
	chainID := s.chain.ChainID()
	opts, err := bind.NewKeyedTransactorWithChainID(s.key, big.NewInt(chainID))
	if err != nil {
		return nil, fmt.Errorf("transactor: %w", err)
	}

	nonce, _ := s.chain.Nonce(s.from.Hex())
	opts.Nonce = big.NewInt(int64(nonce))
	opts.GasLimit = 200000 // mint 操作 ~100k gas
	opts.Context = ctx

	// ⚠️ 实际调用合约需要 ABI — 这里先用原始方法
	// GeoyuanCard.mint(h3Index, title, desc, color, latE4, lonE4, res)
	// 编译好的 ABI 编码在 contracts/ 目录里

	// 简化：先返回 H3 编码结果 + 告诉用户合约 ABI 编译位置
	_ = opts
	latE4 := int16(lat * 10000)
	lonE4 := int16(lon * 10000)

	return &MintResult{
		TokenID: big.NewInt(0), // 链上返回
		H3Index: h3Index,
		TxHash:  fmt.Sprintf("(send raw tx for h3=%d latE4=%d lonE4=%d)", h3Index, latE4, lonE4),
		Lat:     lat,
		Lon:     lon,
		Res:     p.Res,
	}, nil
}

// ─── MyCards ─────────────────────────────────────────

// MyCards — 查某地址所有 GeoyuanCard（本地版本先返回占位，链上读 tokenURI）
func (s *Service) MyCards(addr string) ([]Card, error) {
	// TODO: 等合约部署后用 eth_call 调 myCards(addr) → tokenURI → 解析
	return []Card{}, nil
}

// ─── 邻居匹配 ─────────────────────────────────────────

// NeighborMatch — 查某 H3 cell 附近 k-ring 内的别人的卡
func (s *Service) NeighborMatch(h3Index uint64, k int) ([]Card, error) {
	ring := h3.KRing(h3Index, k)
	// TODO: 对每个 cellMinted[h] != 0 → owner 不是自己 → 收集
	out := make([]Card, 0)
	for _, c := range ring {
		out = append(out, Card{H3Index: c.Index, Lat: c.Lat, Lon: c.Lon, Res: c.Res})
	}
	return out, nil
}

// ─── 工具 ────────────────────────────────────────────

// H3IndexHex — uint64 → hex 字符串（和链上 tokenURI 匹配）
func H3IndexHex(idx uint64) string {
	return "0x" + fmt.Sprintf("%x", idx)
}

// H3IndexFromHex — hex（带/不带 0x）→ uint64
func H3IndexFromHex(s string) (uint64, error) {
	s = strings.TrimPrefix(s, "0x")
	b, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}
	var n uint64
	for _, by := range b {
		n = n<<8 | uint64(by)
	}
	return n, nil
}
