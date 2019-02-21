package main

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"strconv"
)

func pow(targetBits int) *big.Int {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	fmt.Printf("%v\n", target)
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	for {
		data := "hello world " + strconv.Itoa(nonce)
		hash = sha256.Sum256([]byte(data))
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(target) == -1 {
			break
		} else {
			nonce++
		}
	}

	return &hashInt
}
