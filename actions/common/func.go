package common

import (
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/xiusin/pine"
	"github.com/xiusin/pine/di"
	"github.com/xiusin/pine/sessions"
	"github.com/xiusin/web-db-manager/actions/render"
)

func GetServerList() {

}

var lang = GetLang()

func GetLang() map[string]string {
	zh := `{"Use Database":"利用数据库","Drop Database":"删除数据库","Empty Database":"空数据库","Drop all tables from this database":"从该数据库删除所有表","Select statement":"Select语句","Insert statement":"Insert语句","Update statement":"Update语句","Describe":"描述","Show create command":"显示创建命令","View data":"查看数据","Alter Table":"修改表","Indexes":"索引","Engine Type":"存储引擎","More operations":"更多的操作","Truncate":"截断","Drop":"删除","Rename":"重命名","Export table data":"导出表数据","Create Table":"创建表","Create View":"创建视图","Create Procedure":"创建过程","Create Function":"创建函数","Create Trigger":"创建触发器","Create Event":"创建事件","Show\/Hide Panel":"显示\/隐藏面板","Show\/Hide Header":"显示\/隐藏标题","Copy all queries to clipboard":"将所有查询到剪贴板","Clear history":"清除历史记录","Copy Column values":"复制列值","Copy to clipboard":"复制到剪贴板","Generate SQL Filter":"生成SQL过滤器","Database Manager":"数据库管理","Manage databases":"管理数据库","Database":"数据库","Refresh":"刷新","Refresh database object list":"刷新数据库对象名单","Create new":"创建数据库","Export":"导出","Export database to external file":"导出数据库，外部文件","Objects":"对象","Create a new table in the database":"在数据库中创建一个新表","Create a new view in the database":"在数据库中创建一个新视图","Create Stored Procedure":"创建存储过程","Create a new stored procedure in the database":"在数据库中创建一个新的存储过程","Create a new user defined function in the database":"创建一个新用户在数据库中定义的函数","Create a new trigger in the database":"在数据库中创建一个新的触发","Create a new event in the database":"在数据库中创建一个新的事件","Data":"数据","Import batch file":"导入批处理文件","Import multiple queries from batch file":"从批处理文件导入多个查询","Export database":"导出数据库","Export database to batch file as sql dump":"导出数据库与SQL转储文件批量","Export current results":"导出结果","Export query results to clipboard or files":"导出查询结果到剪贴板或文件","Tools":"工具","Process Manager":"进程管理","View and manage database processes":"查看和管理数据库进程","Repair Tables":"修复表","Analyze and repair database tables":"分析和修复数据库表","User Manager":"用户管理","Manage database users":"管理数据库用户","Search in Database":"在数据库中搜索","Search for text in the database":"在数据库中搜索文本","Information":"信息","Server\/Connection Details":"服务器\/连接详细信息","View server configuration":"查看服务器配置","View server and connection details":"查看服务器和连接的详细信息","Server Variables":"服务器变量","Database Summary":"数据库详情","View current database summary stats":"查看当前数据库汇总统计","Interface":"界面","Options":"选项","Configure application options":"配置应用程序选项","UI Theme":"界面主题","Database Objects":"数据库对象","Toggle Object Viewer":"切换对象查看器","Help contents":"帮助内容","Learn the basics of using MyWebSQL":"了解使用MyWebSQL的基础知识","QuickStart Tutorials":"快速入门教程","See quick hands-on tutorial of MyWebSQL interface":"见快动手教程界面的MyWebSQL","Online documentation":"在线文档","View online documentation on project website":"查看联机文档在项目网站","Request a Feature":"请求功能","If you would like your most favourite feature to be part of MyWebSQL, please click here to inform about it":"如果你想你最喜欢的功能是MyWebSQL的一部分，请点击这里了解它","Report a Problem":"报告问题","Check for updates":"检查更新","Check for updated versions of the application online":"检查是否有更新版本网上申请","Logout":"注销","Logout from this session":"退出本次会话","Language":"语言","SQL Editor":"SQL编辑器","Toggle SQL Editor":"切换SQL编辑器","Experimental":"实验","Import table data":"导入表数据","Import table data from external file":"从外部文件导入表数据","Batch operations":"批量操作","Perform one or more batch operations on database":"对数据库执行一个或多个批量操作","Create a new schema in the database":"在数据库中创建一个新的模式","Create Schema":"创建模式","Table Description":"表说明","Backup database on the server as SQL dump":"备份数据库服务器上的SQL转储","Backup Database":"备份数据库","Your browser appears to be very old and does not support all features required to run MyWebSQL.":"您的浏览器似乎很旧，不支持运行MyWebSQL所需的所有功能。","Try using a newer version of the browser to run this application.":"尝试使用新版本的浏览器运行此应用程序","Visit Project website":"访问项目网站","Version":"版本","Loading":"载入中","Quick Edit Options":"快速编辑选项","Press {{KEY}} to set NULL":"按{{KEY}}设置为NULL","MySQL Server":"MySQL服务器","Logged in as: {{USER}}":"作为记录：{{USER}}","Results":"结果","Messages":"消息","History":"历史","There are no results to show in this view":"有没有结果表明在此视图","Refresh results":"刷新结果","Please wait":"请稍候","Query All":"查询所有","Query":"查询","Add Record":"添加记录","Delete Record(s)":"删除记录","Update Record(s)":"更新记录","Generate SQL":"生成SQL","Database name":"数据库名称","Select database after creation":"选择数据库创建后","Database connection failed to the server":"数据库连接到服务器失败","Host":"主机","User":"用户","Select a database to begin":"选择一个数据库，开始","Select\/unselect All records":"选择\/取消选择所有记录","Primary key column":"主键列","Unique key column":"唯一键列","Showing records {{START}} - {{END}}":"显示记录{{START}}-{{END}}","Showing first {{MAX}} records only":"第一只记录显示{{MAX}}","1 query successfully executed":"1查询成功执行","{{NUM}} queries successfully executed":"{{NUM}}个成功执行查询","{{NUM}} record(s) were affected":"{{NUM}}个记录（s）分别为受影响","{{NUM}} record(s) updated":"{{NUM}}个记录（s）更新","Error occurred while executing the query":"错误发生在执行查询","{{NUM}} queries failed to execute":"{{NUM}}个查询未能执行","Click to view\/edit column data [{{NUM}} Bytes]":"点击查看\/编辑列的数据[{{NUM}}个字节]","Blob data is not editable":"BLOB数据是不可编辑","Blob data saved":"Blob数据保存","Failed to save blob data":"无法保存BLOB数据","The process with id [{{PID}}] was killed":"ID为[{{PID}}]的过程被打死","No such process [id = {{PID}}]":"没有此进程[编号={{PID}}]","Select a process and click the button to kill the process":"选择一个进程，然后点击按钮来杀死进程","Process ID":"进程ID","Command":"命令","Time":"时间","Info":"信息","Kill Process":"结束进程","Failed to get process list":"无法获取进程列表","The command executed successfully":"命令执行成功","Invalid server configuration":"无效的服务器配置","Invalid Credentials":"无效的凭证","New database successfully created":"成功创建新的数据库","File upload failed. Please try again":"文件上传失败。请再试一次","No queries were executed during import":"没有问题导入过程中被处决","Maximum upload filesize is {{SIZE}}":"最大上传文件大小是{{SIZE}}","Supported filetypes \/ extensions are: ({{LIST}})":"支持的文件类型\/扩展名是：{{LIST}}）","Database summary":"数据库摘要","Any existing object with the same name should be dropped manually before executing the creation command":"任何具有相同名称的现有对象应该被丢弃之前手动执行创建命令","Only create commands are accepted":"只有接受创建命令","User ID":"账户","Password":"密码","Login":"登录","Create new database object":"创建新的数据库对象","Enter command for object creation":"对象创建输入命令","Submit":"提交","Show blob data as: {{TYPE}}":"BLOB数据显示为：{{TYPE}}","Blob data for column {{NAME}}":"Blob数据列{{NAME}}","Create command for {{TYPE}} {{NAME}}":"创建{{TYPE}}的{{NAME}}的命令","Table Engine (type)":"表引擎（型）","Change Table Type":"更改表型","The requested page is not available on the server":"所请求的页面是不是在服务器上可用","Error":"错误","It appears that you attempted to submit an invalid request to the server":"看来，你尝试提交一个无效的请求到服务器","The request has been denied. Reloading the page might solve the problem":"请求已被拒绝。重新载入页面可能解决问题","Access Denied":"访问被拒绝","Help":"帮助","To see most up-to-date help contents, please visit {{LINK}}":"要查看最先进最新的帮助内容，请访问{{LINK}}","It appears that your browser session has expired":"看来，您的浏览器会话已过期","Please refresh the webpage to re-login":"请刷新网页，重新登录","Table information":"表信息","Edit":"编辑","Save":"保存","Either the database is empty, or there was an error retrieving list of database objects":"无论是数据库是空的，或有一个错误检索数据库对象列表","Please try closing and re-opening this dialog again":"请尝试关闭并重新打开此对话框","Structure":"结构","Table Data":"表数据","Structure and Table Data":"结构和表数据","Set Auto increment field values to NULL":"设置自动递增的字段值为NULL","Add DROP command before create statements":"前添加DROP命令CREATE语句","Import":"导入","Export Database":"导出数据库","Export Table":"导出表","Server information":"服务器的信息","Server":"服务器","Version comment":"版本注释","Character sets":"字符集","Server character set":"服务器字符集","Client character set":"客户端字符集","Database character set":"数据库字符集","Results character set":"结果字符集","Export As":"导出为","Insert Statements":"插入语句","Include field names in query":"在查询中包括字段名","XML":"XML","XHTML":"XHTML","Plain Text (One record per line)":"文本（每行一个记录）","Fields separated by:":"分离领域：","Export Results":"导出结果","Select SQL batch file to import":"选择SQL批处理文件导入","Continue processing even if error occurs":"遇到错误继续处理","Basic Information":"基本信息","Table Properties":"表属性","Table Name":"表名","Field Name":"字段名称","Data Type":"数据类型","Length":"长","Default value":"默认值","Unsigned":"非负","Zero Fill":"填充0","Primary Key":"主键","Auto Increment":"自增","Not NULL":"不为NULL","Character Set":"字符集","Collation":"校勘","Comment":"注释","Waiting for table information to be submitted":"等待的表信息提交","Add field":"添加字段","Delete selected field":"删除选定的字段","Clear Table Information":"清除表信息","List of values":"值列表","Query Results":"查询结果","Data for {{TABLE}}":"数据{{TABLE}}","Select tables to be analyzed\/repaired":"选择表进行分析\/修复","Operation to perform":"要执行的操作","Analyze":"分析","Check":"检查","Optimize":"优化","Repair":"修复","Skip Binary logging":"跳过二进制日志","Default":"默认","Quick":"快速","Fast":"快速","Medium":"中","Extended":"扩展","Changed":"改变","Use Frm files (MyISAM tables)":"使用frm文件（MyISAM表）","User Information":"用户信息","Global Privileges":"全局权限","Database Privileges":"数据库权限","Import File":"导入文件","Index Manager":"经理指数","Edit table structure":"修改表结构","Save All Changes":"保存所有更改","Changes are not saved until you press [Save All Changes]":"不保存更改，直到您按下[储存所有修改]","Field Length":"字段长度","Select objects to include in export":"导出选择的对象","Export type":"导出类型","Select All\/None":"选择全部\/无","Add User":"添加用户","Update User":"更新用户","Delete selected User":"删除选定的用户","Remove Password":"删除密码","Select tables to search":"选择要搜索的表","Search Options":"搜索选项","Search in following field types":"在下面的搜索字段类型","Numeric Fields":"数字字段","Character Fields":"字符字段","Text Fields":"文本字段","Date\/Time Fields":"日期\/时间字段","Comparison Type":"比较类型","Text to search":"要搜寻的文字","Search":"搜索","Search Results":"搜索结果","Number of matches":"匹配数量","Search results for {{KEYWORD}} in the database":"搜索结果关键字 {{KEYWORD}} 在数据库中","{{NUM}} match(es)":"{{NUM}}个匹配","Copy query to editor":"复制查询到编辑器","Done":"完成","Create Copy":"复制表","Minimize All":"最小化所有","Copy Record(s)":"复制记录","YAML":"YAML","Comma Separated (CSV for Excel)":"逗号分隔（用于Excel的CSV）","Select data file to import":"选择要导入的数据文件","Select table for data import":"选择数据导入表","First line contains column names":"第一行包含列名","Select objects to operate upon":"选择对象来操作","Operations to perform":"执行的操作","Add prefix string to name":"添加前缀字符串命名","Delete prefix string from name":"从名称中删除前缀字符串","Command text":"命令文本","DROP selected database objects":"删除选定的数据库对象","Batch operation results":"批量操作结果","Operation":"操作","Status":"状态","{{NUM}} queries generated":"{{NUM}}查询生成","Please select one or more operations to perform":"请选择一个或多个操作执行","Quick Search":"快速搜索","{{NAME}} client library is not installed":"{{NAME}}没有安装客户端库","{{NAME}} extension is not installed":"{{NAME}}扩展未安装","SQLite database folder is inaccessible or not writable":"SQLite数据库文件夹是无法访问或无法写入","Maximize\/Restore Results Pane":"最大化\/还原结果窗格","Select databases to operate upon":"选择数据库来操作","DROP selected databases":"DROP选定的数据库","The following operation is irreversible":"下面的操作是不可逆的","Potential data loss might occur":"潜在的数据可能会丢失","Add TRUNCATE command before insert statements":"truncate命令前插入语句","Select objects to include in backup":"选择对象包括在备份中","Backup folder does not exist or is not writable":"备份文件夹不存在或不可写","Backup type":"备份类型","Database backup successfully created":"成功创建数据库备份","Failed to create database backup":"无法创建数据库备份","Generate Bulk insert statements":"生成BULK INSERT语句","Maximum size of SQL statement":"SQL语句的最大大小","Show record count with table names":"显示的记录数与表名","Sort Table listing by":"排序表上市","Name":"名","Last Update Time":"最后更新时间","Maximum records to display in result set":"最大的记录显示在结果集","Show popup dialog for editing large text data":"编辑大量文本数据显示弹出的对话框","Miscellaneous":"杂项","Reset all confirmation dialogs":"重置所有确认对话框","Record Editing":"编辑记录","Exclude Table type":"不包括表型","Exclude Table Character set":"排除表字符集","Backup filename":"备份文件名","Invalid filename format":"无效的文件名格式","Create Database":"创建数据库","Cancel":"取消","Dialog":"弹窗","Enter new name for the database object":"输入数据库对象名称", "For tables with large dataset, it is recommended to modify and save indexes one by one": "对于数据集较大的表，建议逐个修改和保存索引", "Table Indexes":"表索引", "Select an index to view / edit its details":"查看/编辑索引的详细信息", "Waiting for index information to be submitted": "等待提交索引信息"}`

	var lang = map[string]string{}
	json.Unmarshal([]byte(zh), &lang)
	return lang
}

