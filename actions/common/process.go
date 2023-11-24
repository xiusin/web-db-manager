package common

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/xiusin/pine"
	"github.com/xiusin/web-db-manager/common"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

const trimChar = " \r\n\t;"

type Process struct {
	*pine.Context
	db         *sqlx.DB
	dbname     string
	lastSQL    string
	affectRows int64
	engine     *xorm.Engine
	auth       *Server
	formData   struct {
		id    string
		page  int
		name  string
		table string
		query string
	}
}

func InitProcess(db *sqlx.DB, ctx *pine.Context, auth *Server) *Process {
	p := &Process{
		db:      db,
		Context: ctx,
		auth:    auth,
	}
	p.SelectVersion()

	p.dbname = p.Session().Get("db.name")

	if p.dbname != "" {
		p.db.Exec("USE " + p.dbname)
	}

	query, _ := p.Input().GetString("query")
	p.formData.query = strings.Trim(query, trimChar)
	name, _ := p.Input().GetString("name")
	p.formData.table = strings.Trim(name, trimChar)
	p.formData.name = p.formData.table
	p.formData.id, _ = p.Input().GetString("id")
	p.formData.page = 1

	if match, _ := regexp.MatchString(`^\d+$`, p.formData.table); match || p.formData.table == "" {
		p.formData.page, _ = strconv.Atoi(p.formData.table)
		qs, _ := p.Input().GetString("query")
		p.formData.table = strings.Trim(qs, trimChar)
	}

	return p
}

func (p *Process) Info() string {
	var html []byte
	if p.dbname != "" {
		tip := T("Database summary") + ": [" + p.dbname + "]"
		html = []byte(p.showDbInfoGrid(tip))
	} else {
		html = []byte(p.Infoserver())
	}
	return string(html)
}

// Showinfo 展示数据表或数据库的基本信息
func (p *Process) Showinfo() string {
	typo, _ := p.Input().GetString("id")
	if typo == "table" || typo == "view" {
		p.formData.id = "table"
		return p.Query()
	} else {
		cmd := p.getCreateCommand(p.formData.id, p.formData.table)
		cmd = p.sanitizeCreateCommand(cmd)

		return string(p.Render("showinfo", pine.H{
			"TYPE":    p.formData.id,
			"NAME":    p.formData.table,
			"COMMAND": template.HTML(cmd),
			"SQL":     p.lastSQL,
		}))
	}

}

// Objcreate 创建对象, 如事务,视图,存储过程等
func (p *Process) Objcreate() string {
	typo := "message ui-state-highlight"
	var msg string
	var refresh bool
	objinfo, _ := p.Input().GetString("objinfo")
	if objinfo != "" {
		pine.Logger().Debug("createObj", objinfo)
		msg = p.createDatabaseObject(objinfo)
		if msg == "" {
			msg = T("The command executed successfully")
			typo = "message ui-state-default"
			refresh = true
		} else {
			typo = "message ui-state-error"
		}
	} else {
		msg = T("Any existing object with the same name should be dropped manually before executing the creation command") +
			"!<br/>" + T("Enter command for object creation")
	}

	return p.displayCreateObjectForm(objinfo, msg, typo, refresh)
}

// displayCreateObjectForm 显示创建对象表单 @ref Objcreate
func (p *Process) displayCreateObjectForm(objInfo, msg, typo string, refresh bool) string {
	if objInfo == "" {
		objInfo = p.getObjectCreateCommand()
	}
	form := "</textarea></td></tr>"

	editorLink := template.HTML("<script type=\"text/javascript\" language=\"javascript\" src=\"/mywebsql/cache?script=editor/codemirror\"></script>")

	editorOptions := template.HTML("parserfile: \"mysql.js\", path: \"/mywebsql/js/editor/\"")

	v := pine.H{
		"ID":             p.formData.id,
		"MESSAGE":        template.HTML(msg),
		"MESSAGE_TYPE":   typo,
		"OBJINFO":        objInfo,
		"EDITOR_LINK":    editorLink,
		"EDITOR_OPTIONS": editorOptions,
		"REFRESH":        "0",
	}
	if refresh {
		v["REFRESH"] = "1"
	}

	return form + string(p.Render("objcreate", v))
}

func (p *Process) createDatabaseObject(info string) string {
	cmd := strings.Trim(info, " \t\r\n;")
	if strings.ToLower(cmd[:6]) != "create" {
		return T("Only create commands are accepted")
	}

	if _, err := p.db.Exec(cmd); err != nil {
		return err.Error()
	}

	ws := p.getWarnings()
	if len(ws) > 0 {
		for _, s := range ws {
			return s
		}
	}
	return ""
}

func (p *Process) getObjectCreateCommand() string {
	templates := map[string]string{
		"0": "templates/table",
		"1": "templates/view",
		"2": "templates/procedure",
		"3": "templates/function",
		"4": "templates/trigger",
		"5": "templates/event",
		"6": "templates/schema",
	}

	return string(p.Render(templates[p.formData.id], nil))
}

// Objlist 切换数据库时触发
func (p *Process) Objlist() string {
	grid := `<div id="objlist">`
	grid += GetDatabaseTreeHTML(p.db, []string{}, p.dbname)
	grid += `</div>`
	return grid
}

func (p *Process) Usermanager() string {
	return "用户管理界面 [低优先级]"
}

// Databases 数据库操作管理
func (p *Process) Databases() string {
	dbs, err := GetDbList(p.db)
	if err != nil {
		return p.createErrorGrid("SHOW DATABASES", err)
	}
	byts, _ := json.Marshal(dbs)
	datas := pine.H{"data": pine.H{"objects": template.HTML(byts)}, "objCount": len(dbs), "stats": nil}
	if p.formData.id == "batch" {
		postdata := p.Input().GetForm().Value
		status := map[string]int{"success": 0, "errors": 0}
		databases := postdata["databases[]"]
		pine.Logger().Warning("删除数据库", databases)
		if len(databases) > 0 {
			for _, database := range databases {
				dropcmd, _ := p.Input().GetString("dropcmd")
				if dropcmd == "on" {
					if err := p.dropObject(database, "database"); err == nil {
						status["success"]++
					} else {
						status["errors"]++
					}
				}
			}
			datas["stats"] = pine.H{"drop": status}
			//>' . str_replace('{{ NUM }}', $data['stats']['drop']['success'], __('{{ NUM }} queries successfully executed')) . '
			txt := "<p><span class=\"ui-icon ui-icon-check\"></span>" + strReplace([]string{"{{ NUM }}"}, []string{strconv.Itoa(status["success"])}, T("{{ NUM }} queries successfully executed")) + "</p>"
			if status["errors"] > 0 {
				txt += "<p><span class=\"ui-icon ui-icon-close\"></span>" + strReplace([]string{"{{ NUM }}"}, []string{strconv.Itoa(status["success"])}, T("{{ NUM }} queries failed to execute")) + "</p>"
			}
			datas["statsHtml"] = txt
		}
	}
	return string(p.Render("databases", datas))
}

func (p *Process) dropObject(name, typo string) error {
	query := "drop " + typo + " `" + name + "`"
	p.lastSQL = query
	_, err := p.db.Exec(query)
	return err
}

// func (p *Process) getObjectTypes() []string {
// 	return []string{"tables", "views", "procedures", "functions", "triggers", "events"}
// }

func (p *Process) getWarnings() map[int]string {
	var ret = map[int]string{}
	if rows, err := p.db.Queryx("SHOW WARNINGS"); err != nil {
		pine.Logger().Warning("获取警告失败", err)
	} else {
		for rows.Next() {
			results := make(map[string]interface{})
			rows.MapScan(results)
			code, _ := strconv.Atoi(string(results["Code"].([]byte)))
			ret[code] = string(results["Message"].([]byte))
		}
	}
	return ret
}

func (p *Process) Infodb() string {
	if p.dbname == "" {
		return string(p.Render("invalid_request", nil))
	}
	query := "show table status where Engine is not null"
	return p.createSimpleGrid(T("Database summary")+": ["+p.dbname+"]", query)
}

