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
	"io"
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
		var request struct {
			DepartmentIDs []uuid.UUID `json:"department_ids"`
		}
		if add && c.Request.Body != nil {
			if err := c.ShouldBindJSON(&request); err != nil && err != io.EOF {
				response.BadRequest(c, "无效部门范围")
				return
			}
		}
		err = db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
			var role model.Role
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&role, "id = ?", roleID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return response.NewError(response.CodeNotFound, "角色不存在")
				}
				return err
			}
			leadership := role.Code == "president" || role.Code == "vice_president" || role.Code == "center_director" || role.Code == "minister"
			if leadership && !c.GetBool("is_super_admin") {
				return response.NewForbiddenError("协会职务须由系统管理员登记任命")
			}
			if role.IsSystem && !leadership {
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
			if !leadership && len(request.DepartmentIDs) > 0 {
				return response.NewError(response.CodeBadRequest, "该角色不支持任职范围")
			}
			if leadership && user.Username == "admin" {
				return response.NewError(response.CodeBadRequest, "技术管理员账号不能担任协会职务")
			}
			if leadership {
				if role.Code != "president" && len(request.DepartmentIDs) == 0 {
					return response.NewError(response.CodeBadRequest, "请选择任职部门或中心")
				}
				if role.Code == "president" && len(request.DepartmentIDs) > 0 {
					return response.NewError(response.CodeBadRequest, "会长职务不绑定部门，兼任部长请单独登记")
				}
				if len(request.DepartmentIDs) > 20 {
					return response.NewError(response.CodeBadRequest, "任职范围过多")
				}
				for _, id := range request.DepartmentIDs {
					var department model.Department
					if err := tx.Where("id=? AND status=0", id).First(&department).Error; err != nil {
						return response.NewError(response.CodeBadRequest, "无效部门范围")
					}
					if role.Code == "minister" && department.ParentID == nil || (role.Code == "center_director" || role.Code == "vice_president") && department.ParentID != nil {
						return response.NewError(response.CodeBadRequest, "部长请选择部门，主任或副会长请选择中心")
					}
				}
				if role.Code == "president" {
					var count int64
					if err := tx.Table("user_roles ur").Joins("JOIN users u ON u.id=ur.user_id").Where("ur.role_id=? AND ur.user_id<>? AND u.status=0 AND u.deleted_at IS NULL AND (ur.expired_at IS NULL OR ur.expired_at>NOW())", roleID, userID).Count(&count).Error; err != nil {
						return err
					}
					if count > 0 {
						return response.NewError(response.CodeConflict, "会长只能有一名，请先完成卸任交接")
					}
				}
			}
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "role_id"}}, DoUpdates: clause.Assignments(map[string]interface{}{"expired_at": nil})}).Create(&model.UserRole{ID: uuid.New(), UserID: userID, RoleID: roleID}).Error; err != nil {
				return err
			}
			var assignment model.UserRole
			if err := tx.Where("user_id=? AND role_id=?", userID, roleID).First(&assignment).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM user_role_departments WHERE user_role_id=?", assignment.ID).Error; err != nil {
				return err
			}
			for _, dept := range request.DepartmentIDs {
				if err := tx.Exec("INSERT INTO user_role_departments(user_role_id,department_id) VALUES (?,?) ON CONFLICT DO NOTHING", assignment.ID, dept).Error; err != nil {
					return err
				}
			}
			return nil

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
