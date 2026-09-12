package handler

import (
	"fmt"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	userModel "github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Role membership changes follow the same system-role protection as permission assignment.
func roleMembership(db *gorm.DB, cache rbacService.PermissionCacheService, add bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "无效角色ID")
			return
		}
		userID, err := uuid.Parse(c.Param("user_id"))
		if err != nil {
			response.BadRequest(c, "无效用户ID")
			return
		}
		err = db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
			var role model.Role
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&role, "id = ?", roleID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return response.NewError(response.CodeNotFound, "角色不存在")
				}
				return err
			}
			if role.IsSystem {
				return response.NewForbiddenError("系统内置角色由业务流程管理")
			}
			if add && role.Status != 0 {
				return response.NewError(response.CodeBadRequest, "角色已禁用")
			}
			var user userModel.User
			if err := tx.First(&user, "id = ?", userID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return response.NewError(response.CodeNotFound, "用户不存在")
				}
				return err
			}
			if !add {
				return tx.Where("role_id = ? AND user_id = ?", roleID, userID).Delete(&model.UserRole{}).Error
			}
			if user.Status != 0 {
				return response.NewError(response.CodeBadRequest, "用户不可用")
			}
			return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "role_id"}}, DoUpdates: clause.Assignments(map[string]interface{}{"expired_at": nil})}).Create(&model.UserRole{ID: uuid.New(), UserID: userID, RoleID: roleID}).Error
		})
		if err == nil {
			err = cache.InvalidateUserPermissions(c.Request.Context(), userID)
		}
		if err != nil {
			response.Error(c, fmt.Errorf("role membership: %w", err))
			return
		}
		response.OKWithoutData(c)
	}
}
