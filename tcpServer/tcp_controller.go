package main

import (
	"encoding/binary"
	"go_task/common"
	"go_task/db"
	"go_task/rpc"
	"fmt"
	"io"
	"net"
)

func receivePacket(conn net.Conn) ([]byte, error) {
	head := make([]byte, 4)

	_, err := io.ReadFull(conn, head[:])
	if err != nil {
		return []byte{}, err
	}

	length := binary.BigEndian.Uint32(head)
	packet := make([]byte, length)

	_, err = io.ReadFull(conn, packet)
	if err != nil {
		return []byte{}, err
	}

	return packet, err
}

func sendPacket(respBuf []byte, bufLen int, conn net.Conn) error {
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(bufLen))

	_, err := conn.Write(lenBuf)
	if err != nil {
		return err
	}

	_, err = conn.Write(respBuf)
	if err != nil {
		return err
	}

	return nil
}



func serveTcp(conn net.Conn) {

	defer conn.Close()
	for {
		reqBuf, err := receivePacket(conn)
		if err != nil {
			if err == io.EOF {
				fmt.Println("tcp server read EOF, return")
				return
			} else {
				fmt.Println("tcp server read err", err)
				break
			}
		}

		req := rpc.DecodeReq(reqBuf)
		resp := handleReq(req)

		respBuf, length := rpc.EnncodeResp(resp)
		err = sendPacket(respBuf, length, conn)
		if err != nil {
			if err == io.EOF {
				fmt.Println("tcp server write EOF, return")
				return
			} else {
				fmt.Println("tcp server write err", err)
				break
			}
		}
	}
}

func main() {
	/* 初始化数据库 */
	err := db.InitMySQL()
	common.HandleError(err)
	defer db.SqlDB.Close()

	err = db.InitRedis()
	common.HandleError(err)

	/* 创建listener */
	listener, err := net.Listen("tcp", common.Tcpaddr)
	if err != nil {
		fmt.Println("error listening", err.Error())
		return
	}

	/* 监听客户端 */
	for {
		conn, err := listener.Accept()
		common.HandleError(err)
		go serveTcp(conn)
	}
}

