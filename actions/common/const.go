package common

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/xiusin/web-db-manager/common"
)

const AUTH_TYPE = "LOGIN"

const AUTH_SERVER = "localhost|mysql5"

const AUTH_LOGIN = "test"

const AUTH_PASSWORD = "test"

const ALLOW_CUSTOM_SERVERS = false

const ALLOW_CUSTOM_SERVER_TYPES = "mysql,pgsql"

const MODULE_ACCESS_MODE = "deny"

type Server struct {
	ServerName string
	Host       string
	Driver     string
	User       string
	Password   string
	Port       string
}

func (s *Server) DSN(name string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/"+name+"?charset=utf8&parseTime=true&loc=Local", s.User, s.Password, s.Host, s.Port)
}

const BACKUP_FOLDER = "/backups/"

const BACKUP_FILENAME_FORMAT = "<db>-<date><ext>"

const BACKUP_DATE_FORMAT = "Ymd-His"

const RenderService = "pinecms.mywebsql.plush"

var SERVER_LIST = map[string]Server{
	"Localhost MySQL": {Host: "localhost", Driver: "mysql", Port: "3306"},
	// "Localhost PostgreSQL": {Host: "localhost", Driver: "pgsql"},
	// "SQLite Databases":     {Host: "c:/sqlitedb/", Driver: "sqlite3", User: "root", Password: "sqlite"},
}

var LANGUAGES = map[string]string{
	"zh": "中文(简体)",
	"en": "English",
}

var CODE_EDITORS = map[string]string{
	"simple":      "Plain Text",
	"codemirror":  "CodeMirror",
	"codemirror2": "CodeMirror 2 (Experimental)",
}

var THEMES = map[string]string{
	"default":   "Default",
	"light":     "Light (Gray)",
	"dark":      "Dark",
	"paper":     "Paper",
	"human":     "Humanity (Ubuntu style)",
	"bootstrap": "Bootstrap",
	"chocolate": "Mint-Chocolate",
	"pinky":     "Pinky",
}

const AUTOUPDATE_CHECK = false

var AUTOUPDATE_DAYS = []string{"Mon"}

const TRACE_MESSAGES = false

const TRACE_FILEPATH = ""

const LOG_MESSAGES = false

const MAX_RECORD_TO_DISPLAY = 100

const MAX_TEXT_LENGTH_DISPLAY = 80

const HOTKEYS_ENABLED = true

const DEFAULT_EDITOR = "codemirror"

var DEFAULT_THEME = "bootstrap"

const DEFAULT_LANGUAGE = "zh"

const APP_VERSION = "3.8"

const PROJECT_SITEURL = "http://mywebsql.xiusin.cn/"

const DEVELOPER_EMAIL = "xiusin.chen@gmail.com"

const COOKIE_LIFETIME = 1440

const LIMIT_REGEXP = `(.*)[\s]+(limit[\s]+[\d]+([\s]*(,|offset)[\s]*[\d]+)?)$`
const SORT_REGEXP = "(.*)[\\s]+(ORDER[\\s]+BY[\\s]+([a-zA-z0-9\\._]+|`.*`|\\'.*\\'|\".*\")\\s*(ASC|DESC)?(\\s*\\,\\s*([a-zA-z0-9\\._]+|`.*`|\\'.*\\'|\".*\")\\s*(ASC|DESC)?)*)$"

var DB_LIST = map[string][]string{
	"Test Server": {"test", "wordpress"},
}

var KEY_CODES = map[string][]string{
	"KEYCODE_SETNULL":             {"shift+del", "Shift + Del"},                  // sets value to NULL during edit
	"KEYCODE_QUERY":               {"ctrl+return", "Ctrl + Enter"},               // single query
	"KEYCODE_QUERYALL":            {"ctrl+shift+return", "Ctrl + Shift + Enter"}, // query all
	"KEYCODE_SWITCH_EDITOR1":      {"alt+1", "Alt + 1"},
	"KEYCODE_SWITCH_EDITOR2":      {"alt+2", "Alt + 2"},
	"KEYCODE_SWITCH_EDITOR3":      {"alt+3", "Alt + 3"},
	"KEYCODE_EDITOR_TEXTSIZE_INC": {"ctrl+up", "Ctrl + Up Arrow"},
	"KEYCODE_EDITOR_TEXTSIZE_DEC": {"ctrl+down", "Ctrl + Down Arrow"},
	"KEYCODE_EDITOR_CLEARTEXT":    {"ctrl+shift+del", "Ctrl + Shift + Del"},
}

