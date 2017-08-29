package main

import (
	"github.com/cplusgo/go-weixin"
	"log"
)


func main() {
	//code是微信回传的数据中包含的部分
	userInfo, err := go_weixin.GetUserInfo("code")
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println(userInfo)
}
