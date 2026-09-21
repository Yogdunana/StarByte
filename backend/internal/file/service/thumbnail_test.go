package service

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateThumbnail_RecoversPanic 证明 CVE-2023-36308 的兜底是有效的。
//
// 该告警上游无补丁，唯一的缓解手段就是把 imaging 的 panic 转成 error。
// 这里直接注入 panic，断言：① 不崩溃 ② 返回 error 而不是 panic ③ 返回 nil 字节。
func TestGenerateThumbnail_RecoversPanic(t *testing.T) {
	orig := decodeAndFit
	t.Cleanup(func() { decodeAndFit = orig })

	decodeAndFit = func([]byte) ([]byte, error) {
		panic("runtime error: index out of range [3] with length 1")
	}

	got, err := generateThumbnail([]byte("crafted-tiff-bytes"))
	require.Error(t, err, "panic 必须转成 error，绝不能冲出函数")
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "index out of range")
}

// TestGenerateThumbnail_RecoversNonErrorPanic 覆盖 panic 值不是 error 的情况
// （imaging 那类越界 panic 是 runtime.Error，但 recover 拿到的是 interface{}，
// 兜底不能假设它一定实现了 error）。
func TestGenerateThumbnail_RecoversNonErrorPanic(t *testing.T) {
	orig := decodeAndFit
	t.Cleanup(func() { decodeAndFit = orig })

	decodeAndFit = func([]byte) ([]byte, error) {
		panic(42)
	}

	got, err := generateThumbnail([]byte("whatever"))
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "42")
}

// TestGenerateThumbnail_HappyPath 保证加了 recover 之后正常路径没被改坏。
func TestGenerateThumbnail_HappyPath(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			src.Set(x, y, color.RGBA{R: 200, G: 10, B: 30, A: 255})
		}
	}
	var raw bytes.Buffer
	require.NoError(t, gif.Encode(&raw, src, nil))

	got, err := generateThumbnail(raw.Bytes())
	require.NoError(t, err)
	require.NotEmpty(t, got)

	// 输出必须是 JPEG，且已经被 Fit 到 200x200 的框内。
	decoded, format, err := image.Decode(bytes.NewReader(got))
	require.NoError(t, err)
	assert.Equal(t, "jpeg", format)
	b := decoded.Bounds()
	assert.LessOrEqual(t, b.Dx(), thumbWidth)
	assert.LessOrEqual(t, b.Dy(), thumbHeight)
}

// TestGenerateThumbnail_DecodeErrorIsNotSwallowed 保证 recover 不会把
// 正常的解码失败也吞掉 —— 无效输入就该返回 error。
func TestGenerateThumbnail_DecodeErrorIsNotSwallowed(t *testing.T) {
	got, err := generateThumbnail([]byte("not-an-image-at-all"))
	require.Error(t, err)
	assert.Nil(t, got)
	assert.NotContains(t, err.Error(), "缩略图生成中断")
}
