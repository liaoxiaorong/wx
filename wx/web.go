package wx

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var contactTemp = `<div>
<p>%d %s</p>
<form action="/" method="post">
<input type="hidden" value="%s" name="userid">
<textarea id="msg" name="msg"></textarea>
<button type="submit">发送消息</button>
</form>
</div>
`

func htmlHandler(w http.ResponseWriter, r *http.Request) {
	us, err := GetContacts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if r.Method == "POST" {
		userid := r.FormValue("userid")
		msg := r.FormValue("msg")
		u, ok := us[userid]
		if ok && msg != "" {
			err = SendMsg(userid, msg)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				fmt.Fprintf(w, "send msg to %s success!", u.NickName)
			}
			return
		}
		http.Error(w, "Unknow user or empty msg", http.StatusInternalServerError)
		return
	}

	n := 1
	w.Header().Set("content-type", "text/html; charset=utf-8")
	for _, u := range us {
		if u.Sex != 0 {
			name := u.RemarkName
			if name == "" {
				name = u.NickName
			}
			fmt.Fprintf(w, contactTemp, n, name, u.UserName)
			n = n + 1
		}
	}

}

func listHandler(w http.ResponseWriter, r *http.Request) {
	us, err := GetContacts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	b, err := json.Marshal(us)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("content-type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(b))
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	us, err := GetContacts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var u *User
	var ok bool
	userid := r.FormValue("userid")
	u, ok = us[userid]
	if !ok {
		http.Error(w, "no such user", http.StatusInternalServerError)
		return
	}

	msg := r.FormValue("msg")
	if msg == "" {
		http.Error(w, "msg empty", http.StatusInternalServerError)
		return
	}
	err = SendMsg(u.UserName, msg)
	if err == nil {
		fmt.Fprintf(w, "send msg to %s success!", u.NickName)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WebServe(addr string) error {
	log.Println("web server listen:", addr)

	http.HandleFunc("/", htmlHandler)
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/send", sendHandler)
	return http.ListenAndServe(addr, nil)
}
