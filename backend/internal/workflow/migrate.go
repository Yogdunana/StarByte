package workflow

import (
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"gorm.io/gorm"
)

// AutoMigrate 自动迁移工作流引擎相关的数据库表。
// 应在服务启动时调用，确保所有工作流模型表已创建。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.FlowDefinition{},
		&model.FlowDefinitionVersion{},
		&model.FlowInstance{},
		&model.FlowTask{},
		&model.FlowHistory{},
		&model.FlowVariable{},
	)
}
