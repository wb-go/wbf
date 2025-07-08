package ginext

import (
	"github.com/gin-gonic/gin"
)

type Engine struct {
	*gin.Engine
}

type Context = gin.Context

func New() *Engine {
	return &Engine{gin.New()}
}

func (e *Engine) GET(relativePath string, handlers ...gin.HandlerFunc) {
	e.Engine.GET(relativePath, handlers...)
}

func (e *Engine) Run(addr ...string) error {
	return e.Engine.Run(addr...)
}
