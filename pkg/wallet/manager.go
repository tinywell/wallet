package wallet

import (
	"bewallet/pkg/fab/sdk"
	"bewallet/pkg/keystore"

	"github.com/pkg/errors"
)

// Manager 钱包管理
type Manager struct {
	wallets  map[string]*Wallet
	networks map[string]map[string]*FabNet
	ks       keystore.KeyStore
}

// NewManager ..
func NewManager(ks keystore.KeyStore) (*Manager, error) {
	m := &Manager{
		ks:       ks,
		wallets:  make(map[string]*Wallet, 0),
		networks: make(map[string]map[string]*FabNet),
	}
	err := m.loadWallet()
	if err != nil {
		return nil, err
	}
	return m, nil
}

// AccountList ...
func (m *Manager) AccountList() map[string]string {
	list := make(map[string]string)
	for _, w := range m.wallets {
		list[w.addr] = w.name // 用户名作为可选项
	}
	return list
}

// GetWallet ..
func (m *Manager) GetWallet(addr string) *ModelWallet {
	if w, ok := m.wallets[addr]; ok {
		return &ModelWallet{
			Addr: w.addr,
			Name: w.name,
		}
	}
	return nil
}

// GetNetworks ..
func (m *Manager) GetNetworks(addr string) map[string]*FabNet {
	if n, ok := m.networks[addr]; ok {
		return n
	}
	return nil
}

// GetSigner ..
func (m *Manager) GetSigner(addr, net string) sdk.Signer {
	w, ok := m.wallets[addr]
	if !ok {
		return nil
	}
	nets, ok := m.networks[addr]
	if !ok {
		return nil
	}
	fabnet, ok := nets[net]
	if !ok {
		return nil
	}
	return &FabWallet{
		Wallet: *w,
		FabNet: *fabnet,
	}
}

// LoadWallet 加载钱包
func (m *Manager) loadWallet() error {

	list, err := m.ks.List()
	if err != nil {
		return err
	}

	for _, n := range list {
		w, err := LoadWallet(m.ks, n)
		if err != nil {
			return errors.WithMessagef(err, "加载账户 %s 密钥失败", n)
		}
		m.wallets[w.addr] = w

		nets, err := LoadFabNet(m.ks, n)
		if err != nil {
			return errors.WithMessagef(err, "加载账户网络配置信息失败", n)
		}
		m.networks[n] = nets
	}
	return nil
}
