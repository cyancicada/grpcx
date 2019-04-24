package grpcx

import (
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/coreos/etcd/clientv3"
	"github.com/yakaa/log4g"
	"google.golang.org/grpc"

	"github.com/yakaa/grpcx/config"
	"github.com/yakaa/grpcx/register"
)

var (
	deadSignal = []os.Signal{
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	}
)

type (
	GrpcxServiceFunc func(server *grpc.Server)
	GrpcxServer      struct {
		register       *register.Register
		rpcServiceFunc GrpcxServiceFunc
	}
)

const (
	colon  string = ":"
	dns114 string = "114.114.114.114:80"
)

func MustNewGrpcxServer(conf *config.ServiceConf, rpcServiceFunc GrpcxServiceFunc) (*GrpcxServer, error) {
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
	address := strings.Split(conf.ServerAddress, colon)
	if strings.TrimSpace(address[0]) == "" {
		address[0] = FindLocalAddress()
	}
	conf.ServerAddress = strings.Join(address, colon)
	return &GrpcxServer{
		register: register.NewRegister(
			conf.Schema,
			conf.ServerName,
			conf.ServerAddress,
			client3,
		),
		rpcServiceFunc: rpcServiceFunc,
	}, nil
}

func (s *GrpcxServer) Run(serverOptions ...grpc.ServerOption) error {
	listen, err := net.Listen("tcp", s.register.GetServerAddress())
	if nil != err {
		return err
	}
	log4g.InfoFormat(
		"serverAddress [%s] of %s Rpc server has started and full key [%s]",
		s.register.GetServerAddress(),
		s.register.GetServerName(),
		s.register.GetFullAddress(),
	)
	if err := s.register.Register(); err != nil {
		return err
	}
	server := grpc.NewServer(serverOptions...)
	s.rpcServiceFunc(server)
	s.deadNotify()
	if err := server.Serve(listen); nil != err {
		return err
	}
	return nil

}

func (s *GrpcxServer) deadNotify() {
	ch := make(chan os.Signal, 1) //
	signal.Notify(ch, deadSignal...)
	go func() {
		log4g.InfoFormat(
			"serverAddress [%s] of %s Rpc server has existed with got signal [%v] and full key [%s]",
			s.register.GetServerAddress(),
			s.register.GetServerName(),
			<-ch,
			s.register.GetFullAddress(),
		)
		if err := s.register.UnRegister(); err != nil {
			log4g.InfoFormat(
				"serverAddress [%s] of %s Rpc server UnRegister fail and  err %+v and full key [%s]",
				s.register.GetServerAddress(),
				s.register.GetServerName(),
				s.register.GetFullAddress(),
				err,
				s.register.GetFullAddress(),
			)
		}
		os.Exit(1) //
	}()
	return
}

// find local ip and port by send a udp request to 114.114.114.114:80
func FindLocalAddress() string {
	conn, _ := net.Dial("udp", dns114)
	localAddress := strings.Split(conn.LocalAddr().String(), ",:")
	defer func() {
		if err := conn.Close(); err != nil {
			log4g.ErrorFormat("conn.Close err %+v", err)
		}
	}()
	return localAddress[0]
}
