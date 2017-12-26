package main

import (
	"flag"
	"log"

	"github.com/liaoxiaorong/wx/wx"
)

var addr = flag.String("addr", "0.0.0.0:7001", "listen addr")

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := wx.Init()
	if err != nil {
		log.Fatal(err.Error())
	}
	go wx.Listening()

	log.Fatal(wx.WebServe(*addr))
}
