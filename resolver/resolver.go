package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/yakaa/grpcx/config"
	"google.golang.org/grpc/resolver"
)

type Resolver struct {
	client3    *clientv3.Client
	ClientConn resolver.ClientConn
	target     string // eg: www.xxx.com/auth/user-rpc
	withBlock  bool
}

// NewResolver initialize an etcd client
func NewResolver(conf *config.ClientConf) (*Resolver, error) {
	client3, err := clientv3.New(
		clientv3.Config{
			Endpoints:   conf.Endpoints,
			Username:    conf.UserName,
			Password:    conf.PassWord,
			DialTimeout: config.GrpcxDialTimeout,
		})
	if nil != err {
		return nil, err
	}
	return &Resolver{
		client3:   client3,
		withBlock: conf.WithBlock,
		target:    conf.Target,
	}, nil
}

func (r *Resolver) Build(target resolver.Target, clientConn resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	r.ClientConn = clientConn
	go r.watch(fmt.Sprintf("%s/%s/", target.Scheme, target.Endpoint))
	return r, nil
}

func (r *Resolver) Scheme() string {
	ret := r.parseTarget(r.target)
	return ret.Scheme
}

func (r *Resolver) Target() string {
	return r.target
}

func (r *Resolver) WithBlock() bool {
	return r.withBlock
}

// ResolveNow will be called by gRPC to try to resolve the target name
// again. It's just a hint, resolver can ignore this if it's not necessary.
//
// It could be called multiple times concurrently.
func (r *Resolver) ResolveNow(opt resolver.ResolveNowOption) {

}

// parseTarget splits target into a struct containing scheme, authority and
// endpoint.
//
// If target is not a valid scheme://authority/endpoint, it returns {Endpoint:
// target}.
func (r *Resolver) parseTarget(target string) (ret resolver.Target) {
	var ok bool
	ret.Scheme, ret.Endpoint, ok = r.splitTarget(target, "://")
	if !ok {
		return resolver.Target{Endpoint: target}
	}
	ret.Authority, ret.Endpoint, ok = r.splitTarget(ret.Endpoint, "/")
	if !ok {
		return resolver.Target{Endpoint: target}
	}
	return ret
}

func (r *Resolver) splitTarget(s, sep string) (string, string, bool) {
	spl := strings.SplitN(s, sep, 2)
	if len(spl) < 2 {
		return "", "", false
	}
	return spl[0], spl[1], true
}

// It's just a hint, resolver can ignore this if it's not necessary.
func (r *Resolver) Close() {

}

func (r *Resolver) watch(keyPrefix string) {
	var addrList []resolver.Address
	resp, err := r.client3.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	serverName := r.parseTarget(r.target).Endpoint
	for i := range resp.Kvs {
		addrList = append(
			addrList,
			resolver.Address{
				Addr:       string(resp.Kvs[i].Value),
				ServerName: serverName,
			})
	}
	r.ClientConn.NewAddress(addrList)
	watchChan := r.client3.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	for n := range watchChan {
		for _, ev := range n.Events {
			addr := string(ev.Kv.Value)
			switch ev.Type {
			case mvccpb.PUT:
				if !r.exist(addrList, addr) {
					addrList = append(
						addrList,
						resolver.Address{
							Addr:       addr,
							ServerName: serverName,
						})
					r.ClientConn.NewAddress(addrList)
				}
			case mvccpb.DELETE:
				if s, ok := r.remove(addrList, addr); ok {
					addrList = s
					r.ClientConn.NewAddress(addrList)
				}
			}
		}
	}
}

func (r *Resolver) exist(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}
	return false
}

func (r *Resolver) remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}