func (p *Process) Query() string {
	var querySql string
	if p.formData.id == "table" {
		querySql = p.selectFromTable()
	} else {
		querySql = p.simpleQuery()
	}

	querySql = strings.Trim(querySql, " \r\n")

	if len(querySql) == 0 {
		return p.createErrorGrid(querySql, errors.New("无法生成Query SQL"))
	}

	var html string

	p.loadDbVars()

	//  * For successful SELECT, SHOW, DESCRIBE or EXPLAIN queries, mysqli_query() will return a mysqli_result object.
	//  * For other successful queries mysqli_query() will return TRUE.
	if strings.ToLower(querySql[:6]) == "select" || strings.ToLower(querySql[:4]) == "show" ||
		strings.ToLower(querySql[:8]) == "describe" || strings.ToLower(querySql[:7]) == "explain" {
		queryType := p.getQueryType(querySql)
		if queryType["can_limit"] {
			html = p.createResultGrid(querySql)
		} else {
			html = p.createSimpleGrid(T("Query")+": "+querySql, querySql)
		}
	} else {
		ret, err := p.db.Exec(querySql)
		if err != nil {
			return p.createErrorGrid(querySql, err)
		}

		info := p.getCommandInfo(querySql)
		if info["dbAltered"].(bool) {
			p.Session().Set("db.altered", "true")
		} else if info["setvar"].(bool) && info["variable"].(string) != "" && info["value"].(string) != "" {
			p.setDbVar(info["variable"].(string), info["value"].(string))
		}
		rn, _ := ret.RowsAffected()
		html = p.createDbInfoGrid(querySql, int(rn))
	}

	return html

}

func (p *Process) getCommandInfo(sql string) map[string]interface{} {
	info := map[string]interface{}{
		"db":        "",
		"dbChanged": false,
		"dbAltered": false,
		"setvar":    false,
		"variable":  "",
		"value":     "",
	}
	dbMatch := regexp.MustCompile(`(?i)^[\s]*USE[[:space:]]*([\S]+)`)
	alterMatch := regexp.MustCompile(`(?i)^(CREATE|ALTER|DROP)\s+`)
	setMatch := regexp.MustCompile("(?i)^SET[\\s]+@([a-zA-z0-9_]+|`.*`|\\'.*\\'|\".*\")[\\s]?=[\\s]?(.*)")
	if dbMatch.MatchString(sql) {
		strs := dbMatch.FindAllStringSubmatch(sql, -1)
		info["db"] = strings.Trim(strs[0][1], " ;")
		info["dbChanged"] = true
	} else if alterMatch.MatchString(sql) {
		info["dbAltered"] = true
	} else if setMatch.MatchString(sql) {
		strs := dbMatch.FindAllStringSubmatch(sql, -1)
		info["variable"] = strs[0][1]
		info["value"] = strs[0][2]
		info["setvar"] = true
	}
	return info
}

// loadDbVars 载入会话修改过的变量
func (p *Process) loadDbVars() {
	varByts := []byte(p.Session().Get("vars"))
	var vars = map[string]string{}
	json.Unmarshal(varByts, &vars)
	for variable, value := range vars {
		p.db.Exec("SET @" + variable + " = " + value)
	}
}

func (p *Process) setDbVar(variable, value string) {
	varByts := []byte(p.Session().Get("vars"))
	var vars = map[string]string{}
	json.Unmarshal(varByts, &vars)
	vars[variable] = value
	varByts, _ = json.Marshal(vars)
	p.Session().Set("vars", string(varByts))
}

func (p *Process) sanitizeCreateCommand(cmd string) string {
	cmd = strReplace(
		[]string{" DEFINER=", " FUNCTION ", "BEGIN", "END", " THEN ", "\\n"},
		[]string{"<br>DEFINER=", "<br>FUNCTION ", "<br>BEGIN<br> ", "<br>END", " THEN<br>　", "<br>"},
		cmd)

	cmd = regexp.MustCompile(`;\s*(SET|DECLARE|IF|ELSEIF|END|RETURN)`).ReplaceAllString(cmd, "; <br>　$1")
	return regexp.MustCompile(`[\n|\r]?[\n]+`).ReplaceAllString(cmd, "<br>")
}

func (p *Process) selectFromTable() string {
	recordLimit := p.GetCookie("res-max-count")
	if len(recordLimit) == 0 {
		recordLimit = fmt.Sprintf("%d", MAX_RECORD_TO_DISPLAY)
	}
	recordLimitInt, _ := strconv.Atoi(recordLimit)
	var query string
	if p.formData.page > 1 {
		if p.selectSession("limit") != "true" {
			pine.Logger().Debug("无limit数据")
			return ""
		}
		query = p.selectSession("query")
		table := p.selectSession("table")
		count := p.selectSessionInt("count")
		queryType := p.getQueryType(query)
		if !queryType["result"] || table == "" || count < 1 {
			return ""
		}
		totalPages := int(math.Ceil(float64(count) / float64(recordLimitInt)))
		if totalPages < p.formData.page {
			return ""
		}
		p.Session().Set("select.page", p.n2s(p.formData.page))
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", recordLimitInt, (p.formData.page-1)*recordLimitInt)

	} else {
		keys := []string{"table", "can_limit", "limit", "page", "count"}
		for _, key := range keys {
			p.Session().Remove("select." + key)
		}

		query = "SELECT * FROM `" + p.formData.table + "`"

		p.Session().Set("select.query", query)
		p.Session().Set("select.table", p.formData.table)

		var count int

		if err := p.db.Get(&count, "SELECT COUNT(*) FROM `"+p.formData.table+"`"); err != nil {
			pine.Logger().Warning("获取总数失败", err)
		}

		p.Session().Set("select.count", p.n2s(count))
		p.Session().Set("select.page", "1")
		p.Session().Set("select.can_limit", "true")

		if count > recordLimitInt {
			p.Session().Set("select.limit", "true")
			query += " LIMIT " + recordLimit
		}
	}
	return query
}

func (p *Process) n2s(num int) string {
	return strconv.Itoa(num)
}

func (p *Process) simpleQuery() string {
	query := p.formData.query
	if query == "" {
		query = p.selectSession("query")
	}
	if query == "" {
		return ""
	}

	queryType := p.getQueryType(query)

	if !queryType["result"] || !queryType["can_limit"] {
		return query
	}

	if p.selectSession("can_limit") == "true" {
		p.Session().Set("select.can_limit", "true")
	} else {
		p.Session().Set("select.can_limit", "false")
	}

	if p.formData.id == "sort" {
		field := p.formData.name
		if field != "" {
			query = p.sortQuery(query, field)
		}
	}

	return query
}

func (p *Process) sortQuery(query, field string) string {
	query = strings.Trim(query, " \r\n\t")
	sortType := p.selectSession("sort")
	//limit := ""
	if sortType == "" {
		sortType = "DESC"
	}
	// 匹配LIMIT语句
	// matches := regexp.MustCompile(LIMIT_REGEXP).FindStringSubmatch(query)
	// 匹配sort语句
	matches := regexp.MustCompile(SORT_REGEXP).FindStringSubmatch(query)
	pine.Logger().Debug("matches", matches)

	p.Session().Set("select.sortcol", field)
	p.Session().Set("select.sort", sortType)

	//query += " ORDER BY " + field + " " + sortType + " " + limit

	pine.Logger().Debug("SORT 最终SQL", query)

	return query
}

func (p *Process) getQueryType(query string) map[string]bool {
	types := map[string]bool{"result": false, "can_limit": false, "has_limit": false, "update": false}
	q := strings.ToLower(query[:7])

	if q == "explain" || q[:6] == "select" || q[:4] == "desc" || q[:4] == "show" || q[:4] == "help" {
		types["result"] = true

		if q[:6] == "select" {
			types["can_limit"] = true
		}
	}
	match, err := regexp.MatchString(LIMIT_REGEXP, query)
	if err != nil {
		pine.Logger().Error("匹配sql错误", err)
	}
	if match && strings.Contains(strings.ToLower(query), "limit") {
		types["has_limit"] = true
	}
	return types
}

func (p *Process) exec(module string) (html string) {
	action := common.UcFirst(module)
	defer func() {
		if err := recover(); err != nil {
			pine.Logger().Error(err)
			html = p.createErrorGrid("解析执行方法失败: "+action, fmt.Errorf("%s", err))
		}
	}()
	val := reflect.ValueOf(&p).Elem().MethodByName(action)
	if !val.IsValid() {
		html = p.createErrorGrid("无法解析处理句柄", nil)
	} else {
		html = val.Call([]reflect.Value{})[0].String()
	}
	return
}

