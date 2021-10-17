package actions

import (
	"html/template"

	"github.com/jmoiron/sqlx"
	"github.com/xiusin/web-db-manager/actions/common"

	"github.com/xiusin/pine"
)

type IndexController struct {
	MyWebSql
	hasError     string
	ShareVisitor *int
}

var AjaxResponse = []byte(`<div id="results">1</div>`)

// GetIndex 总入口
func (c *IndexController) GetIndex() {
	theme, _ := c.Ctx().Input().GetString("theme")
	if len(theme) > 0 {
		c.Render().ViewData("THEME_PATH", theme)
		c.Ctx().SetCookie("theme", theme, common.COOKIE_LIFETIME*60*60)
		c.Ctx().Session().Set("theme_path", theme)
		c.Ctx().Write(AjaxResponse)
		return
	} else {
		if theme := c.Ctx().GetCookie("theme"); len(theme) > 0 {
			c.Render().ViewData("THEME_PATH", theme)
		} else {
			c.Render().ViewData("THEME_PATH", common.DEFAULT_THEME)
		}
	}
	editor, _ := c.Ctx().Input().GetString("editor")
	if len(editor) > 0 {
		c.Render().ViewData("SQL_EDITORTYPE", editor)
		c.Ctx().SetCookie("editor", editor, common.COOKIE_LIFETIME*60*60)
		x, _ := c.Ctx().Input().GetString("x")
		if x == "1" {
			c.Ctx().Write(AjaxResponse)
			return
		}

	} else {

		//if editor := c.Ctx().GetCookie("editor"); len(editor) > 0 {
		//	c.Render().ViewData("THEME_PATH", editor)
		//} else {
		//	c.Render().ViewData("THEME_PATH", common.DEFAULT_THEME)
		//}
	}

	if db, authServe := c.GetSQLX(); db == nil {
		c.clearAuthSession()
		q, _ := c.Ctx().Input().GetString("q")
		if q == "wrkfrm" {
			c.Render().HTML("session_expired.php")
			return
		}

		form, _ := c.plush.Exec("auth.php", pine.H{
			"LOGINID":     "root",
			"SERVER_NAME": "",
			"SERVER_TYPE": "mysql",
		})
		formCode := `<div class="login"><form method="post" action="" name="dbform" id="dbform" style="text-align:center">` + string(form) + `</form></div>`
		if len(c.hasError) > 0 {
			c.ViewData("MESSAGE", template.HTML(`<div class="msg">`+c.hasError+`</div>`))
		} else {
			c.ViewData("MESSAGE", "")
		}
		c.ViewData("FORM", template.HTML(formCode))
		c.ViewData("APP_VERSION", "dev.0.0.1")
		c.ViewData("PROJECT_SITEURL", "http://mywebsql.xiusin.cn")
		c.ViewData("EXTRA_SCRIPT", template.JS(`<script language="javascript" type="text/javascript">$(function() {$.jCryption.defaultOptions.getKeysURL = '';$("#dbform").jCryption();});</script>`))
		c.ViewData("SCRIPTS", "jquery")

		c.Render().HTML("splash.php")
	} else {
		defer db.Close()

		q, _ := c.Ctx().Input().GetString("q")
		if q == "wrkfrm" {
			c.Ctx().SetContentType(pine.ContentTypeHTML)
			if err := db.Ping(); err != nil {
				c.Ctx().WriteString(err.Error())
			} else {
				c.Ctx().WriteString(common.ExecuteRequest(db, c.Ctx(), authServe))
			}
			return
		}

		dbname, _ := c.Ctx().Input().GetString("db")
		if dbname != "" && dbname != c.Session().Get("db.name") {
			c.Session().Set("db.change", "1")
			c.Session().Set("db.name", dbname)
			if v, _ := c.Ctx().Input().GetInt("x"); v == 1 {
				c.Ctx().Response.Header.SetContentType(pine.ContentTypeHTML)
				c.Ctx().Write(AjaxResponse)
			} else {
				c.Ctx().Redirect("/mywebsql/index", 302)
			}
			return
		}

		html, dblist, _ := common.PrintDbList(db, c.Session())
		dbname = c.Session().Get("db.name")
		treeHtml := common.GetDatabaseTreeHTML(db, dblist, dbname)
		auth := c.getAuthSession()
		dialogs, _ := c.plush.Exec("dialogs.php", nil)

		c.ViewData("auth", auth)
		c.ViewData("KEY_CODES", common.KEY_CODES)
		c.ViewData("MenuBarHTML", template.HTML(common.GetMenuBarHTML(c.Session().Get("theme_path"))))
		c.ViewData("version", c.Session().Get("db.version"))
		c.ViewData("version_full", c.Session().Get("db.version_full"))
		c.ViewData("version_comment", c.Session().Get("db.version_comment"))
		c.ViewData("dbListHtml", template.HTML(html))
		c.ViewData("treeHtml", template.HTML(treeHtml))
		c.ViewData("contextMenusHTML", template.HTML(common.GetContextMenusHTML()))
		c.ViewData("HotkeysHTML", template.HTML(common.GetHotkeysHTML()))
		c.ViewData("UpdateSqlEditor", template.HTML(common.UpdateSqlEditor()))
		c.ViewData("GetGeneratedJS", template.HTML(common.GetGeneratedJS()))
		c.ViewData("KEYCODE_SETNULL", common.T("Press {{KEY}} to set NULL", common.KEY_CODES["KEYCODE_SETNULL"][1]))
		c.ViewData("LoginUser", common.T("Logged in as: {{USER}}", auth.User))
		c.ViewData("dialogs", template.HTML(dialogs))

		if c.Session().Get("db.change") == "true" {
			c.ViewData("DBCHANGE", template.HTML("document.getElementById(\"messageContainer\").innerHTML = \"Database changed to: "+c.Session().Get("db.name")+"\";"))
			c.Session().Remove("db.change")
		} else {
			c.ViewData("DBCHANGE", template.HTML("document.getElementById(\"messageContainer\").innerHTML = \"Connected to: "+auth.Host+" as "+auth.User+"\";"))
		}

		c.Render().HTML("index.php")
	}

}

// PostIndex 登录提交
func (c *IndexController) PostIndex() {
	authUser, _ := c.Ctx().Input().GetString("auth_user")
	authPwd, _ := c.Ctx().Input().GetString("auth_pwd")
	server, _ := c.Ctx().Input().GetString("server")
	if serve, ok := common.SERVER_LIST[server]; !ok {
		c.hasError = "登录服务器失败"
		c.GetIndex()
	} else {
		serve.User = authUser
		serve.Password = authPwd
		serve.ServerName = server

		if len(serve.User) == 0 {
			c.hasError = "登录账号不能为空"
			c.GetIndex()
			return
		}
		db, err := sqlx.Open(serve.Driver, serve.DSN(""))
		if err != nil {
			c.clearAuthSession()
			c.hasError = err.Error()
			c.GetIndex()
			return
		}

		if err := db.Ping(); err != nil {
			c.clearAuthSession()
			c.hasError = err.Error()
			c.GetIndex()
			return
		}
		defer db.Close()
		c.saveAuthSession(serve)
		common.InitProcess(db, c.Ctx(), &serve)
		c.Ctx().Redirect("/mywebsql/index", 302)
	}

}
