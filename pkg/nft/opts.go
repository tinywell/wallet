package nft

import (
	"bewallet/pkg/fab/sdk"
)

type option struct {
	peers     []Node
	certs     []string
	orderers  []Node
	channel   string
	chaincode string
	ccType    string
	ccVersion string
	signer    sdk.Signer
}

// Option 初始化参数
type Option func(opt *option)

// WithPeer peer 节点参数
func WithPeer(peer Node) Option {
	return func(opt *option) {
		opt.peers = append(opt.peers, peer)
	}
}

// WithOrderer orderer 节点参数
func WithOrderer(peer Node) Option {
	return func(opt *option) {
		opt.peers = append(opt.peers, peer)
	}
}

// WithSigner 签名钱包
func WithSigner(signer sdk.Signer) Option {
	return func(opt *option) {
		opt.signer = signer
	}
}

// WithContract 合约参数
func WithContract(channel, chaincode, cctype, ccversion string) Option {
	return func(opt *option) {
		opt.channel = channel
		opt.chaincode = chaincode
		opt.ccType = cctype
		opt.ccVersion = ccversion
	}
}