func (p *Process) QueryVariables() []Variable {
	query := "SHOW VARIABLES"
	var variables []Variable
	if err := p.db.Select(&variables, query); err != nil {
		pine.Logger().Warning("获取变量失败", err)
	}
	return variables
}

func (p *Process) createErrorGrid(query string, err error, params ...int) string {
	if query == "" {
		query = p.Session().Get("select.query")
	}

	sessionKeys := []string{"result", "pkey", "ukey", "mkey", "unique_table"}
	for _, v := range sessionKeys {
		p.Session().Remove("select." + v)
	}

	numQueries, affectedRows := 0, -1
	if len(params) > 0 {
		numQueries = params[0]
	}
	if len(params) > 1 {
		affectedRows = params[1]
	}
	grid := "<link rel=\"stylesheet\" type=\"text/css\" href=\"/mywebsql/cache?css=theme,default\" />"
	grid += "<div id='results'>\n"
	if numQueries > 0 {
		grid += "<div class=\"message ui-state-default\">"
		var msg string
		if numQueries == 1 {
			msg = T("1 query successfully executed")
		} else {
			msg = strings.ReplaceAll(T("{{NUM}} queries successfully executed"), "{{NUM}}", strconv.Itoa(numQueries))
		}

		msg += "<br/><br/>" + strings.ReplaceAll(T("{{NUM}} record(s) were affected"), "{{NUM}}", strconv.Itoa(affectedRows))
		grid += msg + "</div>"
	}

	formattedQuery := regexp.MustCompile(`[\n|\r]?[\n]+`).ReplaceAllString(query, "<br />")

	grid += "<div class=\"message ui-state-error\">" + T("Error occurred while executing the query") +
		":</div><div class=\"message ui-state-highlight\">" + err.Error() +
		"</div><div class=\"sql-text ui-state-error\">" + formattedQuery + "</div>"

	grid += "</div>"
	grid += "<script type=\"text/javascript\" language='javascript'> parent.transferResultMessage(-1, '&nbsp;', '"
	grid += T("Error occurred while executing the query") + "');\n"
	grid += "parent.addCmdHistory(\"" + template.HTMLEscapeString(formattedQuery) + "\");\n"
	grid += "parent.resetFrame();\n"
	grid += "</script>\n"

	return grid
}

func (p *Process) Infoserver() string {
	grid := ""
	variables := p.QueryVariables()
	if len(variables) == 0 {
		return ""
	}
	v := pine.H{"JS": "", "SERVER_NAME": "MySQL"}

	for _, variable := range variables {
		switch variable.VariableName {
		case "version":
			v["SERVER_VERSION"] = variable.Value
		case "version_comment":
			v["SERVER_COMMENT"] = variable.Value
		case "character_set_server":
			v["SERVER_CHARSET"] = variable.Value
		case "character_set_client":
			v["CLIENT_CHARSET"] = variable.Value
		case "character_set_database":
			v["DATABASE_CHARSET"] = variable.Value
		case "character_set_results":
			v["RESULT_CHARSET"] = variable.Value
		}
		v[variable.VariableName] = variable.Value
	}

	if p.dbname == "" {
		v["JS"] = template.HTML(`parent.$("#main-menu").find(".db").hide();`)
	}
	grid += string(p.Render("infoserver", v))
	return grid
}

func (p *Process) createSimpleGrid(message string, query string) string {
	grid := "<div id='results'>"
	grid += "<div class='message ui-state-default'>" + message + "<span style='float:right'>" + T("Quick Search") +
		"&nbsp;<input type=\"text\" id=\"quick-info-search\" maxlength=\"50\" /></div>"
	grid += "<table cellspacing='0' width='100%' border='0' class='results' id='infoTable'><thead>\n"

	grid += "<tr id='fhead'><th class='th index'><div>#</div></th>\n"

	// 遍历数据
	rows, err := p.db.Query(query)
	if err != nil {
		pine.Logger().Warning("查询异常", query, err)
		return p.createErrorGrid(query, err)
	}

	fields, _ := rows.Columns()
	fieldTypes, _ := rows.ColumnTypes()

	for k, fn := range fields {
		cls, dsrt := "th", "text"
		fieldType := fieldTypes[k]
		if exist, _ := common.InArray(fieldType.DatabaseTypeName(), []string{"DECIMAL", "INT", "BIGINT", "TINYINT", "FLOAT", "DOUBLE"}); exist {
			cls = "th_numeric"
			dsrt = "numeric"
		}
		grid += "<th nowrap=\"nowrap\" class='" + cls + "' data-sort='" + dsrt + "'><div>" + fn + "</div></th>\n"
	}

	grid += "</tr></thead><tbody>\n"

	datas, err := p.row2arrMap(rows)
	if err != nil {
		pine.Logger().Warning("转换数据类型异常", query, err)
		return p.createErrorGrid(query, err)
	}

	for j, table := range datas {
		grid += `<tr id="rc` + strconv.Itoa(j) + `" class="row"><td class="tj">` + strconv.Itoa(j+1) + `</td>`

		for i, field := range fields {
			class := "tl"
			if j == 0 {
				pine.Logger().Print(field, fieldTypes[i].ScanType().Name(), fieldTypes[i].DatabaseTypeName())
			}
			rs := table[field]
			data := "&nbsp;"
			if rs == nil {
				class = "tnl"
				data = "NULL"
			}

			switch fieldTypes[i].DatabaseTypeName() {
			case "VARCHAR", "CHAR", "TEXT":
				class += " text"
				if rs != nil && len(rs.([]byte)) != 0 {
					data = string(rs.([]byte))
				}
			case "INT", "BIGINT":
				class = "tr"
				if rs != nil && len(rs.([]byte)) != 0 {
					data = string(rs.([]byte))
				}
			case "TIMESTAMP", "DATETIME":
				if rs != nil {
					data = rs.(time.Time).Format(common.TimeFormat)
				}
			case "binary", "blob": // blob

			}

			grid += "<td nowrap=\"nowrap\" id=\"r" + p.n2s(j) + "f" + p.n2s(i) + "\" class=\"" + class + "\">" + data + "</td>"
		}
		grid += "</tr>\n"
	}

	grid += "</tbody></table>"
	grid += "</div>"

	grid += "<script type=\"text/javascript\" language=\"javascript\">\n"
	grid += "parent.transferInfoMessage();\n"
	grid += "parent.resetFrame();\n"
	grid += "</script>"
	return grid
}

