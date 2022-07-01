package wechat

import (
	"strings"
)

type WechatOptions struct {
	e *Engine
}

type Option func(e *Engine)

func (e *Engine) SetOptions(opts ...Option) {
	for _, opt := range opts {
		opt(e)
	}
}

func WithRedirectDomain(domain string) Option {
	return func(e *Engine) {
		e.redirectDomain = strings.TrimRight(domain, "/")
	}
}

func WithToggleAgentID(agentID string) Option {
	return func(e *Engine) {
		conf, ok := e.config.(*Qy)
		if ok {
			conf.AgentID = agentID
		}
	}
}
