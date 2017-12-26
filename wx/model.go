package wx

type wxSecret struct {
	Host       string
	BaseUri    string
	PushHost   string
	DeviceID   string
	SyncKeyStr string
	Uin        int64  `xml:"wxuin"`
	Sid        string `xml:"wxsid"`
	Skey       string `xml:"skey"`
	PassTicket string `xml:"pass_ticket"`
	Code       string `xml:"ret"`
	SyncKey    *SyncKey
}

type ticket struct {
	Ticket string
	Scan   string
}

type User struct {
	Uin         int64  `json:"Uin"`         // 777252808,
	UserName    string `json:"UserName"`    // userid
	NickName    string `json:"NickName"`    // 昵称,
	HeadImgUrl  string `json:"HeadImgUrl"`  // "/cgi-bin/mmwebwx-bin/webwxgeticon?seq=1092670198&username=@78c2c1f76b86e1db6a1628f9eeae0f398c1c8a9d49a486ba4f1817c43347218a&skey=@crypt_e9f8e332_78afd5f296b76c1768bcc737088f33e6",
	RemarkName  string `json:"RemarkName"`  // 备注名,
	Sex         int64  `json:"Sex"`         // 1:man, 2:female
	Signature   string `json:"Signature"`   // 个性签名
	ContactFlag int64  `json:"ContactFlag"` // 1:公众号，65537:个人号
}

type BaseRequest struct {
	Uin      int64  `json:"Uin"`
	Sid      string `json:"Sid"`
	Skey     string `json:"Skey"`
	DeviceID string `json:"DeviceID"`
}

type BaseResponse struct {
	Ret    int64  `json:"Ret"`    // 0,
	ErrMsg string `json:"ErrMsg"` // ""
}

type SyncKey struct {
	Count int64     `json:"Count"`
	List  []*KeyVal `json:"List"`
}
type KeyVal struct {
	Key int64 `json:"Key"`
	Val int64 `json:"Val"`
}

type InitResponse struct {
	BaseResponse *BaseResponse `json:"BaseResponse"`
	SyncKey      *SyncKey      `json:"SyncKey"`
	User         *User         `json:"User"`
}

type SyncResponse struct {
	BaseResponse *BaseResponse `json:"BaseResponse"`
	SyncKey      *SyncKey      `json:"SyncCheckKey"`
	MsgCount     int64         `json:"AddMsgCount"`
	MsgList      []*Message    `json:"AddMsgList"`
}

type ContactResponse struct {
	BaseResponse *BaseResponse `json:"BaseResponse"`
	MemberCount  int64         `json:"MemberCount"`
	MemberList   []*User       `json:"MemberList"`
	Seq          int64         `json:"Seq"`
}

type Message struct {
	MsgId                string `json:"MsgId"`                //"MsgId": "5318486354145761421",
	FromUserName         string `json:"FromUserName"`         //"FromUserName": "@8ce30827794deb67aca4885062940013ae1995aba33961be5b646a05cd88d23f",
	ToUserName           string `json"ToUserName"`            //"ToUserName": "filehelper",
	MsgType              int64  `json:"MsgType"`              //"MsgType": 51,
	Content              string `json:"Content"`              //"Content": "",
	Status               int64  `json"Status"`                //"Status": 3,
	ImgStatus            int64  `json:"ImgStatus"`            //"ImgStatus": 1,
	CreateTime           int64  `json:"CreateTime"`           //"CreateTime": 1498531351,
	StatusNotifyCode     int64  `json:""StatusNotifyCode"`    //"StatusNotifyCode": 2,
	StatusNotifyUserName string `json:"StatusNotifyUserName"` // "StatusNotifyUserName": "filehelper"
}
