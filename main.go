package main

import (
	"embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/valyala/fasthttp"

	"github.com/gorilla/securecookie"
	"github.com/xiusin/logger"
	"github.com/xiusin/pine"
	pine_bigcache "github.com/xiusin/pine/cache/providers/bigcache"
	"github.com/xiusin/pine/di"
	"github.com/xiusin/pine/sessions"
	cacheProvider "github.com/xiusin/pine/sessions/providers/cache"
	"github.com/xiusin/reload"
	"github.com/xiusin/web-db-manager/actions"
	"github.com/xiusin/web-db-manager/actions/render"
	"github.com/xiusin/web-db-manager/common"

	. "github.com/xiusin/web-db-manager/actions/common"
)

//go:embed assets/*/**
var assets embed.FS

func main() {

	di.Instance(common.ServiceEmbedAssets, &assets)

	plushEngine := render.New("assets/modules/views", &assets, true)
	plushEngine.AddFunc("T", T)

	plushEngine.AddFunc("getServerList", func() map[string]Server {
		return SERVER_LIST
	})

	pine.RegisterViewEngine(plushEngine)

	di.Set(RenderService, func(builder di.AbstractBuilder) (interface{}, error) {
		return plushEngine, nil
	}, true)

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

	cacheHandler := pine_bigcache.New(bigcache.DefaultConfig(time.Hour))

	di.Instance(di.ServicePineLogger, logger.New())

	di.Instance(common.ServiceICache, cacheHandler)

	di.Set(di.ServicePineSessions, func(builder di.AbstractBuilder) (i interface{}, err error) {
		sess := sessions.New(cacheProvider.NewStore(cacheHandler), &sessions.Config{
			CookieName: common.Appcfg.SessName,
			Expires:    time.Duration(common.Appcfg.SessExpires),
		})
		return sess, nil
	}, true)

	app.SetRecoverHandler(func(ctx *pine.Context) {
		ctx.Logger().Error(ctx.Msg, ctx.Err())
		ctx.Abort(fasthttp.StatusInternalServerError, fasthttp.StatusMessage(fasthttp.StatusInternalServerError))
	})

	// 注册静态地址
	app.StaticFS("/mywebsql/", assets, "assets")

	app.Favicon(common.GetRootPath("assets/favicon.ico"))

	app.ANY("/", func(ctx *pine.Context) { ctx.Redirect("/mywebsql/index") })

	app.ANY("/mywebsql/cache", actions.Cache)

	app.Handle(new(actions.IndexController), "/mywebsql")

	transcoder := securecookie.New([]byte(common.Appcfg.HashKey), []byte(common.Appcfg.BlockKey))

	exec := func() error {
		app.Run(
			pine.Addr(fmt.Sprintf(":%d", common.Appcfg.Port)),
			pine.WithGracefulShutdown(),
			pine.WithCookieTranscoder(transcoder),
			pine.WithoutStartupLog(false),
			pine.WithServerName("xiusin/pine"),
			pine.WithCookie(true),
			pine.WithCompressGzip(true),
		)
		return nil
	}

	if os.Getenv("RELOADABLE") != "" {
		reload.Loop(func() error {
			return exec()
		}, &reload.Conf{Cmd: &reload.CmdConf{}})
	} else {
		exec()
	}

}
