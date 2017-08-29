package go_weixin

import (
	"bytes"
	"log"
	"sort"
	"fmt"
	"crypto/md5"
)

func MD5(content string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(content)))
}

func mapToXML(data map[string]string) string {
	var stringBuffer bytes.Buffer
	stringBuffer.WriteString("<xml>")
	for key, value := range data {
		stringBuffer.WriteString("<")
		stringBuffer.WriteString(key)
		stringBuffer.WriteString("><![CDATA[")
		stringBuffer.WriteString(value)
		stringBuffer.WriteString("]]></")
		stringBuffer.WriteString(key)
		stringBuffer.WriteString(">")
	}
	stringBuffer.WriteString("</xml>")
	return stringBuffer.String()
}

func sortURLParams(params map[string]string) string {
	keys := []string{}
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	endMark := keys[len(keys)-1]
	log.Println(endMark)
	var stringBuffer bytes.Buffer
	for _, key := range keys {
		stringBuffer.WriteString(key)
		stringBuffer.WriteString("=")
		stringBuffer.WriteString(params[key])
		if key != endMark {
			stringBuffer.WriteString("&")
		}
	}
	return stringBuffer.String()
}
