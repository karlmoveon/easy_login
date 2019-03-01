package main

import (
	"crypto/md5"
	"go_task/common"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"
)

type profile struct {
	username string
	nickname string
	picture  string
}

var templates = template.Must(template.ParseFiles(common.LoginHtml, common.UserinfoHtml))

func isInvalidUser(username, password string) bool {
	if ok, _ := regexp.MatchString("^[a-zA-Z0-9]{4,16}$", username); !ok {
		return false
	}

	if ok, _ := regexp.MatchString("^[a-zA-Z0-9]{4,16}$", password); !ok {
		return false
	}

	return true
}

func encryptPwd(originPwd string) string {
	w := md5.New()
	io.WriteString(w, originPwd)
	password := fmt.Sprintf("%x", w.Sum(nil)) + common.MixStr
	return password
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		t, _ := template.ParseFiles(common.LoginHtml)
		t.Execute(w, nil)
	} else {
		username := r.FormValue("username")
		originPwd := r.FormValue("password")
		if !isInvalidUser(username, originPwd) {
			_, _ = fmt.Fprintf(w,"username or password invalid")
			return
		}
		password := encryptPwd(originPwd)

		token, result := login(username, password)
		if result != common.Success {
			_, _ = fmt.Fprintf(w,"login failed, ret:", result)
			return
		}

		/* 设置cookie，然后重定向，下个页面中从redis中读取对比token */
		cookie := http.Cookie{
			Name:     common.TokenName,
			Value:    token,
			HttpOnly: true,
			Expires:  time.Now().Add(time.Second * 180),
		}

		http.SetCookie(w, &cookie)
		setToken(username, token)
		/* todo:根据压测结果决定是否重定向 */
		w.Write([]byte("<!DOCTYPE html><html><body><a href=\"/userinfo\">user_info</a></body></html>"))
		//http.Redirect(w, r, "/userinfo", http.StatusFound)
	}
}

func userinfoHandler(w http.ResponseWriter, r *http.Request) {
	/* 读不到cookie认为未登录，跳转回登录页面  */
	cookie, err := r.Cookie(common.TokenName)
	if err != nil {
		fmt.Println("read cookie error", err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	token := cookie.Value
	/* todo:username从html中解析出来 */
	username := r.FormValue("username")
	resp, ret := getUserinfo(username, token)
	if ret != common.Success {
		fmt.Println(w,"get userinfo failed, ret:", ret)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	p := &profile{resp.Username, resp.Nickname, resp.PictureURL}
	renderTemplate(w, common.UserinfoHtml, p)
}

func modifyNicknameHandler(w http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	nickname := r.FormValue("nickname_submit")
	ret := modifyNickname(username, nickname)
	if ret != common.Success {
		fmt.Println(w,"modify nickname failed, ret:", ret)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	
	http.Redirect(w, r, "/userinfo", http.StatusFound)
}

func uploadPicHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	pictureURL := r.FormValue("picture_submit")

	ret := uploadPic(username, pictureURL)
	if ret != common.Success {
		fmt.Println(w,"upload pic failed, ret:", ret)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/userinfo", http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *profile) {
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
	http.HandleFunc("/userinfo", userinfoHandler)
	http.HandleFunc("/modifyNickname", modifyNicknameHandler)
	http.HandleFunc("/uploadpic", uploadPicHandler)
	err := http.ListenAndServe(common.Httpaddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	fmt.Printf("server started……")
}



