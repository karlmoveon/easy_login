package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"entry_task/adapter"
	"entry_task/util"
	"fmt"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"net"
	"time"
)

/*
	本文件处理内容
   	1.main函数监听rpc client，为每个rpc client起一个goroutine去处理
   	2.goroutine处理函数中进行消息解析、消息分发处理，然后写socket，返回数据库为操作数据库处理结果
*/

var (
	sqlDb         *sql.DB
	redisPool     *redis.Pool
	redisPassword = ""
	stmt1         *sql.Stmt
	stmt2         *sql.Stmt
	stmt3         *sql.Stmt
	stmt4         *sql.Stmt
	testtoken     = "cf9759e6eceb1413380650b4aac0c55d-1"
	localDsn      = "root:@tcp(127.0.0.1:3306)/go_task?charset=utf8"
	remoteDsn     = "root:mysql.3306@tcp(127.0.0.1:3306)/go_task?charset=utf8"
)

func checkUser(user adapter.User) adapter.RpcResp {
	/* 1.访问db */
	var resp adapter.RpcResp
	fmt.Println("-- tcp_server checkUser in")

	var password string
	rows, err := stmt1.Query(user.Username)
	defer rows.Close()

	if err != nil {
		fmt.Println("query data error:%v", err)
		resp.Header.RespCode = adapter.DefaultError
		return resp
	}

	if rows.Next() {
		err = rows.Scan(&password)
		if err != nil {
			log.Println("sql error: %s", err.Error())
		}
	}

	if user.Password == password {
		resp.Header.RespCode = adapter.Success
	} else {
		resp.Header.RespCode = adapter.DefaultError
	}

	return resp
}

