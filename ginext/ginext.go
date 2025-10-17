// Package ginext provides extensions for the Gin web framework.
package ginext

import (
	"github.com/gin-gonic/gin"
)

// Context is an alias for gin.Context.
type Context = gin.Context

// HandlerFunc is an alias for gin.HandlerFunc.
type HandlerFunc = gin.HandlerFunc

// H is an alias for gin.H.
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

// Use attaches middleware to the RouterGroup.
func (g *RouterGroup) Use(middleware ...HandlerFunc) {
	g.RouterGroup.Use(middleware...)
}

// LoadHTMLGlob loads HTML templates.
func (e *Engine) LoadHTMLGlob(pattern string) {
	e.Engine.LoadHTMLGlob(pattern)
}

// Logger returns the default Gin logger middleware.
func Logger() HandlerFunc {
	return gin.Logger()
}

// Recovery returns the default Gin recovery middleware.
func Recovery() HandlerFunc {
	return gin.Recovery()
}

// GET registers a handler for HTTP GET method on the Engine.
func (e *Engine) GET(relativePath string, handlers ...HandlerFunc) {
	e.Engine.GET(relativePath, handlers...)
}

// POST registers a handler for HTTP POST method on the Engine.
func (e *Engine) POST(relativePath string, handlers ...HandlerFunc) {
	e.Engine.POST(relativePath, handlers...)
}

// DELETE registers a handler for HTTP DELETE method on the Engine.
func (e *Engine) DELETE(relativePath string, handlers ...HandlerFunc) {
	e.Engine.DELETE(relativePath, handlers...)
}

// PUT registers a handler for HTTP PUT method on the Engine.
func (e *Engine) PUT(relativePath string, handlers ...HandlerFunc) {
	e.Engine.PUT(relativePath, handlers...)
}

// PATCH registers a handler for HTTP PATCH method on the Engine.
func (e *Engine) PATCH(relativePath string, handlers ...HandlerFunc) {
	e.Engine.PATCH(relativePath, handlers...)
}

// OPTIONS registers a handler for HTTP OPTIONS method on the Engine.
func (e *Engine) OPTIONS(relativePath string, handlers ...HandlerFunc) {
	e.Engine.OPTIONS(relativePath, handlers...)
}

// HEAD registers a handler for HTTP HEAD method on the Engine.
func (e *Engine) HEAD(relativePath string, handlers ...HandlerFunc) {
	e.Engine.HEAD(relativePath, handlers...)
}

// GET registers a handler for HTTP GET method on the RouterGroup.
func (g *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.GET(relativePath, handlers...)
}

// POST registers a handler for HTTP POST method on the RouterGroup.
func (g *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.POST(relativePath, handlers...)
}

// DELETE registers a handler for HTTP DELETE method on the RouterGroup.
func (g *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.DELETE(relativePath, handlers...)
}

// PUT registers a handler for HTTP PUT method on the RouterGroup.
func (g *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.PUT(relativePath, handlers...)
}

// PATCH registers a handler for HTTP PATCH method on the RouterGroup.
func (g *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.PATCH(relativePath, handlers...)
}

// OPTIONS registers a handler for HTTP OPTIONS method on the RouterGroup.
func (g *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.OPTIONS(relativePath, handlers...)
}

// HEAD registers a handler for HTTP HEAD method on the RouterGroup.
func (g *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) {
	g.RouterGroup.HEAD(relativePath, handlers...)
}
