// Package ginext provides extensions for the Gin web framework.
package ginext

import (
	"github.com/gin-gonic/gin"
)

// Type aliases for convenience
type Context = gin.Context
type HandlerFunc = gin.HandlerFunc
type H = gin.H

// Engine extends the standard Gin Engine.
type Engine struct {
	*gin.Engine
}

// RouterGroup groups related routes together.
type RouterGroup struct {
	*gin.RouterGroup
}

// New creates a new Engine instance with the specified Gin mode.
func New(ginMode string) *Engine {
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode("")
	}
	return &Engine{gin.New()}
}

// Run starts the Gin server.
func (e *Engine) Run(addr ...string) error {
	return e.Engine.Run(addr...)
}

// Group creates a new route group.
func (e *Engine) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{e.Engine.Group(relativePath, handlers...)}
}

// Use attaches middleware to the Engine.
func (e *Engine) Use(middleware ...HandlerFunc) {
	e.Engine.Use(middleware...)
}

func (g *RouterGroup) Use(middleware ...HandlerFunc) {
	g.RouterGroup.Use(middleware...)
}

func (e *Engine) LoadHTMLGlob(pattern string) {
	e.Engine.LoadHTMLGlob(pattern)
}

// Default middleware
func Logger() HandlerFunc {
	return gin.Logger()
}

func Recovery() HandlerFunc {
	return gin.Recovery()
}

// HTTP methods for Engine
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

// HTTP methods for RouterGroup
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