func (p *Process) createResultGrid(query string) string {
	sessionKeys := []string{"pkey", "ukey", "mkey", "unique_table"}
	for _, v := range sessionKeys {
		p.Session().Remove("select." + v)
	}

	var grid = ""

	recordLimit := p.GetCookie("res-max-count")
	if len(recordLimit) == 0 {
		recordLimit = fmt.Sprintf("%d", MAX_RECORD_TO_DISPLAY)
	}
	recordLimitInt, _ := strconv.Atoi(recordLimit)

	grid += "<div id='results'>"
	grid += "<table cellspacing=\"0\" width='100%' border=\"0\" class='results' id=\"dataTable\"><thead>\n"

	// 遍历数据
	rows, err := p.db.Query(query)

	if err != nil {
		pine.Logger().Warning("查询异常", query, err)
		return p.createErrorGrid(query, err)
	}

	// TODO 根据查询类型确定查询表字段, 如编辑器查询可能就没有table名称, 需要自己解析出来 目前GO没有此类方法识别字段所属表
	f := p.getFieldInfo()

	isUniqueTable := len(f) > 0 // TODO 非同表不允许添加修改数据

	if len(f) == 0 {
		cts, _ := rows.ColumnTypes()
		for _, ct := range cts {
			nn, _ := ct.Nullable()
			nic, _ := common.InArray(ct.DatabaseTypeName(), []interface{}{"DECIMAL", "INT", "BIGINT", "TINYINT", "FLOAT", "DOUBLE"})
			f = append(f, &Column{
				ColumnName: ct.Name(),
				NotNull:    nn,
				Numeric:    nic,
			})
		}
	}

	if p.Session().Get("select.can_limit") == "true" && isUniqueTable {
		p.Session().Set("select.unique_table", f[0].TableName)
	}

	grid += "<tr id=\"fhead\">"
	grid += "<th class=\"th tch\"><div>#</div></th>"

	// 标记是否可以编辑
	ed := p.Session().Get("select.can_limit") == "true" && p.Session().Get("select.unique_table") != ""

	if ed {
		grid += "<th class=\"th_nosort tch\"><div><input class=\"check-all\" type=\"checkbox\" onclick=\"resultSelectAll()\" title=\"" + T("Select/unselect All records") + "\" /></div></th>"
	}

	var pkey, ukey []string //, mkey
	fieldNames := ""
	fieldInfo, _ := json.Marshal(&f)

	for i, column := range f {
		cls := "th"
		if column.Numeric {
			cls = "th_numeric"
		}
		grid += "<th nowrap=\"nowrap\" class='" + cls + "'><div>"
		if column.PKey {
			pkey = append(pkey, column.ColumnName)
			grid += "<span class='pk' title='" + T("Primary key column") + "'>&nbsp;</span>"
		}
		if column.UKey {
			ukey = append(ukey, column.ColumnName)
			grid += "<span class='uk' title='" + T("Unique key column") + "'>&nbsp;</span>"
		}
		// if column.MKey && !column.Blob {
		// 	mkey = append(mkey, column.ColumnName)
		// }
		grid += column.ColumnName
		// 排序应用
		if p.Session().Get("select.sortcol") == strconv.Itoa(i+1) {
			if p.Session().Get("select.sort") == "DESC" {
				grid += "&nbsp;&#x25BE;"
			} else {
				grid += "&nbsp;&#x25B4"
			}
		}
		grid += "</div></th>"
		fieldNames += "'" + strings.ReplaceAll(column.ColumnName, "'", "\\'") + "',"
	}

	grid += "</tr></thead><tbody>\n"

	datas, err := p.row2arrMap(rows)

	if err != nil {
		pine.Logger().Warning("转换数据类型异常", query, err)
		return p.createErrorGrid(query, err)
	}

	for j, r := range datas {
		grid += "<tr class=\"row\">"
		grid += "<td class=\"tj\">" + strconv.Itoa(j+1) + "</td>"
		if ed {
			grid += "<td class=\"tch\"><input type=\"checkbox\" /></td>"
		}
		for _, column := range f {
			rs := r[column.ColumnName]
			class := "tl"
			if column.Numeric {
				class = "tr"
			}
			if rs == nil {
				class = "tnl"
			}
			if ed {
				class += " edit"
			}

			data := ""
			if !column.Blob {
				if rs == nil {
					data = "NULL"
				} else {
					if v, ok := rs.([]byte); ok && len(v) != 0 {
						data = string(rs.([]byte))
					} else if v, ok := rs.(time.Time); ok {
						data = v.Format(common.TimeFormat)
					} else {
						data = fmt.Sprintf("%s", rs)
					}
				}
				data = template.HTMLEscapeString(data) // 处理带有标签的字段
			} else {
				data = p.getBlobDisplay(rs, column, j, ed) // 大字段单独输入框处理
			}

			grid += "<td nowrap=\"nowrap\" striped=true width=80 class=\"" + class + "\">" + data + "</td>"
		}
		grid += "</tr>\n"
	}

	numRows := len(datas)

	grid += "</tbody></table></div>"

	editTableName := p.Session().Get("select.unique_table")

	gridTitle := T("Query Results")
	if editTableName != "" {
		gridTitle = T("Data for {{TABLE}}", gridTitle)
	}

	grid += "<div id=\"title\">" + gridTitle + "</div>"

	var total_records, current_page, total_pages int
	var message string

	if p.Session().Get("select.can_limit") == "true" {
		if p.Session().Get("select.limit") == "true" {
			total_records = p.selectSessionInt("count")
			total_pages = int(math.Ceil(float64(total_records) / float64(recordLimitInt)))
			current_page = p.selectSessionInt("page")
			from := (current_page-1)*recordLimitInt + 1
			to := from + numRows - 1
			message = "<div class='numrec'>" + T("Showing records {{START}} - {{END}}", p.n2s(from), p.n2s(to)) + "</div>"
		} else {
			total_records = numRows
			total_pages = 1
			current_page = 1
			if recordLimitInt > 0 && total_records > recordLimitInt {
				message = "<div class='numrec'>" + T("Showing first {{MAX}} records only", p.n2s(recordLimitInt)) + "!</div>"
			}
		}
	} else {
		total_records = numRows
		total_pages = 1
		current_page = 1
		message = ""
	}

	js := "<script type=\"text/javascript\" language=\"javascript\">\n"
	if len(pkey) > 0 {
		js += "parent.editKey = " + jsonEncode(&pkey) + ";\n"
	} else if len(ukey) > 0 {
		js += "parent.editKey = " + jsonEncode(&ukey) + ";\n"
	} else {
		js += "parent.editKey = [];\n"
	}

	js += "parent.editTableName = \"" + editTableName + "\";\n"
	js += "parent.fieldInfo = " + string(fieldInfo) + ";\n"
	js += "parent.queryID = '" + common.GetMd5(p.selectSession(query)) + "';\n"
	js += "parent.totalRecords = " + p.n2s(total_records) + ";\n"
	js += "parent.totalPages = " + p.n2s(total_pages) + ";\n"
	js += "parent.currentPage = " + p.n2s(current_page) + ";\n"
	if p.selectSession("table") != "" {
		js += "parent.queryType = \"table\";\n"
	} else {
		js += "parent.queryType = \"query\";\n"
	}
	js += "parent.transferResultGrid(" + p.n2s(numRows) + ", '0', \"" + message + "\");\n"
	js += "parent.addCmdHistory(\"" + strings.ReplaceAll(template.HTMLEscapeString(p.selectSession("query")), "\\n\\r", "<br/>") + "\", 1);\n"
	js += "parent.resetFrame();\n"
	js += "</script>\n"

	grid += js

	return grid

}

func (p *Process) selectSessionInt(key string) int {
	d, _ := strconv.Atoi(p.Session().Get("select." + key))
	return d
}

func (p *Process) selectSession(key string) string {
	return p.Session().Get("select." + key)
}

func (p *Process) getBlobDisplay(rs interface{}, info *Column, numRecord int, editable bool) string {
	binary := info.DataType == "binary"
	span := "<span class=\"i\">"
	var length int
	var size string
	if rs == nil {
		length = 0
		size = "0 B"
		span += "NULL"
	} else {
		length := len(rs.([]byte))
		size = formatBytes(int64(length))
		if length == 0 {
			span += "&nbsp;"
		} else {
			if MAX_TEXT_LENGTH_DISPLAY >= length {
				span += string(rs.([]byte))
			} else if binary {
				span += strings.Replace(T("Blob Data [{{SIZE}}]"), "{{SIZE}}", size, 1)
			} else {
				span += strings.Replace(T("Text Data [{{SIZE}}]"), "{{SIZE}}", size, 1)
			}
		}
	}

	extra, btype := "", "text"
	if binary {
		pine.Logger().Debug("处理为binary")
	}

	span += "</span>"

	// if binary {
	// 	//$span .= "<span title=\"" . str_replace('{{NUM}}', $length, __('Click to view/edit column data [{{NUM}} Bytes]')). "\" class=\"blob $btype\" $extra>&nbsp;</span>";
	// 	//return $span;
	// }
	data := ""
	if rs != nil {
		data = string(rs.([]byte))
	} else {
		rs = "NULL"
	}
	span += "<span class=\"d\" style=\"display:none\">" + template.HTMLEscapeString(data) + "</span>"

	if !editable && rs != nil && MAX_TEXT_LENGTH_DISPLAY < length {
		extra = `onclick="vwTxt(this, &quot;` + size + `&quot;, '` + btype + `')"`
		span += "<span title=\"" + strings.ReplaceAll("Click to view/edit column data [{{NUM}} Bytes]", "{{NUM}}", strconv.Itoa(length)) + "\" class=\"blob " + btype + "\" " + extra + ">&nbsp;</span>"
	}

	return span
}

