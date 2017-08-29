package go_weixin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"
)

type WeixinAccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	CreatedAt   int64  `json:"created_at"`
}

type JsApiTicket struct {
	ErrCode   int64  `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	Ticket    string `json:"ticket"`
	ExpiresIn int64  `json:"expires_in"`
	CreatedAt int64  `json:"created_at"`
}

type Oauth2AccessToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Openid       string `json:"openid"`
	Scope        string `json:"scope"`
}

type WeixinUserInfo struct {
	Openid     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	Headimgurl string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	Unionid    string   `json:"unionid"`
}

const (
	wxServiceAccount        = "xxxx"
	wxServiceSecret         = "xxxx"
	wxServiceToken          = "xxxx"
	wxServiceEncodingAESKey = "xxxx"
	wxAppId                 = "xxxx"
)

func CheckSignature(signature, timestamp, nonce string) bool {
	values := []string{wxServiceToken, timestamp, nonce}
	sort.Strings(values)
	str := ""
	for _, v := range values {
		str += v
	}
	return str == signature
}

func GetUserInfo(code string) (*WeixinUserInfo, error) {
	wxToken, err := GetUserAccessTokenAndOpenId(code)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		wxToken.AccessToken,
		wxToken.Openid,
	)
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	log.Println(string(content))
	userinfo := &WeixinUserInfo{}
	err = json.Unmarshal(content, userinfo)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return userinfo, nil
}

func GetUserAccessTokenAndOpenId(code string) (*Oauth2AccessToken, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		wxAppId,
		wxServiceSecret,
		code,
	)
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	log.Println(fmt.Sprintf("%s", content))
	oauth2AccessToken := &Oauth2AccessToken{}
	err = json.Unmarshal(content, oauth2AccessToken)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return oauth2AccessToken, nil
}

func FlushJsApiTicketAndSave() (string, error) {
	accessToken, err := GetAccessToken()
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=%s&type=jsapi", accessToken)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	jsApiTicket := &JsApiTicket{}
	json.Unmarshal(content, jsApiTicket)
	jsApiTicket.CreatedAt = time.Now().Unix()
	content, err = json.Marshal(jsApiTicket)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	return jsApiTicket.Ticket, nil
}

func FlushWeixinTokenAndSave() (string, error) {
	requestURL := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		wxAppId,
		wxServiceSecret,
	)
	resp, err := http.Get(requestURL)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	wxAccessToken := &WeixinAccessToken{}
	err = json.Unmarshal(content, wxAccessToken)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	wxAccessToken.CreatedAt = time.Now().Unix()
	content, err = json.Marshal(wxAccessToken)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	log.Println(wxAccessToken.AccessToken)
	return wxAccessToken.AccessToken, nil
}

/*
 *access_token的有效期比较长，不需要每次都重新拉取，获取之后存到redis中，可供下次使用，当然需要判断有效期
 *被注释掉的这段代码就是缓存了access_token的信息，因为每个人依赖库不同，所以这里注释掉了
 */
func GetAccessToken() (string, error) {
	return FlushWeixinTokenAndSave()
	/*
	key := "access_token_cache_key"
	tokenInfoString, err := redis.Do("GET", key)
	accessToken := &WeixinAccessToken{}
	err = json.Unmarshal(tokenInfoString.([]byte), accessToken)
	if err != nil {
		return "", err
	}
	now := time.Now().Unix()
	if now-accessToken.CreatedAt > accessToken.ExpiresIn {
		return FlushWeixinTokenAndSave()
	}
	return accessToken.AccessToken, nil
	*/
}
