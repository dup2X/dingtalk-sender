package cron

import (
	"fmt"
	"strings"
	"time"

	"github.com/toolkits/pkg/logger"

	"github.com/dup2X/dingtalk-sender/config"
	"github.com/dup2X/dingtalk-sender/dataobj"
	"github.com/dup2X/dingtalk-sender/redisc"
)

var semaphore chan int

func SendDing() {
	c := config.Get()
	semaphore = make(chan int, c.Consumer.Worker)

	for {
		messages := redisc.Pop(1, c.Consumer.Queue)
		if len(messages) == 0 {
			time.Sleep(time.Duration(300) * time.Millisecond)
			continue
		}

		sendDingTalk(messages)
	}
}

func sendDingTalk(messages []*dataobj.Message) {
	for _, message := range messages {
		semaphore <- 1
		go sendDingDing(message)
	}
}

func sendDingDing(message *dataobj.Message) {
	defer func() {
		<-semaphore
	}()

	subject := genSubject(message)
	var err error
	content := genContent(message)
	logger.Infof("send notify:%v", message)
	msg := &DingTalkMsg{
		Token: message.Tos[0],
	}
	msg.Type = "markdown"
	msg.Title = subject
	msg.Content = "## " + subject + " ## #LINE# " + content
	msg.send()
	logger.Infof("hashid: %d: subject: %s, tos: %v, error: %v", message.Event.HashId, subject, message.Tos, err)
	logger.Infof("hashid: %d: endpoint: %s, metric: %s, tags: %s", message.Event.HashId, message.ReadableEndpoint, strings.Join(message.Metrics, ","), message.ReadableTags)
}

var ET = map[string]string{
	"alert":    "告警",
	"recovery": "告警恢复",
}

func genSubject(message *dataobj.Message) string {
	subject := ""
	if message.IsUpgrade {
		subject = "[报警已升级]" + subject
	}

	return fmt.Sprintf("[P%d %s]%s - %s", message.Event.Priority, ET[message.Event.EventType], message.Event.Sname, message.ReadableEndpoint)
}

func genContent(message *dataobj.Message) string {
	var src string
	if message.IsUpgrade {
		src = " **报警已升级!** #LINE#"
	}
	src += fmt.Sprintf(`
 **事件状态** :%s#LINE#
 **策略名称** :%s #LINE# 
 **挂载节点** :%s #LINE#
 **metric** :%s #LINE#
 **tags** :%s #LINE#
 **当前值** :%v #LINE#
 **报警说明** :%v #LINE#
 **触发时间** :%v #LINE#
 **报警详情** :%s #LINE#
 **报警策略** :%s #LINE#
`,
		ET[message.Event.EventType],
		message.Event.Sname,
		strings.Join(message.Bindings, ","),
		strings.Join(message.Metrics, ","),
		message.ReadableTags,
		message.Event.Value,
		message.Event.Info,
		parseEtime(message.Event.Etime),
		message.EventLink,
		message.StraLink,
	)
	if message.ClaimLink != "" {
		src += " **认领** :" + message.ClaimLink
	}
	return src
}

func parseEtime(etime int64) string {
	t := time.Unix(etime, 0)
	return t.Format("2006-01-02 15:04:05")
}