func newPool() *redis.Pool {
	return &redis.Pool{
		MaxActive:   10000,
		MaxIdle:     10000,
		IdleTimeout: 120 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				return nil, err
			}

			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func searchUser(user adapter.User) adapter.RpcResp {
	/* 打开db */
	var resp adapter.RpcResp
	//fmt.Println("-- tcp_server searchUser in")

	/* prepare参数绑定 */
	stmt, err := sqlDb.Prepare("select * from user where username = ?")
	if err != nil {
		fmt.Println("prepare failed:%v", err)
	}
	defer stmt.Close()

	/* 执行查询 */
	rows, err := stmt.Query(user.Username)
	defer rows.Close()
	defer stmt.Close()

	if rows.Next() {
		fmt.Println("found user:", user.Username)
		resp.Header.RespCode = adapter.Success
		return resp
	}

	fmt.Println("not found user:", user.Username)
	resp.Header.RespCode = adapter.DefaultError
	return resp
}

func changeNickname(user adapter.User) adapter.RpcResp {
	/* 1.访问db */
	var resp adapter.RpcResp
	//fmt.Println("-- tcp_server changeNickname in")

	_, err := stmt2.Exec(user.Nickname, user.Username)
	if err != nil {
		fmt.Println("update data error:%v", err)
		resp.Header.RespCode = adapter.DefaultError
		return resp
	}

	resp.Header.RespCode = adapter.Success
	return resp
}

func setToken(user adapter.User) adapter.RpcResp {
	/* 通过redis保存username和token对应关系 */
	var resp adapter.RpcResp

	conn := redisPool.Get()
	defer conn.Close()
	v, err := conn.Do("SET", user.Username, user.Token)
	if err != nil {
		log.Println("redis error: %s", err.Error())
		resp.Header.RespCode = adapter.DefaultError
		return resp
	}
	fmt.Println(v)

	resp.Header.RespCode = adapter.Success
	return resp
}

func getUserInfo(user adapter.User) adapter.RpcResp {

	var resp adapter.RpcResp
	var userGot adapter.User

	user.Username = util.GetUsernameFromToken(user.Token)

	conn := redisPool.Get()
	defer conn.Close()
	token, err := redis.String(conn.Do("GET", user.Username))
	if err != nil {
		fmt.Println("redis error:", err)
		resp.Header.RespCode = adapter.DefaultError
		return resp
	}

	/* todo:压测时可以使用固定token */
	if token != user.Token && token != testtoken {
		resp.Header.RespCode = adapter.DefaultError
		return resp
	}

	/* 执行查询 */
	rows, err := stmt3.Query(user.Username)
	defer rows.Close()
	if rows.Next() {
		fmt.Println("found user:", user.Username)
		err = rows.Scan(&userGot.Username, &userGot.Nickname, &userGot.PictureUrl)
		if err != nil {
			log.Println("sql error: %s", err.Error())
		}
		data, err := json.Marshal(userGot)
		if err != nil {
			log.Println("json error: %s", err.Error())
		}
		resp.Header.RespLen = int32(len(data))
		resp.RespInfo = string(data)
		resp.Header.RespCode = adapter.Success
		return resp
	}

	resp.Header.RespCode = adapter.DefaultError
	return resp
}

func uploadPicture(user adapter.User) adapter.RpcResp {
	/* 1.访问db */
	var resp adapter.RpcResp
	//fmt.Println("-- tcp_server uploadPicture in")

	_, err := stmt4.Exec(user.PictureUrl, user.Username)
	if err != nil {
		fmt.Println("update data error:%v", err)
		resp.Header.RespCode = adapter.DefaultError
		return resp
	}

	resp.Header.RespCode = adapter.Success
	return resp
}

func handleMethod(body adapter.RpcBody) adapter.RpcResp {
	var (
		resp adapter.RpcResp
		user adapter.User
	)

	/* 解析出userinfo后处理 */
	err := json.Unmarshal([]byte(body.Args), &user)
	if err != nil {
		log.Println("json error: %s", err.Error())
	}

	switch body.Method {
	case adapter.CheckUser:
		{
			resp = checkUser(user)
		}

	case adapter.SearchUser:
		{
			resp = searchUser(user)
		}

	case adapter.ChangeNickname:
		{
			resp = changeNickname(user)
		}

	case adapter.UploadPicture:
		{
			resp = uploadPicture(user)
		}

	case adapter.SetToken:
		{
			resp = setToken(user)
		}

	case adapter.GetUserInfo:
		{
			resp = getUserInfo(user)
		}

	default:
		{
			fmt.Println("incorrect method", body.Method)
			resp.Header.RespCode = adapter.InternalError
		}
	}

	return resp
}

func serveTcp(conn net.Conn) {
	var (
		rpcReqHeader        adapter.RpcHeader
		rpcReqHeaderBuf     = make([]byte, adapter.RpcHeadSize)
		rpcRespHeaderBinBuf = bytes.NewBuffer(make([]byte, 0, adapter.RpcRespHeaderSize))
	)

	defer conn.Close()
	for {
		/* read req header */
		readLen, err := conn.Read(rpcReqHeaderBuf)
		fmt.Println("tcp read req header", readLen, "bytes from client", rpcReqHeaderBuf)
		if err == io.EOF {
			fmt.Println("tcp read EOF 0 , conn", conn.RemoteAddr().String(), "closed")
			return
		}
		checkError(err)

		err = binary.Read(bytes.NewBuffer(rpcReqHeaderBuf), binary.BigEndian, &(rpcReqHeader))
		checkError(err)

		rpcReqBodyLen := int(rpcReqHeader.MethodLen + rpcReqHeader.ArgsLen)

		/* read req body */
		rpcReqBodyBuf := make([]byte, rpcReqBodyLen)
		readLen, err = conn.Read(rpcReqBodyBuf)
		fmt.Println("tcp read req body", readLen, "bytes from client", rpcReqBodyBuf)
		if err == io.EOF {
			fmt.Println("tcp read EOF 1, conn", conn.RemoteAddr().String(), "closed")
			return
		}
		checkError(err)
		rpcReqBody := adapter.Decode(rpcReqBodyBuf, rpcReqHeader.MethodLen, rpcReqHeader.ArgsLen)

		/* get resp */
		resp := handleMethod(rpcReqBody)
		resp.Header.Seq = rpcReqHeader.Seq
		err = binary.Write(rpcRespHeaderBinBuf, binary.BigEndian, resp.Header)
		checkError(err)

		/* write resp header */
		writeLen, err := conn.Write(rpcRespHeaderBinBuf.Bytes())
		fmt.Println("tcp write resp header", writeLen, "bytes to client", resp.Header)
		if err != nil {
			fmt.Println("tcp write 0 error")
		}
		checkError(err)

		/* write resp body */
		if len(resp.RespInfo) != 0 {
			writeLen, err = conn.Write([]byte(resp.RespInfo))
			fmt.Println("tcp write resp body", writeLen, "bytes to client", resp.RespInfo)
			if err != nil {
				fmt.Println("tcp Write 1 error")
			}
			checkError(err)
		}
	}
}

func main() {
	fmt.Println("-- starting tcp server --")

	log.SetPrefix("[tcp_server]:")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)

	/* 创建listener */
	listener, err := net.Listen("tcp", "localhost:50000")
	if err != nil {
		fmt.Println("error listening", err.Error())
		return
	}

	/* 打开mysql和redis */
	sqlDb, err = sql.Open("mysql", localDsn)
	if err != nil {
		fmt.Println("mysql db open failed")
		return
	}
	sqlDb.SetMaxOpenConns(10000)
	sqlDb.SetMaxIdleConns(10000)
	defer sqlDb.Close()

	/* prepare参数绑定 */
	stmt1, err = sqlDb.Prepare("select password from user where username = ?")
	if err != nil {
		fmt.Println("prepare failed:%v", err)
	}

	/* prepare参数绑定 */
	stmt2, err = sqlDb.Prepare("update user set nickname = ?  where username = ?")
	if err != nil {
		fmt.Println("prepare failed:%v", err)
	}

	/* prepare参数绑定 */
	stmt3, err = sqlDb.Prepare("select username, nickname, pictureurl from user where username = ?")
	if err != nil {
		fmt.Println("prepare failed:%v", err)
	}

	/* prepare参数绑定 */
	stmt4, err = sqlDb.Prepare("update user set pictureurl = ?  where username = ?")
	if err != nil {
		fmt.Println("prepare failed:%v", err)
	}

	redisPool = newPool()

	/* 监听客户端 */
	for {
		conn, err := listener.Accept()
		checkError(err)
		go serveTcp(conn)

	}
}

func checkError(err error) {
	if err == nil {
		return
	}

	fmt.Println("TCP server Fatal error: %s", err.Error())
}