func (p *Process) showDbInfoGrid(message string) string {
	grid := "<div id='results'>"
	grid += "<div class='message ui-state-default'>" + message + "<span style='float:right'>" + T("Quick Search") +
		"&nbsp;<input type=\"text\" id=\"quick-info-search\" maxlength=\"50\" /></div>"

	grid += "<table cellspacing='0' width='100%' border='0' class='results' id='infoTable'><thead>\n"

	//fields := p.getFieldInfo()
	dbname := p.Session().Get("db.name")
	grid += "<tr id='fhead'><th class='th index'><div>#</div></th>\n"

	headers := GetTableInfoHeaders()

	// 数字类型的字段
	numericFields := []string{"Version", "Rows", "Avg_row_length", "Data_length", "Max_data_length", "Index_length", "Auto_increment"}

	for _, header := range headers {
		cls, dsrt := "th", "text"
		if ok, _ := common.InArray(header, numericFields); ok {
			cls = "th_numeric"
			dsrt = "numeric"
		}
		grid += "<th nowrap=\"nowrap\" class='" + cls + "' data-sort='" + dsrt + "'><div>" + header + "</div></th>\n"
	}

	grid += "</tr></thead><tbody>\n"

	// 遍历数据
	tables := getTables(p.db, dbname)

	// ------------ print data -----------
	for j, table := range tables {
		grid += `<tr id="rc` + strconv.Itoa(j) + `" class="row"><td class="tj">` + strconv.Itoa(j+1) + `</td>`

		vs := reflect.ValueOf(&table).Elem()

		for i := 0; i < vs.NumField(); i++ {
			rs := vs.Field(i).Interface()
			class := "tl"
			data := ""
			if rs == nil {
				class = "tnl"
				data = "NULL"
			} else if strings.Contains(strings.ToLower(vs.Field(i).String()), "int") {
				class = "tr"
			}

			switch vs.Field(i).Kind() {
			case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
				data = fmt.Sprintf("%d", vs.Field(i).Int())
			case reflect.Uint, reflect.Uint16, reflect.Uint8, reflect.Uint32, reflect.Uint64:
				data = fmt.Sprintf("%d", vs.Field(i).Uint())
			case reflect.String:
				class = "tl"
				data = vs.Field(i).String()
			default:
				sqlField := vs.Field(i)
				// 其他类型格式兼容
				if strings.Contains(vs.Field(i).Type().String(), "*sql.Null") {
					sqlField = sqlField.Elem()
				}

				if !sqlField.IsValid() || sqlField.IsZero() {
					data, class = "NULL", "tnl"
				} else {
					if strings.Contains(sqlField.Type().String(), "sql.Null") {
						val := sqlField.Field(0).Interface()
						if v, ok := val.(time.Time); ok {
							data = v.Format(common.TimeFormat)
						} else if v, ok := val.(int64); ok {
							data = fmt.Sprintf("%d", v)
						} else if v, ok := val.(int32); ok {
							data = fmt.Sprintf("%d", v)
						} else if v, ok := val.(float64); ok {
							data = strconv.FormatFloat(v, 'f', 6, 64)
						} else if v, ok := val.(bool); ok {
							data = strconv.FormatBool(v)
						} else if v, ok := val.(string); ok {
							data = v
						} else {
							data = fmt.Sprintf("%s", vs.Field(i).Interface())
						}
					} else {
						data = fmt.Sprintf("%s", vs.Field(i).Interface())
					}
				}
			}

			class += " text"
			grid += "<td nowrap=\"nowrap\" id=\"r" + strconv.Itoa(j) + "f" + strconv.Itoa(i) + "\" class=\"" + class + "\">" + template.HTMLEscapeString(data) + "</td>\n"
		}
		grid += "</tr>\n"
	}

	grid += "</tbody></table>"
	grid += "</div>"

	grid += "<script type=\"text/javascript\" language=\"javascript\">\n"
	grid += "parent.transferInfoMessage();\n"
	grid += "parent.resetFrame();\n"
	grid += "</script>"
	return grid
}

func (p *Process) createDbInfoGrid(query string, numQueries int) string {
	p.removeSelectSession([]string{"pkey", "ukey", "mkey", "unique_table"})
	if query == "" {
		query = p.formData.query
	}

	grid := "<div id='results'>\n"
	grid += "<div class=\"message ui-state-default\">"

	msg := T("1 query successfully executed")

	if numQueries != 1 {
		msg = strReplace([]string{"{{NUM}}"}, []string{fmt.Sprintf("%d", numQueries)}, T("{{NUM}} queries successfully executed"))
	}

	grid += msg + ".</div>"
	if p.affectRows > 0 {
		grid += "<div class=\"message ui-state-highlight\">"
		grid += strReplace([]string{"{{NUM}}"}, []string{fmt.Sprintf("%d", p.affectRows)}, T("{{NUM}} record(s) were affected"))
		grid += "</div>"
	}

	if numQueries == 1 {
		match := regexp.MustCompile(`[\n|\r]?[\n]+`)
		formattedQuery := match.ReplaceAllString(query, "<br>")
		grid += "<div class='sql-text ui-state-default'>" + formattedQuery + "</div>"
		warnings := p.getWarnings()

		if len(warnings) > 0 {
			grid += "<div class=\"message ui-state-error\">"
			for _, s := range warnings {
				grid += s + "<br />"
			}
			grid += "</div>"
		}
	}
	grid += "</div>"
	grid += "<script type=\"text/javascript\" language='javascript'> parent.transferResultMessage(-1, '0', '" +
		T("{{NUM}} record(s) updated", fmt.Sprintf("%d", p.affectRows)) + "');\n"

	match := regexp.MustCompile(`[\n\r]`)
	grid += "parent.addCmdHistory(\"" + template.HTMLEscapeString(match.ReplaceAllString(query, "<br>")) + "\");\n"

	if p.Session().Get("db.altered") == "true" {
		p.Session().Remove("db.altered")
		grid += "parent.objectsRefresh();\n"
	}

	grid += "parent.resetFrame();\n"
	grid += "</script>\n"
	return grid
}

func (p *Process) getFieldInfo(table ...string) []*Column {
	if p.dbname == "" {
		panic(errors.New("请选择数据库后再进行操作"))
	}
	tbl := p.formData.table
	if len(table) > 0 {
		tbl = table[0]
	}
	query := "SELECT * FROM information_schema.columns WHERE table_schema = '" + p.dbname + "' AND table_name = '" + tbl + "'"
	var columns []*Column
	if err := p.db.Select(&columns, query); err != nil {
		pine.Logger().Warning("获取表字段信息失败", err)
	}
	for _, column := range columns {
		column.Fill()
	}
	return columns
}

func (p *Process) SelectVersion() {
	var variables []Variable

	if err := p.db.Select(&variables, "SHOW VARIABLES LIKE 'version%'"); err != nil {
		pine.Logger().Warning("获取版本信息失败", err)
	}
	if len(variables) > 0 {
		for _, variable := range variables {
			if variable.VariableName == "version" {
				p.Session().Set("db.version", strings.Split(variable.Value, ".")[0])
				p.Session().Set("db.version_full", variable.Value)
			} else if variable.VariableName == "version_comment" {
				p.Session().Set("db.version_comment", variable.Value)
			}
		}
	}
}

// 获取创建命令语句
func (p *Process) getCreateCommand(typo, name string) string {
	sql, cmd := "", ""

	var createCommand interface{}

	if typo == "trigger" {
		sql = "show triggers where `trigger` = '" + name + "'"
		createCommand = &CreateTriggerCommand{}
	} else {
		if typo == "oview" || typo == "view" {
			typo = "view"
			createCommand = &CreateViewCommand{}
		}
		if typo == "table" {
			createCommand = &CreateCommand{}
		}

		if typo == "function" {
			createCommand = &CreateFunctionCommand{}
		}

		if typo == "procedure" {
			createCommand = &CreateProcedureCommand{}
		}

		sql = "show create " + typo + " `" + name + "`"
	}
	if err := p.db.Get(createCommand, sql); err != nil {
		panic(err)
	}
	p.lastSQL = sql
	if typo == "view" {
		cmd = createCommand.(*CreateViewCommand).CreateView
	} else if typo == "table" {
		cmd = createCommand.(*CreateCommand).CreateTable
	} else if typo == "trigger" {
		createCmd := createCommand.(*CreateTriggerCommand)
		cmd = "create trigger `" + createCmd.Trigger + "`\r\n" + createCmd.Timing + " " +
			createCmd.Event + " on `" + createCmd.Table + "`\r\nfor each row\r\n" + createCmd.Statement
	} else if typo == "function" {
		cmd = createCommand.(*CreateFunctionCommand).CreateFunction
	} else if typo == "procedure" {
		cmd = createCommand.(*CreateProcedureCommand).CreateProcedure
	}
	// _, cmd2 := common.FormatSQL(cmd)
	// pine.Logger().Debug(cmd1, cmd2)
	return cmd
}