func GetGeneratedJS() string {
	script := "<script language=\"javascript\" type=\"text/javascript\">\nwindow.lang = "

	script += `{"all":"所有","all selected":"所有选定","An error occured while refreshing the object list.":"刷新时发生错误的对象列表中。","Are you sure you want to clear all field information from table?":"您确定要清除所有表的字段信息？","Are you sure you want to delete this user account?":"你确定要删除此用户帐户？","Are you sure you want to DROP all objects from the database {{NAME}}?":"你确定要删除从数据库{{NAME}}所有对象？","Are you sure you want to DROP selected objects?":"你确定要删除选定的对象吗？","Are you sure you want to DROP the database {{NAME}}?":"你确定要删除数据库{{NAME}}？","Are you sure you want to drop this object? {{NAME}}":"你确定要删除此对象吗？{{NAME}}","Are you sure you want to execute {{SELECTED}} queries?":"您确定要执行{{SELECTED}}查询？","Are you sure you want to logout?":"你确定要登出？","Are you sure you want to truncate the table {{NAME}}?":"你确定要截断表E{{NAME}}？","Blob Data [{{SIZE}}]":"Blob数据[{{SIZE}]","Cancel":"取消","Check for Updates":"检查更新","Clear command history?":"清除命令历史记录？","Close":"关闭","Confirm Action":"确认行动","Confirm and do not ask me again about this choice":"确认，不要问我了这样的选择","Copy Object":"复制对象","Create Database":"创建数据库","Create Primary Index on this field":"创建主索引这一领域","Databases":"数据库","Default value [Use quotes to specify string values]":"默认值[使用引号来指定字符串值]","Disallow NULL values in Field":"不允许空值场","Enter new index name":"输入新的索引名称","Enter new name for the database object":"输入数据库对象的新名称","Enter the text to search in database":"输入要搜寻的文字资料库","Error":"错误","Events":"活动","Execute query":"执行查询","Exports results":"出口业绩","Failed to refresh the results.":"刷新失败的结果。","Field Datatype":"现场数据类型","Field Name":"字段名称","Field value is Auto Incremented":"字段的值是自动递增","Functions":"函数","Indexes Updated":"索引更新","Maximum Length of field value":"最大长度值场","Navigation Error. Try reloading the page":"导航误差。尝试重新载入页面","New settings saved and applied":"保存新的设置和应用","New version is available":"新版本可用","Next":"下一页","No":"无","No Table selected":"没有表中选择","Not enough room to show this pane.":"没有足够的空间来显示此窗格。","OK":"确定","One or more field information is incomplete":"一个或多个领域的信息不完整","Open":"打开","Operation failed":"操作失败","Pad field values with leading zeros":"垫带前导零的字段值","Pane":"窗格","Passwords do not match":"密码不匹配","Pin":"针","Please select one or more operations to perform":"请选择一个或多个操作执行","Please type in one or more queries in the SQL editor!":"请在一个或多个查询在SQL编辑器的类型！","Please wait...":"请稍候...","Previous":"上一页","Procedures":"程序","Prompt":"提示","Refresh results":"刷新结果","Refreshing object list":"刷新对象列表","Rename Object":"重命名对象","Resize":"调整","Results page:":"结果页：","Select a database to view privileges for the user":"选择一个数据库，以查看用户的权限","Select an index to view \/ edit its details":"选择的索引查看\/编辑其细节","Select at least one field type for searching":"至少选择一个搜索字段类型","Select objects to operate upon":"选择对象来操作","selector":"选择","Slide Open":"滑开","Table information requires at least one valid field":"表信息至少需要一个有效的字段","Table name is required":"表的名称是必需的","Table successfully created":"成功创建表","Table successfully modified":"表成功修改","Tables":"表","Text Data [{{SIZE}}]":"文本数据[{{SIZE}]","There is no record in the results to export":"有没有出口的结果记录","This attribute is not required for selected field type":"这个属性不需要选定字段类型","Triggers":"触发器","Un-Pin":"联合国针","Unsigned numbered field only":"未签名的编号字段只","Update check failed":"更新检查失败","User information is incomplete or invalid":"用户信息不完整或无效","User Manager":"用户管理","Views":"点击","WARNING":"警告","Yes":"是的","You have the latest version":"你有最新版本", "OK":"确定", "Cancel": "取消", "Select an index to view / edit its details":"查看/编辑索引的详细信息", "Waiting for index information to be submitted": "等待提交索引信息", "Add Index":"新增索引"}`

	script += "\n</script>\n"
	return script
}

