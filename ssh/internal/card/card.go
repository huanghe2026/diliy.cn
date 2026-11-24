// Package card — GeoyuanCard 业务逻辑
//
// 分层：
//
//	card  (本包)  = 业务逻辑 + abigen 绑定 + RPC 调用
//	chain         = Polygon RPC 原始交互
//	h3            = H3 地理编码
//	store         = SQLite 用户数据
package card

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"geoyuan.cn/ssh/internal/chain"
	"geoyuan.cn/ssh/internal/h3"
)

// Card — 链上 GeoyuanCard 简化表示
type Card struct {
	TokenID   string  `json:"tokenId"` // hex 字符串
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

// Service — 卡片服务（可选 chain 客户端 + abigen 合约绑定）
type Service struct {
	chain    *chain.Client
	contract common.Address
	gyc      *GeoyuanCard // abigen 生成的合约绑定
	hasPriv  bool
	key      *ecdsa.PrivateKey
	from     common.Address
}

// New — 初始化（私钥可选；chain + 合约地址都有才尝试建绑定）
func New(c *chain.Client, privKeyHex string) (*Service, error) {
	s := &Service{chain: c}

	// 私钥（可选）
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

	// abigen 合约绑定（需要 chain + 合约地址 + RPC 可达）
	if c != nil {
		s.contract = c.ContractAddr()
		if s.contract != (common.Address{}) && c.IsUp() {
			gyc, err := NewGeoyuanCard(s.contract, c.Eth())
			if err != nil {
				return nil, fmt.Errorf("bind GeoyuanCard at %s: %w", s.contract.Hex(), err)
			}
			s.gyc = gyc
		}
	}
	return s, nil
}

// FromAddr — 返回签名地址（如果有私钥）
func (s *Service) FromAddr() string { return s.from.Hex() }

// ContractAddr — 返回合约地址
func (s *Service) ContractAddr() common.Address { return s.contract }

// Ready — 是否具备真实链上 Mint 能力（chain + 合约 + 私钥 全齐）
func (s *Service) Ready() bool {
	return s.gyc != nil && s.hasPriv
}

// ─── Mint ────────────────────────────────────────────

// MintParams — mint 一张卡所需参数
type MintParams struct {
	Lat         float64
	Lon         float64
	Res         int
	Title       string
	Description string
	Color       string
}

// MintResult — mint 结果
type MintResult struct {
	TokenID     *big.Int
	H3Index     uint64
	TxHash      string
	Block       uint64
	Lat         float64
	Lon         float64
	Res         int
	Title       string
	Description string
	Color       string
	Preview     bool
}

// Mint — 优先发链上交易；条件不满足时降级为 H3 预览
func (s *Service) Mint(p MintParams) (*MintResult, error) {
	res := p.Res
	if res <= 0 {
		res = h3.DefaultRes
	}
	cell := h3.Encode(p.Lat, p.Lon, res)

	if !s.Ready() {
		return &MintResult{
			TokenID:     big.NewInt(0),
			H3Index:     cell.Index,
			TxHash:      "(offline preview)",
			Lat:         cell.Lat,
			Lon:         cell.Lon,
			Res:         res,
			Title:       p.Title,
			Description: p.Description,
			Color:       p.Color,
			Preview:     true,
		}, nil
	}
	return s.mintOnChain(cell.Index, p, cell.Lat, cell.Lon, res)
}

// mintOnChain — 用 abigen 绑定发真实链上 Mint 交易
func (s *Service) mintOnChain(h3Index uint64, p MintParams, lat, lon float64, res int) (*MintResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// 先查 cell 是否已被 mint（快速失败，省 gas）
	isMinted, existingID, err := s.gyc.IsCellMinted(&bind.CallOpts{Context: ctx}, h3Index)
	if err != nil {
		return nil, fmt.Errorf("isCellMinted check: %w", err)
	}
	if isMinted {
		return nil, fmt.Errorf("h3 cell 0x%x already minted as token %d", h3Index, existingID)
	}

	chainID := s.chain.ChainID()
	opts, err := bind.NewKeyedTransactorWithChainID(s.key, big.NewInt(chainID))
	if err != nil {
		return nil, fmt.Errorf("transactor: %w", err)
	}
	nonce, err := s.chain.Nonce(s.from.Hex())
	if err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}
	opts.Nonce = big.NewInt(int64(nonce))
	opts.GasLimit = 600_000 // Mint 实测 ~137k，留足余量
	opts.Context = ctx

	latE4 := int32(lat * 10000)
	lonE4 := int32(lon * 10000)
	tx, err := s.gyc.Mint(opts,
		h3Index,
		p.Title,
		p.Description,
		p.Color,
		latE4,
		lonE4,
		uint32(res),
	)
	if err != nil {
		return nil, fmt.Errorf("mint tx submit: %w", err)
	}

	receipt, err := bind.WaitMined(ctx, s.chain.Eth(), tx)
	if err != nil {
		return nil, fmt.Errorf("wait for mint receipt: %w", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		// 尝试找 revert reason
		errReason := "unknown"
		for _, l := range receipt.Logs {
			if len(l.Data) >= 4 {
				// 简化：取前 4 bytes 近似
				errReason = fmt.Sprintf("reverted in tx %s", tx.Hash().Hex())
			}
		}
		return nil, fmt.Errorf("mint reverted: %s", errReason)
	}

	// 从 Receipt 的 Minted 事件里提取 tokenId
	var tokenID *big.Int
	for _, log := range receipt.Logs {
		ev, err := s.gyc.ParseMinted(*log)
		if err == nil {
			tokenID = ev.TokenId
			break
		}
	}
	if tokenID == nil {
		// fallback：读 totalSupply
		ts, _ := s.gyc.TotalSupply(&bind.CallOpts{Context: ctx})
		tokenID = ts
	}

	return &MintResult{
		TokenID:     tokenID,
		H3Index:     h3Index,
		TxHash:      receipt.TxHash.Hex(),
		Block:       receipt.BlockNumber.Uint64(),
		Lat:         lat,
		Lon:         lon,
		Res:         res,
		Title:       p.Title,
		Description: p.Description,
		Color:       p.Color,
		Preview:     false,
	}, nil
}

// ─── MyCards ─────────────────────────────────────────

// MyCards — 查某地址所有 GeoyuanCard，读 metadata 拿 title/desc/color/latE4/lonE4/res/timestamp
func (s *Service) MyCards(addrHex string) ([]Card, error) {
	if s.gyc == nil {
		return []Card{}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	addr := common.HexToAddress(addrHex)
	tokenIDs, err := s.gyc.MyCards(&bind.CallOpts{Context: ctx}, addr)
	if err != nil {
		return nil, fmt.Errorf("myCards: %w", err)
	}
	out := make([]Card, 0, len(tokenIDs))
	for _, tid := range tokenIDs {
		md, err := s.gyc.Metadata(&bind.CallOpts{Context: ctx}, tid)
		if err != nil {
			// metadata 查不到时至少保留有 tokenId 的条目
			out = append(out, Card{TokenID: tid.String(), Owner: addrHex})
			continue
		}
		out = append(out, Card{
			TokenID:   tid.String(),
			H3Index:   md.H3Index,
			Title:     md.Title,
			Desc:      md.Description,
			Color:     md.Color,
			Lat:       float64(md.LatE4) / 10000.0,
			Lon:       float64(md.LonE4) / 10000.0,
			Res:       int(md.Res),
			Timestamp: md.Timestamp.Int64(),
			Owner:     addrHex,
		})
	}
	return out, nil
}

// ─── 邻居匹配 ─────────────────────────────────────────

// NeighborMatch — H3 k-ring 里真正被 mint 的卡（owner != self）
// 用合约 neighborMatch(self, cells) 拿已 mint 的 tokenId，再读 metadata
func (s *Service) NeighborMatch(selfAddrHex string, h3Index uint64, k int) ([]Card, error) {
	ring := h3.KRing(h3Index, k)
	cells := make([]uint64, 0, len(ring))
	for _, c := range ring {
		cells = append(cells, c.Index)
	}

	if s.gyc == nil {
		// 无合约时降级：返回全部 ring 作为候选
		out := make([]Card, 0, len(ring))
		for _, c := range ring {
			out = append(out, Card{H3Index: c.Index, Lat: c.Lat, Lon: c.Lon, Res: c.Res})
		}
		return out, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	self := common.HexToAddress(selfAddrHex)
	tokenIDs, err := s.gyc.NeighborMatch(&bind.CallOpts{Context: ctx}, self, cells)
	if err != nil {
		return nil, fmt.Errorf("neighborMatch: %w", err)
	}
	out := make([]Card, 0, len(tokenIDs))
	for _, tid := range tokenIDs {
		md, err := s.gyc.Metadata(&bind.CallOpts{Context: ctx}, tid)
		if err != nil {
			out = append(out, Card{TokenID: tid.String()})
			continue
		}
		out = append(out, Card{
			TokenID:   tid.String(),
			H3Index:   md.H3Index,
			Title:     md.Title,
			Desc:      md.Description,
			Color:     md.Color,
			Lat:       float64(md.LatE4) / 10000.0,
			Lon:       float64(md.LonE4) / 10000.0,
			Res:       int(md.Res),
			Timestamp: md.Timestamp.Int64(),
		})
	}
	return out, nil
}

// ─── 工具 ────────────────────────────────────────────

// H3IndexHex — uint64 → hex 字符串
func H3IndexHex(idx uint64) string { return "0x" + fmt.Sprintf("%x", idx) }

// H3IndexFromHex — hex → uint64
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

var errMintedEventNotFound = errors.New("Minted event not found in receipt")
