package main

import (
	pb "bookstore/book-service/proto"
	"context"
	"crypto/tls"
	"github.com/docker/docker/pkg/pubsub"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/mgo.v2"
	"grpc-hello-world/pkg/util"
	"log"
	"net"
	"net/http"
	"time"
)

const(
	URL = "localhost:27017" //mongodb的地址
)
var (
	ServerPort string
	EndPoint string
	CertName string
	CertPemPath string
	CertKeyPath string
	tlsConfig *tls.Config
)

var c *mgo.Collection //集合

func main() {
	ServerPort = "8080"
	EndPoint = ":" + ServerPort
	CertName = "go-grpc-example"
	CertPemPath = "../certs/server.pem"
	CertKeyPath = "../certs/server.key"

	tlsConfig = GetTLSConfig(CertPemPath, CertKeyPath) //获取证书配置


	//连接数据库
	session, err := mgo.Dial(URL)
	if err != nil{
		log.Fatal("database connect err: ",err)
	}
	defer session.Close()
	c = session.DB("bookstore").C("book")



	lis, err := net.Listen("tcp", EndPoint)
	if err != nil{
		log.Fatal("network listen err : ",err)
	}
	srv := newServer(lis)

	log.Printf("gRPC and https listen on: %s\n", ServerPort)

	//在支持http2之后，再进行Serve
	if err = srv.Serve(NewTLSListener(lis, tlsConfig)); err != nil {
		log.Printf("ListenAndServe: %v\n", err)
	}
	return
}


func newServer(conn net.Listener) (*http.Server) {
	grpcServer := newGrpc()
	gwmux, err := newGateway()
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)


	return &http.Server{
		Addr:      EndPoint,
		Handler:   util.GrpcHandlerFunc(grpcServer, mux),
		TLSConfig: tlsConfig,
	}
}

func newGrpc() *grpc.Server {
	creds, err := credentials.NewServerTLSFromFile(CertPemPath, CertKeyPath)
	if err != nil {
		panic(err)
	}

	opts := []grpc.ServerOption{
		grpc.Creds(creds),
	}

	server := grpc.NewServer(opts...)
	pb.RegisterBookServiceServer(server, &BookService{pub:pubsub.NewPublisher(100*time.Microsecond,10)})//注册服务

	return server
}

func newGateway() (http.Handler, error) {
	ctx := context.Background()
	dcreds, err := credentials.NewClientTLSFromFile(CertPemPath, CertName)
	if err != nil {
		return nil, err
	}

	gwmux := runtime.NewServeMux()
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)} //指定客户端进行访问的时候使用的证书!!!!
	if err := pb.RegisterBookServiceHandlerFromEndpoint(ctx, gwmux, EndPoint, dopts); err != nil {
		return nil, err
	}

	return gwmux, nil
}

