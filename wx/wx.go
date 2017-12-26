package wx

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	fs                   = &FileStore{"/tmp/wx.secret"}
	LoginUri             = "https://login.weixin.qq.com"
	ErrUnknow            = errors.New("Unknow Error")
	ErrUserNotExists     = errors.New("Error User Not Exist")
	ErrNotLogin          = errors.New("Not Login")
	ErrLoginTimeout      = errors.New("Login Timeout")
	ErrWaitingForConfirm = errors.New("Waiting For Confirm")
)

type Weixin struct {
	httpClient  *Client
	secret      *wxSecret
	baseRequest *BaseRequest
	user        *User
	contacts    map[string]*User
}

func NewWeixin() *Weixin {
	return &Weixin{
		httpClient:  NewClient(),
		secret:      &wxSecret{},
		baseRequest: &BaseRequest{},
		user:        &User{},
	}
}

func (wx *Weixin) GetUser(userName string) (*User, error) {
	u, ok := wx.contacts[userName]
	if ok {
		return u, nil
	} else {
		return nil, ErrUserNotExists
	}
}

func (wx *Weixin) GetUserName(userName string) string {
	u, err := wx.GetUser(userName)
	if err != nil {
		return userName
	}
	if u.RemarkName != "" {
		return u.RemarkName
	} else {
		return u.NickName
	}
}

func (wx *Weixin) getUuid() (string, error) {
	values := &url.Values{}
	values.Set("appid", "wx782c26e4c19acffb")
	values.Set("fun", "new")
	values.Set("lang", "zh_CN")
	values.Set("_", TimestampStr())
	uri := fmt.Sprintf("%s/jslogin", LoginUri)
	b, err := wx.httpClient.Get(uri, values)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`"([\S]+)"`)
	find := re.FindStringSubmatch(string(b))
	if len(find) > 1 {
		return find[1], nil
	} else {
		return "", fmt.Errorf("get uuid error, response: %s", b)
	}
}

func (wx *Weixin) ShowQRcodeUrl(uuid string) error {
	uri := fmt.Sprintf("%s/qrcode/%s", LoginUri, uuid)
	log.Println("Please open link in browser: " + uri)
	return nil
}

func (wx *Weixin) WaitingForLoginConfirm(uuid string) (string, error) {
	re := regexp.MustCompile(`window.code=([0-9]*);`)
	tip := "1"
	for {
		values := &url.Values{}
		values.Set("uuid", uuid)
		values.Set("tip", tip)
		values.Set("_", TimestampStr())
		b, err := wx.httpClient.Get("https://login.wx.qq.com/cgi-bin/mmwebwx-bin/login", values)
		if err != nil {
			log.Printf("HTTP GET err: %s", err.Error())
			return "", err
		}
		s := string(b)
		codes := re.FindStringSubmatch(s)
		if len(codes) == 0 {
			log.Printf("find window.code failed, origin response: %s\n", s)
			return "", ErrUnknow
		} else {
			code := codes[1]
			if code == "408" {
				log.Println("login timeout, reconnecting...")
				// }else if code == "400" {
				// 	log.Println("login timeout, need refresh qrcode")
			} else if code == "201" {
				log.Println("scan success, please confirm login on your phone")
				tip = "0"
			} else if code == "200" {
				log.Println("login success")
				re := regexp.MustCompile(`window\.redirect_uri="(.*?)";`)
				us := re.FindStringSubmatch(s)
				if len(us) == 0 {
					log.Println(s)
					return "", errors.New("find redirect uri failed")
				}
				return us[1], nil
			} else {
				log.Printf("unknow window.code %s\n", code)
				return "", ErrUnknow
			}
		}
	}
	return "", nil
}

