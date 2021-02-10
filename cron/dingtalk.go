package cron

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type DingTalkMsg struct {
	Token     string   `json:"token"`
	Type      string   `json:"type"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Receivers []string `json:"receivers"`
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"is_at_all"`
}

func (dm *DingTalkMsg) toDingTalk() []byte {
	var msg = map[string]interface{}{
		"msgtype": dm.Type,
		"markdown": map[string]string{
			"title": strings.Replace(dm.Title, "#LINE#", "\n", -1),
			"text":  strings.Replace(dm.Content, "#LINE#", "\n", -1),
		},
	}
	at := make(map[string]interface{})
	at["isAtAll"] = dm.IsAtAll
	if dm.AtMobiles != nil && len(dm.AtMobiles) > 0 {
		at["atMobiles"] = dm.AtMobiles
	}
	msg["at"] = at
	data, _ := json.Marshal(msg)
	return data
}

func (dm *DingTalkMsg) send() {
	cli := &http.Client{}
	postData := dm.toDingTalk()
	dingTalkURI := "https://oapi.dingtalk.com/robot/send?access_token=" + dm.Token
	println(dingTalkURI, string(postData))
	req, err := http.NewRequest("POST", dingTalkURI, bytes.NewBuffer(postData))
	if err != nil {
		println(err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := cli.Do(req)
	if err != nil {
		println(err.Error())
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println(err.Error())
		return
	}
	println(string(data))
}
