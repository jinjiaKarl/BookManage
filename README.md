# Book Store System

## 主要功能

1. 提供图书的录入功能  //同一本书可以重复录入
2. 提供图书的信息更新功能
3. 提供图书的检索功能，通过图书的书名、作者名、出版日期等信息查询图书
4. 提供图书的删除功能
5. 提供图书的订阅功能，当订阅后可以持续收到图书的信息更新通知、比如图书的数量、状态

## 业务模型

### 图书信息

1. 图书id
2. 书名
3. 作者名
4. 图书数量
5. 出版日期
6. 国家

## 技术要求

1. 基于grpc的client-server模型实现
2. grpc服务端实现所有功能
   1. 图书信息需要**持久化存储**
3. 可以通过grpc客户端访问所有功能
4. 可以通过http访问所有功能
   1. 使用浏览器访问
   2. 使用curl工具访问

---

进入`$GOPATH/book-service/proto/`目录下，运行以下的命令生成.pb.go和.gw.pb.go文件。

```
protoc --go_out=plugins=grpc:.   book.proto
protoc --grpc-gateway_out=logtostderr=true:. ./book.proto
```

进入`$GOPATH/book-service/server/`目录下，运行`go run *.go`运行服务端代码，以下为curl的测试命令。

```
增加图书
curl -X POST -k https://localhost:8080/addbook -d '{"add_book_info": [{"book_name":"Go In Action","book_author":"Brian","book_number":10,"book_publish_time":"2016 - 01 - 01","book_country":"America"},{"book_name":"Effective C++","book_author":"Scott","book_number":3,"book_publish_time":"2006 - 07 - 01","book_country":"America"}]}'

删除图书
curl -X POST -k https://localhost:8080/deletebook -d '{"book_name": "Go In Action","delete_book_one":true}'

更新图书
curl -X POST -k https://localhost:8080/updatebook -d '{"update_book_info": {"book_name":"Go In Action","book_author":"Brian","book_number":20,"book_publish_time":"2016 - 01 - 01","book_country":"America"}}'

检索图书
curl -k https://localhost:8080/retrievebook/Go%20In%20Action
```

进入`$GOPATH/book-service/client/`目录下，

运行`go run main.go -A "booinfo.json"`可以向服务端添加图书信息，

运行`go run main.go -U "file.json"`可以向服务端更新图书信息，

运行`go run main.go -R -name "Go In Action"`可以查询书名为《Go In Action》的信息，

运行`go run main.go -D "Go In Action"`可以删除图书信息，

运行`go run main.go -S -name "Go In Action"`可以订阅《Go In Action》图书信息。





、

