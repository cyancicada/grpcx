package config

import "time"

type (
	ServiceConf struct {
		EtcdAuth
		Schema        string
		ServerName    string
		ServerAddress string
		Endpoints     []string
	}

	ClientConf struct {
		EtcdAuth
		Target    string
		Endpoints []string
		WithBlock bool
	}

	EtcdAuth struct {
		UserName string
		PassWord string
	}
)

const (
	GrpcxDialTimeout = 3 * time.Second
)
