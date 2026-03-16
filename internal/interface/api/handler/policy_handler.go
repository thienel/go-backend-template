package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/thienel/go-backend-template/internal/usecase/service"
	apperror "github.com/thienel/go-backend-template/pkg/error"
	"github.com/thienel/go-backend-template/pkg/response"
)

// PolicyHandler interface
type PolicyHandler interface {
	// Policies
	ListPolicies(c *gin.Context)
	AddPolicy(c *gin.Context)
	RemovePolicy(c *gin.Context)
	GetPoliciesForRole(c *gin.Context)

	// Role hierarchy
	ListRoles(c *gin.Context)
	AddRole(c *gin.Context)
	RemoveRole(c *gin.Context)
}

type policyHandlerImpl struct {
	authzService service.AuthorizationService
}

// NewPolicyHandler creates a new policy handler
func NewPolicyHandler(authzService service.AuthorizationService) PolicyHandler {
	return &policyHandlerImpl{authzService: authzService}
}

func (h *policyHandlerImpl) ListPolicies(c *gin.Context) {
	policies, err := h.authzService.GetAllPolicies()
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrInternalServerError.WithMessage("Không thể lấy danh sách policy").WithError(err))
		return
	}
	response.OK(c, policies, "")
}

func (h *policyHandlerImpl) AddPolicy(c *gin.Context) {
	var req service.PolicyRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteErrorResponse(c, apperror.ErrValidation.WithMessage("Dữ liệu không hợp lệ"))
		return
	}

	if req.Role == "" || req.Path == "" || req.Method == "" {
		response.WriteErrorResponse(c, apperror.ErrValidation.WithMessage("Role, path và method không được để trống"))
		return
	}

	added, err := h.authzService.AddPolicy(req)
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrInternalServerError.WithMessage("Không thể thêm policy").WithError(err))
		return
	}

	if !added {
		response.WriteErrorResponse(c, apperror.ErrConflict.WithMessage("Policy đã tồn tại"))
		return
	}

	response.Created(c, req, "Thêm policy thành công")
}

func (h *policyHandlerImpl) RemovePolicy(c *gin.Context) {
	var req service.PolicyRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteErrorResponse(c, apperror.ErrValidation.WithMessage("Dữ liệu không hợp lệ"))
		return
	}

	removed, err := h.authzService.RemovePolicy(req)
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrInternalServerError.WithMessage("Không thể xóa policy").WithError(err))
		return
	}

	if !removed {
		response.WriteErrorResponse(c, apperror.ErrNotFound.WithMessage("Không tìm thấy policy"))
		return
	}

	response.OK[any](c, nil, "Xóa policy thành công")
}

func (h *policyHandlerImpl) GetPoliciesForRole(c *gin.Context) {
	role := c.Param("role")
	if role == "" {
		response.WriteErrorResponse(c, apperror.ErrValidation.WithMessage("Role không được để trống"))
		return
	}

	policies, err := h.authzService.GetPermissionsForRole(role)
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrInternalServerError.WithMessage("Không thể lấy permissions").WithError(err))
		return
	}

	response.OK(c, policies, "")
}

func (h *policyHandlerImpl) ListRoles(c *gin.Context) {
	roles, err := h.authzService.GetAllRoles()
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrInternalServerError.WithMessage("Không thể lấy danh sách roles").WithError(err))
		return
	}
	response.OK(c, roles, "")
}

func (h *policyHandlerImpl) AddRole(c *gin.Context) {
	var req service.RoleLink
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteErrorResponse(c, apperror.ErrValidation.WithMessage("Dữ liệu không hợp lệ"))
		return
	}

	if req.Child == "" || req.Parent == "" {
		response.WriteErrorResponse(c, apperror.ErrValidation.WithMessage("Child và parent role không được để trống"))
		return
	}

	added, err := h.authzService.AddRoleLink(req.Child, req.Parent)
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrInternalServerError.WithMessage("Không thể thêm role").WithError(err))
		return
	}

	if !added {
		response.WriteErrorResponse(c, apperror.ErrConflict.WithMessage("Role link đã tồn tại"))
		return
	}

	response.Created(c, req, "Thêm role link thành công")
}

func (h *policyHandlerImpl) RemoveRole(c *gin.Context) {
	var req service.RoleLink
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteErrorResponse(c, apperror.ErrValidation.WithMessage("Dữ liệu không hợp lệ"))
		return
	}

	removed, err := h.authzService.RemoveRoleLink(req.Child, req.Parent)
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrInternalServerError.WithMessage("Không thể xóa role").WithError(err))
		return
	}

	if !removed {
		response.WriteErrorResponse(c, apperror.ErrNotFound.WithMessage("Không tìm thấy role link"))
		return
	}

	response.OK[any](c, nil, "Xóa role link thành công")
}
