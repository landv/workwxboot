package main

import "workwxboot/pkg/workwx"

func main() {
	corpid := "xxxx"
	agentid := int64(1000000)
	secretkey := "xxxx"
	client := workwx.Client{CropID: corpid, AgentID: agentid, AgentSecret: secretkey}
	md := workwx.Message{
		ToUser:   "xxxx",
		MsgType:  "markdown",
		Markdown: workwx.Content{Content: "### 测试"},
	}
	client.Send(md)
	text := workwx.Message{
		ToUser:  "xxxx",
		MsgType: "text",
		Text:    workwx.Content{Content: "文本测试"},
	}
	client.Send(text)
	textcard := workwx.Message{
		ToUser:  "xxx",
		MsgType: "textcard",
		Textcard: workwx.TextCard{
			Title:       "hahha",
			Description: "xxxx",
			Url:         "https://jira.baidu.com",
			Btntxt:      "更多",
		},
	}
	client.Send(textcard)
	news := workwx.Message{
		ToUser:  "xxxx",
		MsgType: "news",
		News: workwx.News{
			Articles: []workwx.Article{
				{
					Title:       "中秋节礼品领取",
					Description: "今年中秋节公司有豪礼相送",
					Url:         "https://jira.baidu.com",
					Picurl:      "http://res.mail.qq.com/node/ww/wwopenmng/images/independent/doc/test_pic_msg1.png",
				},
				{
					Title:       "国庆节礼品领取",
					Description: "今年国庆节公司有豪礼相送",
					Url:         "https://wiki.baidu.com",
					Picurl:      "http://res.mail.qq.com/node/ww/wwopenmng/images/independent/doc/test_pic_msg1.png",
				},
			},
		},
	}
	client.Send(news)
}
