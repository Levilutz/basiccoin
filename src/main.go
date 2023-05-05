package main

import (
	"fmt"

	"github.com/levilutz/basiccoin/src/utils"
)

func main() {
	ecdsaKey := utils.Ecdsa256()
	_, pubX, pubY := utils.EcdsaToKeys(ecdsaKey)
	fmt.Printf("%x\n", utils.Dhash(pubX, pubY))
}