func T(s string, replace ...string) string {
	v, ok := lang[s]
	if !ok {
		v = s
	}
	ss := regexp.MustCompile("{{(.+?)}}").FindAllString(v, -1)
	for k, hold := range ss {
		var rep string
		if k < len(replace) {
			rep = replace[k]
		} else {
			rep = hold
		}

		v = strings.Replace(v, hold, rep, 1)
	}
	return v
}

func GetMenuBarHTML(selTheme string) string {
	themeMenu := ""
	if selTheme == "" {
		selTheme = DEFAULT_THEME
	}
	for themeId, theme := range THEMES {
		if selTheme == themeId {
			themeMenu += `<li><a class="check" href="javascript:setPreference('theme', '` + themeId + `')">` + theme + `</a></li>`
		} else {
			themeMenu += `<li><a href="javascript:setPreference('theme', '` + themeId + `')">` + theme + `</a></li>`
		}
	}

	editorMenu := ""
	for editorId, name := range CODE_EDITORS {
		editorMenu += `<li><a href="javascript:setPreference('editor', '` + editorId + `')">` + name + `</a></li>`
	}

	data, _ := di.MustGet(RenderService).(*render.Plush).Exec("menubar.php", pine.H{
		"THEMES_MENU":   template.HTML(themeMenu),
		"LANGUAGE_MENU": "",
		"EDITOR_MENU":   template.HTML(editorMenu),
	})

	return string(data)
}

