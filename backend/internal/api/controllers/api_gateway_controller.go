package controllers

import (
	"net/http"
	"strconv"

	"backend/internal/api/middleware"
	"backend/internal/models/dto/requests"
	"backend/internal/pkg/utils"
	apigatewaysvc "backend/internal/service/api_gateway"

	"github.com/gin-gonic/gin"
)

type APIGatewayController struct {
	apiGatewayService *apigatewaysvc.APIGatewayService
}

func NewAPIGatewayController(apiGatewayService *apigatewaysvc.APIGatewayService) *APIGatewayController {
	return &APIGatewayController{apiGatewayService: apiGatewayService}
}

func (ctl *APIGatewayController) ListInterfaces(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	resp, err := ctl.apiGatewayService.ListInterfaces(claims.UserID)
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.Success(c, resp)
}

func (ctl *APIGatewayController) CreateInterface(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req requests.CreateUserInterfaceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "invalid request")
		return
	}
	resp, err := ctl.apiGatewayService.CreateInterface(claims.UserID, req)
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.Success(c, resp)
}

func (ctl *APIGatewayController) UpdateInterface(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	interfaceID, err := strconv.Atoi(c.Param("id"))
	if err != nil || interfaceID <= 0 {
		utils.Fail(c, http.StatusBadRequest, "invalid interface id")
		return
	}
	var req requests.UpdateUserInterfaceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "invalid request")
		return
	}
	resp, err := ctl.apiGatewayService.UpdateInterface(claims.UserID, uint(interfaceID), req)
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.Success(c, resp)
}

func (ctl *APIGatewayController) DeleteInterface(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	interfaceID, err := strconv.Atoi(c.Param("id"))
	if err != nil || interfaceID <= 0 {
		utils.Fail(c, http.StatusBadRequest, "invalid interface id")
		return
	}
	if err := ctl.apiGatewayService.DeleteInterface(claims.UserID, uint(interfaceID)); err != nil {
		utils.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"ok": true})
}

func (ctl *APIGatewayController) RegenerateInterfaceGatewayKey(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	interfaceID, err := strconv.Atoi(c.Param("id"))
	if err != nil || interfaceID <= 0 {
		utils.Fail(c, http.StatusBadRequest, "invalid interface id")
		return
	}
	resp, err := ctl.apiGatewayService.RegenerateInterfaceGatewayKey(claims.UserID, uint(interfaceID))
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.Success(c, resp)
}

func (ctl *APIGatewayController) TestInterfaceConnection(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	interfaceID, err := strconv.Atoi(c.Param("id"))
	if err != nil || interfaceID <= 0 {
		utils.Fail(c, http.StatusBadRequest, "invalid interface id")
		return
	}
	resp, err := ctl.apiGatewayService.TestInterfaceConnection(claims.UserID, uint(interfaceID))
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.Success(c, resp)
}

func (ctl *APIGatewayController) ListGatewayRequestLogs(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req requests.ListGatewayRequestLogsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "invalid request")
		return
	}
	resp, err := ctl.apiGatewayService.ListGatewayRequestLogs(claims.UserID, req)
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.Success(c, resp)
}

func (ctl *APIGatewayController) ListChatRecords(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req requests.ListChatRecordsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "invalid request")
		return
	}
	resp, err := ctl.apiGatewayService.ListChatRecords(claims.UserID, req)
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.Success(c, resp)
}

func (ctl *APIGatewayController) GetGatewayDisplayConfig(c *gin.Context) {
	utils.Success(c, ctl.apiGatewayService.GetGatewayDisplayConfig())
}
