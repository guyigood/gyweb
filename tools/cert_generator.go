package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

// 证书生成工具
func main() {
	var (
		host     = flag.String("host", "localhost", "证书域名或IP")
		validFor = flag.Duration("duration", 365*24*time.Hour, "证书有效期")
		isCA     = flag.Bool("ca", false, "是否生成CA证书")
		certFile = flag.String("cert", "cert.pem", "证书文件名")
		keyFile  = flag.String("key", "key.pem", "私钥文件名")
	)
	flag.Parse()

	fmt.Printf("正在生成证书...\n")
	fmt.Printf("域名/IP: %s\n", *host)
	fmt.Printf("有效期: %s\n", *validFor)
	fmt.Printf("证书文件: %s\n", *certFile)
	fmt.Printf("私钥文件: %s\n", *keyFile)

	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("生成私钥失败: %v", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"GyWeb"},
			Country:       []string{"CN"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(*validFor),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 解析主机名
	hosts := []string{*host}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// 添加localhost和127.0.0.1
	template.DNSNames = append(template.DNSNames, "localhost")
	template.IPAddresses = append(template.IPAddresses, net.IPv4(127, 0, 0, 1), net.IPv6loopback)

	if *isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	// 创建证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatalf("创建证书失败: %v", err)
	}

	// 保存证书
	certOut, err := os.Create(*certFile)
	if err != nil {
		log.Fatalf("创建证书文件失败: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("写入证书失败: %v", err)
	}

	fmt.Printf("证书已保存到: %s\n", *certFile)

	// 保存私钥
	keyOut, err := os.OpenFile(*keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("创建私钥文件失败: %v", err)
	}
	defer keyOut.Close()

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatalf("序列化私钥失败: %v", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privKeyBytes}); err != nil {
		log.Fatalf("写入私钥失败: %v", err)
	}

	fmt.Printf("私钥已保存到: %s\n", *keyFile)
	fmt.Println("\n证书生成完成！")
	fmt.Println("\n使用方法:")
	fmt.Printf("  tlsConfig := &engine.TLSConfig{\n")
	fmt.Printf("      CertFile: \"%s\",\n", *certFile)
	fmt.Printf("      KeyFile:  \"%s\",\n", *keyFile)
	fmt.Printf("  }\n")
	fmt.Printf("  r.RunTLS(\":8443\", tlsConfig)\n")
}
