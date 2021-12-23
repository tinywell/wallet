package nft

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"

	"bewallet/pkg/fab/sdk"
)

var (
	defaultTimeout = 30 * time.Second
)

type EventCenter interface {
}

// TxEvent transaction event
type TxEvent struct {
	channel string
	txid    string
	clients []peer.Deliver_DeliverFilteredClient
	receive chan *peer.FilteredTransaction
	errChan chan error
	signer  sdk.Signer
}

// NewTxEvent register a new txevent manager
func NewTxEvent(channel, txid string, clients []*sdk.PeerClient) (*TxEvent, error) {
	if len(clients) == 0 {
		return nil, errors.New("no peer clients")
	}
	te := &TxEvent{
		channel: channel,
		txid:    txid,
		clients: make([]peer.Deliver_DeliverFilteredClient, 0, len(clients)),
		receive: make(chan *peer.FilteredTransaction),
		errChan: make(chan error),
	}
	errs := []error{}
	for _, p := range clients {
		fc, err := p.DeliverFilter()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		te.clients = append(te.clients, fc)
	}
	if len(te.clients) == 0 {
		return nil, errsToError(errs)
	}
	return te, nil
}

func (t *TxEvent) Listen() (*peer.FilteredTransaction, error) {
	seek := sdk.CreateNewestSeekInfo()
	seekEnv, err := sdk.CreateSeekEnvelope(t.signer, t.channel, seek)
	if err != nil {
		return nil, errors.WithMessage(err, "创建 deliver 信封失败")
	}
	err = t.connect(seekEnv)
	if err != nil {
		return nil, errors.WithMessage(err, "发送 deliver 信封失败")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return t.wait(ctx)
}

// Connect to target peer and send event request
func (t *TxEvent) connect(env *common.Envelope) error {
	errs := []error{}
	for _, fc := range t.clients {
		err := fc.Send(env)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}
	if len(errs) == len(t.clients) {
		return errsToError(errs)
	}
	return nil
}

// Wait for event response
func (t *TxEvent) wait(ctx context.Context) (*peer.FilteredTransaction, error) {
	done := make(chan struct{})
	for _, c := range t.clients {
		go func(fc peer.Deliver_DeliverFilteredClient, done <-chan struct{}) {
			for {
				select {
				case <-done:
					return
				default:
					{
						resp, err := fc.Recv()
						if err != nil {
							err = errors.Wrap(err, "error receiving from deliver filtered")
							t.errChan <- err
							return
						}
						switch r := resp.Type.(type) {
						case *peer.DeliverResponse_FilteredBlock:
							filteredTransactions := r.FilteredBlock.FilteredTransactions
							for _, tx := range filteredTransactions {
								if tx.Txid == t.txid {
									t.receive <- tx
									return
								}
							}
						case *peer.DeliverResponse_Status:
							err = errors.Errorf("deliver completed with status (%s) before txid received", r.Status)
							t.errChan <- err
							return
						default:
							err = errors.Errorf("received unexpected response type (%T)", r)
							t.errChan <- err
							return
						}
					}
				}
			}
		}(c, done)
	}

	errs := []error{}
	for {
		select {
		case err := <-t.errChan:
			errs = append(errs, err)
			if len(errs) == len(t.clients) {
				return nil, errors.Wrap(errsToError(errs), "failed to receive txid on all peers")
			}
		case tx := <-t.receive:
			close(done)
			return tx, nil
		case <-ctx.Done():
			return nil, errors.Wrap(errsToError(errs), "timed out waiting for txid on all peers")
		}
	}
}

func errsToError(errs []error) error {
	errsstr := []string{}
	for i, err := range errs {
		errsstr = append(errsstr, fmt.Sprintf("[%d] %s", i, err.Error()))
	}
	return errors.New("multiple errors: " + strings.Join(errsstr, " ; "))
}
