package runtime

import (
	"math/big"
	"math/rand"
	"testing"
)

func TestUint128ArithmeticMatchesBigInt(t *testing.T) {
	values := []string{"0", "1", "18446744073709551615", "18446744073709551616", "170141183460469231731687303715884105727", "340282366920938463463374607431768211455"}
	for _, leftText := range values {
		leftBig, _ := new(big.Int).SetString(leftText, 10)
		left, ok := Uint128FromBig(leftBig)
		if !ok {
			t.Fatalf("parse %s", leftText)
		}
		for _, rightText := range values {
			rightBig, _ := new(big.Int).SetString(rightText, 10)
			right, _ := Uint128FromBig(rightBig)
			if sum, ok := left.AddChecked(right); ok {
				want := new(big.Int).Add(leftBig, rightBig)
				if want.BitLen() > 128 || sum.BigInt().Cmp(want) != 0 {
					t.Fatalf("add %s %s", leftText, rightText)
				}
			}
			if product, ok := left.MulChecked(right); ok {
				want := new(big.Int).Mul(leftBig, rightBig)
				if want.BitLen() > 128 || product.BigInt().Cmp(want) != 0 {
					t.Fatalf("mul %s %s", leftText, rightText)
				}
			}
			if !right.IsZero() {
				q, r, ok := left.DivMod(right)
				if !ok {
					t.Fatalf("div %s %s", leftText, rightText)
				}
				wantQ := new(big.Int).Quo(leftBig, rightBig)
				wantR := new(big.Int).Rem(leftBig, rightBig)
				if q.BigInt().Cmp(wantQ) != 0 || r.BigInt().Cmp(wantR) != 0 {
					t.Fatalf("divmod %s %s", leftText, rightText)
				}
			}
		}
	}
}

func TestUint128DivisionRandomMatchesBigInt(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 10000; i++ {
		left := Uint128{High: rng.Uint64(), Low: rng.Uint64()}
		right := Uint128{High: rng.Uint64(), Low: rng.Uint64()}
		if right.IsZero() {
			right.Low = 1
		}
		quotient, remainder, ok := left.DivMod(right)
		if !ok {
			t.Fatal("unexpected division failure")
		}
		wantQuotient := new(big.Int).Quo(left.BigInt(), right.BigInt())
		wantRemainder := new(big.Int).Rem(left.BigInt(), right.BigInt())
		if quotient.BigInt().Cmp(wantQuotient) != 0 || remainder.BigInt().Cmp(wantRemainder) != 0 {
			t.Fatalf("divmod mismatch: %#v / %#v", left, right)
		}
	}
}

func TestWideIntegerCheckedArithmeticRandomMatchesBigInt(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	for i := 0; i < 5000; i++ {
		left := Uint128{High: rng.Uint64(), Low: rng.Uint64()}
		right := Uint128{High: rng.Uint64(), Low: rng.Uint64()}
		leftBig := left.BigInt()
		rightBig := right.BigInt()

		wantAdd := new(big.Int).Add(leftBig, rightBig)
		add, addOK := left.AddChecked(right)
		wantAddOK := wantAdd.BitLen() <= 128
		if addOK != wantAddOK || (addOK && add.BigInt().Cmp(wantAdd) != 0) {
			t.Fatalf("unsigned add mismatch: %#v + %#v", left, right)
		}

		wantSub := new(big.Int).Sub(leftBig, rightBig)
		sub, subOK := left.SubChecked(right)
		wantSubOK := wantSub.Sign() >= 0
		if subOK != wantSubOK || (subOK && sub.BigInt().Cmp(wantSub) != 0) {
			t.Fatalf("unsigned sub mismatch: %#v - %#v", left, right)
		}

		wantMul := new(big.Int).Mul(leftBig, rightBig)
		mul, mulOK := left.MulChecked(right)
		wantMulOK := wantMul.BitLen() <= 128
		if mulOK != wantMulOK || (mulOK && mul.BigInt().Cmp(wantMul) != 0) {
			t.Fatalf("unsigned mul mismatch: %#v * %#v", left, right)
		}

		signedLeft := Int128FromBits(left)
		signedRight := Int128FromBits(right)
		checkSignedArithmetic(t, signedLeft, signedRight)
	}
}

