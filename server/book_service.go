package main

import (
	pb "bookstore/book-service/proto"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/pkg/pubsub"
	"gopkg.in/mgo.v2/bson"
	"log"
)

type BookInfo pb.BookInfo
//实现Stringer接口，自定义输出
func (this *BookInfo)String() string {
	return fmt.Sprintf("[book name: %s   book author: %s  book number: %d    book publish time: %s   book country: %s]",
		this.BookName,this.BookAuthor,this.BookNumber,this.BookPublishTime,this.BookCountry)
}
//定义错误信息
var(
	ErrNoBook = errors.New("this book is not on the database")
	//ErrYesBook = errors.New("this book is on the database")
	ErrAddBook = errors.New("add book false")
	ErrDeleteBook = errors.New("delete book false")
	ErrUpdateBook = errors.New("update book false")
	ErrRetrieveBook = errors.New("retrieve book false")
	ErrSubscribeBook = errors.New("subscribe book false")
)
//实现BookServiceServer接口
type BookService struct {
	pub *pubsub.Publisher
}

//先不判断用户的信息是否合法
func (this *BookService) AddBook(ctx context.Context, in *pb.AddBookRequest) (*pb.AddBookResponse, error) {
	//插入之前，先判断在数据库中是否存在
	bookInfo := in.AddBookInfo //书籍信息
	var temp pb.BookInfo
	for i := 0; i < len(bookInfo); i++{
		err := c.Find(bson.M{"bookname": bookInfo[i].BookName}).One(&temp) //上面必须声明成var temp pb.BookInfo，不能是指针。因此这里必须用&
		if err == nil{
			//已经存在
			err = c.Update(bson.M{"bookname": temp.BookName}, bson.M{"$set": bson.M{"booknumber": temp.BookNumber + temp.BookNumber}})
			if err != nil{
				log.Println(ErrAddBook.Error())
				return &pb.AddBookResponse{AddBookOk:false,AddBookInfoRes:nil},ErrAddBook
			}
		}else{
			//不存在
			err = c.Insert(bookInfo[i])
			if err != nil{
				log.Println(ErrAddBook.Error())
				return &pb.AddBookResponse{AddBookOk:false,AddBookInfoRes:nil},ErrAddBook
			}
		}
	}

	log.Printf("Add book is %t, book information: %v\n", true,in.AddBookInfo) //打印出相关信息

	return &pb.AddBookResponse{AddBookOk:true,AddBookInfoRes:bookInfo},nil
}
//查询书籍
func (this *BookService) RetrieveBook(ctx context.Context,in *pb.RetrieveBookRequest) (*pb.RetrieveBookResponse, error) {
	var temp pb.BookInfo
	var res []*pb.BookInfo
	var err error
	switch  {
	case in.BookNameRetrieve != "":
		err = c.Find(bson.M{"bookname": in.BookNameRetrieve}).One(&temp)
		res = append(res, &temp)
	case in.BookAuthorRetrieve != "": //一个作者可以有好多书
		//c.Find(bson.M{"bookauthor": in.BookAuthorRetrieve}).All(res)
		iter := c.Find(bson.M{"bookauthor": in.BookAuthorRetrieve}).Iter()
		for iter.Next(&temp){
			res = append(res, &temp)
		}
	case in.BookTimeRetrieve != "":
		iter := c.Find(bson.M{"bookpublishtime": in.BookTimeRetrieve}).Iter()
		for iter.Next(&temp){
			res = append(res, &temp)
		}
	default:
		return &pb.RetrieveBookResponse{RetrieveBookOk:false,BookInfoRes:nil},ErrRetrieveBook
	}

	if err != nil{
		log.Println(ErrRetrieveBook.Error())
		return &pb.RetrieveBookResponse{RetrieveBookOk:false,BookInfoRes:nil},ErrRetrieveBook
	}

	return &pb.RetrieveBookResponse{RetrieveBookOk:true,BookInfoRes:res},nil
}


//根据书名删除图书
func (this *BookService) DeleteBook(ctx context.Context, in *pb.DeleteBookRequest) (*pb.DeleteBookResponse, error) {
	//判断是否存在在数据库中
	bookName := in.BookName
	var temp pb.BookInfo
	err := c.Find(bson.M{"bookname": bookName}).One(&temp)
	if err != nil{
		log.Println(ErrNoBook.Error())
		return &pb.DeleteBookResponse{DeleteBookOk:false}, ErrNoBook
	}
	//该确保在数据库中
	//删除一个
	if in.DeleteBookOne {
		temp.BookNumber = temp.BookNumber - 1
		err := c.Update(bson.M{"bookname": bookName}, bson.M{"$set": bson.M{"booknumber": temp.BookNumber}})
		if err != nil{
			return &pb.DeleteBookResponse{DeleteBookOk:false},ErrDeleteBook
		}
		return &pb.DeleteBookResponse{DeleteBookOk:true,BookInfoRes:&temp}, nil
	}else{ //删除全部
		err := c.Remove(bson.M{"bookname": bookName})
		if err != nil{
			return &pb.DeleteBookResponse{DeleteBookOk:false}, ErrDeleteBook
		}
		return &pb.DeleteBookResponse{DeleteBookOk:true,BookInfoRes:&pb.BookInfo{BookName:temp.BookName}}, nil
	}
}
func (this *BookService) UpdateBook(ctx context.Context, in *pb.UpdateBookRequest) (*pb.UpdateBookResponse, error) {
	book := in.UpdateBookInfo
	var temp pb.BookInfo
	err := c.Find(bson.M{"bookname": book.BookName}).One(&temp)
	if err != nil{
		log.Println(ErrNoBook.Error())
		return &pb.UpdateBookResponse{UpdateBookOk:false}, ErrNoBook
	}

	//先删除之前的
	err = c.Remove(bson.M{"bookname": book.BookName})
	if err != nil{
		return &pb.UpdateBookResponse{UpdateBookOk:false}, ErrUpdateBook
	}
	//添加
	err = c.Insert(book)
	if err != nil{
		return &pb.UpdateBookResponse{UpdateBookOk:false}, ErrUpdateBook
	}
	return &pb.UpdateBookResponse{UpdateBookOk:true}, nil
}

func (this *BookService)Publish(ctx context.Context, in *pb.BookInfo) (*pb.PublishResponse, error) {
	this.pub.Publish(in)
	//log.Println(in.BookName + " publish successfully")
	return &pb.PublishResponse{PublishBookOk:true},nil
}

func (this *BookService) Subscribe(in *pb.SubscribeBookRequest,stream pb.BookService_SubscribeServer) error {
	ch := this.pub.SubscribeTopic(func(v interface{}) bool {
		if key, ok := v.(*pb.BookInfo); ok{
			if len(in.BookName) > 0 && key.BookName == in.BookName{ //当书名一样时，订阅
				return true
			}else if len(in.BookAuthor) > 0 && key.BookAuthor == in.BookAuthor{
				return true
			}
		}
		return false
	})
	//如果没有数据就会阻塞。只要 ch 不关闭，就会一直读取。close(ch)才会结束range
	for v := range ch{
		if key, ok := v.(*pb.BookInfo); ok{
			if err := stream.Send(key); err != nil{
				return ErrSubscribeBook
			}
			//log.Println(in.BookName + " send successfully")
		}
	}
	return nil
}