package sdk

import (
	"math"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/crypto"
	utils "github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"bewallet/pkg/fab/comm"
)

// Signer 签名
type Signer interface {
	Serialize() ([]byte, error)
	Sign(object []byte) ([]byte, error)
}

type commonClient struct {
	*comm.ClientConfig
	address string
	sn      string
}

var (
	defaultConnTimeout = 10 * time.Second // 本机测试时，低于 10s 连接会失败，原因未明
)

// CreateGRPCClient fabric grpc 连接客户端
func CreateGRPCClient(certs [][]byte) (*comm.ClientConfig, error) {
	config := comm.ClientConfig{}
	config.DialTimeout = defaultConnTimeout
	config.SecOpts = comm.SecureOptions{
		UseTLS:            false,
		RequireClientCert: false,
		ServerRootCAs:     certs,
	}
	if len(certs) > 0 {
		config.SecOpts.UseTLS = true
	}

	return &config, nil
}

// GetGRPCConn 建立 grpc 连接
func GetGRPCConn(addr string, cert []byte, serverNameOverride string) (*grpc.ClientConn, error) {
	grpcClient, err := CreateGRPCClient([][]byte{cert})
	if err != nil {
		return nil, err
	}
	grpcClient.SecOpts.ServerNameOverride = serverNameOverride
	return grpcClient.Dial(addr)
}

// CreateProposal 构建 proposal
func CreateProposal(signer Signer, channel, ccname, version, cctype string, transientMap map[string][]byte, args ...[]byte) (*peer.Proposal, string, error) {
	spec := &peer.ChaincodeSpec{
		Type:        (peer.ChaincodeSpec_Type)(peer.ChaincodeSpec_Type_value[strings.ToUpper(cctype)]),
		ChaincodeId: &peer.ChaincodeID{Name: ccname, Version: version},
		Input:       &peer.ChaincodeInput{Args: args},
	}
	cis := &peer.ChaincodeInvocationSpec{ChaincodeSpec: spec}
	creator, err := signer.Serialize()
	if err != nil {
		return nil, "", errors.Wrap(err, "get signer serialize error")
	}
	prop, txid, err := utils.CreateChaincodeProposalWithTransient(common.HeaderType_ENDORSER_TRANSACTION, channel, cis, creator, transientMap)
	if err != nil {
		return nil, "", errors.Wrap(err, "create chaincode proposal error")
	}
	return prop, txid, nil
}

// SignProposal proposal 签名
func SignProposal(signer Signer, proposal *peer.Proposal) (*peer.SignedProposal, error) {
	propBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "marshal proposal error")
	}

	sig, err := signer.Sign(propBytes)
	if err != nil {
		return nil, errors.Wrap(err, "sign proposal error")
	}

	return &peer.SignedProposal{ProposalBytes: propBytes, Signature: sig}, nil
}

// CreateSignedProposal 构建签名 proposal
func CreateSignedProposal(signer Signer, channel, ccname, version, cctype string, transientMap map[string][]byte, args ...[]byte) (*peer.SignedProposal, error) {
	prop, _, err := CreateProposal(signer, channel, ccname, version, cctype, transientMap, args...)
	if err != nil {
		return nil, err
	}
	return SignProposal(signer, prop)
}

// CreateEnvelope assembles an Envelope message from proposal, endorsements,
// and a signer. This function should be called by a client when it has
// collected enough endorsements for a proposal to create a transaction and
// submit it to peers for ordering
func CreateEnvelope(
	proposal *peer.Proposal,
	signer Signer,
	resps ...*peer.ProposalResponse,
) (*common.Envelope, error) {
	return utils.CreateSignedTx(proposal, signer, resps...)
}

// CreateTxSeekInfo 最新交易区块
func CreateTxSeekInfo() *orderer.SeekInfo {
	start := &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Newest{
			Newest: &orderer.SeekNewest{},
		},
	}

	stop := &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Specified{
			Specified: &orderer.SeekSpecified{
				Number: math.MaxUint64,
			},
		},
	}
	return seekHelp(start, stop)
}

// CreateSpecifiedSeekInfo 指定区块
func CreateSpecifiedSeekInfo(blockNumber uint64) *orderer.SeekInfo {
	seekPosition := &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Specified{
			Specified: &orderer.SeekSpecified{
				Number: blockNumber,
			},
		},
	}
	return seekHelp(seekPosition, seekPosition)
}

// CreateNewestSeekInfo 获取最新快
func CreateNewestSeekInfo() *orderer.SeekInfo {
	newest := &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Newest{
			Newest: &orderer.SeekNewest{},
		},
	}

	return seekHelp(newest, newest)
}

func seekHelp(start, stop *orderer.SeekPosition) *orderer.SeekInfo {
	return &orderer.SeekInfo{
		Start:    start,
		Stop:     stop,
		Behavior: orderer.SeekInfo_BLOCK_UNTIL_READY,
	}
}

// CreateTxSeekEnvelope 构建 deliver 交易信封
func CreateTxSeekEnvelope(signer Signer, channel string) (*common.Envelope, error) {
	seekInfo := CreateTxSeekInfo()
	return CreateSeekEnvelope(signer, channel, seekInfo)
}

// CreateSeekEnvelope 根据偏移信息构建 deliver 交易信封
func CreateSeekEnvelope(signer Signer, channel string, seekInfo *orderer.SeekInfo) (*common.Envelope, error) {
	payloadChannelHeader := utils.MakeChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, int32(0), channel, uint64(0))
	var err error
	payloadSignatureHeader := &common.SignatureHeader{}

	if signer != nil {
		payloadSignatureHeader, err = GetSignatureHeader(signer)
		if err != nil {
			return nil, errors.Wrap(err, "get signature header error")
		}
	}

	data, err := proto.Marshal(seekInfo)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling")
	}

	paylBytes := utils.MarshalOrPanic(
		&common.Payload{
			Header: utils.MakePayloadHeader(payloadChannelHeader, payloadSignatureHeader),
			Data:   data,
		},
	)

	var sig []byte
	if signer != nil {
		sig, err = signer.Sign(paylBytes)
		if err != nil {
			return nil, errors.Wrap(err, "sign deliver data error")
		}
	}

	env := &common.Envelope{
		Payload:   paylBytes,
		Signature: sig,
	}

	return env, nil
}

// GetSignatureHeader 构建 SignatureHeader
func GetSignatureHeader(signer Signer) (*common.SignatureHeader, error) {
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, err
	}
	seri, err := signer.Serialize()
	if err != nil {
		return nil, err
	}
	return &common.SignatureHeader{
		Creator: seri,
		Nonce:   nonce,
	}, nil
}
