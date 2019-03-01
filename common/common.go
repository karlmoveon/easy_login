package common

import (
	"fmt"
	"os"
)

/* web server const */
const (
	Success           = 0
	DefaultError      = -1
	UsernameInvalid   = -2
	UserNotFound      = -3
	UserPasswordError = -4
	TcpServerError	  = -5

	LoginHtml    = "login.html"
	UserinfoHtml = "userinfo.html"
	TokenName    = "usertoken"

	MixStr = "24d1ec97ad854a43"
	TestToken = "cf9759e6eceb1413380650b4aac0c55d"
)

/* rpc proto const */

const (
	Login      	   = "login"
	ModifyNickname = "modifyNickname"
	UploadPicture  = "uploadPicture"
	SetToken       = "setToken"
	GetUserInfo    = "getUserInfo"
)

/* server addr const */
const (
	Tcpaddr = "localhost:50000"
	Httpaddr = "localhost:9090"
	Redisaddr = "localhost:6379"
	MySQLDsn  = "root:@tcp(127.0.0.1:3306)/go_task?charset=utf8"
)

type User struct {
	Username   string
	Nickname   string
	Password   string
	PictureUrl string
	Token      string
}

/* send in json format */
type RpcReq struct {
	Method     string `json:"method"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	Password   string `json:"password"`
	PictureURL string `json:"pictureURL"`
	Token      string `json:"token"`
}

type RpcResp struct {
	RespCode   int32  `json:"code"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	PictureURL string `json:"pictureURL"`
}

/* 预计会出现无法处理的错误或无法启动服务时使用 */
func HandleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Happened fatal error: %s", err.Error())
		os.Exit(1)
	}
}