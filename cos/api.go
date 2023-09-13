package cos

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type ListFileRes struct {
	Items []FileItem `json:"items"`
	Next  string     `json:"next"`
	Has   bool       `json:"has"`
	Size  int64      `json:"size"`
}

type FileItem struct {
	Key  string `json:"key"`
	Size int64  `json:"size"`
}

// GetCount
// @Description: GetCount
// @param p  			文件夹路径
// @return int			文件数量
func GetCount(p string) int {
	return len(ListFile(p, "", 1000).Items)
}

// ListFile
// @Description: ListFile
// @param p			文件夹路径
// @param nextMarker	下一页标识
// @return []FileItem	文件列表
// @return string		下一页标识
// @return bool			是否有下一页
func ListFile(p string, nextMarker string, size int) ListFileRes {
	var r []FileItem
	opt := &cos.BucketGetOptions{
		Prefix:    p,
		Delimiter: "/",
		MaxKeys:   size,
	}
	opt.Marker = nextMarker
	v, _, err := cc.Bucket.Get(context.Background(), opt)
	if err != nil {
		log.Error(err)
	}
	for _, content := range v.Contents {
		r = append(r, FileItem{
			Key:  content.Key,
			Size: content.Size,
		})
	}

	return ListFileRes{
		Items: r,
		Next:  v.NextMarker,
		Has:   v.IsTruncated,
	}
}
