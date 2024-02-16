package routes

import (
	"messager-server/internal/config"
	"messager-server/internal/messager"
	"messager-server/internal/storage"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/pprofhandler"

	fasthttprouter "github.com/fasthttp/router"
)

const (
	contentTypeJson     = "application/json"
	contentTypeCsv      = "text/csv"
	cacheControlNoStore = "no-store"
	dateLayout          = "2006-01-02"
)

type Router struct {
	storage  *storage.Storage
	messager *messager.Messager
	rtr      *fasthttprouter.Router
	srv      *fasthttp.Server
	logger   *logrus.Logger
	port     string
}

func (r *Router) Start() error {
	return r.srv.ListenAndServe(r.port)
}

func (r *Router) Shutdown() error {
	return r.srv.Shutdown()
}

func New(postgres *storage.Storage, cfg *config.Config, logger *logrus.Logger, messager *messager.Messager) *Router {

	rtr := fasthttprouter.New()

	r := &Router{
		storage:  postgres,
		messager: messager,
		rtr:      rtr,
		srv: &fasthttp.Server{
			Handler:            cors(rtr.Handler),
			MaxRequestBodySize: 100_000_000,
			ReadTimeout:        time.Duration(cfg.Api.ReadTimeout) * time.Second,
			WriteTimeout:       time.Duration(cfg.Api.WriteTimeout) * time.Second,
			IdleTimeout:        time.Duration(cfg.Api.IdleTimeout) * time.Second,
			Logger:             logger,
		},
		logger: logger,
		port:   cfg.Api.HTTPPort,
	}

	registerAuth(r)

	r.rtr.GET("/status", statusHandler)
	r.rtr.GET("/", test)
	r.rtr.GET("/debug/pprof/{profile:*}", pprofhandler.PprofHandler)

	r.rtr.HandleMethodNotAllowed = true
	r.rtr.MethodNotAllowed = methodNotAllowedHandler
	r.rtr.NotFound = notFoundHandler
	return r
}

func test(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusBadRequest)
}
func methodNotAllowedHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
}

func notFoundHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusBadRequest)
}

// Хендлер для прохождения liveness/readiness проб kubernetes.
func statusHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
}

// Мидлвара добавляющая заголовок, который запрещает кеширование.
func noCache(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		handler(ctx)
		ctx.Response.Header.Set(fasthttp.HeaderCacheControl, cacheControlNoStore)
	}
}

// Мидлвара добавляющая CORS заголовки.
func cors(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		handler(ctx)
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowOrigin, "*")
	}
}
