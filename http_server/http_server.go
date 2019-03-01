package main

import (
	"crypto/md5"
	"encoding/json"
	"entry_task/adapter"
	"entry_task/util"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	Success           = 0
	UsernameInvalid   = -1
	UserNotFound      = -2
	UserPasswordError = -3
	UnknownError      = -4

	LoginHtml    = "login.html"
	UserinfoHtml = "userinfo.html"
	TokenName    = "usertoken"
)

type Profile struct {
	Username string
	Nickname string
	Picture  string
}

var templates = template.Must(template.ParseFiles(LoginHtml, UserinfoHtml))

func handleUserInfo(r *http.Request) (adapter.User, int32) {
	/* 1.解析出用户名密码 */
	var user adapter.User

	user.Username = r.FormValue("username")
	originPwd := r.FormValue("password")
	w := md5.New()
	io.WriteString(w, originPwd+"karl")
	user.Password = fmt.Sprintf("%x", w.Sum(nil))

	/* 2.rpccall checkUser(User) */
	resp, errno := adapter.RpcCall(adapter.CheckUser, int32(len(adapter.CheckUser)), user)
	checkRpcError(errno)
	if resp.Header.RespCode != 0 {
		fmt.Println("checkUser failed, errno:", resp.Header.RespCode)
		return user, UserPasswordError
	}

	return user, resp.Header.RespCode
}

func setToken(user adapter.User) {
	resp, errno := adapter.RpcCall(adapter.SetToken, int32(len(adapter.SetToken)), user)
	checkRpcError(errno)
	if resp.Header.RespCode != 0 {
		fmt.Println("SaveToken failed, errno:", resp.Header.RespCode)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-- loginHandler trace in --, handle ", r.URL.Path, r.Method)

	if r.Method == http.MethodGet {
		t, _ := template.ParseFiles(LoginHtml)
		t.Execute(w, nil)
	} else {
		user, errno := handleUserInfo(r)
		switch errno {
		default:
			{
				fmt.Println("user login succeed: ", user.Username)
				/* 计算token，设置cookie，然后重定向，下个页面中从redis中读取对比token */
				token := util.CreatToken(user.Username)
				cookie := http.Cookie{
					Name:     TokenName,
					Value:    token,
					HttpOnly: true,
					Expires:  time.Now().Add(time.Second * 180),
				}

				http.SetCookie(w, &cookie)
				user.Token = token
				setToken(user)
				w.Write([]byte("<!DOCTYPE html><html><body><a href=\"/userinfo\">user_info</a></body></html>"))
				//http.Redirect(w, r, "/userinfo", http.StatusFound)
			}

		case UserNotFound:
			{
				fmt.Println("user not found")
				http.Redirect(w, r, "/usernotfound", http.StatusFound)
			}

		case UserPasswordError:
			{
				fmt.Println("password incorrect")
				http.Redirect(w, r, "/passworderror", http.StatusFound)
			}
		}
	}
}

func userinfoHandler(w http.ResponseWriter, r *http.Request) {
	/* 读不到cookie认为未登录，跳转回登录页面  */
	fmt.Println("-- userinfoHandler trace in --, handle ", r.URL.Path, r.Method)
	cookie, err := r.Cookie(TokenName)
	if err != nil {
		fmt.Println("read cookie error", err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	var user adapter.User
	user.Token = cookie.Value
	resp, errno := adapter.RpcCall(adapter.GetUserInfo, int32(len(adapter.GetUserInfo)), user)
	checkRpcError(errno)
	if resp.Header.RespCode != 0 {
		fmt.Println("Token auth failed, errno:", resp.Header.RespCode)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	err = json.Unmarshal([]byte(resp.RespInfo), &user)
	checkError(err)
	fmt.Println(user.Username, user.Nickname, user.PictureUrl)
	p := &Profile{user.Username, user.Nickname, user.PictureUrl}
	renderTemplate(w, UserinfoHtml, p)
}

func editNicknameHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-- editNicknameHandler trace in --, handle ", r.URL.Path, r.Method)
	/* 调用rpc修改完成后，重定向到userinfo */
	var user adapter.User
	user.Username = r.FormValue("username")
	user.Nickname = r.FormValue("nickname_submit")
	resp, errno := adapter.RpcCall(adapter.ChangeNickname, int32(len(adapter.ChangeNickname)), user)
	checkRpcError(errno)
	if resp.Header.RespCode != 0 {
		fmt.Println("ChangeNickname failed, errno:", resp.Header.RespCode)
		http.Redirect(w, r, "/userinfo", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/userinfo", http.StatusFound)
}

func uploadPicHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-- uploadPicHandler trace in --, handle ", r.URL.Path, r.Method)
	/* 调用rpc修改完成后，重定向到userinfo */
	var user adapter.User
	user.Username = r.FormValue("username")
	user.PictureUrl = r.FormValue("picture_submit")

	resp, errno := adapter.RpcCall(adapter.UploadPicture, int32(len(adapter.UploadPicture)), user)
	checkRpcError(errno)
	if resp.Header.RespCode != 0 {
		fmt.Println("UploadPicture failed, errno:", resp.Header.RespCode)
		http.Redirect(w, r, "/userinfo", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/userinfo", http.StatusFound)
}

func userNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("usernotfound.html")
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, nil)
}

func passwordErrorHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("passworderror.html")
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Profile) {
	fmt.Println("render html:", tmpl)
	err := templates.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	fmt.Println("-- starting http server --")

	http.HandleFunc("/", loginHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/usernotfound", userNotFoundHandler)
	http.HandleFunc("/passworderror", passwordErrorHandler)
	http.HandleFunc("/userinfo", userinfoHandler)
	http.HandleFunc("/editnickname", editNicknameHandler)
	http.HandleFunc("/uploadpic", uploadPicHandler)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	fmt.Printf("server start at ")
}

func checkRpcError(err int32) {
	if err != 0 {
		fmt.Println("RPC HANDLE ERROR")
		os.Exit(1)
	}
}

/* 预计会出现无法处理的错误时使用 */
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "HTTP server Fatal error: %s", err.Error())
	}
}