var DOCUMENT_KEYS = map[string]string{
	"KEYCODE_SETNULL":        "closeEditor(true, null)",
	"KEYCODE_SWITCH_EDITOR1": "switchEditor(0)",
	"KEYCODE_SWITCH_EDITOR2": "switchEditor(1)",
	"KEYCODE_SWITCH_EDITOR3": "switchEditor(2)",
}

var SIMPLE_KEYS = map[string]string{
	"KEYCODE_QUERY":    "queryGo(0)",
	"KEYCODE_QUERYALL": "queryGo(1)",
}

var CODEMIRROR_KEYS = map[string]string{
	"KEYCODE_QUERY":               "queryGo(0)",
	"KEYCODE_QUERYALL":            "queryGo(1)",
	"KEYCODE_SWITCH_EDITOR1":      "switchEditor(0)",
	"KEYCODE_SWITCH_EDITOR2":      "switchEditor(1)",
	"KEYCODE_SWITCH_EDITOR3":      "switchEditor(2)",
	"KEYCODE_EDITOR_TEXTSIZE_INC": "editorTextSize(0.2)",
	"KEYCODE_EDITOR_TEXTSIZE_DEC": "editorTextSize(-0.2)",
	"KEYCODE_EDITOR_CLEARTEXT":    "editorClear()",
}

var CODEMIRROR2_KEYS = map[string]string{
	"KEYCODE_QUERY":    "queryGo(0)",
	"KEYCODE_QUERYALL": "queryGo(1)",
}

type Database struct {
	Database string `db:"Database"`
}

type Table struct {
	Name          string          `json:"name" db:"Name"`
	Engine        *sql.NullString `json:"engine" db:"Engine"`
	Version       *sql.NullInt64  `json:"version" db:"Version"`
	RowFormat     *sql.NullString `json:"row_format" db:"Row_format"`
	Rows          *sql.NullInt64  `json:"rows" db:"Rows"`
	AvgRowLength  *sql.NullInt64  `json:"avg_row_length" db:"Avg_row_length"`
	DataLength    *sql.NullInt64  `json:"data_length" db:"Data_length"`
	MaxDataLength *sql.NullInt64  `json:"max_data_length" db:"Max_data_length"`
	IndexLength   *sql.NullInt64  `json:"index_length" db:"Index_length"`
	DataFree      *sql.NullInt64  `json:"data_free" db:"Data_free"`
	AutoIncrement *sql.NullInt64  `json:"auto_increment" db:"Auto_increment"`
	CreateTime    *sql.NullTime   `json:"create_time" db:"Create_time"`
	UpdateTime    *sql.NullTime   `json:"update_time" db:"Update_time"`
	CheckTime     *sql.NullTime   `json:"check_time" db:"Check_time"`
	Collation     *sql.NullString `json:"collation" db:"Collation"`
	CheckSum      *sql.NullString `json:"check_sum" db:"Checksum"`
	CreateOptions *sql.NullString `json:"create_options" db:"Create_options"`
	Comment       string          `json:"comment" db:"Comment"`
}

func GetTableInfoHeaders() []string {
	var headers []string
	vs := reflect.TypeOf(&Table{}).Elem()

	for i := 0; i < vs.NumField(); i++ {
		headers = append(headers, vs.Field(i).Tag.Get("db"))
	}

	return headers
}

type ProcedureOrFunction struct {
	Db                  string        `json:"db" db:"Db"`
	Name                string        `json:"name" db:"Name"`
	Type                string        `json:"type" db:"Type"`
	Definer             string        `json:"definer" db:"Definer"`
	Modified            *sql.NullTime `json:"modified" db:"Modified"`
	Created             *sql.NullTime `json:"created" db:"Created"`
	SecurityType        string        `json:"security_type" db:"Security_type"`
	Comment             string        `json:"comment" db:"Comment"`
	CharacterSetClient  string        `json:"character_set_client" db:"character_set_client"`
	CollationConnection string        `json:"collation_connection" db:"collation_connection"`
	DatabaseCollation   string        `json:"database_collation" db:"Database Collation"`
}

