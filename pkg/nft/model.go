package nft

// Node fabric 节点信息（peer、orderer）
type Node struct {
	URL          string
	TLSCert      string
	OverrideName string
}
