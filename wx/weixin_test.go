package wx

import (
	"os"
	"testing"
)

func TestFileCache(t *testing.T) {
	filePath := "/tmp/wx.json"
	fs := &FileStore{filePath}
	err := fs.SaveToFile("corp", "best of corp")
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Log("name", fs.GetFromFile("name"))
	t.Log("corp", fs.GetFromFile("corp"))
	t.Log("nothing", fs.GetFromFile("nothing"))
}

func TestMain(m *testing.M) {
	// init
	os.Exit(m.Run())
}