func (p *Process) GetDropCommand(table string) string {
	return "drop table if exists `" + table + "`"
}

func (p *Process) GetTruncateCommand(table string) string {
	return "truncate table `" + table + "`"
}

// GetEngines 获取数据库支持的引擎
func (p *Process) GetEngines() []string {
	var engines []Engine
	if err := p.db.Select(&engines, "SHOW ENGINES"); err != nil {
		pine.Logger().Warning("获取存储引擎失败", err)
	}
	var ret []string
	for _, engine := range engines {
		if engine.Support != "NO" {
			ret = append(ret, engine.Engine)
		}
	}
	return ret
}

func (p *Process) GetCharsets() []string {
	var ret []string
	if rows, err := p.db.Queryx("SHOW CHARACTER SET"); err != nil {
		pine.Logger().Warning("获取字符集失败", err)
	} else {
		for rows.Next() {
			results := make(map[string]interface{})
			rows.MapScan(results)
			ret = append(ret, string(results["Charset"].([]byte)))
		}
		sort.Strings(ret)
	}
	return ret
}

func (p *Process) GetCollations() []string {
	var collations []Collation
	p.db.Select(&collations, "SHOW COLLATION")
	var ret []string
	for _, collation := range collations {
		ret = append(ret, collation.Collation)
	}
	sort.Strings(ret)
	return ret
}

func (p *Process) Logout() string {
	p.Session().Destroy() // 销毁session
	return string(p.Render("logout", nil))
}

// Infovars 服务器变量
func (p *Process) Infovars() string {
	return p.createSimpleGrid(T("Server Variables"), "SHOW VARIABLES")
}

// Search 搜索数据
func (p *Process) Search() string {
	return "搜索数据 [低优先级]"
}

// Indexes 索引设置
func (p *Process) Indexes() string {
	if p.formData.id == "alter" {

		return ""
	}
	return string(p.displayIndexesForm())
}

func (p *Process) displayIndexesForm() []byte {
	indexes := p.getIndexes()
	fields := p.getFields()

	return p.Render("indexes", pine.H{
		"ID":         p.formData.id,
		"MESSAGE":    T("Changes are not saved until you press [Save All Changes]"),
		"INDEXES":    template.HTML(jsonEncode(&indexes)),
		"FIELDS":     template.HTML(jsonEncode(&fields)),
		"TABLE_NAME": p.formData.name,
	})
}

func (p *Process) getIndexes() map[string][]Index {
	sql := "show indexes from `" + p.formData.name + "`"
	var indexes []Index
	p.lastSQL = sql
	if err := p.db.Select(&indexes, sql); err != nil {
		pine.Logger().Warning("获取"+p.formData.name+"索引失败", err)
	}
	var indexMap = map[string][]Index{}
	for k := range indexes {
		im := indexMap[indexes[k].KeyName]
		im = append(im, indexes[k])
		indexMap[indexes[k].KeyName] = im
	}

	return indexMap
}

func (p *Process) getFields() []*Field {
	sql := "show fields from `" + p.formData.name + "`"
	var ff []*Field
	p.lastSQL = sql

	if err := p.db.Select(&ff, sql); err != nil {
		pine.Logger().Warning("获取"+p.formData.name+"字段列表失败", err)
	}

	for _, field := range ff {
		field.fetchFieldInfo()
	}
	return ff
}

// Enginetype 存储引擎切换
func (p *Process) Enginetype() string {
	return "存储引擎"
}

func (p *Process) Altertbl() string {
	// if p.formData.id == "alter" {

	// } else {

	// }
	return "修改结构"
}

// Options 选项配置 用于配置系统内数据
func (p *Process) Options() string {
	pk, _ := p.Input().GetString("p", "ui")

	pagesort := []string{"results", "editing", "misc", "ui"}

	pages := pine.H{
		"results": T("Results"),
		"editing": T("Record Editing"),
		"misc":    T("Miscellaneous"),
		"ui":      T("Interface"),
	}

	if _, exist := pages[pk]; !exist {
		pk = "ui"
	}

	content := string(p.Render("options/"+pk, nil))

	lis := ""
	for _, x := range pagesort {
		y := pages[x]
		if pk == x {
			lis += "<li class=\"current\"><img border=\"0\" align=\"absmiddle\" src='/mywebsql/img/options/o_" + x + ".gif' alt=\"\" />" + y.(string) + "</li>"
		} else {
			lis += "<li><a href=\"#" + x + "\"><img border=\"0\" align=\"absmiddle\" src='/mywebsql/img/options/o_" + x + ".gif' alt=\"\" />" + y.(string) + "</a></li>"
		}
	}

	return string(p.Render("options", pine.H{
		"CONTENT": template.HTML(content),
		"lis":     template.HTML(lis),
		"data": pine.H{
			"pages": pages,
			"page":  pk,
		},
	}))

}

// Queryall 执行语句, 如插入数据, 提交数据 此方法不支持限制表格, 查询要使用query
func (p *Process) Queryall() string {
	// 按照每行以;换行为一个执行语句
	qs := strings.Split(strings.Trim(p.formData.query, trimChar)+";", ";\r\n")

	var success int64  // 影响行数
	var sqlSuccess int // 执行成功数

	for _, query := range qs {
		query = strings.Trim(query, trimChar)
		if len(query) > 0 {
			if ret, err := p.db.Exec(query); err != nil {
				return p.createErrorGrid(query, err, sqlSuccess, int(success))
			} else {
				row, _ := ret.RowsAffected()
				success += row
				sqlSuccess++
			}
		}
	}
	p.affectRows = success
	return p.createDbInfoGrid(p.formData.name, sqlSuccess)
}

// Truncate 截断数据表
func (p *Process) Truncate() string {
	if p.formData.name == "" {
		return p.createErrorGrid("", errors.New("参数错误"))
	}
	p.lastSQL = p.GetTruncateCommand(p.formData.name)
	if ret, err := p.db.Exec(p.lastSQL); err != nil {
		return p.createErrorGrid(p.lastSQL, err)
	} else {
		p.affectRows, _ = ret.RowsAffected()
		return p.createDbInfoGrid(p.lastSQL, 1)
	}
}

// Drop 删除数据表
func (p *Process) Drop() string {
	if p.formData.name == "" {
		return p.createErrorGrid("", errors.New("没有删除的表参数"))
	}
	if err := p.dropObject(p.formData.name, p.formData.id); err != nil {
		return p.createErrorGrid(p.lastSQL, err)
	}
	p.Session().Set("db.altered", "true")
	return p.createDbInfoGrid(p.lastSQL, 1)
}

// Rename 重命名
func (p *Process) Rename() string {
	newName := p.formData.query

	if newName == "" || p.formData.name == "" {
		return p.createErrorGrid("", errors.New("缺少参数"))
	}

	if err := p.renameObject(newName); err != nil {
		return p.createErrorGrid(p.lastSQL, err)
	}
	p.Session().Set("db.altered", "true")
	numQueries := 1
	if p.formData.id == "table" {
		numQueries = 2
	}
	return p.createDbInfoGrid(p.lastSQL, numQueries)
}

func (p *Process) renameObject(newName string) error {
	if p.formData.id == "table" {
		p.lastSQL = "rename table `" + p.formData.name + "` to `" + newName + "`"
		if rows, err := p.db.Exec(p.lastSQL); err == nil {
			p.affectRows, _ = rows.RowsAffected()
		} else {
			return err
		}
	}
	//  else {
	//cmd := p.getCreateCommand(p.formData.id, p.formData.name)
	//search := ""	// TODO 忽略大小写匹配
	// }
	return nil
}

// Dbrepair 修复表
func (p *Process) Dbrepair() string {
	// TODO tables 必须得数组
	optype, _ := p.Input().GetString("optype")
	tables, _ := p.Input().GetString("tables[]")
	if optype != "" && tables != "" {
		return p.checkTables()
	} else {
		tableStrs := getTables(p.db, p.dbname)
		var tables []string
		for _, v := range tableStrs {
			tables = append(tables, v.Name)
		}
		byts, _ := json.Marshal(&tables)
		extra := ""
		if len(tables) > 0 {
			extra += "$('#db_objects').html('');\n"
			extra += "uiShowObjectList(tables, 'tables', '" + T("Tables") + "');\n"
		}

		return string(p.Render("dbrepair", pine.H{
			"tables":  string(byts),
			"extraJs": template.HTML(extra)}))
	}
}

