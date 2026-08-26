package handler

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

// parseBindingError 解析参数绑定错误，返回具体的字段错误信息。
// 对于 validator 验证错误，提取字段名和失败的 tag；
// 对于其他类型错误，返回通用的"参数错误"。
func parseBindingError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		var msgs []string
		for _, fe := range ve {
			field := fe.Field()
			tag := fe.Tag()
			msgs = append(msgs, field+": "+tag+" 验证失败")
		}
		return strings.Join(msgs, "; ")
	}
	return "参数错误"
}
