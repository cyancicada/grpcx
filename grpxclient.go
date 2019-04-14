package grpcx

import (
	"context"

	"github.com/yakaa/grpcx/config"
	"github.com/yakaa/grpcx/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	rpcResolver "google.golang.org/grpc/resolver"
)

type (
	GrpcxClient struct {
		resolver *resolver.Resolver
	}
)

func MustNewGrpcxClient(conf *config.ClientConf) (*GrpcxClient, error) {
	r, err := resolver.NewResolver(conf)
	if err != nil {
		return nil, err
	}
	rpcResolver.Register(r)
	return &GrpcxClient{resolver: r}, nil
}

func (c *GrpcxClient) NextConnection(options ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name)}
	if c.resolver.WithBlock() {
		opts = append(opts, grpc.WithBlock())
	}
	opts = append(opts, options...)
	return grpc.DialContext(
		context.Background(),
		c.resolver.Target(),
		opts...,
	)
}
