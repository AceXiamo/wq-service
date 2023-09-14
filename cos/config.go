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
	init, _, err := cc.Object.InitiateMultipartUpload(context.Background(), p, nil)
	if err != nil {
		fmt.Print(err)
		panic(err)
	}
	UploadID := init.UploadID
	log.Infof("📦 [COS] 分片上传 [%s]", p)
	f, err := os.Open(lp)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	defer f.Close()

	var parts []string
	var offset int64 = 0
	var fileChunk = 200 * 1024 * 1024
	fileSize, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	for offset < fileSize {
		_, err := f.Seek(offset, io.SeekStart)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		remainingSize := fileSize - offset
		if remainingSize < int64(fileChunk) {
			fileChunk = int(remainingSize)
		}
		buffer := make([]byte, fileChunk)
		n, err := f.Read(buffer)
		if err != nil {
			log.Error(err)
			panic(err)
		}

		directory := "./parts"
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			err := os.MkdirAll(directory, 0755)
			if err != nil {
				log.Println("无法创建文件夹:", err)
				return
			}
		}

		// save to file, path: ./parts/xxx
		partPath := fmt.Sprintf("%s/%d", directory, offset)
		os.Create(partPath)
		err = ioutil.WriteFile(partPath, buffer[:n], 0644)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		parts = append(parts, partPath)
		offset += int64(n)
	}

	var result = make(map[int]string)
	for i, path := range parts {
		partNumber := i + 1
		f, _ = os.Open(path)
		content, _ := ioutil.ReadAll(f)
		log.Infof("📦 [COS] 上传分片 %d", partNumber)
		resp, err := cc.Object.UploadPart(context.Background(), p, UploadID, partNumber, bytes.NewReader(content), nil)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		ETag := resp.Header.Get("ETag")
		result[partNumber] = ETag
	}
	opt := &cos.CompleteMultipartUploadOptions{}
	for i := 0; i < len(result); i++ {
		opt.Parts = append(opt.Parts, cos.Object{
			PartNumber: i,
			ETag:       result[i],
		})
	}
	_, _, err = cc.Object.CompleteMultipartUpload(
		context.Background(), p, UploadID, opt,
	)
	if err != nil {
		panic(err)
	}

	// 删除分片
	log.Infof("📦 [COS] Upload Complete, Delete Parts")
	for _, path := range parts {
		err = os.Remove(path)
		if err != nil {
			log.Error(err)
			panic(err)
		}
	}
}
