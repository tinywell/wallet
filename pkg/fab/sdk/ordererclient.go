package sdk

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/pkg/errors"
)

// OrdererClient 调用 orderer 服务接口的客户端
type OrdererClient struct {
	commonClient
	client          orderer.AtomicBroadcastClient
	broadCastClient orderer.AtomicBroadcast_BroadcastClient
	deliverClient   orderer.AtomicBroadcast_DeliverClient
}

// NewOrdererClient 生成新的 OrdererClient 实例
func NewOrdererClient(addr, override string, tlsCaCert []byte) (*OrdererClient, error) {
	grpcClient, err := CreateGRPCClient([][]byte{tlsCaCert})
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client error")
	}
	return &OrdererClient{
		commonClient: commonClient{
			ClientConfig: grpcClient,
			address:      addr,
			sn:           override,
		},
	}, nil
}

// AtomicBroadCast 生成 AtomicBroadcastClient 实例
func (o *OrdererClient) AtomicBroadCast() (orderer.AtomicBroadcastClient, error) {
	if o.client != nil {
		return o.client, nil
	}
	o.SecOpts.ServerNameOverride = o.sn
	conn, err := o.Dial(o.address)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc connection error")
	}
	o.client = orderer.NewAtomicBroadcastClient(conn)
	return o.client, nil
}

// BroadCast 生成 AtomicBroadcast_BroadcastClient 实例
func (o *OrdererClient) BroadCast() (orderer.AtomicBroadcast_BroadcastClient, error) {
	if o.broadCastClient != nil {
		return o.broadCastClient, nil
	}
	client, err := o.AtomicBroadCast()
	if err != nil {
		return nil, err
	}
	bc, err := client.Broadcast(context.TODO())
	if err != nil {
		return nil, errors.Wrap(err, "get broadcast client error")
	}
	o.broadCastClient = bc
	return o.broadCastClient, nil
}

// Deliver 生成 AtomicBroadcast_DeliverClient 实例
func (o *OrdererClient) Deliver() (orderer.AtomicBroadcast_DeliverClient, error) {
	if o.deliverClient != nil {
		return o.deliverClient, nil
	}
	client, err := o.AtomicBroadCast()
	if err != nil {
		return nil, err
	}
	bc, err := client.Deliver(context.TODO())
	if err != nil {
		return nil, errors.Wrap(err, "get broadcast client error")
	}
	o.deliverClient = bc
	return o.deliverClient, nil
}

// SendBroadCast 发送 Broadcast 交易信封到 orderer
func (o *OrdererClient) SendBroadCast(ctx context.Context, env *common.Envelope) error {
	bc, err := o.BroadCast()
	if err != nil {
		return err
	}
	err = bc.Send(env)
	if err != nil {
		return errors.Wrap(err, "send broadcast error")
	}
	resp, err := bc.Recv()
	if err != nil {
		return errors.Wrap(err, "receive broadcast response error")
	}
	if resp.Status != common.Status_SUCCESS {
		return errors.Errorf("receive broadcast ressponse with invalid status = %d:%s",
			resp.Status, resp.Status.String())
	}
	return nil
}

// SendDeliver 发送 deliver 请求信封到 orderer
func (o *OrdererClient) SendDeliver(ctx context.Context, seekEnv *common.Envelope) (*common.Block, error) {
	dc, err := o.Deliver()
	if err != nil {
		return nil, err
	}
	err = dc.Send(seekEnv)
	if err != nil {
		return nil, errors.Wrap(err, "send deliver envelope error")
	}
	resp, err := dc.Recv()
	if err != nil {
		return nil, errors.Wrap(err, "receive deliver response error")
	}
	switch t := resp.Type.(type) {
	case *orderer.DeliverResponse_Status:
		return nil, errors.Errorf("get block with status = %d:%s", t.Status, t.Status.String())
	case *orderer.DeliverResponse_Block:
		dc.Recv() // Flush the success message
		return t.Block, nil
	default:
		return nil, errors.Errorf("response error: unknown type %T", t)
	}
}
