package main

import (
	"bewallet/pkg/wallet"
	"fmt"
)

func main() {
	ks := wallet.NewFilKeyStore("./wallet", "wallet123")

	// w, err := wallet.CreateWallet(ks, "tinywell")
	w, err := wallet.LoadWallet(ks, "tinywell")
	if err != nil {
		panic(err)
	}

	fmt.Println(w.ShowMnemonic())
	fmt.Println(w.Address())
}
