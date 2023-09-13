package live

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
	"wq-service/cos"
)

type RoomInfo struct {
	RoomId      int    `json:"room_id"`
	Title       string `json:"title"`
	LiveTime    string `json:"live_time"`
	Description string `json:"description"`
	UserCover   string `json:"user_cover"`
}

var streamApi = "https://api.live.bilibili.com/room/v1/Room/playUrl"
var infoApi = "https://api.live.bilibili.com/room/v1/Room/get_info"
var userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"

// DoRecord
// @Description: 录制视频
// @param roomId 	房间号
func DoRecord(roomId int) {
	info := info(roomId)
	url := getVideoUrl(roomId)

	directory := "./output"
	fileName := directory + "/🎥 [" + info.Title + "] - " + info.LiveTime + ".flv"
	download(url, directory, fileName, info)
}

// download
// @Description: 下载视频
// @param url  		视频地址
// @param directory	目录
// @param fileName	文件名
func download(url string, directory string, fileName string, info RoomInfo) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("创建请求失败:", err)
		return
	}
	req.Header.Add("User-Agent", userAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("发送请求失败:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		// 文件夹不存在则创建
		err := os.MkdirAll(directory, 0755)
		if err != nil {
			fmt.Println("无法创建文件夹:", err)
			return
		}
	}
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("无法创建文件:", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	log.Infof("🎄 [直播录制已开启][%s] %s", info.LiveTime, info.Title)
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Infof("🔴 [录制已结束][%s] %s", info.LiveTime, info.Title)
	}

	// TODO: 上传到COS
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	cos.UploadLocalFile("/wq/live/["+info.LiveTime+"~"+formattedTime+
		"] ["+info.Title+"].flv", fileName)
}

// getVideoUrl
// @Description: 获取视频地址
// @param roomId	房间号
// @return string	视频地址
func getVideoUrl(roomId int) string {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", streamApi, nil)
	req.Header.Add("User-Agent", userAgent)
	q := req.URL.Query()
	q.Add("cid", strconv.Itoa(roomId))
	q.Add("quality", "4")
	req.URL.RawQuery = q.Encode()
	resp, _ := client.Do(req)

	if resp.StatusCode != 200 {
		fmt.Println("获取视频地址失败:", resp.Status)
		return ""
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)
	v := gjson.ParseBytes(body)
	u := v.Get("data.durl.0.url").String()
	return u
}

// info
// @Description: 获取房间信息
// @param roomId	房间号
// @return RoomInfo	房间信息
func info(roomId int) RoomInfo {
	var roomInfo RoomInfo
	client := &http.Client{}
	req, _ := http.NewRequest("GET", infoApi, nil)
	req.Header.Add("User-Agent", userAgent)
	q := req.URL.Query()
	q.Add("room_id", strconv.Itoa(roomId))
	req.URL.RawQuery = q.Encode()

	resp, _ := client.Do(req)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Println("获取视频地址失败:", resp.Status)
	} else {
		v := gjson.ParseBytes(body)
		d := v.Get("data").String()
		_ = json.Unmarshal([]byte(d), &roomInfo)
	}
	return roomInfo
}
