package nft

import (
	"bewallet/pkg/fab/sdk"
	"context"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

// Client ...
type Client struct {
	opt         *option
	peerClis    []*sdk.PeerClient
	ordererClis []*sdk.OrdererClient
}

// NewClient ...
func NewClient(opts ...Option) *Client {
	opt := &option{}
	for _, o := range opts {
		o(opt)
	}
	// TODO: 参数检查
	if opt.signer == nil {
		return nil //TODO: return error
	}
	c := &Client{
		opt: opt,
	}
	err := c.initClients() //TODO:
	if err != nil {
		return nil
	}
	return c
}

func (c *Client) initClients() error {
	for _, p := range c.opt.peers {
		pc, err := sdk.NewPeerClient(p.URL, p.OverrideName, []byte(p.TLSCert))
		if err != nil {
			return errors.WithMessagef(err, "创建 peer client 失败，peer=%s", p.URL)
		}
		c.peerClis = append(c.peerClis, pc)
	}
	for _, o := range c.opt.orderers {
		oc, err := sdk.NewOrdererClient(o.URL, o.OverrideName, []byte(o.TLSCert))
		if err != nil {
			return errors.WithMessagef(err, "创建 orderer client 失败，orderer=%s", o.URL)
		}
		c.ordererClis = append(c.ordererClis, oc)
	}
	return nil
}

func (c *Client) createProposal(args [][]byte) (*fabProposal, error) {
	proposal, txid, err := sdk.CreateProposal(c.opt.signer, c.opt.channel, c.opt.chaincode, c.opt.ccVersion, c.opt.ccType, nil, args...)
	if err != nil {
		return nil, errors.WithMessagef(err, "构造交易提案失败, txid=%s", txid)
	}
	signedPropoal, err := sdk.SignProposal(c.opt.signer, proposal)
	if err != nil {
		return nil, errors.WithMessagef(err, "提案签名失败, txid=%s", txid)
	}
	return &fabProposal{
		txid:       txid,
		prop:       proposal,
		signedProp: signedPropoal,
	}, nil
	// return proposal, signedPropoal, nil
}

func (c *Client) createEnvelope(proposal *peer.Proposal, resps ...*peer.ProposalResponse) (*common.Envelope, error) {
	return sdk.CreateEnvelope(proposal, c.opt.signer, resps...)
}

// Send ...
func (c *Client) sendProposal(ctx context.Context, proposal *peer.SignedProposal) ([]*peer.ProposalResponse, error) {

	resps := make([]*peer.ProposalResponse, 0)
	errs := make([]error, 0)
	for _, p := range c.peerClis {
		resp, err := p.SendProposal(ctx, proposal)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		resps = append(resps, resp)
	}
	if len(resps) == 0 {
		return nil, errors.Errorf("提交提案出错: %s", mutilError(errs).Error())
	}
	return resps, nil
}

// Broadcast ...
func (c *Client) broadcast(ctx context.Context, env *common.Envelope) error {
	errs := make([]error, 0)
	for _, o := range c.ordererClis {
		err := o.SendBroadCast(ctx, env)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		return nil
	}
	return mutilError(errs)
}

// Invoke 共识交易
func (c *Client) Invoke(args ...[]byte) (peer.TxValidationCode, error) {
	// 背书
	prop, err := c.createProposal(args)
	if err != nil {
		return -1, err
	}
	ctx := context.Background()
	resps, err := c.sendProposal(ctx, prop.signedProp)
	if err != nil {
		return -1, err
	}
	if len(resps) == 0 {
		return -1, errors.New("未预期异常，返回结果为空")
	}
	resp := resps[0].Response
	if resp.Status != 200 {
		return -1, errors.Errorf("查询出错：[状态码 %d] %s", resp.Status, resp.Message)
	}
	// 广播
	env, err := c.createEnvelope(prop.prop, resps...)
	if err != nil {
		return -1, errors.WithMessagef(err, "构造交易信封出错,txid=%s", prop.txid)
	}
	err = c.broadcast(ctx, env)
	if err != nil {
		return -1, errors.WithMessagef(err, "交易广播出错.txid=%s", prop.txid)
	}
	// TODO:监听
	txcli, err := NewTxEvent(c.opt.channel, prop.txid, c.peerClis)
	if err != nil {
		return -1, errors.WithMessage(err, "创建交易事件客户端失败")
	}
	tx, err := txcli.Listen()
	if err != nil {
		return -1, errors.WithMessage(err, "监听交易事件失败")
	}
	return tx.TxValidationCode, nil
}

// Query 查询交易
func (c *Client) Query(args ...[]byte) ([]byte, error) {
	prop, err := c.createProposal(args)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	resps, err := c.sendProposal(ctx, prop.signedProp)
	if err != nil {
		return nil, err
	}
	if len(resps) == 0 {
		return nil, errors.New("未预期异常，交易返回结果为空")
	}
	resp := resps[0].Response
	if resp.Status != 200 {
		return nil, errors.Errorf("查询出错交易 txid=%s：[状态码 %d] %s", prop.txid, resp.Status, resp.Message)
	}
	return resp.Payload, nil
}

func mutilError(errs []error) error {
	errmsg := "出现多个错误："
	for i, e := range errs {
		errmsg += fmt.Sprintf("[%d] %s;", i, e.Error())
	}
	return errors.New(errmsg)
}
