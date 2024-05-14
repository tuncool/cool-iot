package common

import "net/http"

const ()

type CodeEnum uint16

func (m CodeEnum) String() string {
	res := http.StatusText(int(m))
	if len(res) != 0 {
		return res
	}
	switch m {
	// 连接参数

	default:
		return "unknown type"
	}
}
