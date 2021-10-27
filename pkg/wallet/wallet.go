package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
	"github.com/tyler-smith/go-bip39/wordlists"

	"bewallet/pkg/keystore"
	"bewallet/pkg/utils"
)

var (
	// Curve 椭圆曲线
	Curve = elliptic.P256()
)

// Secret ..
type Secret struct {
	Key      []byte
	Mnemonic string
}

// Wallet .
type Wallet struct {
	keystore.KeyStore
	private  *ecdsa.PrivateKey
	mnemonic string
	addr     string
	name     string
}

// Sign 私钥签名 (fabric 签名)
func (w *Wallet) Sign(data []byte) ([]byte, error) {
	ri, si, err := ecdsa.Sign(rand.Reader, w.private, digest(data))
	if err != nil {
		return nil, err
	}

	si, _, err = utils.ToLowS(&w.private.PublicKey, si)
	if err != nil {
		return nil, err
	}
	return utils.MarshalECDSASignature(ri, si)
}

// Verify 签名验证
func (w *Wallet) Verify(sig []byte, data []byte) (bool, error) {

	r, s, err := utils.UnmarshalECDSASignature(sig)
	if err != nil {
		return false, fmt.Errorf("Failed unmashalling signature [%s]", err)
	}

	lowS, err := utils.IsLowS(&w.private.PublicKey, s)
	if err != nil {
		return false, err
	}

	if !lowS {
		return false, fmt.Errorf("Invalid S. Must be smaller than half the order [%s][%s]", s, utils.GetCurveHalfOrdersAt(w.private.PublicKey.Curve))
	}

	return ecdsa.Verify(&w.private.PublicKey, digest(data), r, s), nil
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
	opt := &keystore.SecretStoreOpt{
		Content: secRaw,
		Name:    w.name,
	}
	return w.Store(opt)
}

// ShowMnemonic 展示助记词
func (w *Wallet) ShowMnemonic() string {
	return w.mnemonic
}

// Address 地址
func (w *Wallet) Address() string {
	return w.addr
}

// PublicKey 返回钱包公钥
func (w *Wallet) PublicKey() *ecdsa.PublicKey {
	return &w.private.PublicKey
}

func (w *Wallet) initByMnemonic(mnemonic string) error {
	w.mnemonic = mnemonic
	pri, err := genKey(mnemonic)
	if err != nil {
		return err
	}
	w.private = pri
	w.addr = genAddr(w.private)
	if len(w.name) == 0 {
		w.name = w.addr
	}
	err = w.store()
	if err != nil {
		return err
	}
	return nil
}

// CreateWallet ..
func CreateWallet(keystore keystore.KeyStore, name string) (*Wallet, error) {
	w := &Wallet{
		KeyStore: keystore,
		name:     name,
	}
	mnemonic := genMnemonic()
	err := w.initByMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}
	return w, nil
}

// LoadWallet ..
func LoadWallet(ks keystore.KeyStore, name string) (*Wallet, error) {
	opt := &keystore.SecretLoadOpt{
		Name: name,
	}
	secRaw, err := ks.Load(opt)
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
		KeyStore: ks,
		private:  pri,
		mnemonic: sec.Mnemonic,
		addr:     genAddr(pri),
		name:     name,
	}

	return w, nil
}

// RecoverWallet ...
func RecoverWallet(ks keystore.KeyStore, name string, mnemonic string) (*Wallet, error) {
	w := &Wallet{
		KeyStore: ks,
		name:     name,
	}
	err := w.initByMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func genAddr(pri *ecdsa.PrivateKey) string {
	// pubBytes := crypto.FromECDSAPub(&pri.PublicKey)
	// fmt.Println(len(pubBytes))
	// fmt.Printf("0x%2x\n", crypto.Keccak256(pubBytes[1:]))

	return publicToAddress(&pri.PublicKey) // 与以太坊的 crypto.PubkeyToAddress() 等同
	// return crypto.PubkeyToAddress(pri.PublicKey).String()
}

func publicToAddress(pub *ecdsa.PublicKey) string {
	pubBytes := fromECDSAPub(pub)
	addr := common.BytesToAddress(crypto.Keccak256(pubBytes[1:])[12:])
	return addr.String()
}

func fromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(pub.Curve, pub.X, pub.Y)
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
	pri, err := ecdsa.GenerateKey(Curve, buf) // secp256r1
	// pri, err := ecdsa.GenerateKey(crypto.S256(), buf) // secp256k1 golang 不支持
	if err != nil {
		return nil, err
	}
	return pri, nil
}

func digest(in []byte) []byte {
	h := sha256.New()
	h.Write(in)
	return h.Sum(nil)
}