func GetDbList(db *sqlx.DB) (ret []string, err error) {
	var dbs = []Database{}
	if err = db.Select(&dbs, "SHOW DATABASES"); err != nil {
		pine.Logger().Error(err)
		return nil, err
	}
	for _, s := range dbs {
		ret = append(ret, s.Database)
	}
	return
}

func PrintDbList(db *sqlx.DB, sess sessions.AbstractSession) (string, []string, error) {
	dblist, err := GetDbList(db)
	if err != nil {
		return "", nil, err
	}
	// 不自动选择数据库
	// if sess.Get("db.name") == "" {
	// 	count := 0
	// 	selDb := ""
	// 	stmtDbs := []string{"information_schema", "performance_schema", "mysql", "test"}
	// 	for _, s := range dblist {
	// 		if exist, _ := common.InArray(s, stmtDbs); exist {
	// 			count++
	// 		} else {
	// 			selDb = s
	// 		}
	// 	}
	// 	sess.Set("db.name", selDb)
	// }
	html := ""
	if curdb := sess.Get("db.name"); curdb != "" {
		data, _ := di.MustGet(RenderService).(*render.Plush).Exec("dblist.php", pine.H{
			"dblist": dblist,
			"curdb":  curdb,
		})
		html += string(data)
	} else {
		html += "<span>" + T("Select a database to begin") + "</span>"
	}

	return html, dblist, nil
}

