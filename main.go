package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/xiusin/logger"

	"github.com/gorilla/securecookie"
	"github.com/xiusin/pine"
	pine_bigcache "github.com/xiusin/pine/cache/providers/pine-bigcache"
	"github.com/xiusin/pine/di"
	"github.com/xiusin/pine/sessions"
	cacheProvider "github.com/xiusin/pine/sessions/providers/cache"
	"github.com/xiusin/web-db-manager/actions"
	"github.com/xiusin/web-db-manager/common"
)

func main() {
	app := pine.New()

	app.Use(func(ctx *pine.Context) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", strings.TrimRight(string(ctx.RequestCtx.Referer()), "/"))
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "X-TOKEN, Content-Type, Origin, Referer, Content-Length, Access-Control-Allow-Headers, Authorization")
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if !ctx.IsOptions() {
			ctx.Next()
		}
	})

	di.Set(di.ServicePineLogger, func(builder di.AbstractBuilder) (i interface{}, err error) {
		loggers := logger.New()
		loggers.SetOutput(os.Stdout)
		logger.SetDefault(loggers)
		loggers.SetReportCaller(true, 3)
		loggers.SetLogLevel(common.Appcfg.LogLevel)
		return loggers, nil
	}, false)

	cacheHandler := pine_bigcache.New(bigcache.DefaultConfig(24 * time.Hour))

	di.Set(common.ServiceICache, func(builder di.AbstractBuilder) (i interface{}, err error) {
		return cacheHandler, nil
	}, true)

	di.Set(di.ServicePineSessions, func(builder di.AbstractBuilder) (i interface{}, err error) {
		sess := sessions.New(cacheProvider.NewStore(cacheHandler), &sessions.Config{
			CookieName: common.Appcfg.SessName,
			Expires:    time.Duration(common.Appcfg.SessExpires),
		})
		return sess, nil
	}, true)

	pine.RegisterErrorCodeHandler(500, func(ctx *pine.Context) {
		if ctx.IsAjax() {
			_ = ctx.WriteJSON(pine.H{"code": http.StatusInternalServerError, "message": ctx.Msg})
		} else {
			ctx.Abort(http.StatusInternalServerError, ctx.Msg)
		}
	})

	// 注册静态地址
	app.Static("/mywebsql/", common.GetRootPath("assets"), 1)
	app.StaticFile("favicon.ico", common.GetRootPath("assets/favicon.ico"))

	app.ANY("/", func(ctx *pine.Context) { ctx.Redirect("/mywebsql/index") })
	app.ANY("/mywebsql/cache", actions.Cache)
	app.Handle(new(actions.IndexController), "/mywebsql")

	app.Run(
		pine.Addr(fmt.Sprintf(":%d", common.Appcfg.Port)),
		pine.WithGracefulShutdown(),
		pine.WithCookieTranscoder(securecookie.New([]byte(common.Appcfg.HashKey), []byte(common.Appcfg.BlockKey))),
		pine.WithoutStartupLog(true),
		pine.WithServerName("xiusin/pinecms"),
		pine.WithCookie(true),
	)
}
