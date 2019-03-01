package util

import (
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"time"
)

/* token的样式为token-username */
func CreatToken() string {
	curtime := time.Now().Unix()

	h := md5.New()

	io.WriteString(h, strconv.FormatInt(curtime, 10))

	token := fmt.Sprintf("%x", h.Sum(nil))
	token = token + string(rand.Int())
	return token
}


