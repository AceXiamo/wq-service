package cos

import (
	"bytes"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"net/url"
	"wq-service/core"
)

type Cos struct {
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
	SecretId  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
}

var cc *cos.Client
var cosConfig *Cos

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

	core.Log.Infof("📦 [COS] %s", config.Cos.Bucket)
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
	core.Log.Infof("📦 [COS] 上传文件 %s", p)
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
	core.Log.Infof("📦 [COS] 上传文件 %s", p)
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
	core.Log.Infof("📦 [COS] 分片上传 %s", p)
	opt := &cos.MultiUploadOptions{
		PartSize:       100,
		ThreadPoolSize: 2,
	}
	_, _, err := cc.Object.Upload(
		context.Background(), p, lp, opt,
	)
	if err != nil {
		panic(err)
	}
	core.Log.Infof("📦 [COS] 上传完毕 %s", p)
}
