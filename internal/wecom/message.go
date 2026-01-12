package wecom

type IncomingMessage struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	MsgType      string `xml:"MsgType"`
	Content      string `xml:"Content"`
	Event        string `xml:"Event"`
	EventKey     string `xml:"EventKey"`
	TaskId       string `xml:"TaskId"`
	CardType     string `xml:"CardType"`
}

const (
	EventKeyUnraidRestart     = "unraid.action.restart"
	EventKeyUnraidStop        = "unraid.action.stop"
	EventKeyUnraidForceUpdate = "unraid.action.force_update"

	EventKeyConfirm = "core.action.confirm"
	EventKeyCancel  = "core.action.cancel"
)

type TextMessage struct {
	ToUser  string
	Content string
}

type TemplateCardMessage struct {
	ToUser string
	Card   TemplateCard
}

type TemplateCard map[string]interface{}

func NewUnraidActionCard() TemplateCard {
	return TemplateCard{
		"card_type": "button_interaction",
		"main_title": map[string]interface{}{
			"title": "Unraid 容器操作",
			"desc":  "请选择动作",
		},
		"button_list": []map[string]interface{}{
			{
				"text":  "重启容器",
				"style": 1,
				"key":   EventKeyUnraidRestart,
			},
			{
				"text":  "停止容器",
				"style": 2,
				"key":   EventKeyUnraidStop,
			},
			{
				"text":  "强制更新",
				"style": 2,
				"key":   EventKeyUnraidForceUpdate,
			},
		},
	}
}

func NewConfirmCard(actionDisplayName, containerName string) TemplateCard {
	return TemplateCard{
		"card_type": "button_interaction",
		"main_title": map[string]interface{}{
			"title": "确认执行",
			"desc":  actionDisplayName + "：" + containerName,
		},
		"button_list": []map[string]interface{}{
			{
				"text":  "确认",
				"style": 2,
				"key":   EventKeyConfirm,
			},
			{
				"text":  "取消",
				"style": 1,
				"key":   EventKeyCancel,
			},
		},
	}
}
