package wx

var wx *Weixin

func Init() error {
	wx = NewWeixin()
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

	err = wx.GetContacts()
	if err != nil {
		return err
	}
	return nil
}

func Listening() error {
	return wx.Listening()
}

func GetContacts() (map[string]*User, error) {
	return wx.contacts, nil
}

func SendMsg(userId, msg string) error {
	return wx.SendMsg(userId, msg)
}
