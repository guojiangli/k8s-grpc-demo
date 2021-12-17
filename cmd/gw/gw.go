package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"

	pb "github.com/wwcd/grpc-lb/cmd/helloworld"
	grpclb "github.com/wwcd/grpc-lb/etcdv3"
)

var (
	svc  = flag.String("service", "hello_service", "service name")
	host = flag.String("host", GetLocalIP(), "listening host")
	port = flag.String("port", "60001", "listening port")
	reg  = flag.String("reg", "http://etcd3:2379", "register etcd address")
)

func main() {
	flag.Parse()
	r := grpclb.NewResolver(*reg, *svc)
	resolver.Register(r)

	ctx, cancel := context.WithTimeout(context.Background(), 1000000*time.Second)
	// https://github.com/grpc/grpc/blob/master/doc/naming.md
	// The gRPC client library will use the specified scheme to pick the right resolver plugin and pass it the fully qualified name string.
	conn, err := grpc.DialContext(ctx, r.Scheme()+"://authority/"+*svc, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name), grpc.WithBlock())
	cancel()
	if err != nil {
		panic(err)
	}

	mux := runtime.NewServeMux()
	err = pb.RegisterGreeterHandler(ctx, mux, conn)
	if err != nil {
		panic(err)
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	logrus.Fatal(http.ListenAndServe(net.JoinHostPort(*host, *port), mux))
}

//获取本地eth0 IP
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil { //本地网卡eth0的ip
				return ipnet.IP.String()
			}
		}
	}

	return ""
}