func checkSignedArithmetic(t *testing.T, left Int128, right Int128) {
	t.Helper()
	leftBig := left.BigInt()
	rightBig := right.BigInt()

	wantAdd := new(big.Int).Add(leftBig, rightBig)
	wantAddCarrier, wantAddOK := Int128FromBig(wantAdd)
	add, addOK := left.AddChecked(right)
	if addOK != wantAddOK || (addOK && add != wantAddCarrier) {
		t.Fatalf("signed add mismatch: %s + %s", leftBig, rightBig)
	}

	wantSub := new(big.Int).Sub(leftBig, rightBig)
	wantSubCarrier, wantSubOK := Int128FromBig(wantSub)
	sub, subOK := left.SubChecked(right)
	if subOK != wantSubOK || (subOK && sub != wantSubCarrier) {
		t.Fatalf("signed sub mismatch: %s - %s", leftBig, rightBig)
	}

	wantMul := new(big.Int).Mul(leftBig, rightBig)
	wantMulCarrier, wantMulOK := Int128FromBig(wantMul)
	mul, mulOK := left.MulChecked(right)
	if mulOK != wantMulOK || (mulOK && mul != wantMulCarrier) {
		t.Fatalf("signed mul mismatch: %s * %s", leftBig, rightBig)
	}
}

func TestInt128EuclideanDivisionMatchesBigInt(t *testing.T) {
	values := []string{"-170141183460469231731687303715884105728", "-18446744073709551617", "-7", "-1", "0", "1", "3", "18446744073709551617", "170141183460469231731687303715884105727"}
	for _, leftText := range values {
		leftBig, _ := new(big.Int).SetString(leftText, 10)
		left, _ := Int128FromBig(leftBig)
		for _, rightText := range values {
			rightBig, _ := new(big.Int).SetString(rightText, 10)
			if rightBig.Sign() == 0 {
				continue
			}
			right, _ := Int128FromBig(rightBig)
			q, r, nonzero, inRange := left.DivMod(right)
			wantQ := new(big.Int).Quo(leftBig, rightBig)
			wantR := new(big.Int).Rem(leftBig, rightBig)
			if wantR.Sign() < 0 {
				if rightBig.Sign() > 0 {
					wantQ.Sub(wantQ, big.NewInt(1))
					wantR.Add(wantR, rightBig)
				} else {
					wantQ.Add(wantQ, big.NewInt(1))
					wantR.Sub(wantR, rightBig)
				}
			}
			wantQ128, wantInRange := Int128FromBig(wantQ)
			if !nonzero || inRange != wantInRange {
				t.Fatalf("status %s %s", leftText, rightText)
			}
			if inRange && (q != wantQ128 || r.BigInt().Cmp(wantR) != 0) {
				t.Fatalf("divmod %s %s", leftText, rightText)
			}
		}
	}
}

func TestWideIntegerRuntimeValueBoundaries(t *testing.T) {
	signedWant := Int128{High: ^uint64(0), Low: ^uint64(4)}
	signedValue := signedWant.IntegerValue()
	if got, ok := Int128FromValue(&signedValue); !ok || got != signedWant {
		t.Fatalf("pointer i128 round trip = %#v, %v", got, ok)
	}

	unsignedWant := Uint128{High: 17, Low: 23}
	wrapped := InterfaceValue{Underlying: unsignedWant.IntegerValue()}
	if got, ok := Uint128FromValue(wrapped); !ok || got != unsignedWant {
		t.Fatalf("interface u128 round trip = %#v, %v", got, ok)
	}
}
