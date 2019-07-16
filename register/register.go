package register

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/satori/go.uuid"
	"github.com/yakaa/log4g"
)

type (
	Register struct {
		schema        string
		client3       *clientv3.Client
		stop          chan bool
		interval      time.Duration
		leaseTime     int64
		serverName    string
		serverAddress string
		fullAddress   string
	}
)

//naming.Update{Op: naming.Add, Addr: "1.2.3.4", Metadata: "..."})
func NewRegister(
	schema string,
	serverName string,
	serverAddress string,
	client3 *clientv3.Client,
) *Register {
	return &Register{
		schema:        schema, // eg : wwww.grpcx.com
		serverName:    serverName,
		serverAddress: serverAddress,
		client3:       client3,
		interval:      3 * time.Second,
		leaseTime:     6,
		stop:          make(chan bool, 1),
		fullAddress:   fmt.Sprintf("%s/%s/%s", schema, serverName, uuid.NewV4().String()),
	}
}

func (r *Register) SetInterval(interval time.Duration) {
	r.interval = interval
}

func (r *Register) GetServerAddress() string {
	return r.serverAddress
}

func (r *Register) GetServerName() string {
	return r.serverName
}

func (r *Register) SetServerName(serverName string) {
	r.serverName = serverName
}

func (r *Register) GetFullAddress() string {
	return r.fullAddress
}

// set interval
func (r *Register) SetLeaseTime(leaseTime int64) {
	r.leaseTime = leaseTime
}

// Register register service with name as prefix to etcd, multi etcd addr should use ; to split
func (r *Register) Register() error {
	ticker := time.NewTicker(r.interval)
	go func() {
		for {
			if getResp, err := r.client3.Get(context.Background(), r.fullAddress); err != nil {
				log4g.Error(err)
			} else if getResp.Count == 0 {
				if err = r.withAlive(); err != nil {
					log4g.Error(err)
				}
			}
			select {
			case <-ticker.C:
			case <-r.stop:
				return
			}
		}
	}()
	return nil
}

func (r *Register) withAlive() error {
	leaseResp, err := r.client3.Grant(context.Background(), r.leaseTime)
	if err != nil {
		return err
	}
	_, err = r.client3.Put(context.Background(), r.fullAddress, r.serverAddress, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	_, err = r.client3.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}
	return nil
}

// UnRegister remove service from etcd
func (r *Register) UnRegister() error {
	if r.client3 != nil {
		r.stop <- true
		_, err := r.client3.Delete(context.Background(), r.fullAddress)
		if err != nil {
			return err
		}
	}
	return nil
}
