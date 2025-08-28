// Package ginext предоставляет расширения для веб-фреймворка Gin.
package ginext

import (
	"github.com/gin-gonic/gin"
)

// Engine расширяет стандартный Gin Engine.
type Engine struct {
	*gin.Engine
}

// Context представляет контекст HTTP запроса.
type Context = gin.Context

// New создает новый экземпляр Engine.
func New() *Engine {
	return &Engine{gin.New()}
}

// GET регистрирует GET маршрут.
func (e *Engine) GET(relativePath string, handlers ...gin.HandlerFunc) {
	e.Engine.GET(relativePath, handlers...)
}

// Run запускает HTTP сервер на указанном адресе.
func (e *Engine) Run(addr ...string) error {
	return e.Engine.Run(addr...)
}
