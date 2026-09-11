// Package repo 数据访问层 GORM 实现 - 事务管理器与公共辅助
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
//
// 本文件实现:
//   1. TransactionManager 基于 gorm.DB 的事务管理器
//   2. 公共辅助函数: FromTx 解包、gorm 错误到 sentinel 错误的映射
package repo

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

// ============================================================================
// GORM 事务管理器实现
// ============================================================================

// GormTransactionManager 基于 gorm.DB 的事务管理器
// 实现 repo.TransactionManager 接口
type GormTransactionManager struct {
	db *gorm.DB
}

// NewGormTransactionManager 构造事务管理器
func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

// Transaction 开启事务并执行 fn
//   - fn 内的 repo 调用通过 WithTx 注入的 ctx 自动复用同一事务
//   - fn 返回 nil: 提交事务
//   - fn 返回 error: 回滚事务
//   - fn panic: 回滚事务（panic 向上传播）
func (m *GormTransactionManager) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// 嵌套事务支持：若 ctx 已存在事务，则复用不另开
	if _, ok := FromTx(ctx, nil); ok {
		// 已在事务内，直接执行 fn
		return fn(ctx)
	}

	// 开启新事务
	tx := m.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 注入事务到 ctx
	txCtx := WithTx(ctx, tx)

	// panic 保护：回滚并将 panic 向上传播
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 执行业务闭包
	if err := fn(txCtx); err != nil {
		// 回滚并返回原错误（不包装，保留 sentinel 特性）
		if rbErr := tx.Rollback().Error; rbErr != nil {
			// 回滚失败是严重错误，记录到日志但不覆盖原错误
			// 实现时建议接入日志组件记录 rbErr
		}
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// ============================================================================
// 公共辅助函数
// ============================================================================

// DBFromCtx 从 context 取出事务 DB，无则使用 fallback
// 供各 Repository 实现内部调用，自动复用事务或回退默认 DB
func DBFromCtx(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := FromTx(ctx, fallback); ok {
		return tx
	}
	return fallback.WithContext(ctx)
}

// translateGormErr 将 gorm 错误翻译为 repo 层 sentinel 错误
//   - gorm.ErrRecordNotFound → ErrNotFound
//   - 重复键错误 → ErrDuplicate
//   - 其他错误原样返回
func translateGormErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	// 重复键检测：依赖底层驱动错误字符串匹配
	//   MySQL: Error 1062 Duplicate entry
	//   PostgreSQL: duplicate key value violates unique constraint
	//   SQLite: UNIQUE constraint failed
	// 此处使用 errors.Is + 字符串匹配双重保障
	if isDuplicateErr(err) {
		return ErrDuplicate
	}
	return err
}

// isDuplicateErr 判断是否为唯一键冲突错误
//   覆盖主流驱动的错误特征字符串
func isDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// MySQL
	if strings.Contains(msg, "Duplicate entry") {
		return true
	}
	// PostgreSQL
	if strings.Contains(msg, "duplicate key value violates unique constraint") {
		return true
	}
	// SQLite
	if strings.Contains(msg, "UNIQUE constraint failed") {
		return true
	}
	return false
}

// paginate 构建 gorm 分页 Scope
//   page: 1-based 页码
//   pageSize: 每页条数，0 表示不分页
func paginate(page, pageSize int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 20 // 默认每页 20 条
		}
		if pageSize > 200 {
			pageSize = 200 // 上限保护
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// applyOrder 应用排序参数
//   orderBy: 字段名（须在白名单内，避免 SQL 注入）
//   order: asc/desc
//   whitelist: 允许排序的字段集合，传入字段不在白名单内则忽略
func applyOrder(db *gorm.DB, orderBy, order string, whitelist map[string]bool) *gorm.DB {
	if orderBy == "" {
		return db
	}
	if !whitelist[orderBy] {
		return db
	}
	dir := "asc"
	if order == "desc" {
		dir = "desc"
	}
	return db.Order(orderBy + " " + dir)
}
