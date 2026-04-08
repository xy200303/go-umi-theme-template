package middleware

import (
	"log"
	"time"

	"backend/internal/models/entities"
	accesssvc "backend/internal/service/access"
	adminsvc "backend/internal/service/admin"
	"github.com/gin-gonic/gin"
)

const (
	contextAuditUserIDKey   = "audit_user_id"
	contextAuditUsernameKey = "audit_username"
)

func SetAuditActor(c *gin.Context, userID uint, username string) {
	c.Set(contextAuditUserIDKey, userID)
	c.Set(contextAuditUsernameKey, username)
}

func getAuditActor(c *gin.Context) (uint, string, bool) {
	userIDValue, hasUserID := c.Get(contextAuditUserIDKey)
	usernameValue, hasUsername := c.Get(contextAuditUsernameKey)
	if !hasUserID && !hasUsername {
		return 0, "", false
	}

	var userID uint
	if value, ok := userIDValue.(uint); ok {
		userID = value
	}
	username, _ := usernameValue.(string)
	return userID, username, true
}

func AuditLogMiddleware(adminService *adminsvc.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		method := c.Request.Method
		requestPath := c.Request.URL.Path
		routePath := c.FullPath()
		if routePath == "" {
			routePath = requestPath
		}

		c.Next()

		userID := uint(0)
		username := "anonymous"
		if claims, ok := GetClaims(c); ok {
			userID = claims.UserID
			username = claims.Username
		} else if auditUserID, auditUsername, ok := getAuditActor(c); ok {
			userID = auditUserID
			if auditUsername != "" {
				username = auditUsername
			}
		}

		item := &entities.AuditLog{
			UserID:      userID,
			Username:    username,
			Method:      method,
			RoutePath:   routePath,
			RequestPath: requestPath,
			StatusCode:  c.Writer.Status(),
			ClientIP:    c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			DurationMS:  time.Since(startedAt).Milliseconds(),
		}

		if template, matched := accesssvc.ResolveOperationTemplate(method, routePath); matched {
			item.OperationID = template.Key
			item.OperationName = template.ActionLabel
			item.MenuKey = template.MenuKey
			item.MenuLabel = template.MenuLabel
		}

		if err := adminService.CreateAuditLog(item); err != nil {
			log.Printf("failed to create audit log: %v", err)
		}
	}
}
