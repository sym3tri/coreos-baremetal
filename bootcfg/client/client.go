package client

import (
	"crypto/tls"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/coreos/coreos-baremetal/bootcfg/rpc/rpcpb"
)

var (
	errNoEndpoints = errors.New("client: No endpoints provided")
	errNoTLSConfig = errors.New("client: No TLS Config provided")
)

// Config configures a Client.
type Config struct {
	// List of endpoint URLs
	Endpoints []string
	// Client TLS credentials
	TLS *tls.Config
}

// Client provides a bootcfg client RPC session.
type Client struct {
	Groups   rpcpb.GroupsClient
	Profiles rpcpb.ProfilesClient
	Ignition rpcpb.IgnitionClient
	conn     *grpc.ClientConn
}

// New creates a new Client from the given Config.
func New(config *Config) (*Client, error) {
	if len(config.Endpoints) == 0 {
		return nil, errNoEndpoints
	}
	return newClient(config)
}

// Close closes the client's connections.
func (c *Client) Close() error {
	return c.conn.Close()
}

func newClient(config *Config) (*Client, error) {
	conn, err := dialEndpoints(config)
	if err != nil {
		return nil, err
	}
	client := &Client{
		conn:     conn,
		Groups:   rpcpb.NewGroupsClient(conn),
		Profiles: rpcpb.NewProfilesClient(conn),
		Ignition: rpcpb.NewIgnitionClient(conn),
	}
	return client, nil
}

// dialEndpoints attemps to Dial each endpoint in order to establish a
// connection.
func dialEndpoints(config *Config) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if config.TLS != nil {
		creds := credentials.NewTLS(config.TLS)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		return nil, errNoTLSConfig
	}

	var err error
	for _, endpoint := range config.Endpoints {
		conn, err := grpc.Dial(endpoint, opts...)
		if err == nil {
			return conn, nil
		}
	}
	return nil, err
}
