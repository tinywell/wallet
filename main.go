package main

import (
	"bewallet/cmd"
	"fmt"
	"os"
)

func main() {
	if err := cmd.WalletCMD.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}
