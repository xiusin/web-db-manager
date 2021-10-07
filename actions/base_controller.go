package actions

import (
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xiusin/pine"
	"github.com/xiusin/pine/di"
	"github.com/xiusin/web-db-manager/actions/common"
	"github.com/xiusin/web-db-manager/actions/render"
	commonFn "github.com/xiusin/web-db-manager/common"
)

type MyWebSql struct {
	pine.Controller
	plush *render.Plush
}

func init() {
	plushEngine := render.New(commonFn.GetRootPath("assets/modules/views"), true)
	plushEngine.AddFunc("T", common.T)

	plushEngine.AddFunc("getServerList", func() map[string]common.Server {
		return common.SERVER_LIST
	})

	pine.RegisterViewEngine(plushEngine)

	di.Set(common.RenderService, func(builder di.AbstractBuilder) (interface{}, error) {
		return plushEngine, nil
	}, true)
}

func (c *MyWebSql) Construct() {

	if c.Session().Get("theme_path") == "" {
		c.Session().Set("theme_path", "bootstrap")
	}
	c.ViewData("THEME_PATH", c.Session().Get("theme_path"))
	c.ViewData("MAX_TEXT_LENGTH_DISPLAY", common.MAX_TEXT_LENGTH_DISPLAY)
	c.ViewData("APP_VERSION", common.APP_VERSION)
	c.ViewData("EXTERNAL_PATH", "/mywebsql/")
	c.ViewData("PRODUCT_URL", "https://github.com/xiusin/db-web-manager.git")

	c.plush = di.MustGet(common.RenderService).(*render.Plush)

}

func (c *MyWebSql) saveAuthSession(serve common.Server) {
	sess, _ := json.Marshal(&serve)
	c.Session().Set("auth", string(sess))
}

func (c *MyWebSql) getAuthSession() common.Server {
	auth := c.Session().Get("auth")
	var authServer common.Server
	json.Unmarshal([]byte(auth), &authServer)
	return authServer
}

func (c *MyWebSql) clearAuthSession() {
	c.Session().Remove("auth")
}

func (c *MyWebSql) GetSQLX() (*sqlx.DB, *common.Server) {
	var serve common.Server
	sess := c.Session().Get("auth")

	if len(sess) == 0 {
		return nil, nil
	}

	json.Unmarshal([]byte(sess), &serve)
	db, _ := sqlx.Open(serve.Driver, fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8&parseTime=true&loc=Local", serve.User, serve.Password, serve.Host, serve.Port))
	return db, &serve
}
