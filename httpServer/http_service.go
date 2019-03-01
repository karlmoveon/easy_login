package httpServer

import (
	"entry_task/rpc"
	"entry_task/common"
	"fmt"
	"entry_task/util"
)

func login(username, password string) (string, int) {

	token := util.CreatToken()
	req := common.RpcReq{
		Method:common.Login,
		Username:username,
		Password:password,
	}

	resp, ret := rpc.Call(req)
	if ret != 0 {
		fmt.Println(username, "login return", ret)
		return token, ret
	}

	return token, int(resp.RespCode)
}

func setToken(username, token string) int {
	req := common.RpcReq{
		Method:common.SetToken,
		Username:username,
		Token:token,
	}

	resp, ret := rpc.Call(req)
	if ret != 0 {
		fmt.Println(username, "set token return", ret)
		return ret
	}

	return int(resp.RespCode)
}

func modifyNickname(username, nickname string) int {
	req := common.RpcReq{
		Method:common.ModifyNickname,
		Username:username,
		Nickname:nickname,
	}

	resp, ret := rpc.Call(req)
	if ret != 0 {
		fmt.Println(username, "modify nickname return", ret)
		return ret
	}

	return int(resp.RespCode)
}

func uploadPic(username, pictureURL string) int {
	req := common.RpcReq{
		Method:common.UploadPicture,
		Username:username,
		PictureURL:pictureURL,
	}

	resp, ret := rpc.Call(req)
	if ret != 0 {
		fmt.Println(username, "upload picture return", ret)
		return ret
	}

	return int(resp.RespCode)
}

func getUserinfo(username, token string) (common.RpcResp, int) {
	req := common.RpcReq{
		Method:common.GetUserInfo,
		Username:username,
		Token:token,
	}

	resp, ret := rpc.Call(req)
	if ret != 0 {
		fmt.Println(username, "set token return", ret, resp)
		return resp, ret
	}

	return resp, int(resp.RespCode)
}