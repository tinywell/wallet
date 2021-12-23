package nft

import "github.com/hyperledger/fabric-protos-go/peer"

// Node fabric 节点信息（peer、orderer）
type Node struct {
	URL          string
	TLSCert      string
	OverrideName string
}

type fabProposal struct {
	txid       string
	prop       *peer.Proposal
	signedProp *peer.SignedProposal
}
