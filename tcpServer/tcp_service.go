package main

import (
	"go_task/common"
	"go_task/db"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
)

type Handler func(req common.RpcReq) (resp common.RpcResp)

var handlers = map[string]Handler {
	common.Login : handleLogin,
	common.ModifyNickname : handleNick,
	common.SetToken : handleSetToken,
	common.UploadPicture : handlePic,
	common.GetUserInfo : handleUserinfo,
}

func handleLogin(req common.RpcReq) (resp common.RpcResp) {

	var password string
	rows, err := db.StmtLogin.Query(req.Username)
	defer rows.Close()

	if err != nil {
		fmt.Println("query data error:%v", err)
		resp.RespCode = common.DefaultError
		return resp
	}

	if rows.Next() {
		err = rows.Scan(&password)
		if err != nil {
			log.Println("sql error: %s", err.Error())
		}
	}

	if req.Password == password {
		resp.RespCode = common.Success
	} else {
		resp.RespCode = common.DefaultError
	}

	return resp
}

func handleNick(req common.RpcReq) (resp common.RpcResp) {
	_, err := db.StmtNick.Exec(req.Nickname, req.Username)
	if err != nil {
		fmt.Println("update data error:%v", err)
		resp.RespCode = common.DefaultError
		return resp
	}

	resp.RespCode = common.Success
	return resp
}

func handleSetToken(req common.RpcReq) (resp common.RpcResp) {

	conn := db.RedisPool.Get()
	defer conn.Close()
	v, err := conn.Do("SET", req.Username, req.Token)
	if err != nil {
		fmt.Println("redis error: %s", err.Error())
		resp.RespCode = common.DefaultError
		return resp
	}

	resp.RespCode = common.Success
	return resp
}

func handlePic(req common.RpcReq) (resp common.RpcResp) {
	_, err := db.StmtPic.Exec(req.PictureUrl, req.Username)
	if err != nil {
		fmt.Println("update data error:%v", err)
		resp.RespCode = common.DefaultError
		return resp
	}

	resp.RespCode = common.Success
	return resp
}

func handleUserinfo(req common.RpcReq) (resp common.RpcResp) {

	resp.Username = req.Username

	conn := db.RedisPool.Get()
	defer conn.Close()
	token, err := redis.String(conn.Do("GET", req.Username))
	if err != nil {
		fmt.Println("redis error:", err)
		resp.RespCode = common.DefaultError
		return resp
	}

	/* todo:压测时可以使用固定token */
	if token != req.Token && token != common.TestToken {
		resp.RespCode = common.DefaultError
		return resp
	}

	/* 执行查询 */
	rows, err := db.StmtUserinfo.Query(req.Username)
	defer rows.Close()
	if !rows.Next() {

		resp.RespCode = common.DefaultError
		return resp
	}

	err = rows.Scan(&resp.Username, &resp.Nickname, &resp.PictureURL)
	if err != nil {
		log.Println("sql error: %s", err.Error())
		resp.RespCode = common.DefaultError
		return resp
	}

	resp.RespCode = common.Success
	return resp

}

func handleReq(req common.RpcReq) (resp common.RpcResp) {
	for method, handler := range handlers {
		if method == req.Method {
			fmt.Println("start handle method", method)
			resp = handler(req)
		}
	}

	return resp
}