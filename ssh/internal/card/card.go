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
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"geoyuan.cn/ssh/internal/chain"
	"geoyuan.cn/ssh/internal/h3"
)

// ─── 市场类型 ────────────────────────────────────────

// MarketListing — 链上挂售信息（含 metadata 拼装）
type MarketListing struct {
	TokenID     string  `json:"tokenId"`
	Seller      string  `json:"seller"`
	PriceWei    string  `json:"priceWei"`
	PriceMATIC  string  `json:"priceMATIC"`
	ListedBlock uint64  `json:"listedBlock"`
	Title       string  `json:"title"`
	Color       string  `json:"color"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	H3Index     uint64  `json:"h3Index"`
}

// priceMATIC — wei → MATIC 显示（18 位精度，2 位小数）
func priceMATIC(wei *big.Int) string {
	if wei == nil {
		return "0.00"
	}
	// MATIC 是 18 位精度
	denom := big.NewInt(1000000000000000000) // 1e18
	q := new(big.Int).Div(wei, denom)
	r := new(big.Int).Mod(wei, denom)
	// r / 1e16 取 2 位小数
	denom2 := big.NewInt(10000000000000000) // 1e16
	cents := new(big.Int).Div(r, denom2)
	return fmt.Sprintf("%s.%02d", q.String(), cents.Uint64())
}

// ─── Card 类型（不变）───────────────────────────────
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

// Chain — 底层链客户端（供 wallet 包做 gas 充值）
func (s *Service) Chain() *chain.Client { return s.chain }

// SignerKey — 部署钱包私钥（供 wallet 包做 gas 充值签名）
func (s *Service) SignerKey() *ecdsa.PrivateKey { return s.key }

// ContractReady — 合约绑定是否可用（读操作即可用，不需要私钥）
func (s *Service) ContractReady() bool { return s.gyc != nil }

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

// Mint — 用服务自带（部署）私钥 mint；条件不满足时降级为 H3 预览
func (s *Service) Mint(p MintParams) (*MintResult, error) {
	if !s.Ready() {
		return s.preview(p), nil
	}
	return s.MintAs(s.key, p)
}

// MintAs — 用指定私钥签名 mint（用户独立钱包走这里）
// 无合约绑定时返回 H3 预览（Preview=true），不报错。
func (s *Service) MintAs(key *ecdsa.PrivateKey, p MintParams) (*MintResult, error) {
	if s.gyc == nil || key == nil {
		return s.preview(p), nil
	}
	res := p.Res
	if res <= 0 {
		res = h3.DefaultRes
	}
	cell := h3.Encode(p.Lat, p.Lon, res)
	return s.mintOnChain(key, cell.Index, p, cell.Lat, cell.Lon, res)
}

// preview — 离线 H3 预览结果
func (s *Service) preview(p MintParams) *MintResult {
	res := p.Res
	if res <= 0 {
		res = h3.DefaultRes
	}
	cell := h3.Encode(p.Lat, p.Lon, res)
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
	}
}

// mintOnChain — 用指定私钥发真实链上 Mint 交易
func (s *Service) mintOnChain(key *ecdsa.PrivateKey, h3Index uint64, p MintParams, lat, lon float64, res int) (*MintResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	from := crypto.PubkeyToAddress(key.PublicKey)

	// 先查 cell 是否已被 mint（快速失败，省 gas）
	isMinted, existingID, err := s.gyc.IsCellMinted(&bind.CallOpts{Context: ctx}, h3Index)
	if err != nil {
		return nil, fmt.Errorf("isCellMinted check: %w", err)
	}
	if isMinted {
		return nil, fmt.Errorf("h3 cell 0x%x already minted as token %d", h3Index, existingID)
	}

	chainID := s.chain.ChainID()
	opts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(chainID))
	if err != nil {
		return nil, fmt.Errorf("transactor: %w", err)
	}
	nonce, err := s.chain.Nonce(from.Hex())
	if err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}
	opts.Nonce = big.NewInt(int64(nonce))
	opts.GasLimit = 700_000 // 实测 400-550k（字符串 SSTORE 占大头），留足余量
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

	// 从 Receipt 的 Minted 事件里提取 tokenId（失败直接报错，不要用 totalSupply 冒充）
	var tokenID *big.Int
	for _, log := range receipt.Logs {
		ev, err := s.gyc.ParseMinted(*log)
		if err == nil {
			tokenID = ev.TokenId
			break
		}
	}
	if tokenID == nil {
		return nil, fmt.Errorf("mint succeeded (tx=%s) but Minted event not found in receipt", tx.Hash().Hex())
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

// metadataConcurrency — 并发拉 metadata 的上限（避免 RPC 限速）
const metadataConcurrency = 6

// MyCards — 查某地址所有 GeoyuanCard（并发拉 metadata）
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
	return s.fetchMetadataBatch(ctx, tokenIDs, addrHex), nil
}

// fetchMetadataBatch — 并发拉一组 tokenId 的 metadata
// 单张失败不阻塞整体，用占位 Card 顶替
func (s *Service) fetchMetadataBatch(ctx context.Context, tokenIDs []*big.Int, ownerHex string) []Card {
	out := make([]Card, len(tokenIDs))
	sem := make(chan struct{}, metadataConcurrency)
	var wg sync.WaitGroup
	for i, tid := range tokenIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, tid *big.Int) {
			defer wg.Done()
			defer func() { <-sem }()
			md, err := s.gyc.Metadata(&bind.CallOpts{Context: ctx}, tid)
			if err != nil {
				out[i] = Card{TokenID: tid.String(), Owner: ownerHex}
				return
			}
			out[i] = Card{
				TokenID:   tid.String(),
				H3Index:   md.H3Index,
				Title:     md.Title,
				Desc:      md.Description,
				Color:     md.Color,
				Lat:       float64(md.LatE4) / 10000.0,
				Lon:       float64(md.LonE4) / 10000.0,
				Res:       int(md.Res),
				Timestamp: md.Timestamp.Int64(),
				Owner:     ownerHex,
			}
		}(i, tid)
	}
	wg.Wait()
	return out
}

// ─── 邻居匹配 ─────────────────────────────────────────

// NeighborMatch — H3 k-ring 里真正被 mint 的卡（owner != self）
// 用合约 neighborMatch(self, cells) 拿已 mint 的 tokenId，再并发读 metadata
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
	return s.fetchMetadataBatch(ctx, tokenIDs, ""), nil
}

// ─── 市场：挂售 / 下架 / 购买 ──────────────────────────

// List — 把自己的一张卡挂到市场（token 转合约托管）
func (s *Service) List(key *ecdsa.PrivateKey, tokenID *big.Int, priceWei *big.Int) (txHash string, block uint64, err error) {
	if s.gyc == nil || key == nil {
		return "", 0, fmt.Errorf("合约未就绪或无私钥")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	chainID := s.chain.ChainID()
	opts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(chainID))
	if err != nil {
		return "", 0, fmt.Errorf("transactor: %w", err)
	}
	from := crypto.PubkeyToAddress(key.PublicKey)
	nonce, err := s.chain.Nonce(from.Hex())
	if err != nil {
		return "", 0, fmt.Errorf("nonce: %w", err)
	}
	opts.Nonce = big.NewInt(int64(nonce))
	opts.GasLimit = 300_000 // list 比 mint 便宜（不需要 SSTORE 字符串）
	opts.Context = ctx

	tx, err := s.gyc.List(opts, tokenID, priceWei)
	if err != nil {
		return "", 0, fmt.Errorf("list tx: %w", err)
	}
	receipt, err := bind.WaitMined(ctx, s.chain.Eth(), tx)
	if err != nil {
		return "", 0, fmt.Errorf("wait receipt: %w", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return tx.Hash().Hex(), receipt.BlockNumber.Uint64(), fmt.Errorf("list reverted (tx=%s)", tx.Hash().Hex())
	}
	return receipt.TxHash.Hex(), receipt.BlockNumber.Uint64(), nil
}

// Unlist — 下架自己挂的卡（token 回到 seller）
func (s *Service) Unlist(key *ecdsa.PrivateKey, tokenID *big.Int) (txHash string, block uint64, err error) {
	if s.gyc == nil || key == nil {
		return "", 0, fmt.Errorf("合约未就绪或无私钥")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	chainID := s.chain.ChainID()
	opts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(chainID))
	if err != nil {
		return "", 0, fmt.Errorf("transactor: %w", err)
	}
	from := crypto.PubkeyToAddress(key.PublicKey)
	nonce, err := s.chain.Nonce(from.Hex())
	if err != nil {
		return "", 0, fmt.Errorf("nonce: %w", err)
	}
	opts.Nonce = big.NewInt(int64(nonce))
	opts.GasLimit = 300_000
	opts.Context = ctx

	tx, err := s.gyc.Unlist(opts, tokenID)
	if err != nil {
		return "", 0, fmt.Errorf("unlist tx: %w", err)
	}
	receipt, err := bind.WaitMined(ctx, s.chain.Eth(), tx)
	if err != nil {
		return "", 0, fmt.Errorf("wait receipt: %w", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return tx.Hash().Hex(), receipt.BlockNumber.Uint64(), fmt.Errorf("unlist reverted (tx=%s)", tx.Hash().Hex())
	}
	return receipt.TxHash.Hex(), receipt.BlockNumber.Uint64(), nil
}

// Buy — 买下一张挂售卡（合约里 MATIC 给 seller，token 给 buyer）
// valueWei — 实际随交易送的金额（必须 >= listing.price）
func (s *Service) Buy(key *ecdsa.PrivateKey, tokenID *big.Int, valueWei *big.Int) (txHash string, block uint64, err error) {
	if s.gyc == nil || key == nil {
		return "", 0, fmt.Errorf("合约未就绪或无私钥")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	chainID := s.chain.ChainID()
	opts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(chainID))
	if err != nil {
		return "", 0, fmt.Errorf("transactor: %w", err)
	}
	from := crypto.PubkeyToAddress(key.PublicKey)
	nonce, err := s.chain.Nonce(from.Hex())
	if err != nil {
		return "", 0, fmt.Errorf("nonce: %w", err)
	}
	opts.Nonce = big.NewInt(int64(nonce))
	opts.GasLimit = 300_000
	opts.Value = valueWei // payable
	opts.Context = ctx

	tx, err := s.gyc.Buy(opts, tokenID)
	if err != nil {
		return "", 0, fmt.Errorf("buy tx: %w", err)
	}
	receipt, err := bind.WaitMined(ctx, s.chain.Eth(), tx)
	if err != nil {
		return "", 0, fmt.Errorf("wait receipt: %w", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return tx.Hash().Hex(), receipt.BlockNumber.Uint64(), fmt.Errorf("buy reverted (tx=%s)", tx.Hash().Hex())
	}
	return receipt.TxHash.Hex(), receipt.BlockNumber.Uint64(), nil
}

// ListingOf — 查某张卡的挂售状态（空 seller 表示未挂售）
func (s *Service) ListingOf(tokenID *big.Int) (seller string, priceWei *big.Int, listedBlock uint64, err error) {
	if s.gyc == nil {
		return "", big.NewInt(0), 0, fmt.Errorf("合约未就绪")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := s.gyc.ListingOf(&bind.CallOpts{Context: ctx}, tokenID)
	if err != nil {
		return "", big.NewInt(0), 0, fmt.Errorf("listingOf: %w", err)
	}
	return res.Seller.Hex(), res.Price, res.ListedBlock.Uint64(), nil
}

// MarketListings — 拉全市场挂售列表（分页）
func (s *Service) MarketListings(limit, offset uint64) ([]MarketListing, error) {
	if s.gyc == nil {
		return []MarketListing{}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	res, err := s.gyc.MarketListings(&bind.CallOpts{Context: ctx}, new(big.Int).SetUint64(limit), new(big.Int).SetUint64(offset))
	if err != nil {
		return nil, fmt.Errorf("marketListings: %w", err)
	}
	return s.fetchMarketMeta(ctx, res.TokenIds, res.Prices, res.Sellers), nil
}

// ListingCount — 全市场挂售总数
func (s *Service) ListingCount() (uint64, error) {
	if s.gyc == nil {
		return 0, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	n, err := s.gyc.ListingCount(&bind.CallOpts{Context: ctx})
	if err != nil {
		return 0, fmt.Errorf("listingCount: %w", err)
	}
	return n.Uint64(), nil
}

// fetchMarketMeta — 并发拉市场 listing 的 metadata，拼出完整 MarketListing
func (s *Service) fetchMarketMeta(ctx context.Context, tokenIDs []*big.Int, prices []*big.Int, sellers []common.Address) []MarketListing {
	out := make([]MarketListing, len(tokenIDs))
	sem := make(chan struct{}, metadataConcurrency)
	var wg sync.WaitGroup
	for i, tid := range tokenIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, tid *big.Int) {
			defer wg.Done()
			defer func() { <-sem }()
			md, err := s.gyc.Metadata(&bind.CallOpts{Context: ctx}, tid)
			if err != nil {
				out[i] = MarketListing{
					TokenID:    tid.String(),
					Seller:     sellers[i].Hex(),
					PriceWei:   prices[i].String(),
					PriceMATIC: priceMATIC(prices[i]),
				}
				return
			}
			out[i] = MarketListing{
				TokenID:     tid.String(),
				Seller:      sellers[i].Hex(),
				PriceWei:    prices[i].String(),
				PriceMATIC:  priceMATIC(prices[i]),
				Title:       md.Title,
				Color:       md.Color,
				Lat:         float64(md.LatE4) / 10000.0,
				Lon:         float64(md.LonE4) / 10000.0,
				H3Index:     md.H3Index,
			}
		}(i, tid)
	}
	wg.Wait()
	return out
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
