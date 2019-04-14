# grpcx

## Quick start

# NOTE
 
 ## 使用 `golang 1.11.1  `以上版本开发，使用 go mod <br>
 
 ### 找一个非gopath 目录拉取到本地后进行如下操作<br>
 1. 添加GOPROXY到你的环境变量中<br>
    `window` : `set GOPROXY=https://goproxy.io` <br>
    `linux` : `export GOPROXY=https://goproxy.io`<br>
 2. 进行你的项目根目录 执行 命令 `go mod tidy` 来拉取所有依赖包<br>
 
 
## 启动一个 knowing 的rpc服务<br>

```javascript 1.8
function main (){
    // 启动3个knowing的rpc服务
    for i := 0; i < 3; i++ {
         conf := &config.ServiceConf{
                EtcdAuth:      config.EtcdAuth{}, //1
                Schema:        "www.vector.com", // //2
                ServerName:    "knowing",//3
                Endpoints:     []string{"127.0.0.1:2379"},//4
                ServerAddress: "127.0.0.1:2000"+strconv.Itoa(i),
            }
            demo := &RegionHandlerServer{ServerAddress: conf.ServerAddress}
            rpcServer, err := grpcx.MustNewGrpcxServer(conf, func(server *grpc.Server) {
                proto.RegisterRegionHandlerServer(server, demo)
            })
            if err != nil {
                panic(err)
            }
            log.Fatal(rpcServer.Run())
    }
}
```
`1`  `EtcdAuth` 这是配置etcd的访问时权限设置`auth`，UserName 和 PassWord，如果没有设置etcd的auth就不加 <br>
`2` `Schema` 随便什么都可以，一般会写成您的公司网址或代号<br>
`3` `ServerName` 这个rpc服务的名称<br>
`4` `Endpoints` Etcd的集群地址<br>



## 启动访问knowing rpc服务的客户端

```javascript 1.8

func main() {
	conf := &config.ClientConf{
		EtcdAuth:  config.EtcdAuth{},
		Target:    "www.vector.com:///knowing",
		Endpoints: []string{"127.0.0.1:2379"},
	}

	r, err := grpcx.MustNewGrpcxClient(conf)
	if err != nil {
		panic(err)
	}
	conn, err := r.NextConnection()
	if err != nil {
		panic(err)
	}
	regionHandlerClient := proto.NewRegionHandlerClient(conn)
	for {
		res, err := regionHandlerClient.GetListenAudio(
			context.Background(),
			&proto.FindRequest{Tokens: []string{"a_"}},
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res)
		time.Sleep(2 * time.Second)  // 每两次请求一次knowing的rpc服务
	}
}
```



## 可以看grpc负载均衡的效果

```cgo

Items:<token:"a_127.0.0.1:20002" listen:1 > 
Items:<token:"a_127.0.0.1:20001" listen:1 > 
Items:<token:"a_127.0.0.1:20000" listen:1 > 
Items:<token:"a_127.0.0.1:20002" listen:1 > 
Items:<token:"a_127.0.0.1:20001" listen:1 > 
Items:<token:"a_127.0.0.1:20000" listen:1 > 
Items:<token:"a_127.0.0.1:20002" listen:1 > 
Items:<token:"a_127.0.0.1:20001" listen:1 > 
Items:<token:"a_127.0.0.1:20000" listen:1 > 
Items:<token:"a_127.0.0.1:20002" listen:1 > 
Items:<token:"a_127.0.0.1:20001" listen:1 > 
Items:<token:"a_127.0.0.1:20000" listen:1 > 
Items:<token:"a_127.0.0.1:20002" listen:1 > 
Items:<token:"a_127.0.0.1:20001" listen:1 > 
Items:<token:"a_127.0.0.1:20000" listen:1 > 
Items:<token:"a_127.0.0.1:20002" listen:1 > 
```


## 更多详细请看 `example` 文件夹中的实例