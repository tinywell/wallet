package wallet

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/pkg/errors"

	"bewallet/pkg/keystore"
)

// FabMSP fabric 身份
type FabMSP struct {
	Network  string
	OrgMSP   string
	Org      string
	SignCert string
}

// Serialize 证书序列化
func (fm FabMSP) Serialize() ([]byte, error) {
	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   fm.OrgMSP,
		IdBytes: []byte(fm.SignCert),
	}
	identity, err := proto.Marshal(serializedIdentity)
	if err != nil {
		return nil, errors.Wrap(err, "marshal serializedIdentity failed")
	}
	return identity, nil
}

// FabNet 网络信息
type FabNet struct {
	FabMSP
	Network
}

// Network fabric 网络
type Network struct {
	Name     string  `json:"name,omitempty"`
	Peers    []*Node `json:"peers,omitempty"`
	Orderers []*Node `json:"orderers,omitempty"`
}

// Node fabric 节点
type Node struct {
	Address        string `json:"address,omitempty" `
	ServerOverride string `json:"server_override,omitempty"`
	TLSCA          string `json:"tlsca,omitempty" `
}

// FabWallet 连接 fabric 网络的钱包
type FabWallet struct {
	Wallet
	FabNet
}

// // SKI
// func (fw *FabWallet) SKI() []byte {

// 	raw := elliptic.Marshal(fw.private.Curve, fw.private.PublicKey.X, fw.private.PublicKey.Y)

// 	// Hash it
// 	hash := sha256.New()
// 	hash.Write(raw)
// 	return hash.Sum(nil)
// }

// CertRequest 基于网络信息生成证书请求
func (fw FabWallet) CertRequest() ([]byte, error) {
	template := &x509.CertificateRequest{}
	subject := pkix.Name{
		OrganizationalUnit: []string{"client"},
		Organization:       []string{fw.Org},
		CommonName:         fmt.Sprintf("%s@%s", fw.Address(), fw.Org),
	}
	template.Subject = subject
	return x509.CreateCertificateRequest(rand.Reader, template, fw.private)
}

// SetNetwork 加入新网络
func (fw *FabNet) SetNetwork(net Network) {
	fw.Network = net
}

// AddPeers ..
func (fw *FabNet) AddPeers(peers []*Node) {
	fw.Peers = append(fw.Peers, peers...)
}

// AddOrderers ..
func (fw *FabNet) AddOrderers(orderers []*Node) {
	fw.Peers = append(fw.Orderers, orderers...)
}

// SetSignCert  设置钱包认证证书
func (fw *FabNet) SetSignCert(orgmsp string, cert string) {
	fw.OrgMSP = orgmsp
	fw.SignCert = cert
}

// Serialize 。

// SaveFabNet 网络信息保存
func SaveFabNet(ks keystore.KeyStore, name string, nets map[string]*FabNet) error {
	data, err := json.Marshal(nets)
	if err != nil {
		return err
	}
	opt := &keystore.NetworkStoreOpt{
		Content: data,
		Name:    name,
	}
	return ks.Store(opt)
}

// LoadFabNet 加载网络信息
func LoadFabNet(ks keystore.KeyStore, name string) (map[string]*FabNet, error) {
	opt := &keystore.NetworkLoadOpt{
		Name: name,
	}
	data, err := ks.Load(opt)
	if err != nil {
		return nil, err
	}
	nets := make(map[string]*FabNet)
	err = json.Unmarshal(data, &nets)
	if err != nil {
		return nil, err
	}
	return nets, nil
}

func x509Template() *x509.Certificate {

	// generate a serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)

	// set expiry to around 10 years
	expiry := 3650 * 24 * time.Hour
	// backdate 5 min
	notBefore := time.Now().Add(-5 * time.Minute).UTC()

	//basic template to use
	x509 := &x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notBefore.Add(expiry).UTC(),
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature,
	}
	return x509
}

// Additional for X509 subject
func subjectTemplateAdditional(country, province, locality, orgUnit, streetAddress, postalCode string) pkix.Name {
	name := subjectTemplate()
	if len(country) >= 1 {
		name.Country = []string{country}
	}
	if len(province) >= 1 {
		name.Province = []string{province}
	}

	if len(locality) >= 1 {
		name.Locality = []string{locality}
	}
	if len(orgUnit) >= 1 {
		name.OrganizationalUnit = []string{orgUnit}
	}
	if len(streetAddress) >= 1 {
		name.StreetAddress = []string{streetAddress}
	}
	if len(postalCode) >= 1 {
		name.PostalCode = []string{postalCode}
	}
	return name
}

// default template for X509 subject
func subjectTemplate() pkix.Name {
	return pkix.Name{
		Country:  []string{"CN"},
		Locality: []string{"Beijing"},
		Province: []string{"Beijing"},
	}
}
