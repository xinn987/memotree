// Package buckets 提供本地运维脚本使用的 bucket 初始化编排。
package buckets

import (
	"context"
	"strings"
)

// Ensurer 是对象存储 bucket 幂等创建能力的最小接口。
type Ensurer interface {
	EnsureBucket(ctx context.Context, bucket string) error
}

// EnsureAll 按配置确保所有 bucket 存在；空值和重复值会被忽略。
func EnsureAll(ctx context.Context, ensurer Ensurer, names []string) error {
	seen := map[string]bool{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		if err := ensurer.EnsureBucket(ctx, name); err != nil {
			return err
		}
	}
	return nil
}
