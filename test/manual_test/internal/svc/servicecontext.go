package svc

import (
	"github.com/linabellbiu/goctl-validate/test/manual_test/internal/config"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
