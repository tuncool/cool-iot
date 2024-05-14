package config

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/exp/slog"
)

func Test(t *testing.T) {
	base := "conf/server.toml"
	dr := New(base)
	conf, err := dr.Load()
	if err != nil {
		return
	}
	fmt.Println(conf)
	conf.Remark = "XNps config file"
	conf.InitTime = time.Now().Unix()
	err = dr.Update(conf)
	if err != nil {
		return
	}

}
func Test2(t *testing.T) {
	base := "./conf/server2.toml"
	if fileTool.DirExisted("./conf") {
		fileTool.CreateFolder("./conf")
	}
	err := CreateNewInitFile(base)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	dr := New(base)
	conf, err := dr.Load()
	if err != nil {
		slog.Error(err.Error())
		return
	}
	conf.InitTime = time.Now().Unix()

}
