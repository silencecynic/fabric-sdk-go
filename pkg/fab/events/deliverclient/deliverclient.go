/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deliverclient

import (
	"math"
	"time"

	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client"
	deliverconn "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/connection"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/dispatcher"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/endpoint"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

// deliverProvider is the connection provider used for connecting to the Deliver service
var deliverProvider = func(context fabcontext.Client, chConfig fab.ChannelCfg, peer fab.Peer) (api.Connection, error) {
	eventEndpoint, ok := peer.(api.EventEndpoint)
	if !ok {
		panic("peer is not an EventEndpoint")
	}
	return deliverconn.New(context, chConfig, deliverconn.Deliver, peer.URL(), eventEndpoint.Opts()...)
}

// deliverFilteredProvider is the connection provider used for connecting to the DeliverFiltered service
var deliverFilteredProvider = func(context fabcontext.Client, chConfig fab.ChannelCfg, peer fab.Peer) (api.Connection, error) {
	eventEndpoint, ok := peer.(api.EventEndpoint)
	if !ok {
		panic("peer is not an EventEndpoint")
	}
	return deliverconn.New(context, chConfig, deliverconn.DeliverFiltered, peer.URL(), eventEndpoint.Opts()...)
}

// Client connects to a peer and receives channel events, such as bock, filtered block, chaincode, and transaction status events.
type Client struct {
	client.Client
	params
}

// New returns a new deliver event client
func New(context fabcontext.Client, chConfig fab.ChannelCfg, opts ...options.Opt) (*Client, error) {
	params := defaultParams()
	options.Apply(params, opts)

	// Use a context that returns a custom Discovery Provider which
	// produces event endpoints containing additional GRPC options.
	deliverCtx := newDeliverContext(context)

	client := &Client{
		Client: *client.New(
			dispatcher.New(deliverCtx, chConfig, params.connProvider, opts...),
			opts...,
		),
		params: *params,
	}
	client.SetAfterConnectHandler(client.seek)
	client.SetBeforeReconnectHandler(client.setSeekFromLastBlockReceived)

	if err := client.Start(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) seek() error {
	logger.Debugf("Sending seek request....")

	seekInfo, err := c.seekInfo()
	if err != nil {
		return err
	}

	errch := make(chan error)
	c.Submit(dispatcher.NewSeekEvent(seekInfo, errch))

	select {
	case err = <-errch:
	case <-time.After(c.respTimeout):
		err = errors.New("timeout waiting for deliver status response")
	}

	if err != nil {
		logger.Errorf("Unable to send seek request: %s", err)
		return err
	}

	logger.Debugf("Successfully sent seek")
	return nil
}

func (c *Client) setSeekFromLastBlockReceived() error {
	c.Lock()
	defer c.Unlock()

	// Make sure that, when we reconnect, we receive all of the events that we've missed
	lastBlockNum := c.Dispatcher().LastBlockNum()
	if lastBlockNum < math.MaxUint64 {
		c.seekType = seek.FromBlock
		c.fromBlock = c.Dispatcher().LastBlockNum() + 1
	} else {
		// We haven't received any blocks yet. Just ask for the newest
		c.seekType = seek.Newest
	}
	return nil
}

func (c *Client) seekInfo() (*ab.SeekInfo, error) {
	c.RLock()
	defer c.RUnlock()

	switch c.seekType {
	case seek.Newest:
		return seek.InfoNewest(), nil
	case seek.Oldest:
		return seek.InfoOldest(), nil
	case seek.FromBlock:
		return seek.InfoFrom(c.fromBlock), nil
	default:
		return nil, errors.Errorf("unsupported seek type:[%s]", c.seekType)
	}
}

// deliverContext overrides the DiscoveryProvider
type deliverContext struct {
	fabcontext.Client
}

func newDeliverContext(ctx fabcontext.Client) fabcontext.Client {
	return &deliverContext{
		Client: ctx,
	}
}

// DiscoveryProvider returns a custom discovery provider which produces
// event endpoints with additional GRPC options
func (ctx *deliverContext) DiscoveryProvider() fab.DiscoveryProvider {
	return endpoint.NewDiscoveryProvider(ctx.Client)
}
