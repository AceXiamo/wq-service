package cos

import (
	"bytes"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type Cos struct {
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
	SecretId  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
}

var cc *cos.Client
var cosConfig *Cos

const chunkSize = 1024 * 1024 * 1024 // 分片大小，1GB

// Init
// @Description: 初始化COS
func Init() {
	yamlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatal(err)
	}
	var config struct {
		Cos Cos `yaml:"cos"`
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatal(err)
	}
	cosConfig = &config.Cos

	log.Infof("📦 [COS] %s", config.Cos.Bucket)
	CreateCosClient()
}

// CreateCosClient
// @Description: 创建COS客户端
func CreateCosClient() {
	u, _ := url.Parse("https://" + cosConfig.Bucket + ".cos." + cosConfig.Region + ".myqcloud.com")
	b := &cos.BaseURL{BucketURL: u}
	cc = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cosConfig.SecretId,
			SecretKey: cosConfig.SecretKey,
		},
	})
}

// UploadFile
// @Description: 上传文件
// @param p 		路径
// @param bts		文件内容
func UploadFile(p string, bts []byte) {
	log.Infof("📦 [COS] 上传文件 %s", p)
	_, err := cc.Object.Put(context.Background(), p, bytes.NewReader(bts), nil)
	if err != nil {
		return
	}
}

// UploadLocalFile
// @Description: 上传本地文件
// @param p			路径
// @param lp		本地文件路径
func UploadLocalFile(p string, lp string) {
	log.Infof("📦 [COS] 上传文件 %s", p)
	_, err := cc.Object.PutFromFile(context.Background(), p, lp, nil)
	if err != nil {
		fmt.Print(err)
		return
	}
}

// MultipartUpload
// @Description: 分片上传
// @param p			路径
// @param lp		本地文件路径
func MultipartUpload(p string, lp string) {
	log.Infof("📦 [COS] 上传文件 %s", p)
	init, _, err := cc.Object.InitiateMultipartUpload(context.Background(), p, nil)
	if err != nil {
		fmt.Print(err)
		panic(err)
	}
	UploadID := init.UploadID
	f, err := os.Open(lp)
	if err != nil {
		fmt.Print(err)
		panic(err)
	}
	defer f.Close()

	var parts []int
	chunk := make([]byte, chunkSize)
	for {
		n, err := f.Read(chunk)
		if err != nil && err != io.EOF {
			fmt.Println(err)
			break
		}
		if n == 0 {
			break
		}
		parts = append(parts, n)
	}

	var ec = make(chan string, len(parts))
	for i, part := range parts {
		go func(partNumber int, content *bytes.Reader) {
			log.Infof("📦 [COS] 上传分片 %d", partNumber)
			resp, err := cc.Object.UploadPart(context.Background(), p, UploadID, partNumber, content, nil)
			if err != nil {
				fmt.Println(err)
				panic(err)
			}
			log.Infof("📦 [COS] 分片 %d 上传完毕", partNumber)
			ec <- resp.Header.Get("ETag")
		}(i+1, bytes.NewReader(chunk[:part]))
	}

	// 等待所有分片上传完成
	opt := &cos.CompleteMultipartUploadOptions{}
	index := 1
	for s := range ec {
		opt.Parts = append(opt.Parts, cos.Object{
			PartNumber: index,
			ETag:       s,
		})
		if index == len(parts) {
			break
		}
		index++
	}
	_, _, err = cc.Object.CompleteMultipartUpload(
		context.Background(), p, UploadID, opt,
	)
	if err != nil {
		panic(err)
	}
}
