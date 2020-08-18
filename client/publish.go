package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"gopkg.in/mgo.v2"
	"log"
	pb "bookstore/book-service/proto"
)

func main()  {
	session, err := mgo.Dial(URL)
	if err != nil{
		log.Fatal("database connect err: ",err)
	}
	defer session.Close()
	c = session.DB("bookstore").C("book")

	conn, err := grpc.Dial(":8080",grpc.WithInsecure())
	if err != nil{
		log.Fatalf("grpc.Dial err: %v", err)
	}
	defer conn.Close()

	client := pb.NewBookServiceClient(conn) //创建grpc客户端对象
	err = PublishFuncc(client)
	if err != nil{
		log.Fatal(err)
	}
}

func PublishFuncc(client pb.BookServiceClient )  error{
	response, err := client.Publish(context.Background(), &pb.BookInfo{
		BookName:        "Go In Action",
		BookAuthor:      "Brian",
		BookNumber:      5,
		BookPublishTime: "2016 - 01 - 01",
		BookCountry:     "America",
	})
	if err != nil{
		return err
	}
	log.Println("Publish book is ",response.PublishBookOk)
	return nil
}