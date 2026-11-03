package tools

// 取出SQL语句的第一个关键字
import "strings"

// firstKeyword 剥掉前导空白和注释，返回真正的第一个词（小写）
func firstKeyword(sql string) string {
	s := sql

	for {
		// 去两头空白：空格/制表符/换行
		s = strings.TrimSpace(s)

		// 去掉 `-- 行注释`（到行尾）
		if strings.HasPrefix(s, "--") {
			idx := strings.IndexAny(s, "\n") // 找行尾
			if idx < 0 {
				break // 全是注释，没内容了
			}
			s = s[idx:] // 从换行处开始（换行本身留着，下一轮 TrimSpace 会清）
			continue    // 继续循环，再看开头
		}

		// 去掉 `/* 块注释 */`
		if strings.HasPrefix(s, "/*") {
			idx := strings.Index(s, "*/") // 找注释结束
			if idx < 0 {
				break // 注释没闭合，放弃（异常输入）
			}
			s = s[idx+2:] // 从 */ 之后开始
			continue      // 继续循环
		}

		// 前面既没空白也没注释 ,真正到开头了
		break
	}

	// ⑤ 取出第一个词并转小写
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	return strings.ToLower(words[0])
}

// 权限判断
// 判断当前SQL语句是否是只读的
func isReadOnly(sql string) bool {
	// 去SQL第一个关键字
	first := firstKeyword(sql)
	// 判断是否是只读的关键字
	switch first {
	case "select", "show", "describe", "desc":
		return true
	default:
		return false
	}

}