func getTables(db *sqlx.DB, currentDbName string) []Table {
	if currentDbName == "" {
		return nil
	}
	querySql := "show table status from `" + currentDbName + "` where engine is NOT null"
	rets := []Table{}
	if err := db.Select(&rets, querySql); err != nil {
		pine.Logger().Error("获取数据库表异常", err)
	}
	return rets
}

func getViews(db *sqlx.DB, currentDbName string) []Table {
	if currentDbName == "" {
		return nil
	}
	querySql := `show table status where comment='view'`
	rets := []Table{}
	if err := db.Select(&rets, querySql); err != nil {
		pine.Logger().Error("获取视图表异常", err)
	}
	return rets
}

func GetProcedures(db *sqlx.DB, currentDbName string) []ProcedureOrFunction {
	if currentDbName == "" {
		return nil
	}
	querySql := "show procedure status where db = '" + currentDbName + "'"
	rets := []ProcedureOrFunction{}
	if err := db.Select(&rets, querySql); err != nil {
		pine.Logger().Error("获取数据库Procedures异常", err)
	}
	return rets
}

func GetFunctions(db *sqlx.DB, currentDbName string) []ProcedureOrFunction {
	if currentDbName == "" {
		return nil
	}
	querySql := "show function status where db = '" + currentDbName + "'"
	rets := []ProcedureOrFunction{}
	if err := db.Select(&rets, querySql); err != nil {
		pine.Logger().Error("获取数据库函数异常", err)
	}
	return rets
}