type TriggerOrEvent struct {
	TriggerName string `json:"trigger_name" db:"TRIGGER_NAME"`
	EventName   string `json:"event_name" db:"EVENT_NAME"`
}

type Column struct {
	TableCataLog string `json:"-" db:"TABLE_CATALOG"`
	TableSchema  string `json:"-" db:"TABLE_SCHEMA"`

	OrdinalPosition        string          `json:"-" db:"ORDINAL_POSITION"`
	ColumnDefault          *sql.NullString `json:"-" db:"COLUMN_DEFAULT"`
	IsNullAble             string          `json:"-" db:"IS_NULLABLE"`
	DataType               string          `json:"-" db:"DATA_TYPE"`
	CharacterMaximumLength *sql.NullInt64  `json:"-" db:"CHARACTER_MAXIMUM_LENGTH"`
	CharacterOctetLength   *sql.NullInt64  `json:"-" db:"CHARACTER_OCTET_LENGTH"`
	NumericPrecision       *sql.NullInt64  `json:"-" db:"NUMERIC_PRECISION"`
	NumericScale           *sql.NullInt64  `json:"-" db:"NUMERIC_SCALE"`
	DatetimePrecision      *sql.NullString `json:"-" db:"DATETIME_PRECISION"`
	CharacterSetName       *sql.NullString `json:"-" db:"CHARACTER_SET_NAME"`
	CollationName          *sql.NullString `json:"-" db:"COLLATION_NAME"`
	ColumnType             string          `json:"-" db:"COLUMN_TYPE"`
	ColumnKey              string          `json:"-" db:"COLUMN_KEY"`
	Extra                  string          `json:"-" db:"EXTRA"`
	Privileges             string          `json:"-" db:"PRIVILEGES"`
	SRS_ID                 *sql.NullString `json:"srs_id" db:"SRS_ID"`
	ColumnComment          string          `json:"-" db:"COLUMN_COMMENT"`
	GenerationExpression   string          `json:"-" db:"GENERATION_EXPRESSION"`

	// 前端使用
	TableName  string   `json:"table" db:"TABLE_NAME"`
	ColumnName string   `json:"name" db:"COLUMN_NAME"`
	NotNull    bool     `json:"not_null"`
	Blob       bool     `json:"blob"`
	PKey       bool     `json:"pkey"`
	UKey       bool     `json:"ukey"`
	MKey       bool     `json:"mkey"`
	ZeroFill   bool     `json:"zerofill"`
	Unsigned   bool     `json:"unsigned"`
	Autoinc    bool     `json:"autoinc"`
	Numeric    bool     `json:"numeric"`
	Type       string   `json:"type"`
	List       []string `json:"list"`
}

func (c *Column) Fill() {
	c.NotNull = c.IsNullAble == "NO"
	c.Numeric, _ = common.InArray(c.DataType, []string{"float", "double", "decimal", "tinyint", "int", "bigint", "mediumint", "numeric"})
	c.Blob, _ = common.InArray(c.DataType, []string{"binary", "blob", "text", "longtext", "mediumtext"})
	c.Autoinc = c.Extra == "auto_increment"
	c.PKey = c.ColumnKey == "PRI"
	c.UKey = c.ColumnKey == "UNI"
	c.ZeroFill = strings.Contains(c.ColumnType, "zerofill")
	c.Unsigned = strings.Contains(c.ColumnType, "unsigned")
	c.Type = c.ColumnType
}

type Variable struct {
	VariableName string `db:"Variable_name" json:"variable_name"`
	Value        string `db:"Value" json:"value"`
}

type CreateCommand struct {
	Table       string `db:"Table"`
	CreateTable string `db:"Create Table"`
}

type CreateViewCommand struct {
	View                string `db:"View"`
	CreateView          string `db:"Create View"`
	CharacterSetClient  string `db:"character_set_client"`
	CollationConnection string `db:"collation_connection"`
}

