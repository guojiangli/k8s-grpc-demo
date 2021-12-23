package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/sirupsen/logrus"
	pb "github.com/wwcd/grpc-lb/cmd/helloworld"
	grpclb "github.com/wwcd/grpc-lb/etcdv3"
)

var (
	serv = flag.String("service", "hello_service", "service name")
	host = flag.String("host", GetLocalIP(), "listening host")
	port = flag.String("port", "50001", "listening port")
	reg  = flag.String("reg", "http://etcd3:2379", "register etcd address") //如果是本地就把etcd3替换成127.0.0.1
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", net.JoinHostPort(*host, *port))
	if err != nil {
		panic(err)
	}

	err = grpclb.Register(*reg, *serv, *host, *port, time.Second*10, 15)
	if err != nil {
		panic(err)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		logrus.Infof("receive signal '%v'", s)
		grpclb.UnRegister()
		os.Exit(1)
	}()

	logrus.Infof("starting hello service at: %s", *port)
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	s.Serve(lis)
}

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	logrus.Infof("%v: Receive is %s\n", time.Now(), in.Name)
	return &pb.HelloReply{Message: "[upgrade] Hello " + in.Name + " from " + net.JoinHostPort(*host, *port)}, nil
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