func GetTriggers(db *sqlx.DB, currentDbName string) []TriggerOrEvent {
	if currentDbName == "" {
		return nil
	}
	querySql := "select `TRIGGER_NAME` from `INFORMATION_SCHEMA`.`TRIGGERS` where `TRIGGER_SCHEMA` = '" + currentDbName + "'"
	rets := []TriggerOrEvent{}
	if err := db.Select(&rets, querySql); err != nil {
		pine.Logger().Error("获取数据库Trigger异常", err)
	}
	return rets
}

func GetEvents(db *sqlx.DB, currentDbName string) []TriggerOrEvent {
	if currentDbName == "" {
		return nil
	}
	querySql := "select `EVENT_NAME` from `INFORMATION_SCHEMA`.`EVENTS` where `EVENT_SCHEMA` = '" + currentDbName + "'"
	rets := []TriggerOrEvent{}
	if err := db.Select(&rets, querySql); err != nil {
		pine.Logger().Error("获取数据库GetEvents异常", err)
	}
	return rets
}

func GetDatabaseTreeHTML(db *sqlx.DB, dblist []string, currentDbName string) string {
	var html []byte
	if currentDbName != "" {
		db.Exec("USE " + currentDbName)
		tables := getTables(db, currentDbName)
		views := getViews(db, currentDbName)
		procedures := GetProcedures(db, currentDbName)
		functions := GetFunctions(db, currentDbName)
		triggers := GetTriggers(db, currentDbName)
		events := GetEvents(db, currentDbName)
		html, _ = GetPlush().Exec("objtree.php", pine.H{
			"tables":     tables,
			"views":      views,
			"procedures": procedures,
			"functions":  functions,
			"triggers":   triggers,
			"events":     events,
		})
	} else {
		html, _ = GetPlush().Exec("dbtree.php", pine.H{
			"dblist": dblist,
		})
	}
	return string(html)
}

