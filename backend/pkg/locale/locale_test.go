package locale

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLanguageNegotiationAndMessages(t *testing.T) {
	require.Equal(t, "ru-RU", FromHeader("en;q=0.3,ru-RU;q=0.9"))
	require.Equal(t, "zh-CN", FromHeader("unsupported"))
	require.Equal(t, "en-US", FromContext(WithContext(context.Background(), "en-GB")))
	require.Equal(t, "Заявка не найдена", Message("ru-RU", "申请不存在"))
	require.Equal(t, "Workflow node not found: abc", Message("en-US", "流程节点不存在: abc"))
	require.Equal(t, "Field age has an invalid pattern regular expression", Message("en-US", "字段 age 的 pattern 不是合法正则"))
	require.Equal(t, "email: неверный формат", Message("ru-RU", "email 格式不正确"))
	require.Equal(t, "张三的自定义内容", Text("ru-RU", "张三的自定义内容"))
}

func TestCatalogParityAndFormatTokens(t *testing.T) {
	require.Len(t, catalogs["ru-RU"], len(catalogs["en-US"]))
	for source, english := range catalogs["en-US"] {
		russian, ok := catalogs["ru-RU"][source]
		require.True(t, ok, source)
		require.NotEmpty(t, english, source)
		require.NotEmpty(t, russian, source)
		require.Equal(t, formatToken.FindAllString(source, -1), formatToken.FindAllString(english, -1), source)
		require.Equal(t, formatToken.FindAllString(source, -1), formatToken.FindAllString(russian, -1), source)
	}
}
