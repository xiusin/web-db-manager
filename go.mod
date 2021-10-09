module github.com/xiusin/web-db-manager

go 1.16

require (
	github.com/allegro/bigcache/v3 v3.0.1
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gobuffalo/helpers v0.6.2 // indirect
	github.com/gobuffalo/plush v3.8.3+incompatible
	github.com/gobuffalo/tags v2.1.7+incompatible // indirect
	github.com/gorilla/securecookie v1.1.1
	github.com/jmoiron/sqlx v1.3.4
	github.com/xiusin/logger v0.0.6-0.20210624030332-1618e61b92ce
	github.com/xiusin/pine v0.0.0-20211009022508-0d0017aeccc4
	xorm.io/xorm v1.2.5
)

replace github.com/xiusin/pine => ../pine
