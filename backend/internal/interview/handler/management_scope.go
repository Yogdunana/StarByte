package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func denyManagement(c *gin.Context) {
	response.Error(c, response.NewError(response.CodeForbidden, "无权管理该部门的面试"))
	c.Abort()
}

func managementIdentity(c *gin.Context) bool {
	v := viewer(c)
	if v.ID == uuid.Nil || v.Scope == nil || v.Scope.IsSelf {
		denyManagement(c)
		return false
	}
	return true
}

func (h *InterviewHandler) canManageSession(c *gin.Context, id uuid.UUID) bool {
	if !managementIdentity(c) {
		return false
	}
	sess, err := h.svc.GetSession(c.Request.Context(), viewer(c), id)
	if err != nil {
		response.Error(c, err)
		c.Abort()
		return false
	}
	// Assigned interviewers may read a session; mutating it still requires
	// department management scope, matching meeting attendance vs manage.
	return canManageDepartment(c, sess.DepartmentID)
}

func (h *InterviewHandler) requireManagedSession(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		c.Abort()
		return
	}
	if h.canManageSession(c, id) {
		c.Next()
	}
}

func (h *InterviewHandler) requireManagedInterview(c *gin.Context) {
	if !managementIdentity(c) {
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		c.Abort()
		return
	}
	item, err := h.svc.GetInterview(c.Request.Context(), viewer(c), id)
	if err != nil {
		response.Error(c, err)
		c.Abort()
		return
	}
	if item.Applicant.ID == viewer(c).ID.String() {
		denyManagement(c)
		return
	}
	c.Next()
}

func requireGlobalManagement(c *gin.Context) {
	if !managementIdentity(c) {
		return
	}
	if !dataScope(c).IsEmpty() {
		denyManagement(c)
		return
	}
	c.Next()
}

func canManageDepartment(c *gin.Context, raw string) bool {
	var departmentID *uuid.UUID
	if raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.BadRequest(c, "无效的部门ID")
			return false
		}
		departmentID = &id
	}
	if !viewer(c).CanManageDepartment(departmentID) {
		denyManagement(c)
		return false
	}
	return true
}
