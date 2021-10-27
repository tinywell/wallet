package keystore

// StoreOpts ..
type StoreOpts interface {
	Data() []byte
	StoreType() string
	Identity() string
}

// LoadOpts ..
type LoadOpts interface {
	Identity() string
	LoadType() string
}

// KeyType
const (
	KeyTypeSecret  = "secret"
	KeyTypeNetwork = "network"
)

// SecretStoreOpt 密钥存储
type SecretStoreOpt struct {
	Content []byte
	Name    string
}

// Data 存储内容
func (so *SecretStoreOpt) Data() []byte {
	return so.Content
}

// StoreType 存储数据类别
func (so *SecretStoreOpt) StoreType() string {
	return KeyTypeSecret
}

// Identity 存储数据标识
func (so *SecretStoreOpt) Identity() string {
	return so.Name
}

// NetworkStoreOpt 网络信息存储
type NetworkStoreOpt struct {
	Content []byte
	Name    string
}

// Data 存储内容
func (no *NetworkStoreOpt) Data() []byte {
	return no.Content
}

// StoreType 存储数据类别
func (no *NetworkStoreOpt) StoreType() string {
	return KeyTypeNetwork
}

// Identity 存储数据标识
func (no *NetworkStoreOpt) Identity() string {
	return no.Name
}

type SecretLoadOpt struct {
	Name string
}

func (sl *SecretLoadOpt) Identity() string {
	return sl.Name
}

func (sl *SecretLoadOpt) LoadType() string {
	return KeyTypeSecret
}

type NetworkLoadOpt struct {
	Name string
}

func (nl *NetworkLoadOpt) Identity() string {
	return nl.Name
}

func (nl *NetworkLoadOpt) LoadType() string {
	return KeyTypeNetwork
}
