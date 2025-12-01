package common

import (
	"math/rand"
	"strconv"
	"time"
)

func GetNowAndLenRandomString(len int) string {
	return time.Now().Format("20060102") + strconv.Itoa(GetRandomInt(len))
}

// smcTODO:
func getDefaultGroupAvatar() string {

	return ""
}