package rpc

import (
	"encoding/binary"
	"encoding/json"
	"go_task/common"
	"fmt"
	"io"
	"net"
	"os"
)

func Call(req common.RpcReq) (common.RpcResp, int){
	reqBytes, length := EncodeReq(req)

	var resp common.RpcResp

	/* 连接tcp server 发送数据 todo:用tcp pool处理, 多个客户端连接 */
	conn, err := net.Dial("tcp", common.Tcpaddr)
	defer conn.Close()
	if err != nil {
		fmt.Println("-- RPC : Error dialing", err.Error())
		return resp, common.TcpServerError
	}

	lenBuf := make([]byte, 4)

	binary.BigEndian.PutUint32(lenBuf, uint32(length))
	_, err = conn.Write(lenBuf)
	n, err := conn.Write(reqBytes)
	if n != length {
		fmt.Println("-- RPC : write length error")
		return resp, common.TcpServerError
	}

	lenBuf = make([]byte, 4)
	n, err = conn.Read(lenBuf)
	if err == io.EOF {
		fmt.Println("-- RPC : Read EOF from conn", conn.RemoteAddr().String())
		return resp, common.TcpServerError
	}

	respLen := binary.BigEndian.Uint32(lenBuf)
	respBuf := make([]byte, respLen)
	n, err = conn.Read(respBuf)
	if err == io.EOF {
		fmt.Println("-- RPC : Read EOF from conn", conn.RemoteAddr().String())
		return resp, common.TcpServerError
	}

	resp = DecodeResp(respBuf)
	return resp, common.Success
}

func EnncodeResp(resp common.RpcResp) ([]byte, int) {

	data, err := json.Marshal(resp)
	checkError(err)
	return data, len(data)
}

func EncodeReq(req common.RpcReq) ([]byte, int) {

	data, err := json.Marshal(req)
	checkError(err)
	return data, len(data)
}

func DecodeReq(reqBytes []byte) (req common.RpcReq) {

	err := json.Unmarshal(reqBytes, req)
	checkError(err)
	return req
}

func DecodeResp(respBytes []byte) (resp common.RpcResp) {

	err := json.Unmarshal(respBytes, resp)
	checkError(err)
	return resp
}


func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "-- RPC Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

