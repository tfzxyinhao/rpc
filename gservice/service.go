package gservice

import (
	"log"
	"net"
	"strconv"
	"time"

	"github.com/tfzxyinhao/rpc/gservice/calc"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/naming"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	gn "google.golang.org/grpc/naming"
)

const (
	port         = 8000
	Host         = "192.168.0.45"
	endpoint     = "http://192.168.0.45:2379"
	service_name = "/service"
)

type Server struct {
	calc.CalcServer
}

func (s *Server) CalcResult(ctx context.Context, param *calc.CalcRequest) (*calc.CalcReply, error) {
	return &calc.CalcReply{IResult: param.IResult * 2, SResult: "result:" + param.SResult + ":OK"}, nil
}

func GetLocalAddrs() []string {
	result := make([]string, 0)
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println("lookup addr err:", err)
		return result
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Println("get interface err:", err.Error())
			continue
		}

		for _, addr := range addrs {
			ip, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			if !ip.IP.IsLoopback() && ip.IP.IsGlobalUnicast() {
				result = append(result, ip.IP.String())
			}
		}
	}
	return result
}

func ServService() {
	addr := Host + ":" + strconv.Itoa(port)
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("listen err:", err.Error())
		return
	}

	s := Server{}
	server := grpc.NewServer()
	calc.RegisterCalcServer(server, &s)
	server.Serve(listen)
}

func RegisterService() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: time.Second * 5,
	})

	if err != nil {
		log.Println("etcd err:", err)
		return
	}

	cli.Delete(cli.Ctx(), service_name)
	r := &naming.GRPCResolver{Client: cli}
	for _, addr := range GetLocalAddrs() {
		service_node := addr + ":" + strconv.Itoa(port)
		err = r.Update(cli.Ctx(), service_name, gn.Update{Op: gn.Add, Addr: service_node})
		log.Println("register node :", service_name, service_node, err)
	}
}

func ClientTestService() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: time.Second * 5,
	})

	if err != nil {
		log.Println("connect etcd err:", err.Error())
		return
	}

	defer cli.Close()
	r := &naming.GRPCResolver{Client: cli}
	b := grpc.RoundRobin(r)

	conn, err := grpc.Dial(service_name, grpc.WithBalancer(b), grpc.WithInsecure())
	if err != nil {
		log.Println("dial err:", err.Error())
		return
	}

	defer conn.Close()
	c := calc.NewCalcClient(conn)
	req := calc.CalcRequest{IResult: 1, SResult: "req"}
	resp, err := c.CalcResult(cli.Ctx(), &req)
	if err != nil {
		log.Println("calc err:", err)
		return
	}

	log.Println(resp.IResult, resp.SResult)
}

func ClientTestServiceDirect() {
	conn, err := grpc.Dial("192.168.0.45:8000", grpc.WithInsecure())
	if err != nil {
		log.Println("dial err:", err.Error())
		return
	}

	defer conn.Close()
	c := calc.NewCalcClient(conn)
	req := calc.CalcRequest{IResult: 1, SResult: "req"}
	start := time.Now()
	resp, err := c.CalcResult(context.Background(), &req)
	cost := time.Now().Sub(start)
	if err != nil {
		log.Println("calc err:", err)
		return
	}

	log.Println(cost, resp.IResult, resp.SResult)
}
