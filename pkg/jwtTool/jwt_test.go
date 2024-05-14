package jwtTool

import (
	"cool-iot/pkg/crypt"
	"testing"
	"time"
)

const MaxC = 100000

func Test(t *testing.T) {
	driver := NewToken()
	tokens := make(map[string]int64)
	for i := 0; i < MaxC; i++ {
		uid := crypt.SnowID(1)
		token := driver.Generate(uid, 3*time.Second)
		tokens[token] = uid
	}
	for k, _ := range tokens {
		_, err := driver.Parse(k)
		if err != nil {
			continue
		}
	}
}
