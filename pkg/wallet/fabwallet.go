package wallet

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/pkg/errors"

	"bewallet/pkg/keystore"
)

// FabMSP fabric 身份
type FabMSP struct {
	Network  string
	OrgMSP   string
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
func SaveFabNet(ks keystore.KeyStore, name string, nets map[string]FabNet) error {
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
func LoadFabNet(ks keystore.KeyStore, name string) (map[string]FabNet, error) {
	opt := &keystore.NetworkLoadOpt{
		Name: name,
	}
	data, err := ks.Load(opt)
	if err != nil {
		return nil, err
	}
	nets := make(map[string]FabNet)
	err = json.Unmarshal(data, &nets)
	if err != nil {
		return nil, err
	}
	return nets, nil
}
