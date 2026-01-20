package email

import (
	"github.com/KOMKZ/go-yogan-framework/logger"
	"github.com/samber/do/v2"
)

// ProvideManager 注册 Email Manager 到 DI 容器
// 配置结构: email.Config
// 依赖: *logger.CtxZapLogger
//
// 用法:
//
//	email.ProvideManager(injector, &emailConfig)
//	manager := do.MustInvoke[*email.Manager](injector)
func ProvideManager(injector do.Injector, config *Config) {
	do.Provide(injector, func(i do.Injector) (*Manager, error) {
		log := do.MustInvoke[*logger.CtxZapLogger](i)
		return NewManager(config, log, nil)
	})
}

// ProvideManagerWithRegistry 注册 Email Manager 到 DI 容器（自定义 Registry）
func ProvideManagerWithRegistry(injector do.Injector, config *Config, registry *Registry) {
	do.Provide(injector, func(i do.Injector) (*Manager, error) {
		log := do.MustInvoke[*logger.CtxZapLogger](i)
		return NewManager(config, log, registry)
	})
}
