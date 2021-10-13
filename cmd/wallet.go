package cmd

import (
	"bewallet/pkg/wallet"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// subcommand name
const (
	SubCMDCreate = "create"
)

const (
	defaultSubBaseDir = "fabric/wallet"
)

var (
	// WalletCMD .
	WalletCMD = cobra.Command{
		Use:   "wallet",
		Short: "fabric wallet",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				printHelp()
				return
			}
			fn := args[0]
			switch fn {
			case SubCMDCreate:
				create()
			}
		},
	}

	name     string
	password string
	mnemonic string
	basedir  string
)

func init() {
	WalletCMD.Flags().StringVarP(&name, "name", "n", "", "账户名称")
	WalletCMD.Flags().StringVarP(&password, "password", "p", "", "账户口令")
	WalletCMD.Flags().StringVarP(&mnemonic, "mnemonic", "m", "", "助记词（空格连接）")
	WalletCMD.Flags().StringVarP(&basedir, "basedir", "d", "", "账户缓存目录")
}

func create() error {
	if len(basedir) == 0 {
		userdir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		basedir = filepath.Join(userdir, defaultSubBaseDir)
		fmt.Println("密钥缓存目录:", basedir)
	}
	ks := wallet.NewFilKeyStore(basedir, password)
	w, err := wallet.CreateWallet(ks, name)
	if err != nil {
		return err
	}
	fmt.Println("钱包创建成功！")
	fmt.Println("  钱包助记词:", w.ShowMnemonic())
	fmt.Println("  钱包地址:", w.Address())
	return nil
}

func printHelp() {
	fmt.Println("wallet 是一个基于 fabric 体系的钱包客户端")
	fmt.Println("Usage:")
	fmt.Println("    wallet <command> [arguments]")
	fmt.Println()
	fmt.Println("The commands are:")
	fmt.Println("  create - 创建钱包")
	fmt.Println()
	fmt.Println("The arguments are:")
	fmt.Println("    -n  name       账户名称")
	fmt.Println("    -p  password   账户口令")
	fmt.Println("    -d  basedir    缓存目录")
	fmt.Println("    -m  mnemonic   助记词")
}

// ks := wallet.NewFilKeyStore("./wallet", "wallet123")

// 	// w, err := wallet.CreateWallet(ks, "tinywell323")
// 	w, err := wallet.LoadWallet(ks, "tinywell")
// 	// w, err := wallet.RecoverWallet(ks, "retiynwell", "造 促 桃 作 最 包 送 郑 综 于 张 讯")
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println(w.ShowMnemonic())
// 	fmt.Println(w.Address())

// 	sig, err := w.Sign([]byte("hello world"))
// 	if err != nil {
// 		panic(err)
// 	}

// 	res, err := w.Verify(sig, []byte("hello world"))
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("sig is", res)
