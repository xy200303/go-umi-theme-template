package controllers

import (
	"net/http"
	"strings"

	"backend/internal/api/middleware"
	"backend/internal/models/dto/requests"
	"backend/internal/pkg/utils"
	fileservice "backend/internal/service/file"

	"github.com/gin-gonic/gin"
)

type FileController struct {
	fileService *fileservice.FileService
}

func NewFileController(fileService *fileservice.FileService) *FileController {
	return &FileController{fileService: fileService}
}

// UploadFile godoc
// @Summary 上传文件
// @Description 统一文件上传入口，返回 file_id 和访问地址供后续业务绑定
// @Tags files
// @ID files.upload
// @Accept multipart/form-data
// @Router /api/v1/user/files/upload [post]
func (ctl *FileController) UploadFile(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, "file is required")
		return
	}

	item, err := ctl.fileService.UploadMultipartFile(file, fileservice.UploadFileOptions{
		UserID:   claims.UserID,
	})
	if err != nil {
		utils.FailError(c, http.StatusBadRequest, err)
		return
	}

	utils.Success(c, ctl.fileService.BuildUploadResp(item))
}

// InitDirectUpload godoc
// @Summary 初始化直传
// @Description 统一文件直传初始化接口，OSS 模式下返回预签名上传信息
// @Tags files
// @ID files.direct.init
// @Router /api/v1/user/files/direct/init [post]
func (ctl *FileController) InitDirectUpload(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req requests.InitDirectUploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "invalid request")
		return
	}

	resp, err := ctl.fileService.InitDirectUpload(req, claims.UserID)
	if err != nil {
		utils.FailError(c, http.StatusBadRequest, err)
		return
	}

	utils.Success(c, resp)
}

// CompleteDirectUpload godoc
// @Summary 完成直传
// @Description 统一文件直传完成接口，OSS 模式下回写文件元数据
// @Tags files
// @ID files.direct.complete
// @Router /api/v1/user/files/direct/complete [post]
func (ctl *FileController) CompleteDirectUpload(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req requests.CompleteDirectUploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "invalid request")
		return
	}

	item, err := ctl.fileService.CompleteDirectUpload(req, claims.UserID)
	if err != nil {
		utils.FailError(c, http.StatusBadRequest, err)
		return
	}

	utils.Success(c, ctl.fileService.BuildUploadResp(item))
}

func (ctl *FileController) DownloadFile(c *gin.Context) {
	if err := ctl.fileService.ValidateDownloadSignature(c.Param("id"), c.Query("expires"), c.Query("sig")); err != nil {
		utils.FailError(c, http.StatusForbidden, err)
		return
	}

	item, err := ctl.fileService.GetByID(c.Param("id"))
	if err != nil {
		utils.FailError(c, http.StatusNotFound, err)
		return
	}

	storageService := ctl.fileService.StorageService()
	if storageService == nil {
		utils.FailWithCode(c, http.StatusInternalServerError, utils.ErrCodeInternal, "file service unavailable")
		return
	}

	switch strings.ToLower(strings.TrimSpace(item.StorageDriver)) {
	case "local":
		fullPath, err := storageService.ResolveLocalFilePath(item.StoragePath)
		if err != nil {
			utils.FailError(c, http.StatusBadRequest, err)
			return
		}
		fallback := utils.FallbackDownloadFilename(item.StoragePath)
		displayName := utils.SanitizeDownloadFilename(item.OriginalName, fallback)
		c.FileAttachment(fullPath, displayName)
	case "cos":
		target := storageService.BuildCOSDownloadRedirectURL(item.StoragePath, item.OriginalName)
		if strings.TrimSpace(target) == "" {
			utils.FailWithCode(c, http.StatusNotFound, utils.ErrCodeNotFound, "file not found")
			return
		}
		c.Redirect(http.StatusFound, target)
	default:
		utils.FailWithCode(c, http.StatusBadRequest, utils.ErrCodeInvalidRequest, "unsupported storage driver")
	}
}