func findTicket(s string) (*ticket, error) {
	re := regexp.MustCompile(`window\.redirect_uri="(.*?)";`)
	us := re.FindStringSubmatch(s)
	if len(us) == 0 {
		log.Println(s)
		return nil, errors.New("find redirect_uri error")
	}
	u, err := url.Parse(us[1])
	if err != nil {
		return nil, err
	}
	q := u.Query()
	return &ticket{
		Ticket: q.Get("ticket"),
		Scan:   q.Get("scan"),
	}, nil
}

func (wx *Weixin) NewLoginPage(newLoginUri string) error {
	b, err := wx.httpClient.Get(newLoginUri+"&fun=new", nil)
	if err != nil {
		log.Printf("HTTP GET err: %s", err.Error())
		return err
	}
	err = xml.Unmarshal(b, wx.secret)
	if err != nil {
		log.Printf("parse wxSecret from xml failed: %v", err)
		return err
	}
	if wx.secret.Code == "0" {
		u, _ := url.Parse(newLoginUri)
		wx.secret.BaseUri = newLoginUri[:strings.LastIndex(newLoginUri, "/")]
		wx.secret.Host = u.Host
		wx.secret.DeviceID = "e" + RandNumbers(15)
		return nil
	} else {
		return errors.New("Get wxSecret Error")
	}

}

func (wx *Weixin) Init() error {
	values := &url.Values{}
	values.Set("r", TimestampStr())
	values.Set("lang", "en_US")
	values.Set("pass_ticket", wx.secret.PassTicket)
	url := fmt.Sprintf("%s/webwxinit?%s", wx.secret.BaseUri, values.Encode())
	wx.baseRequest = &BaseRequest{
		Uin:      wx.secret.Uin,
		Sid:      wx.secret.Sid,
		Skey:     wx.secret.Skey,
		DeviceID: wx.secret.DeviceID,
	}
	b, err := wx.httpClient.PostJson(url, map[string]interface{}{
		"BaseRequest": wx.baseRequest,
	})
	if err != nil {
		log.Printf("HTTP GET err: %s", err.Error())
		return err
	}
	var r InitResponse
	err = json.Unmarshal(b, &r)
	if err != nil {
		return err
	}
	if r.BaseResponse.Ret == 0 {
		wx.user = r.User
		wx.updateSyncKey(r.SyncKey)
		return nil
	}
	return fmt.Errorf("Init error: %+v", r.BaseResponse)
}

func (wx *Weixin) updateSyncKey(s *SyncKey) {
	wx.secret.SyncKey = s
	syncKeys := make([]string, s.Count)
	for n, k := range s.List {
		syncKeys[n] = fmt.Sprintf("%d_%d", k.Key, k.Val)
	}
	wx.secret.SyncKeyStr = strings.Join(syncKeys, "|")
}

func (wx *Weixin) GetNewLoginUrl() (string, error) {
	uuid, err := wx.getUuid()
	if err != nil {
		return "", err
	}
	err = wx.ShowQRcodeUrl(uuid)
	if err != nil {
		return "", err
	}
	newLoginUri, err := wx.WaitingForLoginConfirm(uuid)
	if err != nil {
		return "", err
	}
	return newLoginUri, nil
}

type syncStatus struct {
	Retcode  string
	Selector string
}

func (wx *Weixin) StatusNotify() error {
	values := &url.Values{}
	values.Set("lang", "zh_CN")
	values.Set("pass_ticket", wx.secret.PassTicket)
	url := fmt.Sprintf("%s/webwxstatusnotify?%s", wx.secret.BaseUri, values.Encode())
	b, err := wx.httpClient.PostJson(url, map[string]interface{}{
		"BaseRequest":  wx.baseRequest,
		"code":         3,
		"FromUserName": wx.user.UserName,
		"ToUserName":   wx.user.UserName,
		"ClientMsgId":  TimestampMicroSecond(),
	})
	if err != nil {
		return err
	}
	return wx.CheckCode(b, "Status Notify error")
}

