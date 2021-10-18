package comm

import "google.golang.org/grpc"

type CommonClient struct {
	clientConfig ClientConfig
	address      string
}

func NewCommonClient(clientConfig ClientConfig) (*CommonClient, error) {
	return &CommonClient{
		clientConfig: clientConfig,
	}, nil
}

func (cc *CommonClient) Dial(address string, serverOverride string) (*grpc.ClientConn, error) {
	cc.clientConfig.SecOpts.ServerNameOverride = serverOverride
	return cc.clientConfig.Dial(address)
}
