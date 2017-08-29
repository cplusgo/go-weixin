package go_weixin

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
	"encoding/xml"
	"log"
)

/*
 *参考官方文档:
 *统一下单接口:https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_1
 */

type WxPaymentSigned struct {
	appId      string
	appKey     string
	mchId      string
	nonceStr   string
	body       string
	desc       string
	fee        int
	notifyUrl  string
	outTradeNo string
	createIp   string
	tradeType  string
	attach     string
}

type UnifiedorderResult struct {
	ReturnCode string `xml:"return_code"`
	ReturnMsg  string `xml:"return_msg"`
	AppId      string `xml:"appid"`
	MchId      string `xml:"mch_id"`
	NonceStr   string `xml:"nonce_str"`
	Sign       string `xml:"sign"`
	ResultCode string `xml:"result_code"`
	PrepayId   string `xml:"prepay_id"`
	TradeType  string `xml:"trade_type"`
}

const (
	APP_ID            = ""
	APP_KEY           = ""
	MCH_ID            = ""
	wxUnifiedorderURL = "https://api.mch.weixin.qq.com/pay/unifiedorder"
	TRADE_TYPE = "APP"
)

func NewWxPaymentSigned(nonceStr string, body string,
	desc string, fee int, notifyUrl string, outTradeNo string, createIp string, attach string) *WxPaymentSigned {
	payment := &WxPaymentSigned{
		appId:      APP_ID,
		appKey:     APP_KEY,
		mchId:      MCH_ID,
		nonceStr:   nonceStr,
		body:       body,
		desc:       desc,
		fee:        fee,
		notifyUrl:  notifyUrl,
		outTradeNo: outTradeNo,
		createIp:   createIp,
		tradeType:  TRADE_TYPE,
		attach:     attach,
	}
	return payment
}

func (this *WxPaymentSigned) Signed() ([]byte, error) {
	prepayid, err := this.unifiedorder()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	timestamp := strconv.Itoa(int(time.Now().Unix()))
	presignData := make(map[string]string)
	presignData["appid"] = APP_ID
	presignData["partnerid"] = MCH_ID
	presignData["prepayid"] = prepayid
	presignData["package"] = "Sign=WXPay"
	presignData["noncestr"] = MD5(timestamp)
	presignData["timestamp"] = timestamp

	params := sortURLParams(presignData)
	params = params + "&key=" + APP_KEY
	presignData["sign"] = MD5(params)
	content, err := json.Marshal(presignData)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return content, nil
}

func (this *WxPaymentSigned) createPresign() map[string]string {
	presignData := make(map[string]string)
	presignData["appid"] = this.appId
	presignData["attach"] = this.attach
	presignData["body"] = this.body
	presignData["mch_id"] = this.mchId
	presignData["nonce_str"] = this.nonceStr
	presignData["notify_url"] = this.notifyUrl
	presignData["out_trade_no"] = this.outTradeNo
	presignData["spbill_create_ip"] = this.createIp
	presignData["total_fee"] = strconv.Itoa(this.fee)
	presignData["trade_type"] = this.tradeType
	return presignData
}

func (this *WxPaymentSigned) unifiedorder() (string, error) {
	presignData := this.createPresign()
	params := sortURLParams(presignData)
	params = params + "&key=" + APP_KEY
	presignData["sign"] = MD5(params)
	xmlString := mapToXML(presignData)
	log.Println(xmlString)
	reader := bytes.NewReader([]byte(xmlString))
	resp, err := http.Post(wxUnifiedorderURL, "application/xml", reader)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	var result UnifiedorderResult
	err = xml.Unmarshal(respBytes, &result)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	return result.PrepayId, nil
}