func (p *Process) checkTables() string {
	typo, _ := p.Input().GetString("optype")
	options := map[string]interface{}{}

	postdata := p.Input().GetForm().Value
	pine.Logger().Debug("postData", postdata)
	skiplog, _ := p.Input().GetString("skiplog")
	if skiplog == "on" {
		options["skiplog"] = true
	} else {
		options["skiplog"] = false
	}
	options["checktype"], _ = p.Input().GetString("checktype")
	options["repairtype"] = postdata["repairtype"]
	tables := postdata["tables[]"]
	checker := NewTableChecker(p.db)
	checker.SetOperation(typo)
	checker.SetOptions(options)
	checker.SetTables(tables)
	checker.Runcheck()
	results := checker.GetResults()
	byts, _ := json.Marshal(&results)
	return string(p.Render("dbrepair_results", pine.H{"RESULTS": template.HTML(string(byts))}))
}

// Dbcreate 创建表
func (p *Process) Dbcreate() string {
	p.removeSelectSession([]string{"result", "pkey", "ukey", "mkey", "unique_table"})

	name, _ := p.Input().GetString("name")
	dbSelect, _ := p.Input().GetString("query")

	sql := "create database `" + name + "`"

	if _, err := p.db.Exec(sql); err != nil {
		return p.createErrorGrid("", err)
	}
	redirect := "0"

	if dbSelect != "" {
		p.Session().Set("db.change", "true")
		p.Session().Set("db.name", name)
		redirect = "1"
	}

	return string(p.Render("dbcreate", pine.H{
		"DB_NAME":  name,
		"SQL":      sql,
		"TIME":     0,
		"REDIRECT": redirect,
	}))

}

// Tableinsert 插入表数据
func (p *Process) Tableinsert() string {
	sql, err := p.getInsertStatement(p.formData.name)
	if err != nil {
		return p.createErrorGrid(p.lastSQL, err)
	}
	msg := "<div id='results'>" + sql + "</div>"
	msg += "<script type=\"text/javascript\" language='javascript'> parent.transferQuery(); </script>\n"
	return msg
}

func (p *Process) getInsertStatement(tbl string) (string, error) {
	query := "show full fields from `" + tbl + "`"
	p.lastSQL = query
	if rows, err := p.db.Queryx(query); err != nil {
		return "", err
	} else {
		str := "INSERT INTO `" + tbl + "` ("
		str2 := ""
		defer rows.Close()
		var i = 0
		for rows.Next() {
			results := make(map[string]interface{})
			rows.MapScan(results)
			if i == 0 {
				str += "`" + string(results["Field"].([]byte)) + "`"
				if string(results["Extra"].([]byte)) == "auto_increment" {
					str2 += " VALUES (NULL"
				} else {
					str2 += " VALUES (\"\""
				}
			} else {
				str += ",`" + string(results["Field"].([]byte)) + "`"
				if string(results["Extra"].([]byte)) == "auto_increment" {
					str2 += ",NULL"
				} else {
					str2 += ",\"\""
				}
			}
			i++
		}
		str += ")"
		str2 += ")"
		return str + str2, nil
	}
}

// Tableupdate 更新表数据
func (p *Process) Tableupdate() string {
	sql, err := p.getUpdateStatement(p.formData.name)
	if err != nil {
		return p.createErrorGrid(p.lastSQL, err)
	}
	msg := "<div id='results'>" + sql + "</div>"
	msg += "<script type=\"text/javascript\" language='javascript'> parent.transferQuery(); </script>\n"
	return msg
}

// getUpdateStatement 获取更新语句
func (p *Process) getUpdateStatement(tbl string) (string, error) {
	query := "show full fields from `" + tbl + "`"
	p.lastSQL = query
	if rows, err := p.db.Queryx(query); err != nil {
		return "", err
	} else {
		str := "UPDATE `" + tbl + "` SET "
		str2 := ""
		pKey := ""
		defer rows.Close()
		var i = 0
		for rows.Next() {
			results := make(map[string]interface{})
			rows.MapScan(results)
			if i == 0 {
				str += "`" + string(results["Field"].([]byte)) + "`=\"\""
				if string(results["Key"].([]byte)) != "" {
					str2 += "`" + string(results["Field"].([]byte)) + "`=\"\""
				}
				if string(results["Key"].([]byte)) == "PRI" {
					pKey = string(results["Field"].([]byte))
				}
			} else {
				str += ",`" + string(results["Field"].([]byte)) + "`=\"\""
				if string(results["Key"].([]byte)) == "" {
					if string(results["Key"].([]byte)) == "PRI" {
						pKey = string(results["Field"].([]byte))
					}
					if str2 != "" {
						str2 += " AND "
					}
					str2 += "`" + string(results["Field"].([]byte)) + "`=\"\""
				}
			}
			i++
		}

		if pKey != "" {
			str2 = "`" + pKey + "`=\"\""
		}
		if str2 != "" {
			str2 = " WHERE " + str2
		}

		return str + str2, nil
	}
}

// Showcreate 展示创建语句
func (p *Process) Showcreate() string {
	dels := []string{"result", "pkey", "ukey", "mkey", "unique_table"}
	for _, v := range dels {
		p.Session().Remove("select." + v)
	}

	cmd := p.sanitizeCreateCommand(p.getCreateCommand(p.formData.id, p.formData.name))
	// cmd := p.getCreateCommand(p.formData.id, p.formData.name)

	v := pine.H{
		"TYPE":    p.formData.id,
		"NAME":    p.formData.name,
		"COMMAND": template.HTML(cmd),
		"TIME":    0,
		"SQL":     template.HTML(p.lastSQL),
		"MESSAGE": T("Create command for {{TYPE}} {{NAME}}", p.formData.id, p.formData.name),
	}

	return string(p.Render("showcreate", v))
}

func (p *Process) Describe() string {
	if p.dbname == "" || p.formData.name == "" {
		return string(p.Render("invalid_request", nil))
	}

	p.lastSQL = "DESCRIBE `" + p.formData.name + "`"
	return p.createSimpleGrid(T("Table Description")+": ["+p.formData.name+"]", p.lastSQL)
}

func (p *Process) Copy() string {
	if p.formData.name == "" || p.formData.query == "" {
		return p.createErrorGrid("", errors.New("参数错误或不足"))
	}

	if err := p.copyObject(p.formData.query); err != nil {
		pine.Logger().Warning("copy failed", err)
		return p.createErrorGrid(p.lastSQL, err)
	} else {
		p.Session().Set("db.altered", "true")
		numQueries := 1
		if p.formData.id == "table" {
			numQueries = 2
		}
		return p.createDbInfoGrid(p.lastSQL, numQueries)
	}
}

func (p *Process) copyObject(newName string) error {
	var query string
	var err error
	if p.formData.id == "table" {
		query = "CREATE table `" + newName + "` LIKE `" + p.formData.name + "`"
		p.lastSQL = query
		if _, err = p.db.Exec(query); err != nil {
			return err
		}
		query = "INSERT INTO `" + newName + "` SELECT * FROM `" + p.formData.name + "`"
		p.lastSQL = query
		if _, err = p.db.Exec(query); err != nil {
			return err
		}
	}
	// else {
	//cmd := p.getCreateCommand(p.formData.id, p.formData.name)
	//$command = $this->getCreateCommand($type, $name);
	//$search = '/(create.*'.$type. ' )('.$name.'|\`'.$name.'\`)/i';
	//$replace = '${1} `'.$new_name.'`';
	//$query = preg_replace($search, $replace, $command, 1);
	//$result = $this->query($query);
	// }

	return err

}

// Processes 进程管理器
func (p *Process) Processes() string {
	html := "<link href='/mywebsql/cache?css=theme,default,alerts,results' rel=\"stylesheet\" />\n"
	typo := "message ui-state-highlight"
	msg := T("Select a process and click the button to kill the process")
	if val := p.Input().GetForm().Value; val != nil {
		if prcids := val["prcid[]"]; len(prcids) > 0 {
			killed := []string{}
			missed := []string{}

			for _, v := range prcids {
				if err := p.killProcess(v); err == nil {
					pine.Logger().Warning("kill 进程失败", err)
					killed = append(killed, v)
				} else {
					missed = append(missed, v)
				}
			}
			if len(killed) > 0 {
				msg = T("The process with id [{{PID}}] was killed", strings.Join(killed, ","))
				typo = "message ui-state-default"
			} else {
				msg = T("No such process [id = {{PID}}]", strings.Join(missed, ","))
				typo = "message ui-state-error"
			}
		}
	}

	return html + p.displayProcessList(msg, typo)

}

