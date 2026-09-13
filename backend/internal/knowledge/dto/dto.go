package dto

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AttachmentResponse struct {
	ID        string `json:"id"`
	FileID    string `json:"file_id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
}

type CategoryResponse struct {
	ID        string              `json:"id"`
	ParentID  string              `json:"parent_id,omitempty"`
	Name      string              `json:"name"`
	Slug      string              `json:"slug"`
	SortOrder int                 `json:"sort_order"`
	Children  []*CategoryResponse `json:"children,omitempty"`
}

type DocResponse struct {
	AllowedRoles   []string             `json:"allowed_roles"`
	ID             string               `json:"id"`
	Kind           string               `json:"kind"`
	Slug           string               `json:"slug"`
	Path           string               `json:"path"`
	Title          string               `json:"title"`
	Summary        string               `json:"summary"`
	Content        string               `json:"content,omitempty"`
	CategoryID     string               `json:"category_id,omitempty"`
	CategoryName   string               `json:"category_name,omitempty"`
	Visibility     string               `json:"visibility"`
	PermissionCode string               `json:"permission_code,omitempty"`
	Status         int16                `json:"status"`
	Version        int                  `json:"version"`
	Author         Person               `json:"author"`
	PublishedAt    string               `json:"published_at,omitempty"`
	CreatedAt      string               `json:"created_at"`
	UpdatedAt      string               `json:"updated_at"`
	Attachments    []AttachmentResponse `json:"attachments,omitempty"`
}

type TreeNode struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"` // category | doc
	Title    string      `json:"title"`
	Slug     string      `json:"slug,omitempty"`
	Path     string      `json:"path,omitempty"`
	Kind     string      `json:"kind,omitempty"`
	Children []*TreeNode `json:"children,omitempty"`
}

type VersionResponse struct {
	Version   int    `json:"version"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Content   string `json:"content,omitempty"`
	Editor    Person `json:"editor"`
	CreatedAt string `json:"created_at"`
	IsCurrent bool   `json:"is_current"`
}

type CreateCategoryRequest struct {
	ParentID  *string `json:"parent_id"`
	Name      string  `json:"name" binding:"required,max=100"`
	Slug      string  `json:"slug" binding:"required,max=80"`
	SortOrder int     `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	ParentID    *string `json:"parent_id"`
	ClearParent bool    `json:"clear_parent"`
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	SortOrder   *int    `json:"sort_order"`
}

type CreateDocRequest struct {
	AllowedRoles   []string `json:"allowed_roles"`
	Kind           string   `json:"kind" binding:"required"`
	Slug           string   `json:"slug" binding:"required,max=80"`
	Title          string   `json:"title" binding:"required,max=200"`
	Summary        string   `json:"summary" binding:"omitempty,max=500"`
	Content        string   `json:"content"`
	CategoryID     *string  `json:"category_id"`
	Visibility     string   `json:"visibility"`
	PermissionCode string   `json:"permission_code"`
}

type UpdateDocRequest struct {
	AllowedRoles   *[]string `json:"allowed_roles"`
	Slug           *string   `json:"slug"`
	Title          *string   `json:"title"`
	Summary        *string   `json:"summary"`
	Content        *string   `json:"content"`
	CategoryID     *string   `json:"category_id"`
	ClearCategory  bool      `json:"clear_category"`
	Visibility     *string   `json:"visibility"`
	PermissionCode *string   `json:"permission_code"`
	Kind           *string   `json:"kind"`
}

type ListDocRequest struct {
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
	Kind        string `form:"kind"`
	Status      *int16 `form:"status"`
	Visibility  string `form:"visibility"`
	CategoryID  string `form:"category_id"`
	Keyword     string `form:"keyword"`
	IncludeBody bool   `form:"include_body"`
}

type SearchRequest struct {
	Query    string `form:"q" binding:"required"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Kind     string `form:"kind"`
}

type AttachRequest struct {
	FileID string `json:"file_id" binding:"required"`
}

type RollbackRequest struct {
	Version int `json:"version" binding:"required"`
}
