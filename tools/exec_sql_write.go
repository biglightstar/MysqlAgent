package tools

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/openai/openai-go/v3"
)

func NewSqlWrite() *ToolDef {
	return &ToolDef{
		Name: "SqlWrite",
		Description: `执行修改数据库的 SQL 语句(INSERT、UPDATE、DELETE 等),
			例如新增记录、修改现有数据、删除记录,
			当用户想要写入、更新或删除数据库中的数据时使用,
			本工具在执行前会向用户确认是否同意修改;只有用户明确同意后才真正执行,
			调用本工具得到查询结果后,请用带边框的 ASCII 表格形式返回给用户,让每一列对齐列头
			切记生成的 SQL 必须写全库名和表名,例如"SELECT * FROM test.user",不要省略库名`,
		Params: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"sqlsentence": map[string]string{
					"type":        "string",
					"description": "需要写权限的 SQL 语句(INSERT、UPDATE、DELETE 等写操作)",
				},
			},
			"required": []string{"sqlsentence"},
		},
		Fn: func(ctx context.Context, args map[string]any) (string, error) {
			// 取SQL语句
			sqlsentence, ok := args["sqlsentence"].(string)
			if !ok || sqlsentence == "" {
				return "", errors.New("缺少sqlsentence参数")
			}
			// 查询
			res, err := RunWriteQuery(ctx, sqlsentence)
			if err != nil {
				return "", err
			}
			return res, nil
		},
	}
}

func RunWriteQuery(ctx context.Context, sqlsentence string) (string, error) {
	// 判断是否是关于写的SQL语句
	isread := isReadOnly(sqlsentence)
	if isread {
		return "", errors.New("工具 exec_sql_write.go 不能执行读操作,如果你是读操作,请改用exec_sql_write.go工具")
	}
	// 拿到数据库连接句柄
	db, err := sql.Open("mysql", os.Getenv("MYSQL_DSN"))
	if err != nil {
		return "", errors.New("sql.Open()出错:" + err.Error())
	}
	// 验证是否能够连接上MySQL
	err = db.Ping()
	if err != nil {
		return "", errors.New("执行db.Ping()出现错误,连接MySQL失败" + err.Error())
	}
	// 连接成功就开始执行sql语句
	res, err := db.Exec(sqlsentence)
	if err != nil {
		return "", errors.New("执行db.Exec(sqlsentence)出现错误,写操作失败" + err.Error())
	}
	// 执行写操作之后,拿到被修改的行数
	affected, err := res.RowsAffected()
	if err != nil {
		return "", errors.New("执行res.RowsAffected()出错,获取影响行数失败" + err.Error())
	}
	// 定义要返回给大模型的执行结果
	var result strings.Builder
	result.WriteString(fmt.Sprintf("写操作执行成功,已影响%d行\n", affected))
	// 如果是insert插入语句,拿到新加入的数据的id
	insertId, err := res.LastInsertId()
	if err != nil {
		return "", errors.New("执行res.LastInsertId()出错: " + err.Error())
	}
	if insertId != 0 {
		result.WriteString(fmt.Sprintf("增加的新记录的 ID : %d\n", insertId))
	}
	return result.String(), nil
}