func (p *Process) killProcess(pid string) error {
	query := "KILL " + pid
	_, err := p.db.Exec(query)
	return err
}

func (p *Process) displayProcessList(msg, typo string) string {
	html := "<input type='hidden' name='q' value='wrkfrm' />"
	html += "<input type='hidden' name='type' value='processes' />"
	html += "<input type='hidden' name='id' value='' />"
	html += "<table border=0 cellspacing=2 cellpadding=2 width='100%'>"
	if len(msg) > 0 {
		html += "<tr><td height=\"25\"><div class=\"" + typo + "\">" + msg + "</div></td></tr>"
	}
	html += "<tr><td colspan=2 valign=top>"

	if rows, err := p.db.Queryx("show full processlist"); err != nil {
		pine.Logger().Warning("获取进程列表失败", err)
		html += T("Failed to get process list")
	} else {
		html += "<table class='results postsort' border=0 cellspacing=1 cellpadding=2 width='100%' id='processes'><tbody>"
		html += "<tr id='fhead'><th></th><th class='th'>" + T("Process ID") + "</th><th class='th'>" + T("Command") + "</th><th class='th'>" + T("Time") +
			"</th><th class='th'>" + T("Info") + "</th></tr>"

		for rows.Next() {
			results := make(map[string]interface{})
			rows.MapScan(results)

			id := string(results["Id"].([]byte))
			command := string(results["Command"].([]byte))
			var timed string
			if results["Time"] != nil {
				timed = string(results["Time"].([]byte))
			}
			var info string
			if results["Info"] != nil {
				info = string(results["Info"].([]byte))
			}

			html += "<tr class='row'><td class=\"tch\"><input type=\"checkbox\" name='prcid[]' value='" + id + "' /></td>" +
				"<td class='tl'>" + id + "</td><td class='tl'>" + command + "</td>" +
				"<td class='tl'>" + timed + "</td><td class='tl'>" + info + "</td></tr>"
		}

		html += "</tbody></table>"

		html += "<tr><td colspan=2 align=right><div id=\"popup_buttons\"><input type='submit' id=\"btn_kill\" name='btn_kill' value='" + T("Kill Process") + "' /></div></td></tr>"

		html += "<script type=\"text/javascript\" language='javascript' src=\"/mywebsql/cache?script=common,jquery,ui,query,sorttable,tables\"></script>\n"

		html += `<script type="text/javascript" language='javascript'>
			window.title = "` + T("Process Manager") + `";
			$('#btn_kill').button().click(function() { document.frmquery.submit(); });
			setupTable('processes', {sortable:'inline', highlight:true, selectable:true});
			</script>`
	}
	return html
}

// Help 帮助页面
func (p *Process) Help() string {
	page, _ := p.Input().GetString("p")
	if page == "" {
		page = "queries"
	}
	pages := map[string]string{
		"queries":  "Executing queries",
		"results":  "Working with results",
		"keyboard": "Keyboard shortcuts",
		"prefs":    "Preferences",
		"misc":     "Miscellaneous",
		"credits":  "Credits",
		"about":    "About",
	}
	if _, ok := pages[page]; !ok {
		page = "queries"
	}

	msg := strReplace([]string{"{{LINK}}"}, []string{""}, T("To see most up-to-date help contents, please visit {{LINK}}"))

	contents := p.Render("help/"+page, nil)
	return string(p.Render("help", pine.H{
		"indexs":          [7]string{"queries", "results", "keyboard", "prefs", "misc", "credits", "about"},
		"pages":           pages,
		"MSG":             msg,
		"PROJECT_SITEURL": template.HTML(PROJECT_SITEURL),
		"page":            page, "contents": template.HTML(contents)}))
}

// Createtbl 创建表
func (p *Process) Createtbl() string {
	action := p.formData.id
	var html string
	if action == "create" || action == "alter" {

	} else {
		engines := Html.ArrayToOptions(p.GetEngines(), "", "Default")
		charsets := Html.ArrayToOptions(p.GetCharsets(), "", "Default")
		collations := Html.ArrayToOptions(p.GetCollations(), "", "Default")
		html = string(p.Render("editable", pine.H{
			"ID":          action,
			"MESSAGE":     "",
			"ROWINFO":     "[]",
			"ALTER_TABLE": "false",
			"TABLE_NAME":  "",
			"ENGINE":      engines,
			"CHARSET":     charsets,
			"COLLATION":   collations,
			"COMMENT":     "",
		}))
	}
	return html
}

// Backup 备份数据
func (p *Process) Backup() string {
	return ""
}

// removeSelectSession 移除选择状态的session数据
func (p *Process) removeSelectSession(sessionKeys []string) {
	for _, v := range sessionKeys {
		p.Session().Remove("select." + v)
	}
}

// Render 渲染模板
func (p *Process) Render(name string, data pine.H) []byte {
	var byts []byte
	var err error
	if byts, err = GetPlush().Exec(name+".php", data); err != nil {
		pine.Logger().Warning("渲染模板"+name+".php失败", err)
	}
	return byts
}

func (p *Process) row2arrMap(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, _ := rows.Columns()
	columnLength := len(columns)
	cache := make([]interface{}, columnLength)
	for index := range cache {
		var a interface{}
		cache[index] = &a
	}
	var list []map[string]interface{}
	for rows.Next() {
		_ = rows.Scan(cache...)
		item := make(map[string]interface{})
		for i, data := range cache {
			item[columns[i]] = *data.(*interface{})
		}
		list = append(list, item)
	}
	err := rows.Close()
	return list, err
}

func (p *Process) Export() string {

	if p.formData.id == "export" {
		orm := p.getXORM()
		pd := p.Input().GetForm().Value
		tables := pd["tables[]"]

		metas, _ := orm.DBMetas()

		schemaTables := []*schemas.Table{}
		for _, v := range metas {
			if exist, _ := common.InArray(v.Name, tables); exist {
				schemaTables = append(schemaTables, v)
			}
		}
		if len(schemaTables) == 0 {
			return "请选择要导出的表"
		}

		p.Response.Header.Set("Content-Disposition", "attachment;filename="+p.dbname+".sql")

		orm.DumpTables(schemaTables, p.Response.BodyWriter())
		return ""
	}

	p.db.Exec("USE " + p.dbname)
	tables := getTables(p.db, p.dbname)
	// views := getViews(p.db, p.dbname)
	// procedures := GetProcedures(p.db, p.dbname)
	// functions := GetFunctions(p.db, p.dbname)
	// triggers := GetTriggers(p.db, p.dbname)
	// events := GetEvents(p.db, p.dbname)
	tableNames := []string{}
	for _, v := range tables {
		tableNames = append(tableNames, v.Name)
	}

	// viewNames := []string{}
	// for _, v := range views {
	// 	viewNames = append(viewNames, v.Name)
	// }

	// procNames := []string{}
	// for _, v := range procedures {
	// 	procNames = append(procNames, v.Name)
	// }

	// funcNames := []string{}
	// for _, v := range functions {
	// 	funcNames = append(funcNames, v.Name)
	// }

	// triggerNames := []string{}
	// for _, v := range triggers {
	// 	triggerNames = append(triggerNames, v.TriggerName)
	// }

	// eventNames := []string{}
	// for _, v := range events {
	// 	eventNames = append(eventNames, v.EventName)
	// }
	return string(p.Render("export", pine.H{
		"list": pine.H{
			"tables": template.HTML(jsonEncode(&tableNames)),
			// "views":      template.HTML(jsonEncode(&viewNames)),
			// "procedures": template.HTML(jsonEncode(&procNames)),
			// "functions":  template.HTML(jsonEncode(&funcNames)),
			// "triggers":   template.HTML(jsonEncode(&triggerNames)),
			// "events":     template.HTML(jsonEncode(&eventNames)),
		},
	}))
}

func (p *Process) Exportres() string {
	return string(p.Render("exportres", nil))
}

func (p *Process) getXORM() *xorm.Engine {
	var err error
	if p.engine == nil {
		p.engine, err = xorm.NewEngine(p.auth.Driver, p.auth.DSN(p.dbname))
		if err != nil {
			panic(err)
		}
	}
	return p.engine
}
