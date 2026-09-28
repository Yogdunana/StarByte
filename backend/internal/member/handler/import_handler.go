package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// PreviewMemberImport 批量录入预检。
// @Summary 批量录入预检
// @Description 逐行校验待录入成员，返回每行会新建/更新/跳过还是出错，不写库
// @Tags 会员
// @Accept json
// @Produce json
// @Param request body dto.MemberImportRequest true "待录入名单"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /member/profiles/import/preview [post]
// @Security BearerAuth
func (h *MemberHandler) PreviewMemberImport(c *gin.Context) {
	h.handleMemberImport(c, false)
}

// ImportMembers 批量录入成员。
// @Summary 批量录入成员
// @Description 一次建好账号、档案与角色；已有的人按学号幂等更新
// @Tags 会员
// @Accept json
// @Produce json
// @Param request body dto.MemberImportRequest true "待录入名单"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /member/profiles/import [post]
// @Security BearerAuth
func (h *MemberHandler) ImportMembers(c *gin.Context) {
	h.handleMemberImport(c, true)
}

func (h *MemberHandler) handleMemberImport(c *gin.Context, commit bool) {
	if h.importer == nil {
		response.Error(c, response.NewError(response.CodeInternalError, "批量录入服务未配置"))
		return
	}
	var req dto.MemberImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	// 协会职务（部长、中心副主任）跟手动任命同一条规则：只认系统管理员。
	canAppoint := c.GetBool("is_super_admin")
	var result *dto.MemberImportResponse
	var err error
	if commit {
		result, err = h.importer.Import(c.Request.Context(), canAppoint, &req)
	} else {
		result, err = h.importer.Preview(c.Request.Context(), canAppoint, &req)
	}
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}