func GetPlush() *render.Plush {
	return di.MustGet(RenderService).(*render.Plush)
}

func GetContextMenusHTML() string {
	data, _ := GetPlush().Exec("menuobjects.php", nil)
	return string(data)
}

func GetHotkeysHTML() string {
	hotkeysHTML := `<script type="text/javascript" language="javascript" src="/mywebsql/cache?script=hotkeys"></script><script type="text/javascript" language="javascript"> $(function() { `

	for name, fun := range DOCUMENT_KEYS {
		code := KEY_CODES[name][0]
		hotkeysHTML += "\n$(document).bind('keydown', '" + code + "', function (evt) { " + fun + "; return false; }); \n"
	}
	var editorKeys = map[string]string{}
	switch strings.ToLower(DEFAULT_EDITOR) {
	case "simple":
		editorKeys = SIMPLE_KEYS
	case "codemirror":
		editorKeys = CODEMIRROR_KEYS
	case "codemirror2":
		editorKeys = CODEMIRROR2_KEYS
	}
	for name, fun := range editorKeys {
		code := KEY_CODES[name][0]
		hotkeysHTML += "\n$(document).bind('keydown', '" + code + "', function (evt) { " + fun + "; return false; }); \n"
	}

	hotkeysHTML += " }); </script>"
	return hotkeysHTML
}

func UpdateSqlEditor() string {
	html := ""
	switch strings.ToLower(DEFAULT_EDITOR) {
	case "simple":
		html = `<script type="text/javascript" language="javascript" src="/mywebsql/cache?script=texteditor"></script><script type="text/javascript" language="javascript">
			function editorHotkey(code, fn) {
				$('#commandEditor').bind('keydown', code, fn);
				$('#commandEditor2').bind('keydown', code, fn);
				$('#commandEditor3').bind('keydown', code, fn);
			}
			$(function() {
				commandEditor = new textEditor("#commandEditor");
				commandEditor2 = new textEditor("#commandEditor2");
				commandEditor3 = new textEditor("#commandEditor3");
				initStart();
			}); </script>`
	case "codemirror":
		html += `<link rel="stylesheet" type="text/css" href="/mywebsql/cache?css=editor" />`
		html += `<script type="text/javascript" language="javascript" src="/mywebsql/cache?script=editor/codemirror"></script><script type="text/javascript" language="javascript">
			function editorHotkey(code, fn) {
				$(document.getElementById('sqlEditFrame').contentWindow.document).bind('keydown', code, fn);
				$(document.getElementById('sqlEditFrame2').contentWindow.document).bind('keydown', code, fn);
				$(document.getElementById('sqlEditFrame3').contentWindow.document).bind('keydown', code, fn);
			}
			$(function() {`
		html += sqlEditorJs("commandEditor", "sqlEditFrame", "initEditor(0);")
		html += sqlEditorJs("commandEditor2", "sqlEditFrame2", "initEditor(1);")
		html += sqlEditorJs("commandEditor3", "sqlEditFrame3", "initEditor(2);")
		html += "}); </script>"

	case "codemirror2":
		html += `<link rel="stylesheet" type="text/css" href="/mywebsql/cache?css=codemirror2" />`
		html += `<script type="text/javascript" language="javascript" src="cache.php?script=codemirror2,mysql"></script>
			<script type="text/javascript" language="javascript">
			function editorHotkey(code, fn) {
				$(document).bind('keydown', code, fn);
				$(document).bind('keydown', code, fn);
				$(document).bind('keydown', code, fn);
			}
			$(function() {
			` + sqlEditor2Js("commandEditor", "initStart();") +
			sqlEditor2Js("commandEditor2", "") +
			sqlEditor2Js("commandEditor3", "") + `
}); </script>`
	}

	return html
}

