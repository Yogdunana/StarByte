package service

import (
	"bytes"
	"fmt"
	"image/jpeg"

	"github.com/disintegration/imaging"
)

const (
	thumbWidth   = 200
	thumbHeight  = 200
	thumbQuality = 80
)

// decodeAndFit 是真正处理不可信输入的部分：解码 + 重采样。
//
// 做成包级变量而不是直接内联，是为了让测试能把 panic 注入进来，
// 从而验证 generateThumbnail 的 recover 真的把它转成了 error ——
// 否则这段兜底没有任何测试覆盖，改坏了也不会有人发现。
var decodeAndFit = func(data []byte) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	thumb := imaging.Fit(img, thumbWidth, thumbHeight, imaging.Lanczos)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: thumbQuality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// generateThumbnail 把上传的图片缩成 JPEG 缩略图。
//
// 为什么这里必须 recover：
//
// 解码走 imaging.Decode → image.Decode（按魔数分派），注册的格式里包含 TIFF，
// 而 disintegration/imaging 的扫描路径在处理构造过的 TIFF 时会 panic
// （CVE-2023-36308，scanner.go 里 Grayscale 的整数索引越界）。
// 这条告警上游**没有补丁** —— v1.6.2 已经是该库的最新版本，
// first_patched_version 为 null —— 所以升级路线走不通，只能在这一侧兜住。
//
// 不兜的后果不只是 500：panic 发生在 Upload() 的第 93 行，
// 也就是**对象已经写进对象存储（88 行）之后、DB 记录创建之前**，
// 一旦 panic 冲出 handler，就会留下一个没人引用的孤儿对象，
// 并且请求以连接重置收场。转成 error 之后 uploadThumbnail 会照常跳过缩略图，
// 上传主流程正常完成 —— 缩略图本来就是可降级的东西。
func generateThumbnail(data []byte) (thumb []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			thumb = nil
			err = fmt.Errorf("缩略图生成中断: %v", r)
		}
	}()
	return decodeAndFit(data)
}
