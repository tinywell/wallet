package sdk

import (
	"context"
	"io"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

// PeerClient 调用 peer 服务接口的客户端
type PeerClient struct {
	commonClient
	peerDeliverClient     peer.DeliverClient
	peerEndorserClient    peer.EndorserClient
	deliverClient         peer.Deliver_DeliverClient
	deliverFilteredClient peer.Deliver_DeliverFilteredClient
}

// NewPeerClient 生成新的 PeerClient 实例
func NewPeerClient(addr, override string, tlsCaCert []byte) (*PeerClient, error) {
	grpcClient, err := CreateGRPCClient([][]byte{tlsCaCert})
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client error")
	}
	return &PeerClient{
		commonClient: commonClient{
			ClientConfig: grpcClient,
			address:      addr,
			sn:           override,
		},
	}, nil
}

// Addr 返回实例访问地址
func (p *PeerClient) Addr() string {
	return p.address
}

// Endorser 生成 EndorserClient 实例
func (p *PeerClient) Endorser() (peer.EndorserClient, error) {
	if p.peerEndorserClient != nil {
		return p.peerEndorserClient, nil
	}
	p.SecOpts.ServerNameOverride = p.sn
	conn, err := p.Dial(p.address)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc connection error")
	}
	p.peerEndorserClient = peer.NewEndorserClient(conn)
	return p.peerEndorserClient, nil
}

// PeerDeliver 生成 DeliverClient 实例
func (p *PeerClient) PeerDeliver() (peer.DeliverClient, error) {
	if p.peerDeliverClient != nil {
		return p.peerDeliverClient, nil
	}
	p.SecOpts.ServerNameOverride = p.sn
	conn, err := p.Dial(p.address)
	if err != nil {
		return nil, errors.Wrapf(err, "deliver client failed to connect to %s", p.address)
	}
	p.peerDeliverClient = peer.NewDeliverClient(conn)
	return p.peerDeliverClient, nil
}

// Deliver 生成 Deliver_DeliverClient 实例
func (p *PeerClient) Deliver() (peer.Deliver_DeliverClient, error) {
	if p.deliverClient != nil {
		return p.deliverClient, nil
	}
	dc, err := p.PeerDeliver()
	if err != nil {
		return nil, err
	}
	p.deliverClient, err = dc.Deliver(context.TODO())
	if err != nil {
		return nil, errors.Wrap(err, "create deliver client error")
	}
	return p.deliverClient, nil
}

// DeliverFilter 生成 Deliver_DeliverFilteredClient 实例
func (p *PeerClient) DeliverFilter() (peer.Deliver_DeliverFilteredClient, error) {
	if p.deliverFilteredClient != nil {
		return p.deliverFilteredClient, nil
	}
	dc, err := p.PeerDeliver()
	if err != nil {
		return nil, err
	}
	p.deliverFilteredClient, err = dc.DeliverFiltered(context.TODO())
	if err != nil {
		return nil, errors.Wrap(err, "create deliverfiltered client error")
	}
	return p.deliverFilteredClient, nil
}

// SendProposal 发送交易 proposal 到 peer
func (p *PeerClient) SendProposal(ctx context.Context, signedProposal *peer.SignedProposal) (*peer.ProposalResponse, error) {
	ec, err := p.Endorser()
	if err != nil {
		return nil, errors.Wrap(err, "get peer deliver client error")
	}
	resp, err := ec.ProcessProposal(ctx, signedProposal)
	if err != nil {
		return nil, errors.Wrap(err, "process proposal error")
	}
	if resp.Response.Status < 200 || resp.Response.Status > 400 {
		return nil, errors.Errorf("process proposal return invalid status = %d", resp.Response.Status)
	}
	return resp, nil
}

// DeliverBlock 从 peer 接收 block
func (p *PeerClient) DeliverBlock(ctx context.Context, seekEnv *common.Envelope) (<-chan *peer.DeliverResponse, <-chan error) {
	respChan := make(chan *peer.DeliverResponse)
	errChan := make(chan error)
	go func() {
		defer close(respChan)
		defer close(errChan)
		dc, err := p.Deliver()
		if err != nil {
			errChan <- errors.Wrap(err, "get deliver client error")
			return
		}
		err = dc.Send(seekEnv)
		if err != nil {
			errChan <- errors.Wrap(err, "send deliver request error")
			return
		}
		for {
			select {
			case <-ctx.Done():
				errChan <- errors.Wrapf(err, "context done ")
				return
			default:
				resp, err := dc.Recv()
				if err != io.EOF {
					if err != nil {
						errChan <- errors.Wrap(err, "receive delvier response error")
						return
					}
				} else {
					errChan <- err
					return
				}
				switch t := resp.Type.(type) {
				case *peer.DeliverResponse_Status:
					errChan <- errors.Wrapf(err, "receive delvier response with unexpected status=[%d]%s", t.Status, t.Status.String())
					return
				case *peer.DeliverResponse_Block:
					respChan <- resp
				default:
					errChan <- errors.Wrapf(err, "receive delvier response with unexcept type=%T", t)
					return
				}
			}
		}
	}()

	return respChan, errChan
}

// DeliverFilteredBlock 从 peer 接收 FilteredBlock
func (p *PeerClient) DeliverFilteredBlock(ctx context.Context, seekEnv *common.Envelope) (<-chan *peer.DeliverResponse, <-chan error) {
	respChan := make(chan *peer.DeliverResponse)
	errChan := make(chan error)
	go func() {
		defer close(respChan)
		defer close(errChan)
		dc, err := p.DeliverFilter()
		if err != nil {
			errChan <- errors.Wrap(err, "get deliver filtered client error")
			return
		}
		err = dc.Send(seekEnv)
		if err != nil {
			errChan <- errors.Wrap(err, "send deliver request error")
			return
		}
		for {
			select {
			case <-ctx.Done():
				errChan <- errors.Wrapf(err, "context done ")
				return
			default:
				resp, err := dc.Recv()
				if err != io.EOF {
					if err != nil {
						errChan <- errors.Wrap(err, "receive delvier response error")
						return
					}
				} else {
					errChan <- err
					return
				}
				switch t := resp.Type.(type) {
				case *peer.DeliverResponse_Status:
					errChan <- errors.Wrapf(err, "receive delvier response with unexpected status=[%d]%s", t.Status, t.Status.String())
					return
				case *peer.DeliverResponse_FilteredBlock:
					respChan <- resp
				default:
					errChan <- errors.Wrapf(err, "receive delvier response with unexcept type=%T", t)
					return
				}
			}
		}
	}()

	return respChan, errChan
}