type CreateTriggerCommand struct {
	Trigger             string        `db:"Trigger"`
	Event               string        `db:"Event"`
	Table               string        `db:"Table"`
	Statement           string        `db:"Statement"`
	Timing              string        `db:"Timing"`
	Created             *sql.NullTime `db:"Created"`
	SqlMode             string        `db:"sql_mode"`
	Definer             string        `db:"Definer"`
	CharacterSetClient  string        `db:"character_set_client"`
	CollationConnection string        `db:"collation_connection"`
	DatabaseCollation   string        `db:"Database Collation"`
}

type CreateFunctionCommand struct {
	Function            string `db:"Function"`
	SqlMode             string `db:"sql_mode"`
	CreateFunction      string `db:"Create Function"`
	CharacterSetClient  string `db:"character_set_client"`
	CollationConnection string `db:"collation_connection"`
	DatabaseCollation   string `db:"Database Collation"`
}

type CreateProcedureCommand struct {
	Procedure           string `db:"Procedure"`
	SqlMode             string `db:"sql_mode"`
	CreateProcedure     string `db:"Create Procedure"`
	CharacterSetClient  string `db:"character_set_client"`
	CollationConnection string `db:"collation_connection"`
	DatabaseCollation   string `db:"Database Collation"`
}

type Engine struct {
	Engine       string          `db:"Engine"`
	Support      string          `db:"Support"`
	Comment      string          `db:"Comment"`
	Transactions *sql.NullString `db:"Transactions"`
	XA           *sql.NullString `db:"XA"`
	Savepoints   *sql.NullString `db:"Savepoints"`
}

type Charset struct {
	Charset          string `db:"Charset"`
	Description      string `db:"Description"`
	DefaultCollation string `db:"Default collation"`
	Maxlen           int    `db:"Maxlen"`
}

type Collation struct {
	Id        int64  `db:"Id"`
	Collation string `db:"Collation"`
	Charset   string `db:"Charset"`
	Default   string `db:"Default"`
	Compiled  string `db:"Compiled"`
	Sortlen   int64  `db:"Sortlen"`
}

type Index struct {
	Table        string      `db:"Table"`
	NonUnique    int         `db:"Non_unique" json:"unique"`
	KeyName      string      `db:"Key_name"`
	SeqInIndex   int         `db:"Seq_in_index" json:"order"`
	ColumnName   string      `db:"Column_name" json:"column"`
	Collation    string      `db:"Collation"`
	Cardinality  int         `db:"Cardinality"`
	SubPart      interface{} `db:"Sub_part"`
	Packed       interface{} `db:"Packed"`
	Null         interface{} `db:"Null"`
	IndexType    string      `db:"Index_type"`
	Comment      string      `db:"Comment"`
	IndexComment string      `db:"Index_comment"`
	Visible      string      `db:"Visible"`
	Expression   interface{} `db:"Expression"`
}

type Field struct {
	Field   string          `db:"Field"`
	Type    string          `db:"Type"`
	Null    string          `db:"Null"`
	Key     string          `db:"Key"`
	Default *sql.NullString `db:"Default"`
	Extra   string          `db:"Extra"`

	Fname string `json:"fname"`
	Ftype string `json:"ftype"`
	Flen  string `json:"flen"`
	Flist string `json:"flist"`
	Fval  string `json:"fval"`
	Fsign bool   `json:"fsign"`
	Fzero bool   `json:"fzero"`
	Fpkey bool   `json:"fpkey"`
	Fauto bool   `json:"fauto"`
	Fnull bool   `json:"fnull"`
}

func (f *Field) fetchFieldInfo() {
	f.Fname = f.Field
	start := strings.Index(f.Type, "(")
	if start == -1 {
		f.Ftype = f.Type
	} else {
		f.Ftype = f.Type[:start]

		if f.Ftype == "enum" || f.Ftype == "set" {
			f.Flen = ""
		} else {
			f.Flen = f.Type[start+1 : strings.Index(f.Type, ")")]
		}
	}
	if f.Default != nil {
		f.Fval = f.Default.String
	}

	f.Fsign = strings.Contains(f.Type, "unsigned")
	f.Fzero = strings.Contains(f.Ftype, "zerofill")
	f.Fpkey = f.Key == "PRI"
	f.Fauto = strings.Contains(f.Extra, "auto_increment")
	f.Fnull = f.Null == "NO"
}
