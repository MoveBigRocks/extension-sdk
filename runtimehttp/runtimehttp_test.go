package runtimehttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestForwardedContextMiddlewareSetsCommonKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(ForwardedContextMiddleware())
	engine.GET("/test", func(c *gin.Context) {
		payload := gin.H{
			"extension_id":   ExtensionID(c),
			"extension_slug": ExtensionSlug(c),
			"package_key":    ExtensionPackageKey(c),
			"mode":           stringValue(c, "mode"),
			"user_id":        c.GetString("user_id"),
			"workspace_id":   c.GetString("workspace_id"),
			"name":           c.GetString("name"),
			"email":          c.GetString("email"),
			"user_role":      c.GetString("user_role"),
			"workspace_name": c.GetString("workspace_name"),
			"workspace_slug": c.GetString("workspace_slug"),
			"analytics":      boolValue(c, "admin_feature_analytics"),
			"errorTracking":  boolValue(c, "admin_feature_error_tracking"),
		}
		c.JSON(http.StatusOK, payload)
	})

	workspaceID := "ws_123"
	workspaceName := "Demand Ops"
	workspaceSlug := "demand-ops"
	sessionJSON, err := json.Marshal(SessionContext{
		Type:          "workspace",
		WorkspaceID:   &workspaceID,
		WorkspaceName: &workspaceName,
		WorkspaceSlug: &workspaceSlug,
		Role:          "admin",
	})
	if err != nil {
		t.Fatalf("marshal session context: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(HeaderExtensionID, "ext_123")
	req.Header.Set(HeaderExtensionSlug, "sales-pipeline")
	req.Header.Set(HeaderExtensionPackageKey, "demandops/sales-pipeline")
	req.Header.Set(HeaderExtensionConfigJSON, `{"mode":"agency","showTotals":true}`)
	req.Header.Set(HeaderUserID, "usr_123")
	req.Header.Set(HeaderWorkspaceID, "ws_header")
	req.Header.Set(HeaderUserName, "Ada")
	req.Header.Set(HeaderUserEmail, "ada@example.com")
	req.Header.Set(HeaderSessionContextJSON, string(sessionJSON))
	req.Header.Set(HeaderShowAnalytics, "true")
	req.Header.Set(HeaderShowErrorTracking, "false")

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got := payload["user_id"]; got != "usr_123" {
		t.Fatalf("expected user_id usr_123, got %#v", got)
	}
	if got := payload["extension_id"]; got != "ext_123" {
		t.Fatalf("expected extension_id ext_123, got %#v", got)
	}
	if got := payload["extension_slug"]; got != "sales-pipeline" {
		t.Fatalf("expected extension_slug sales-pipeline, got %#v", got)
	}
	if got := payload["package_key"]; got != "demandops/sales-pipeline" {
		t.Fatalf("expected package_key demandops/sales-pipeline, got %#v", got)
	}
	if got := payload["mode"]; got != "agency" {
		t.Fatalf("expected mode agency, got %#v", got)
	}
	if got := payload["workspace_id"]; got != "ws_123" {
		t.Fatalf("expected workspace_id from session context, got %#v", got)
	}
	if got := payload["name"]; got != "Ada" {
		t.Fatalf("expected name Ada, got %#v", got)
	}
	if got := payload["email"]; got != "ada@example.com" {
		t.Fatalf("expected email ada@example.com, got %#v", got)
	}
	if got := payload["user_role"]; got != "admin" {
		t.Fatalf("expected user_role admin, got %#v", got)
	}
	if got := payload["workspace_name"]; got != "Demand Ops" {
		t.Fatalf("expected workspace_name Demand Ops, got %#v", got)
	}
	if got := payload["workspace_slug"]; got != "demand-ops" {
		t.Fatalf("expected workspace_slug demand-ops, got %#v", got)
	}
	if got := payload["analytics"]; got != true {
		t.Fatalf("expected analytics true, got %#v", got)
	}
	if got := payload["errorTracking"]; got != false {
		t.Fatalf("expected errorTracking false, got %#v", got)
	}
}

func TestBuildBasePageDataDetectsAdminRoles(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set("user_role", "super_admin")
	ctx.Set("workspace_id", "ws_123")
	ctx.Set("workspace_name", "Fleet")

	data := BuildBasePageData(ctx, "fleet", "Fleet", "Overview")
	if data["CanManageUsers"] != true {
		t.Fatalf("expected CanManageUsers true, got %#v", data["CanManageUsers"])
	}
	if data["IsWorkspaceScoped"] != true {
		t.Fatalf("expected IsWorkspaceScoped true, got %#v", data["IsWorkspaceScoped"])
	}
	if data["CurrentWorkspace"] != "Fleet" {
		t.Fatalf("expected CurrentWorkspace Fleet, got %#v", data["CurrentWorkspace"])
	}
}

func boolValue(c *gin.Context, key string) bool {
	value, ok := c.Get(key)
	if !ok {
		return false
	}
	parsed, ok := value.(bool)
	return ok && parsed
}

func stringValue(c *gin.Context, key string) string {
	value, ok := ExtensionConfigString(c, key)
	if !ok {
		return ""
	}
	return value
}