func (wx *Weixin) CheckCode(b []byte, errmsg string) error {
	var r InitResponse
	err := json.Unmarshal(b, &r)
	if err != nil {
		return err
	}
	if r.BaseResponse.Ret != 0 {
		return errors.New("Status Notify error")
	}
	return nil
}

func (wx *Weixin) GetContacts() error {
	values := &url.Values{}
	values.Set("seq", "0")
	values.Set("pass_ticket", wx.secret.PassTicket)
	values.Set("skey", wx.secret.Skey)
	values.Set("r", TimestampStr())
	url := fmt.Sprintf("%s/webwxgetcontact?%s", wx.secret.BaseUri, values.Encode())
	b, err := wx.httpClient.PostJson(url, map[string]interface{}{})
	if err != nil {
		return err
	}
	var r ContactResponse
	err = json.Unmarshal(b, &r)
	if err != nil {
		return err
	}
	if r.BaseResponse.Ret != 0 {
		return errors.New("Get Contacts error")
	}
	log.Printf("update %d contacts", r.MemberCount)
	wx.contacts = make(map[string]*User, r.MemberCount)
	return wx.updateContacts(r.MemberList)
}

func (wx *Weixin) updateContacts(us []*User) error {
	for _, u := range us {
		wx.contacts[u.UserName] = u
	}
	b, err := json.Marshal(us)
	if err != nil {
		log.Println("save contacts json encode error:", err)
	}
	err = ioutil.WriteFile("wx-contacts.json", b, 0644)
	if err != nil {
		log.Println("save json write to file error:", err)
	}
	return nil
}

func (wx *Weixin) TestSyncCheck() error {
	for _, h := range []string{"webpush.", "webpush2."} {
		wx.secret.PushHost = h + wx.secret.Host
		syncStatus, err := wx.SyncCheck()
		if err == nil {
			if syncStatus.Retcode == "0" {
				return nil
			}
		}
	}
	return errors.New("Test SyncCheck error")
}

func (wx *Weixin) SyncCheck() (*syncStatus, error) {
	uri := fmt.Sprintf("https://%s/cgi-bin/mmwebwx-bin/synccheck", wx.secret.PushHost)
	values := &url.Values{}
	values.Set("r", TimestampStr())
	values.Set("sid", wx.secret.Sid)
	values.Set("uin", strconv.FormatInt(wx.secret.Uin, 10))
	values.Set("skey", wx.secret.Skey)
	values.Set("deviceid", wx.secret.DeviceID)
	values.Set("synckey", wx.secret.SyncKeyStr)
	values.Set("_", TimestampStr())

	b, err := wx.httpClient.Get(uri, values)
	if err != nil {
		return nil, err
	}
	s := string(b)
	re := regexp.MustCompile(`window.synccheck=\{retcode:"(\d+)",selector:"(\d+)"\}`)
	matchs := re.FindStringSubmatch(s)
	if len(matchs) == 0 {
		log.Println(s)
		return nil, errors.New("find Sync check code error")
	}
	syncStatus := &syncStatus{Retcode: matchs[1], Selector: matchs[2]}
	return syncStatus, nil
}

func (wx *Weixin) Sync() ([]*Message, error) {
	values := &url.Values{}
	values.Set("sid", wx.secret.Sid)
	values.Set("skey", wx.secret.Skey)
	values.Set("lang", "en_US")
	values.Set("pass_ticket", wx.secret.PassTicket)
	url := fmt.Sprintf("%s/webwxsync?%s", wx.secret.BaseUri, values.Encode())
	b, err := wx.httpClient.PostJson(url, map[string]interface{}{
		"BaseRequest": wx.baseRequest,
		"SyncKey":     wx.secret.SyncKey,
		"rr":          ^int(time.Now().Unix()) + 1,
	})
	if err != nil {
		return nil, err
	}

	var r SyncResponse
	err = json.Unmarshal(b, &r)
	if err != nil {
		return nil, err
	}
	if r.BaseResponse.Ret != 0 {
		log.Println(string(b))
		// log.Printf("%+v", r.BaseResponse)
		return nil, errors.New("sync error")
	}
	wx.updateSyncKey(r.SyncKey)
	return r.MsgList, nil
}

