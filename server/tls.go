package main

import (
	"crypto/tls"
	"golang.org/x/net/http2"
	"io/ioutil"
	"log"
	"net"
)
//获取TLS配置
func GetTLSConfig(certPemPath, certKeyPath string) *tls.Config {
	var certKeyPair *tls.Certificate
	cert, _ := ioutil.ReadFile(certPemPath)
	key, _ := ioutil.ReadFile(certKeyPath)

	pair, err := tls.X509KeyPair(cert, key) //一对PEM编码的数据中解析公钥/私钥对
	if err != nil {
		log.Println("TLS KeyPair err: %v\n", err)
	}

	certKeyPair = &pair

	return &tls.Config{
		Certificates: []tls.Certificate{*certKeyPair},
		NextProtos:   []string{http2.NextProtoTLS},
	}
}

func NewTLSListener(inner net.Listener, config *tls.Config) net.Listener {
	return tls.NewListener(inner, config)
}