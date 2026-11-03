package core

import (
	"github.com/openai/openai-go/v3/responses"
)

type Session struct {
	Messages responses.ResponseInputParam
	Len      int
}

func NewSession() *Session {
	return &Session{
		Messages: responses.ResponseInputParam{},
		Len:      0,
	}
}

func (s *Session) GetMessages() responses.ResponseInputParam {
	return s.Messages
}

func (s *Session) GetLen() int {
	return s.Len
}
func (s *Session) AddMessage(item ...responses.ResponseInputItemUnionParam) {
	s.Messages = append(s.Messages, item...)
	s.Len++
}

// NewAgentSession 创建一个带系统提示的会话(系统提示只加一次)
func NewAgentSession() *Session {
	s := NewSession()
	s.AddMessage(
		responses.ResponseInputItemParamOfMessage(
			`你是一个MySQL专家,帮助和辅导人类学习MySQL;
			默认回答简洁干练,只回答和用户的问题相关的问题,并依据工具返回的事实作答;
			当用户想查看、检索数据时,把需求转换成合适的只读 sql,调用 SqlSelect 工具执行;
			当用户想新增、修改、删除数据时,转换成对应的写操作 sql,调用 SqlWrite 工具执行;
			所有 sql 必须写全库名和表名,例如 SELECT * FROM test.user,不要省略库名;
			在执行任何写操作（增删改）之前,必须先向用户说明要执行的内容并征得同意,
			用户明确同意后才执行,用户拒绝则不要执行;
			查询结果的呈现,以工具返回内容为准;
			`,
			responses.EasyInputMessageRoleSystem,
		),
	)
	return s
}
