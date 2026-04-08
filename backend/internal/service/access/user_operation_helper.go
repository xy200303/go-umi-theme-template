package accesssvc

import (
	"backend/internal/models/dto/response"
	"backend/internal/models/entities"
	"regexp"
	"sort"
	"strings"

	policytemplate "backend/generate"
)

var auditOperationTemplates = []response.PolicyTemplateResp{
	{
		Key:         "auth.options",
		MenuKey:     "auth",
		MenuLabel:   "认证中心",
		ActionLabel: "查看认证配置",
		Description: "查看当前系统启用的认证与验证码配置。",
		Method:      "GET",
		Path:        "/api/v1/auth/options",
	},
	{
		Key:         "auth.sms.send",
		MenuKey:     "auth",
		MenuLabel:   "认证中心",
		ActionLabel: "发送短信验证码",
		Description: "向指定手机号发送登录、注册等场景的短信验证码。",
		Method:      "POST",
		Path:        "/api/v1/auth/sms/send",
	},
	{
		Key:         "auth.register",
		MenuKey:     "auth",
		MenuLabel:   "认证中心",
		ActionLabel: "用户注册",
		Description: "创建新用户账号并完成首次登录。",
		Method:      "POST",
		Path:        "/api/v1/auth/register",
	},
	{
		Key:         "auth.login.password",
		MenuKey:     "auth",
		MenuLabel:   "认证中心",
		ActionLabel: "密码登录",
		Description: "使用账号和密码登录系统。",
		Method:      "POST",
		Path:        "/api/v1/auth/login/password",
	},
	{
		Key:         "auth.login.sms",
		MenuKey:     "auth",
		MenuLabel:   "认证中心",
		ActionLabel: "短信验证码登录",
		Description: "使用手机号和短信验证码登录系统。",
		Method:      "POST",
		Path:        "/api/v1/auth/login/sms",
	},
	{
		Key:         "auth.refresh",
		MenuKey:     "auth",
		MenuLabel:   "认证中心",
		ActionLabel: "刷新令牌",
		Description: "使用刷新令牌换取新的访问令牌。",
		Method:      "POST",
		Path:        "/api/v1/auth/refresh",
	},
	{
		Key:         "auth.logout",
		MenuKey:     "auth",
		MenuLabel:   "认证中心",
		ActionLabel: "退出登录",
		Description: "注销当前会话并使刷新令牌失效。",
		Method:      "POST",
		Path:        "/api/v1/auth/logout",
	},
}

func BuildUserOperationIDs(roles []entities.Role, casbin *CasbinService) []string {
	templates := policytemplate.List()
	if len(templates) == 0 {
		return nil
	}

	for _, role := range roles {
		if role.Name == "admin" {
			return collectAllOperationIDs(templates)
		}
	}

	granted := make(map[string]struct{})
	for _, role := range roles {
		for _, policy := range casbin.GetRolePolicies(role.Name) {
			for _, template := range templates {
				if operationMatchesPolicy(template.Path, template.Method, policy.Path, policy.Method) {
					granted[template.Key] = struct{}{}
				}
			}
		}
	}

	operationIDs := make([]string, 0, len(granted))
	for operationID := range granted {
		operationIDs = append(operationIDs, operationID)
	}
	sort.Strings(operationIDs)
	return operationIDs
}

func ResolveOperationTemplate(method string, routePath string) (response.PolicyTemplateResp, bool) {
	for _, template := range auditOperationTemplates {
		if strings.EqualFold(template.Method, method) && template.Path == routePath {
			return template, true
		}
	}

	for _, template := range policytemplate.List() {
		if strings.EqualFold(template.Method, method) && template.Path == routePath {
			return template, true
		}
	}
	return response.PolicyTemplateResp{}, false
}

func collectAllOperationIDs(templates []response.PolicyTemplateResp) []string {
	operationIDs := make([]string, 0, len(templates))
	for _, template := range templates {
		operationIDs = append(operationIDs, template.Key)
	}
	sort.Strings(operationIDs)
	return operationIDs
}

func operationMatchesPolicy(operationPath string, operationMethod string, policyPath string, policyMethod string) bool {
	return matchPolicyPattern(policyMethod, operationMethod) && matchPolicyPattern(policyPath, operationPath)
}

func matchPolicyPattern(pattern string, value string) bool {
	if pattern == "*" {
		return true
	}

	regexPattern := "^" + strings.ReplaceAll(strings.ReplaceAll(pattern, ".", "\\."), "*", ".*") + "$"
	matched, err := regexpMatchString(regexPattern, value)
	if err != nil {
		return pattern == value
	}
	return matched
}

func regexpMatchString(pattern string, value string) (bool, error) {
	return regexp.MatchString(pattern, value)
}