func sqlEditorJs(id, frameId, init string) string {
	return id + ` = CodeMirror.fromTextArea("` + id + `", { parserfile: "mysql.js", path: "/mywebsql/js/editor/", iframeId: "` + frameId + `", iframeClass: "sqlEditFrame", autoMatchParens: true,
				height: "100%", tabMode : "default", stylesheet: "/mywebsql/cache?css=editor",
				lineNumbers: true, tabFunction : function() { document.getElementById("nav_query").focus(); },
				onLoad : function() { ` + init + ` } });` + "\n"
}

func sqlEditor2Js(id, init string) string {
	return id + ` = CodeMirror.fromTextArea(document.getElementById("` + id + `"), { 
				mode: "text/x-mysql",
				lineNumbers: true, matchBrackets: true, indentUnit: 3,
				height: "100%", tabMode : "default",
				tabFunction : function() { document.getElementById("nav_query").focus(); },
				onLoad : function() { ` + init + ` }
			});`
}

func ExecuteRequest(db *sqlx.DB, ctx *pine.Context, auth *Server) string {
	html := ""
	output, _ := ctx.Input().GetString("id")
	if output != "export" {
		html += startForm(db)
	}
	queryType, _ := ctx.Input().GetString("type")
	pine.Logger().Debug("Process::" + queryType)
	if queryType != "" {
		html += InitProcess(db, ctx, auth).exec(queryType)
	}
	if output != "export" {
		html += "</form></body></html>"
	}
	return html
}

func startForm(db *sqlx.DB) string {
	return "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Transitional//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd\">" +
		"<html xmlns=\"http://www.w3.org/1999/xhtml\" style=\"overflow:hidden;width:100%;height:100%\">\n" +
		"<head><title>Go WebDbManager</title>\n" +
		"</head><body class=\"dialogbody\" style=\"margin:0px;overflow:hidden;width:100%;height:100%\">\n" +
		"<div id=\"popup_overlay\" class=\"ui-widget-overlay\">" +
		"<div><span><img src=\"/mywebsql/themes/default/images/loading.gif\" alt=\"\" /></span></div>" +
		"</div>" +
		"<script language='javascript' type='text/javascript' src='/mywebsql/cache?script=mysql,common'></script>\n" +
		"<!--[if lt IE 8]>" +
		"<script type=\"text/javascript\" language=\"javascript\" src=\"/mywebsql/cache?script=json2\"></script>" +
		"<![endif]-->" +
		`<script language='javascript' type='text/javascript'>
	var EXTERNAL_PATH = '';
	var THEME_PATH = '';
	var DB = db_mysql;
	</script>` +
		"<form name='frmquery' id='frmquery' method='post' action='#' enctype='multipart/form-data' onsubmit='return false'>" +
		"<input type='hidden' name='type' value='query' />" +
		"<input type='hidden' name='id' value='' />" +
		"<input type='hidden' name='name' value='' />" +
		"<input type='hidden' name='query' value='' />"

}

func formatBytes(length int64) string {
	if length < 1024 {
		return fmt.Sprintf("%d B", length)
	}

	if length < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(length)/1024)
	}

	return fmt.Sprintf("%.2f MB", float64(length)/1024/1024)
}

func strReplace(repaces []string, to []string, str string) string {
	for i, repace := range repaces {
		str = strings.ReplaceAll(str, repace, to[i])
	}
	return str
}

func jsonEncode(data interface{}) string {
	byts, _ := json.Marshal(data)
	return string(byts)
}
