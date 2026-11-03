package tools

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/openai/openai-go/v3"
)

func valueToString(v any) string {
	if v == nil {
		return "NULL"
	}
	// 最常见：直接把值转回它的字符串形式
	switch val := v.(type) {
	case []byte: // MySQL 很多列（如字符串）驱动返回的是 []byte
		return string(val)
	case string:
		return val
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", val) // 其他类型兜底
	}
}

func NewSqlSelect() *ToolDef {
	return &ToolDef{
		Name: "SqlSelect",
		Description: `执行只读的 SQL 查询(SELECT、SHOW、DESCRIBE 等)，并返回查询结果
			当用户想要查看、检索、列出数据库中的数据或结构时使用,例如:
			"有哪些数据库" "这个表里有什么数据" "看看这张表的结构"。
			收到这类需求时,你把用户的意思转换成合适的 SQL 语句,交给本工具执行
			本工具只能执行读取操作,禁止用于修改或删除数据,
			调用本工具得到查询结果后,请用带边框的 ASCII 表格形式返回给用户,让每一列对齐列头
			切记生成的 SQL 必须写全库名和表名,例如"SELECT * FROM test.user",不要省略库名`,
		Params: openai.FunctionParameters{
			"type": "object",
			// 列出这个工具所需要的所有参数
			// 这里我只需要一个参数: SQL语句
			"properties": map[string]any{
				// 参数名以及对应的 参数类型和描述
				"sqlsentence": map[string]string{
					"type":        "string",
					"description": "只读的 SQL 查询语句(SELECT、SHOW、DESCRIBE 等)",
				},
			},
			"required": []string{"sqlsentence"},
		},
		// 使用MySQL的Go驱动，调用相关函数完成SQL语句的执行
		// args arguements 参数列表
		Fn: func(ctx context.Context, args map[string]any) (string, error) {
			// 从参数列表里面取出SQL语句
			sqlsentence, ok := args["sqlsentence"].(string)
			if !ok || sqlsentence == "" {
				return "", errors.New("缺少sqlsentence参数")
			}
			// 调用查询函数
			res, err := RunSelectQuery(ctx, sqlsentence)
			if err != nil {
				return "", err
			}
			return res, nil
		},
	}
}

func RunSelectQuery(ctx context.Context, sqlsentence string) (string, error) {
	// 判断是否是SQL语句是否是只读数据库
	isread := isReadOnly(sqlsentence)
	// 如果当前SQL是写操作,返回错误
	if !isread {
		return "", errors.New("工具 exec_sql_select.go 不能执行写操作(非 SELECT/SHOW/DESCRIBE/DESC)如果你是写操作,请改用exec_sql_write.go工具")
	}
	// 调用MySQL驱动函数
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
	// 连接成功,开始查询,得到多行数据
	rows, err := db.Query(sqlsentence)
	if err != nil {
		return "", errors.New("执行查询出错:" + err.Error())
	}
	// 关闭rows,防泄露
	defer rows.Close()
	// 取一行一行的数据
	// 取列名,columns的类型: []string
	columns, err := rows.Columns()
	if err != nil {
		return "", errors.New("执行rows.Columns()出现错误,获取列信息失败" + err.Error())
	}
	// 定义一个动态容器
	values := make([]any, len(columns))
	scanArgs := make([]any, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	// 定义一个result建造器,在这里面不断的添加字符串,最后调用string函数取出结果
	var result strings.Builder
	// 第一行,表头
	// 格式: 列名: *,*,*...
	result.WriteString("列名:")
	result.WriteString(strings.Join(columns, ","))
	result.WriteString("\n")
	// 逐行取数据
	rowIndex := 1
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return "", errors.New("rows.Scan(scanArgs...)返回错误,扫描行失败" + err.Error())
		}
		// 成功取出一行(总共有len(values)行)
		line := make([]string, len(values))
		for i, v := range values {
			line[i] = valueToString(v)
		}
		// 将该行数据写到result里面
		// 格式: 行n: 值1,值2...(逗号分隔)
		result.WriteString(fmt.Sprintf("行%d:%s\n", rowIndex, strings.Join(line, ",")))
		rowIndex++
	}
	// 检查整个遍历过程中是否有错
	if err := rows.Err(); err != nil {
		return "", errors.New("遍历结果出错" + err.Error())
	}
	// 调用函数一次性返回结果
	return result.String(), nil

}
