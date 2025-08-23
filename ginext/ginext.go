package ginext

import (
	"github.com/gin-gonic/gin"
)

// Аналоги библиотечных типов
type Context = gin.Context
type HandlerFunc = gin.HandlerFunc
type H = gin.H

// Engine основная структура gin
type Engine struct {
	*gin.Engine
}

// RouterGroup позволяет объединять хэндлеры в группы
type RouterGroup struct {
	*gin.RouterGroup
}

// New - конструктор роутера
func New() *Engine {
	return &Engine{gin.New()}
}

// Запуск сервера
func (e *Engine) Run(addr ...string) error {
	return e.Engine.Run(addr...)
}

// Group используется для создания группы роутов
func (e *Engine) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {

	return &RouterGroup{e.Engine.Group(relativePath, handlers...)}
}

// Use используется для установки middleware на хэндлер/хэндлеры
func (e *Engine) Use(middleware ...HandlerFunc) {
	e.Engine.Use(middleware...)
}

// Стандартные middleware
func Logger() HandlerFunc {
	return gin.Logger()
}

func Recovery() HandlerFunc {
	return gin.Recovery()
}

// HTTP-методы для Engine

func (e *Engine) GET(relativePath string, handlers ...HandlerFunc) {
	e.Engine.GET(relativePath, handlers...)
}

func (e *Engine) POST(relativePath string, handlers ...HandlerFunc) {
	e.Engine.POST(relativePath, handlers...)
}

func (e *Engine) DELETE(relativePath string, handlers ...HandlerFunc) {
	e.Engine.DELETE(relativePath, handlers...)
}

func (e *Engine) PUT(relativePath string, handlers ...HandlerFunc) {
	e.Engine.PUT(relativePath, handlers...)
}

func (e *Engine) PATCH(relativePath string, handlers ...HandlerFunc) {
	e.Engine.PATCH(relativePath, handlers...)
}

func (e *Engine) OPTIONS(relativePath string, handlers ...HandlerFunc) {
	e.Engine.OPTIONS(relativePath, handlers...)
}

func (e *Engine) HEAD(relativePath string, handlers ...HandlerFunc) {
	e.Engine.HEAD(relativePath, handlers...)
}

// HTTP-методы для для RouterGroup

func (g *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.GET(relativePath, handlers...)
}

func (g *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.POST(relativePath, handlers...)
}

func (g *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.DELETE(relativePath, handlers...)
}

func (g *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.PUT(relativePath, handlers...)
}

func (g *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.PATCH(relativePath, handlers...)
}

func (g *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.OPTIONS(relativePath, handlers...)
}

func (g *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.HEAD(relativePath, handlers...)
}

func (g *RouterGroup) Use(middleware ...HandlerFunc) {
	g.RouterGroup.Use(middleware...)
}
