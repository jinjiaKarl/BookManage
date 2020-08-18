package main

import (
	"github.com/docker/docker/pkg/pubsub"
	"google.golang.org/grpc"
	pb "bookstore/book-service/proto"
	"gopkg.in/mgo.v2"
	"log"
	"net"
	"time"
)

const(
	Port  = ":8080"
	URL = "localhost:27017" //mongodb的地址
)

var c *mgo.Collection //集合

func main() {
	//连接数据库
	session, err := mgo.Dial(URL)
	if err != nil{
		log.Fatal("database connect err: ",err)
	}
	defer session.Close()
	c = session.DB("bookstore").C("book")


	server := grpc.NewServer()
	pb.RegisterBookServiceServer(server, &BookService{pub:pubsub.NewPublisher(100*time.Microsecond,10)})//注册服务
	lis, err := net.Listen("tcp", Port)
	if err != nil{
		log.Fatal("network listen err : ",err)
	}

	server.Serve(lis)


}

