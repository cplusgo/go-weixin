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

/*
<xml>
   <appid>wx2421b1c4370ec43b</appid>
   <attach>支付测试</attach>
   <body>APP支付测试</body>
   <mch_id>10000100</mch_id>
   <nonce_str>1add1a30ac87aa2db72f57a2375d8fec</nonce_str>
   <notify_url>http://wxpay.wxutil.com/pub_v2/pay/notify.v2.php</notify_url>
   <out_trade_no>1415659990</out_trade_no>
   <spbill_create_ip>14.23.150.211</spbill_create_ip>
   <total_fee>1</total_fee>
   <trade_type>APP</trade_type>
   <sign>0CB01533B8C1EF103065174F50BCA001</sign>
</xml>
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
)

func NewWxPaymentSigned(nonceStr string, body string,
	desc string, fee int, notifyUrl string, outTradeNo string, createIp string,
	tradeType string, attach string) *WxPaymentSigned {
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
		tradeType:  tradeType,
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

/*
 *在统一下单接口执行完毕之后会返回一个关键的数据,prepayid
 */
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
