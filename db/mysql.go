package db

import (
	"database/sql"
	"entry_task/common"
	"fmt"
)

var SqlDB *sql.DB
var StmtLogin, StmtNick, StmtPic, StmtUserinfo *sql.Stmt

func initStmt() error {
	var err error = nil

	StmtLogin, err = SqlDB.Prepare("select password from user where username = ?")
	common.HandleError(err)

	StmtNick, err = SqlDB.Prepare("update user set nickname = ?  where username = ?")
	common.HandleError(err)

	StmtPic, err = SqlDB.Prepare("update user set pictureurl = ?  where username = ?")
	common.HandleError(err)

	StmtUserinfo, err = SqlDB.Prepare("select username, nickname, pictureurl from user where username = ?")
	common.HandleError(err)

	return err
}

func InitMySQL() error {

	var err error
	SqlDB, err = sql.Open("mysql", common.MySQLDsn)
	if err != nil {
		fmt.Println("mysql db open failed")
		return err
	}

	SqlDB.SetMaxOpenConns(100)
	SqlDB.SetMaxIdleConns(100)
	err = initStmt()
	return err
}

