package main

import (
	pb "bookstore/book-service/proto"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const(
	ADDRESS = ":8080"
	DEFAULT_INFO_FILE = "d:\\go\\src\\bookstore\\client\\bookinfo.json"
	DEFAULT_UPDATE_FILE  = "d:\\go\\src\\bookstore\\client\\file.json"
)

type Book struct {
	Books []*pb.BookInfo `json:"books"`  //指定tag，用于和json中的格式匹配
}
//解析json文件
func parseFile(FileName string) (Book, error){
	data, err := ioutil.ReadFile(FileName) //需要使用绝对路径
	if err != nil{
		return Book{},err
	}
	var bookInfo Book

	err = json.Unmarshal(data, &bookInfo) //反序列化
	/*fmt.Println(bookinfo.Books[0].BookName)
	a,_ := json.Marshal(bookinfo) //序列化

	fmt.Println(string(a)) //直接使用string(a) 就可以输出json格式的对象 */
	return bookInfo,nil
}

var (
	A string
	D string
	U string
	R bool //是否查询数据  -name  -author -publishtime
	h bool
	S bool	//是否订阅  -name  -author
	name string
	author string
	publishtime string
)

func flagInit()  {
	flag.StringVar(&A,"A","","add book information provided ")//提供文件名
	flag.StringVar(&D,"D","","delete a book with the book name provided") //提供书名
	flag.StringVar(&U,"U","","update book information") //提供文件名
	flag.BoolVar(&R,"R",false,"delete book information")
	flag.StringVar(&name,"name","","retrieve name")
	flag.StringVar(&author,"author","","retrieve author")
	flag.StringVar(&publishtime,"publishtime","","retrieve publish time")
	flag.BoolVar(&h,"h",false,"get help") //帮助
	flag.BoolVar(&S,"S",false,"subscribe book")
	flag.Parse()
	if h {
		flag.PrintDefaults()
	}
}
func main()  {
	flagInit()

	conn, err := grpc.Dial(ADDRESS,grpc.WithInsecure())
	if err != nil{
		log.Fatalf("grpc.Dial err: %v", err)
	}
	defer conn.Close()

	client := pb.NewBookServiceClient(conn) //创建grpc客户端对象

	if len(A) > 0 {
		//添加书籍
		err = AddBookFunc(client)
		if err != nil{
			log.Fatal(err)
		}
	}

	if len(D) > 0{
		//删除书籍
		err = DeleteBookFunc(client)
		if err != nil{
			log.Fatal(err)
		}
	}

	if len(U) > 0 {
		//更新
		err = UpdateBookFunc(client)
		if err != nil{
			log.Fatal(err)
		}
	}

	if R {
		//查询
		err = RetrieveBookFunc(client)
		if err != nil{
			log.Fatal(err)
		}
	}
	
	if S {
		err := SubscribeFunc(client)
		if err != nil{
			log.Fatal(err)
		}
	}
}

func AddBookFunc(client pb.BookServiceClient) error {
	dir, err := os.Getwd() //获取当前的路径
	if err != nil{
		return err
	}
	path := filepath.Join(dir, A)
	//添加书籍 之后就发布出去
	bookInfo,err := parseFile(path)
	if err != nil{
		return err
	}

	response, err := client.AddBook(context.Background(), &pb.AddBookRequest{
		AddBookInfo:      bookInfo.Books,
		AdminAddBookInfo: nil,
	})
	if err != nil{
		return err
	}
	if response.AddBookOk{
		//对每一本书都发布
		for i := 0 ; i < len(bookInfo.Books); i++{
			err = PublishFunc(client,bookInfo.Books[i])
			if err != nil{
				log.Fatal(err)
			}
		}
	}
	//打印返回的值
	log.Println("Add book is ",response.AddBookOk)
	bt, _ := json.Marshal(response.AddBookInfoRes)
	log.Println(string(bt))

	return nil
}

//发布
func PublishFunc(client pb.BookServiceClient, bookinfo *pb.BookInfo )  error{
	response, err := client.Publish(context.Background(),bookinfo)
	if err != nil{
		return err
	}
	log.Println(bookinfo.BookName + " publish book is ",response.PublishBookOk)
	return nil
}
//订阅   什么时候会触发订阅
func SubscribeFunc(client pb.BookServiceClient )error{
	var stream pb.BookService_SubscribeClient
	var err error
	//根据书名和作者进行订阅
	switch  {
	case len(name) > 0:
		stream, err = client.Subscribe(context.Background(), &pb.SubscribeBookRequest{
			BookName: name,
		})
		if err != nil{
			return err
		}
	case len(author) > 0:
		stream, err = client.Subscribe(context.Background(), &pb.SubscribeBookRequest{
			BookAuthor: author,
		})
		if err != nil{
			return err
		}
	}

	for{
		info, err := stream.Recv()
		if err != nil{
			if err == io.EOF{
				return nil
			}
			return err
		}
		//消息有变化的时候 打印消息
		log.Println("subscribe message: [ " + fmt.Sprintln(info) + "]")
	}
	return nil
}

func DeleteBookFunc(client pb.BookServiceClient)  error {
	bookName := D
	response, err := client.DeleteBook(context.Background(), &pb.DeleteBookRequest{
		BookName:            bookName,
		DeleteBookOne:       true, //删除一个
		AdminDeleteBookInfo: nil,
	})
	if err != nil{
		return err
	}

	log.Println("delete book is ",response.DeleteBookOk)
	if response.DeleteBookOk{
		err = PublishFunc(client,response.BookInfoRes)
		if err != nil{
			log.Fatal(err)
		}
	}
	return nil
}


func UpdateBookFunc(client pb.BookServiceClient)  error {
	dir, err := os.Getwd() //获取当前的路径
	if err != nil{
		return err
	}
	path := filepath.Join(dir, U)
	var bookInfo pb.BookInfo
	//从文件中解析出结构体
	date, err := ioutil.ReadFile(path)
	if err != nil{
		return err
	}
	err = json.Unmarshal(date, &bookInfo)
	if err != nil{
		return err
	}
	response, err := client.UpdateBook(context.Background(), &pb.UpdateBookRequest{
		UpdateBookInfo:   &bookInfo,
		AdminAddBookInfo: nil,
	})

	if err != nil{
		return err
	}
	log.Println("Update book is ",response.UpdateBookOk)
	//发布
	if response.UpdateBookOk{
		err = PublishFunc(client,&bookInfo)
		if err != nil{
			log.Fatal(err)
		}
	}
	return nil
}
//根据书名查询书籍  
func RetrieveBookFunc(c pb.BookServiceClient)  error {
	var response *pb.RetrieveBookResponse
	var err error
	switch  {
	case len(name) > 0:
		bookName := name
		response, err = c.RetrieveBook(context.Background(), &pb.RetrieveBookRequest{
			BookNameRetrieve:bookName,
			BookAuthorRetrieve:"",
			BookTimeRetrieve:"",
			AdminRetrieveBookInfo:nil,
		})

	case len(author) > 0:
		bookAuthor := author
		response, err = c.RetrieveBook(context.Background(), &pb.RetrieveBookRequest{
			BookNameRetrieve:"",
			BookAuthorRetrieve:bookAuthor,
			BookTimeRetrieve:"",
			AdminRetrieveBookInfo:nil,
		})
	case len(publishtime) > 0:
		bookPublishTime := publishtime
		response, err = c.RetrieveBook(context.Background(), &pb.RetrieveBookRequest{
			BookNameRetrieve:"",
			BookAuthorRetrieve:"",
			BookTimeRetrieve: bookPublishTime,
			AdminRetrieveBookInfo:nil,
		})
	default:
		return errors.New("err command argument")
	}
	if err != nil{
		return err
	}
	log.Println("Retrieve book is ",response.RetrieveBookOk)
	//打印查询到的书籍
	bt, _ := json.Marshal(response.BookInfoRes)
	log.Println(string(bt))
	return nil
}
