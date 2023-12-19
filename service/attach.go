package service

import (
	"fmt"
	"ginchat/utils"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Upload(c *gin.Context) {
	w := c.Writer
	req := c.Request
	srcFile, head, err := req.FormFile("file") //获取上传文件
	if err != nil {
		fmt.Println(w, err.Error())
	}
	suffix := ".png"          //默认文件后缀
	ofilName := head.Filename //获取文件名
	tem := strings.Split(ofilName, ".")
	if len(tem) > 1 {
		suffix = "." + tem[len(tem)-1] //获取文件后缀
	}
	fileName := fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Int31(), suffix)
	dstFile, err := os.Create("F:/BaiduNetdiskDownload/ginchat/asset/upload/" + fileName)
	if err != nil {
		fmt.Println(w, err.Error())
	}
	_, err = io.Copy(dstFile, srcFile) //复制文件
	if err != nil {
		utils.RespFail(w, err.Error())
	}
	url := "./assets/upload/" + fileName
	utils.RespOK(w, url, "发送图片success")
}
