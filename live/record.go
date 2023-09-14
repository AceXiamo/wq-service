package live

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"wq-service/cos"
)

const (
	streamApi = "https://api.live.bilibili.com/room/v1/Room/playUrl"
	infoApi   = "https://api.live.bilibili.com/room/v1/Room/get_info"
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
)

var reconnectMax = 5

type RoomInfo struct {
	RoomId      int    `json:"room_id"`
	Title       string `json:"title"`
	LiveTime    string `json:"live_time"`
	Description string `json:"description"`
	UserCover   string `json:"user_cover"`
	LiveStatus  int    `json:"live_status"`
}

type DownloadInfo struct {
	Url       string
	Directory string
	FileName  string
	RoomInfo  RoomInfo
}

var lck sync.Mutex

// DoRecord
// @Description: 录制视频
// @param roomId 	房间号
func DoRecord(roomId int) {
	lck.Lock()
	defer lck.Unlock()

	log.Infof("🟢 [直播开始] %d", roomId)
	AsyncFun(roomId)
}

func AsyncFun(roomId int) {
	info := getInfo(roomId)
	if info.LiveStatus == 1 {
		url := getVideoUrl(roomId)
		directory := "./output"
		fileName := getFormattedFileName(directory, info)
		downloadInfo := DownloadInfo{
			Url:       url,
			Directory: directory,
			FileName:  fileName,
			RoomInfo:  info,
		}
		go download(downloadInfo)
	} else {
		log.Info("🔴 [录制已结束]")
	}
}

// download
// @Description: 下载视频
// @param downloadInfo	下载信息
func download(downloadInfo DownloadInfo) {
	req, err := http.NewRequest("GET", downloadInfo.Url, nil)
	if err != nil {
		log.Println("创建请求失败:", err)
		return
	}
	req.Header.Add("User-Agent", userAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("发送请求失败:", err)
		return
	}
	defer resp.Body.Close()

	if _, err := os.Stat(downloadInfo.Directory); os.IsNotExist(err) {
		err := os.MkdirAll(downloadInfo.Directory, 0755)
		if err != nil {
			log.Println("无法创建文件夹:", err)
			return
		}
	}
	file, err := os.Create(downloadInfo.FileName)
	if err != nil {
		log.Println("无法创建文件:", err)
		return
	}
	defer file.Close()

	log.Infof("🎄 [直播录制已开启][%s] %s", downloadInfo.RoomInfo.LiveTime, downloadInfo.RoomInfo.Title)
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Infof(err.Error())
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")
	startTime, _ := time.Parse("2006-01-02 15:04:05", downloadInfo.RoomInfo.LiveTime)
	endTime, _ := time.Parse("2006-01-02 15:04:05", formattedTime)
	if endTime.Sub(startTime).Seconds() < 60 && reconnectMax > 0 {
		reconnectMax--
		AsyncFun(downloadInfo.RoomInfo.RoomId)
	} else {
		log.Infof("🔴 [录制已结束][%s] %s", downloadInfo.RoomInfo.LiveTime, downloadInfo.RoomInfo.Title)
		cos.MultipartUpload(getFormattedCosFileName(downloadInfo.RoomInfo.LiveTime, formattedTime, downloadInfo.RoomInfo.Title), downloadInfo.FileName)

		// 删除本地文件
		os.Remove(downloadInfo.FileName)
	}
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
		log.Println("获取视频地址失败:", resp.Status)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	v := gjson.ParseBytes(body)
	u := v.Get("data.durl.0.url").String()
	return u
}

// getInfo
// @Description: 获取房间信息
// @param roomId	房间号
// @return RoomInfo	房间信息
func getInfo(roomId int) RoomInfo {
	var roomInfo RoomInfo
	client := &http.Client{}
	req, _ := http.NewRequest("GET", infoApi, nil)
	req.Header.Add("User-Agent", userAgent)
	q := req.URL.Query()
	q.Add("room_id", strconv.Itoa(roomId))
	req.URL.RawQuery = q.Encode()

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Println("获取视频地址失败:", resp.Status)
	} else {
		v := gjson.ParseBytes(body)
		d := v.Get("data").String()
		_ = json.Unmarshal([]byte(d), &roomInfo)
	}
	return roomInfo
}

// getFormattedFileName
// @Description: 格式化文件名
// @param directory	目录
// @param info		房间信息
// @return string	格式化后的文件名
func getFormattedFileName(directory string, info RoomInfo) string {
	fileName := directory + "/[" + info.Title + "]_" + info.LiveTime + ".flv"
	fileName = strings.ReplaceAll(fileName, " ", "_")
	fileName = strings.ReplaceAll(fileName, ":", "_")
	return fileName
}

// getFormattedCosFileName
// @Description: 格式化COS存储的文件名
// @param startTime	直播开始时间
// @param endTime	直播结束时间
// @param title		直播标题
// @return string	格式化后的文件名
func getFormattedCosFileName(startTime, endTime, title string) string {
	return "/wq/live/[" + startTime + "~" + endTime + "] [" + title + "].flv"
}
