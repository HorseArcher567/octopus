package api

import "github.com/gin-gonic/gin"

// RouterRegistrar 定义路由注册约定。
// 业务模块可以实现该接口来集中注册自身的路由。
type RouterRegistrar interface {
	RegisterRoutes(engine *gin.Engine)
}

// Register 批量注册多个 RouterRegistrar。
func Register(engine *gin.Engine, registrars ...RouterRegistrar) {
	for _, r := range registrars {
		if r == nil {
			continue
		}
		r.RegisterRoutes(engine)
	}
}