func (wx *Weixin) HandleMsgs(ms []*Message) {
	for _, m := range ms {
		wx.HandleMsg(m)
	}
}

func (wx *Weixin) SendMsgToMyself(msg string) error {
	return wx.SendMsg(wx.user.UserName, msg)
}

func (wx *Weixin) SendMsg(userName, msg string) error {
	values := &url.Values{}
	values.Set("pass_ticket", wx.secret.PassTicket)
	url := fmt.Sprintf("%s/webwxsendmsg?%s", wx.secret.BaseUri, values.Encode())
	msgId := fmt.Sprintf("%d%s", Timestamp()*1000, RandNumbers(4))
	b, err := wx.httpClient.PostJson(url, map[string]interface{}{
		"BaseRequest": wx.baseRequest,
		"Msg": map[string]interface{}{
			"Type":         1,
			"Content":      msg,
			"FromUserName": wx.user.UserName,
			"ToUserName":   userName,
			"LocalID":      msgId,
			"ClientMsgId":  msgId,
		},
		"Scene": 0,
	})
	if err != nil {
		return err
	}
	return wx.CheckCode(b, "发送消息失败")
}

func (wx *Weixin) HandleMsg(m *Message) {
	if m.MsgType == 1 { // 文本消息
		log.Printf("%s: %s", wx.GetUserName(m.FromUserName), m.Content)
	} else if m.MsgType == 3 { // 图片消息
	} else if m.MsgType == 34 { // 语音消息
	} else if m.MsgType == 43 { // 表情消息
	} else if m.MsgType == 47 { // 表情消息
	} else if m.MsgType == 49 { // 链接消息
	} else if m.MsgType == 51 { // 用户在手机进入某个联系人聊天界面时收到的消息
	} else {
		log.Printf("%s: MsgType: %d", wx.GetUserName(m.FromUserName), m.MsgType)
	}
}

func (wx *Weixin) Listening() error {
	err := wx.TestSyncCheck()
	if err != nil {
		return err
	}
	for {
		syncStatus, err := wx.SyncCheck()
		if err != nil {
			log.Printf("sync check error: %s", err.Error())
			time.Sleep(3 * time.Second)
			continue
		}
		if syncStatus.Retcode == "1100" {
			return errors.New("从微信客户端上登出")
		} else if syncStatus.Retcode == "1101" {
			return errors.New("从其它设备上登了网页微信")
		} else if syncStatus.Retcode == "0" {
			if syncStatus.Selector == "0" { // 无更新
				continue
			} else if syncStatus.Selector == "2" { // 有新消息
				ms, err := wx.Sync()
				if err != nil {
					log.Printf("sync err: %s", err.Error())
				}
				wx.HandleMsgs(ms)
			} else { // 可能有其他类型的消息，直接丢弃
				log.Printf("New Message, Unknow type: %+v", syncStatus)
				_, err := wx.Sync()
				if err != nil {

				}
			}
		} else if syncStatus.Retcode == "1102" {
			return fmt.Errorf("Sync Error %+v", syncStatus)
		} else {
			log.Printf("sync check Unknow Code: %+v", syncStatus)
		}
	}
}

func (wx *Weixin) Start() error {
	newLoginUri, err := wx.GetNewLoginUrl()
	if err != nil {
		return err
	}

	err = wx.NewLoginPage(newLoginUri)
	if err != nil {
		return err
	}

	err = wx.Init()
	if err != nil {
		return err
	}

	// err = wx.StatusNotify()
	// if err != nil {
	// 	return err
	// }

	err = wx.GetContacts()
	if err != nil {
		return err
	}
	return wx.Listening()
}
