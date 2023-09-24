package core

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

var (
	Log *logrus.Logger
)

type LogFormatter struct {
}

func (t LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	//根据不同的level展示颜色
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}
	//字节缓冲区
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	//自定义日期格式
	timestamp := entry.Time.Format("2006-01-02 15:04:06")
	fmt.Fprintf(b, "[%s] \033[%dm[%s]\033[0m %s \n", timestamp, levelColor, entry.Level, entry.Message)
	return b.Bytes(), nil
}
func InitLogger() {
	mLog := logrus.New() //新建一个实例
	file, _ := os.OpenFile("gin.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	mLog.SetOutput(io.MultiWriter(os.Stdout, file)) //设置输出类型
	mLog.SetReportCaller(true)                      //开启返回函数名和行号
	//mLog.SetFormatter(&logrus.JSONFormatter{})//设置自定义的Formatter
	mLog.SetFormatter(&LogFormatter{}) //设置自定义的Formatter
	mLog.SetLevel(logrus.DebugLevel)   //设置最低等级
	Log = mLog
}
