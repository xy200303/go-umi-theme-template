package controllers

import (
	"backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
)

func (ctl *APIGatewayController) ProxyOpenAIChatCompletions(c *gin.Context) {
	ctl.proxy(c, "openai_api")
}

func (ctl *APIGatewayController) ProxyOpenAIResponses(c *gin.Context) {
	ctl.proxy(c, "openai_response")
}

func (ctl *APIGatewayController) ProxyClaudeMessages(c *gin.Context) {
	ctl.proxy(c, "claude")
}

func (ctl *APIGatewayController) ProxyGemini(c *gin.Context) {
	ctl.proxy(c, "gemini")
}

func (ctl *APIGatewayController) proxy(c *gin.Context, interfaceType string) {
	if err := ctl.apiGatewayService.ForwardRequest(c.Writer, c.Request, interfaceType); err != nil {
		utils.FailError(c, 502, err)
	}
}
