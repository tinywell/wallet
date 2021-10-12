package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/json"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
	"github.com/tyler-smith/go-bip39/wordlists"
)

// Secret ..
type Secret struct {
	Key      []byte
	Mnemonic string
}

// Wallet .
type Wallet struct {
	KeyStore
	private  *ecdsa.PrivateKey
	mnemonic string
	addr     string
	name     string
}

// Sign 私钥签名
func (w *Wallet) Sign(data []byte) ([]byte, error) {
	return nil, nil
}

// Verify 签名验证
func (w *Wallet) Verify(data []byte) bool {
	return false
}

func (w *Wallet) store() error {
	priRaw, err := x509.MarshalECPrivateKey(w.private)
	if err != nil {
		return err
	}
	sec := Secret{
		Key:      priRaw,
		Mnemonic: w.mnemonic,
	}
	secRaw, err := json.Marshal(sec)
	if err != nil {
		return err
	}
	return w.Store(w.name, secRaw)
}

// ShowMnemonic 展示助记词
func (w *Wallet) ShowMnemonic() string {
	return w.mnemonic
}
func (w *Wallet) Address() string {
	return w.addr
}

// CreateWallet ..
func CreateWallet(keystore KeyStore, name string) (*Wallet, error) {
	w := &Wallet{
		KeyStore: keystore,
		name:     name,
	}
	mnemonic := genMnemonic()
	w.mnemonic = mnemonic
	pri, err := genKey(mnemonic)
	if err != nil {
		return nil, err
	}
	w.private = pri
	err = w.store()
	if err != nil {
		return nil, err
	}
	w.addr = genAddr(w.private)
	return w, nil
}

// LoadWallet ..
func LoadWallet(keystore KeyStore, name string) (*Wallet, error) {
	secRaw, err := keystore.Load(name)
	if err != nil {
		return nil, err
	}
	sec := &Secret{}
	err = json.Unmarshal(secRaw, sec)
	if err != nil {
		return nil, err
	}
	pri, err := x509.ParseECPrivateKey(sec.Key)
	if err != nil {
		return nil, err
	}
	w := &Wallet{
		KeyStore: keystore,
		private:  pri,
		mnemonic: sec.Mnemonic,
		addr:     genAddr(pri),
		name:     name,
	}
	return w, nil
}

func genAddr(pri *ecdsa.PrivateKey) string {
	return crypto.PubkeyToAddress(pri.PublicKey).String()
}

func genMnemonic() string {
	bip39.SetWordList(wordlists.ChineseSimplified)
	entropy, _ := bip39.NewEntropy(128)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	return mnemonic
}

func genKey(mnemonic string) (*ecdsa.PrivateKey, error) {
	seed := bip39.NewSeed(mnemonic, "")
	buf := bytes.NewBuffer(seed)
	pri, err := ecdsa.GenerateKey(elliptic.P256(), buf) // secp256r1
	// pri, err := ecdsa.GenerateKey(crypto.S256(), buf) // secp256k1 golang 不支持
	if err != nil {
		return nil, err
	}
	return pri, nil
}
