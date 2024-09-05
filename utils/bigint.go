package utils

import "math/big"

// MaxBigInt returns the largest big integer given two values
func MaxBigInt(a, b *big.Int) *big.Int {
	if a.Cmp(b) > 0 {
		return new(big.Int).Set(a)
	}
	return new(big.Int).Set(b)
}

// MinBigInt returns the smalllest big integer given two values
func MinBigInt(a, b *big.Int) *big.Int {
	if a.Cmp(b) < 0 {
		return new(big.Int).Set(a)
	}
	return new(big.Int).Set(b)
}
