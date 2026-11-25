package wallet

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

// 固定 seed 下派生必须稳定，且不同 userID 地址不同
func TestDeriveDeterministic(t *testing.T) {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}
	a1, err := derive(seed, 1)
	if err != nil {
		t.Fatal(err)
	}
	a1b, err := derive(seed, 1)
	if err != nil {
		t.Fatal(err)
	}
	if a1.D.Cmp(a1b.D) != 0 {
		t.Fatal("derivation not deterministic")
	}

	a2, err := derive(seed, 2)
	if err != nil {
		t.Fatal(err)
	}
	if crypto.PubkeyToAddress(a1.PublicKey) == crypto.PubkeyToAddress(a2.PublicKey) {
		t.Fatal("different userID produced same address")
	}
	t.Logf("user 1 → %s", crypto.PubkeyToAddress(a1.PublicKey).Hex())
	t.Logf("user 2 → %s", crypto.PubkeyToAddress(a2.PublicKey).Hex())
}

// 不同 seed 派生出不同地址
func TestDeriveDifferentSeed(t *testing.T) {
	s1 := make([]byte, 32)
	s2 := make([]byte, 32)
	s2[0] = 1
	k1, _ := derive(s1, 1)
	k2, _ := derive(s2, 1)
	if crypto.PubkeyToAddress(k1.PublicKey) == crypto.PubkeyToAddress(k2.PublicKey) {
		t.Fatal("different seed produced same address")
	}
}

// 无效 userID 必须报错
func TestDeriveInvalidID(t *testing.T) {
	m := &Manager{seed: make([]byte, 32)}
	if _, err := m.keyFor(0); err == nil {
		t.Fatal("expected error for userID=0")
	}
	if k, err := m.keyFor(1); err != nil || k == nil {
		t.Fatalf("userID=1 should derive fine: %v", err)
	}
}
