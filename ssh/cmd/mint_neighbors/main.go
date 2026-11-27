// cmd/mint_neighbors — 一次性测试工具：用部署钱包给 3 个临时钱包转 gas，
// 让它们分别在 centerCell 的 k=1 邻居 cell 上 mint GeoyuanCard，
// 用于验证 SSH TUI 邻居匹配功能。
//
// ⚠️ 仅用于 Amoy 测试网。不要在主网跑 — 会真实消耗 MATIC。
//
// 用法：
//
//	cd ssh && go run ./cmd/mint_neighbors
//
// 私钥来源：myweb/diliy/.env 的 CARD_CHAIN_KEY（部署钱包）
package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	card "geoyuan.cn/ssh/internal/card"
	h3pkg "geoyuan.cn/ssh/internal/h3"
)

const (
	rpcURL      = "https://polygon-amoy-bor-rpc.publicnode.com"
	contractHex = "0xCB441824e3A67Fb30e89A78d9dcFf4F9E618D748"
	chainID     = int64(80002)
	centerCell  = uint64(0x89309a553dbffff)
)

type wallet struct {
	name string
	key  *ecdsa.PrivateKey
	addr common.Address
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	privHex := loadEnv()["CARD_CHAIN_KEY"]
	privHex = strings.TrimPrefix(privHex, "0x")
	deployerKey, _ := crypto.HexToECDSA(privHex)
	deployerAddr := crypto.PubkeyToAddress(deployerKey.PublicKey)

	c, _ := ethclient.DialContext(ctx, rpcURL)
	defer c.Close()
	fmt.Printf("RPC connected chain_id=%d\n", chainID)

	contractAddr := common.HexToAddress(contractHex)
	gyc, _ := card.NewGeoyuanCard(contractAddr, c)
	ts, _ := gyc.TotalSupply(nil)
	fmt.Printf("totalSupply before=%s\n", ts.String())

	// 邻居 cells
	neighbors := h3pkg.Neighbors(centerCell)
	fmt.Printf("centerCell=0x%x → %d neighbors (k=1)\n", centerCell, len(neighbors))
	for i, n := range neighbors {
		fmt.Printf("  [%d] 0x%x lat=%.4f lon=%.4f res=%d\n", i, n.Index, n.Lat, n.Lon, n.Res)
	}

	// 3 个测试钱包
	wallets := []wallet{}
	for _, name := range []string{"alice", "bob", "carol"} {
		k, _ := crypto.GenerateKey()
		w := wallet{name: name, key: k, addr: crypto.PubkeyToAddress(k.PublicKey)}
		wallets = append(wallets, w)
		fmt.Printf("  wallet %s addr=%s\n", name, w.addr.Hex())
	}

	// 转 0.1 MATIC（足够 mint gas ~54M wei）
	deployerNonce, _ := c.PendingNonceAt(ctx, deployerAddr)
	gasPrice, _ := c.SuggestGasPrice(ctx)
	transferAmt := big.NewInt(1e17) // 0.1 MATIC
	for _, w := range wallets {
		transfer, _ := types.SignNewTx(deployerKey, types.NewLondonSigner(big.NewInt(chainID)), &types.LegacyTx{
			Nonce:    deployerNonce,
			GasPrice: gasPrice,
			Gas:      21000,
			To:       &w.addr,
			Value:    transferAmt,
		})
		c.SendTransaction(ctx, transfer)
		receipt, _ := bind.WaitMined(ctx, c, transfer)
		fmt.Printf("  transfer %s ✓ block=%d\n", w.name, receipt.BlockNumber.Uint64())
		deployerNonce++
	}

	// mint
	for i, w := range wallets {
		nei := neighbors[i]
		auth, _ := bind.NewKeyedTransactorWithChainID(w.key, big.NewInt(chainID))
		auth.Context = ctx
		nonce, _ := c.PendingNonceAt(ctx, w.addr)
		auth.Nonce = big.NewInt(int64(nonce))
		auth.GasLimit = 600_000

		title := fmt.Sprintf("Neighbor-%s", w.name)
		desc := fmt.Sprintf("Mint by %s at neighbor cell", w.name)
		tx, err := gyc.Mint(auth, nei.Index, title, desc, "#00f5ff",
			int32(nei.Lat*10000), int32(nei.Lon*10000), uint32(nei.Res))
		if err != nil {
			fmt.Printf("  %s mint FAIL: %v\n", w.name, err)
			continue
		}
		receipt, _ := bind.WaitMined(ctx, c, tx)
		fmt.Printf("  %s mint OK block=%d tx=%s\n", w.name, receipt.BlockNumber.Uint64(), tx.Hash().Hex()[:20]+"...")
	}

	ts, _ = gyc.TotalSupply(nil)
	fmt.Printf("totalSupply after=%s\n", ts.String())

	// 快速验证 neighborMatch
	fmt.Println("\n--- neighborMatch(self=0x0000, cells=all 6 neighbors) ---")
	anyAddr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	cells := []uint64{}
	for _, n := range neighbors {
		cells = append(cells, n.Index)
	}
	matched, err := gyc.NeighborMatch(nil, anyAddr, cells)
	if err != nil {
		fmt.Printf("  query err: %v\n", err)
	} else {
		fmt.Printf("  matched tokenIds (all owners): %d\n", len(matched))
		for _, tid := range matched {
			owner, _ := gyc.OwnerOf(nil, tid)
			md, _ := gyc.Metadata(nil, tid)
			fmt.Printf("    tokenId=%s owner=%s cell=0x%x title=%q\n",
				tid.String(), owner.Hex()[:10]+"...", md.H3Index, md.Title)
		}
	}
}

func loadEnv() map[string]string {
	data, err := os.ReadFile("/home/huanghe/myweb/diliy/.env")
	if err != nil {
		return nil
	}
	out := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			out[strings.TrimSpace(parts[0])] = strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		}
	}
	return out
}
