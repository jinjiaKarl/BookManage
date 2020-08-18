package main

import (
	pb "bookstore/book-service/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
)


func main()  {
	conn, err := grpc.Dial(":8080",grpc.WithInsecure())
	if err != nil{
		log.Fatalf("grpc.Dial err: %v", err)
	}
	defer conn.Close()

	client := pb.NewBookServiceClient(conn) //创建grpc客户端对象
	err = SubscribeFunc(client)
	if err != nil{
		log.Fatal(err)
	}
}

func SubscribeFuncc(client pb.BookServiceClient )error  {
	stream, err := client.Subscribe(context.Background(), &pb.SubscribeBookRequest{
		BookName: "Go In Action",
	})
	if err != nil{
		return err
	}
	for{
		info, err := stream.Recv()
		if err != nil{
			if err == io.EOF{
				return nil
			}
			return err
		}
		log.Println(info)
	}
	return nil
}