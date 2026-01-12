package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"daily-help/internal/unraid"
	"daily-help/internal/wecom"
)

type RouterDeps struct {
	WeCom         *wecom.Client
	Unraid        *unraid.Client
	AllowedUserID map[string]struct{}
}

type Router struct {
	WeCom         *wecom.Client
	Unraid        *unraid.Client
	AllowedUserID map[string]struct{}

	state *StateStore
}

func NewRouter(deps RouterDeps) *Router {
	return &Router{
		WeCom:         deps.WeCom,
		Unraid:        deps.Unraid,
		AllowedUserID: deps.AllowedUserID,
		state:         NewStateStore(30 * time.Minute),
	}
}

func (r *Router) HandleMessage(ctx context.Context, msg wecom.IncomingMessage) error {
	userID := strings.TrimSpace(msg.FromUserName)
	if userID == "" {
		return nil
	}
	if _, ok := r.AllowedUserID[userID]; !ok {
		_ = r.WeCom.SendText(ctx, wecom.TextMessage{
			ToUser:  userID,
			Content: "无权限：该账号未加入白名单。",
		})
		return nil
	}

	switch msg.MsgType {
	case "text":
		return r.handleText(ctx, userID, strings.TrimSpace(msg.Content))
	case "event":
		return r.handleEvent(ctx, userID, msg)
	default:
		_ = r.WeCom.SendText(ctx, wecom.TextMessage{
			ToUser:  userID,
			Content: "暂不支持的消息类型。",
		})
		return nil
	}
}

func (r *Router) handleText(ctx context.Context, userID string, content string) error {
	if content == "" {
		return nil
	}

	if isEntryKeyword(content) {
		r.state.Clear(userID)
		return r.WeCom.SendTemplateCard(ctx, wecom.TemplateCardMessage{
			ToUser: userID,
			Card:   wecom.NewUnraidActionCard(),
		})
	}

	state, ok := r.state.Get(userID)
	if !ok || state.Step != StepAwaitingContainerName {
		return r.WeCom.SendText(ctx, wecom.TextMessage{
			ToUser:  userID,
			Content: "请输入“容器”打开操作菜单。",
		})
	}

	containerName, err := validateContainerName(content)
	if err != nil {
		return r.WeCom.SendText(ctx, wecom.TextMessage{
			ToUser:  userID,
			Content: fmt.Sprintf("容器名不合法：%s", err.Error()),
		})
	}

	state.ContainerName = containerName
	state.Step = StepAwaitingConfirm
	r.state.Set(userID, state)

	return r.WeCom.SendTemplateCard(ctx, wecom.TemplateCardMessage{
		ToUser: userID,
		Card:   wecom.NewConfirmCard(state.Action.DisplayName(), state.ContainerName),
	})
}

func (r *Router) handleEvent(ctx context.Context, userID string, msg wecom.IncomingMessage) error {
	if msg.Event == "enter_agent" {
		r.state.Clear(userID)
		return r.WeCom.SendTemplateCard(ctx, wecom.TemplateCardMessage{
			ToUser: userID,
			Card:   wecom.NewUnraidActionCard(),
		})
	}

	if msg.Event != "template_card_event" {
		return nil
	}

	key := strings.TrimSpace(msg.EventKey)
	switch key {
	case wecom.EventKeyUnraidRestart, wecom.EventKeyUnraidStop, wecom.EventKeyUnraidForceUpdate:
		action := ActionFromEventKey(key)
		r.state.Set(userID, ConversationState{
			Step:   StepAwaitingContainerName,
			Action: action,
		})
		return r.WeCom.SendText(ctx, wecom.TextMessage{
			ToUser:  userID,
			Content: fmt.Sprintf("已选择动作：%s\n请输入容器名：", action.DisplayName()),
		})

	case wecom.EventKeyConfirm:
		state, ok := r.state.Get(userID)
		if !ok || state.Step != StepAwaitingConfirm {
			return r.WeCom.SendText(ctx, wecom.TextMessage{
				ToUser:  userID,
				Content: "会话已过期，请输入“容器”重新开始。",
			})
		}
		r.state.Clear(userID)

		start := time.Now()
		err := r.execAction(ctx, state.Action, state.ContainerName)
		cost := time.Since(start).Milliseconds()
		if err != nil {
			return r.WeCom.SendText(ctx, wecom.TextMessage{
				ToUser:  userID,
				Content: fmt.Sprintf("执行失败（%dms）：%s", cost, err.Error()),
			})
		}

		return r.WeCom.SendText(ctx, wecom.TextMessage{
			ToUser:  userID,
			Content: fmt.Sprintf("执行成功（%dms）：%s %s", cost, state.Action.DisplayName(), state.ContainerName),
		})

	case wecom.EventKeyCancel:
		r.state.Clear(userID)
		return r.WeCom.SendText(ctx, wecom.TextMessage{
			ToUser:  userID,
			Content: "已取消。",
		})
	default:
		return nil
	}
}

func (r *Router) execAction(ctx context.Context, action Action, containerName string) error {
	switch action {
	case ActionRestart:
		return r.Unraid.RestartContainerByName(ctx, containerName)
	case ActionStop:
		return r.Unraid.StopContainerByName(ctx, containerName)
	case ActionForceUpdate:
		return r.Unraid.ForceUpdateContainerByName(ctx, containerName)
	default:
		return fmt.Errorf("未知动作: %s", action)
	}
}

func isEntryKeyword(content string) bool {
	switch strings.ToLower(strings.TrimSpace(content)) {
	case "help", "菜单", "容器", "docker", "unraid":
		return true
	default:
		return false
	}
}
