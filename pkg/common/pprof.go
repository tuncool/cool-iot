package common

import (
	"github.com/astaxie/beego/logs"
	"net/http"
	_ "net/http/pprof"
)

func InitPProfFromArg(arg string) {
	if len(arg) > 0 {
		runPProf(arg)
	}
}

func runPProf(ipPort string) {
	go func() {
		_ = http.ListenAndServe(ipPort, nil)
	}()
	logs.Info("PProf debug listen on", ipPort)
}
