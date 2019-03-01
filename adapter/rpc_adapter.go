package adapter

/* rpc中间层封装数据 */

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

const (
	RpcHeadSize       = 12
	RpcBufSize        = 256
	RpcRespSize       = 512
	RpcRespHeaderSize = 12
)

const (
	CheckUser      = "checkUser"
	SearchUser     = "searchUser"
	ChangeNickname = "changeNickame"
	UploadPicture  = "uploadPicture"
	SetToken       = "setToken"
	GetUserInfo    = "getUserInfo"
)

const (
	Success        int32 = 0
	DefaultError   int32 = -1
	InternalError  int32 = -2
	ServerError    int32 = -3
	MethodNotFound int32 = -4
)

/* todo：seq是否需要锁保护 */
var rpcSeqNum int32

type User struct {
	Username   string
	Nickname   string
	Password   string
	PictureUrl string
	Token      string
}

type RpcHeader struct {
	Seq       int32
	MethodLen int32
	ArgsLen   int32
}

type RpcBody struct {
	Method string
	Args   string
}

type RpcReq struct {
	Header RpcHeader
	Body   RpcBody
}

type RpcRespHeader struct {
	Seq      int32
	RespCode int32
	RespLen  int32
}

type RpcResp struct {
	Header   RpcRespHeader
	RespInfo string
}

/* 请求与响应都需要解包组包 */

func Encode(header RpcHeader, body RpcBody) ([]byte, uint32) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, (int32)(header.Seq))
	checkError(err)
	length := 4

	err = binary.Write(buf, binary.BigEndian, (int32)(header.MethodLen))
	checkError(err)
	length += 4

	err = binary.Write(buf, binary.BigEndian, (int32)(header.ArgsLen))
	checkError(err)
	length += 4

	err = binary.Write(buf, binary.BigEndian, []byte(body.Method))
	checkError(err)
	length += len(body.Method)

	err = binary.Write(buf, binary.BigEndian, []byte(body.Args))
	checkError(err)
	length += len(body.Args)

	fmt.Println("--------------rpc debug body", body, "buff: ", buf)
	return buf.Bytes(), uint32(length)
}

/* decode中只解body部分 */
func Decode(buff []byte, methodLen, argsLen int32) RpcBody {

	var body RpcBody

	methodBuf := buff[:methodLen]
	ArgsBuf := buff[methodLen:]

	body.Method = (string)(methodBuf)
	body.Args = (string)(ArgsBuf)

	return body
}

func RpcCall(method string, methodLen int32, user User) (RpcResp, int32) {
	/* 1.数据组包 */
	rpcSeqNum++
	if rpcSeqNum >= 65535 {
		rpcSeqNum = 0
	}

	var (
		body                          RpcBody
		header                        RpcHeader
		resp                          RpcResp
		n                             int
		bytesRead, bytesWrite, length uint32
	)

	header.Seq = rpcSeqNum
	header.MethodLen = methodLen

	data, err := json.Marshal(user)
	checkError(err)
	header.ArgsLen = int32(len(data))

	body.Args = string(data)
	body.Method = method
	fmt.Println("--------------rpc debug" + user.Username + "before Encode")
	/* buf为要发送给tcp server的字节流 */
	buf, length := Encode(header, body)

	fmt.Println("--------------rpc debug" + user.Username + "before conn tcp")
	/* 2.连接tcp server */
	conn, err := net.Dial("tcp", "localhost:50000")
	defer conn.Close()

	if err != nil {
		fmt.Println("-- RPC : Error dialing", err.Error())
		return resp, ServerError
	}
	fmt.Println("--------------rpc debug" + user.Username + "before write tcp")
	/* 3.发送buf */
	for bytesWrite < length {
		n, err = conn.Write(buf)
		if err != nil {
			fmt.Println("-- RPC : Error sending msg to tcp server", err.Error())
			return resp, ServerError
		}

		bytesWrite += uint32(n)
	}
	fmt.Println("--------------rpc debug" + user.Username + "after write tcp")
	fmt.Println("-- RPC : write", bytesWrite, "data to tcp server", buf)

	/* 4.读响应, 解包,RespInfo为json */
	headerBuf := make([]byte, RpcRespHeaderSize)
	for bytesRead < uint32(len(headerBuf)) {
		n, err := conn.Read(headerBuf)
		if err == io.EOF {
			fmt.Println("rpc read EOF 0, conn", conn.RemoteAddr().String(), "closed")
			return resp, DefaultError
		}
		checkError(err)

		bytesRead += uint32(n)
	}

	buff := bytes.NewBuffer(headerBuf)

	err = binary.Read(buff, binary.BigEndian, &(resp.Header.Seq))
	checkError(err)

	err = binary.Read(buff, binary.BigEndian, &(resp.Header.RespCode))
	checkError(err)

	err = binary.Read(buff, binary.BigEndian, &(resp.Header.RespLen))
	checkError(err)

	fmt.Println("-- RPC : read header", bytesRead, "bytes from tcp server", headerBuf)

	bodyBuf := make([]byte, resp.Header.RespLen)

	bytesRead = 0
	for bytesRead < uint32(len(bodyBuf)) {
		n, err := conn.Read(bodyBuf)
		if err == io.EOF {
			fmt.Println("rpc read EOF 0, conn", conn.RemoteAddr().String(), "closed")
			return resp, DefaultError
		}
		checkError(err)

		bytesRead += uint32(n)
	}

	fmt.Println("-- RPC : read body", bytesRead, len(bodyBuf), "bytes from tcp server", bodyBuf)
	resp.RespInfo = string(bodyBuf)
	checkError(err)

	fmt.Println("--------------rpc debug", user.Username, "req header:", header, "Body:", body, "resp:", resp)

	return resp, Success
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "-- RPC Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
