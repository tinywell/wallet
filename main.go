package main

import (
	"bewallet/pkg/fab/sdk"
	"bewallet/pkg/wallet"
)

var _ sdk.Signer = &wallet.FabWallet{}

func main() {
	// if err := cmd.WalletCMD.Execute(); err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(0)
	// }

}
