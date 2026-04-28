package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"hypervisor/internal/errorx"
	"hypervisor/pkg/apires"
	"hypervisor/pkg/logger"

	"github.com/gin-gonic/gin"
)

const (
	AdminSessionCookieName      = "__Host-aurora_admin_session"
	AdminDeviceIDCookieName     = "__Host-aurora_admin_device_id"
	AdminDeviceSecretCookieName = "__Host-aurora_admin_device_secret"

	CtxKeyAdminUserID      = "admin_user_id"
	CtxKeyAdminDisplayName = "admin_display_name"
	CtxKeyAdminCredential  = "admin_credential_id"
	CtxKeyAdminDeviceID    = "admin_device_id"
	CtxKeyAdminSessionID   = "admin_session_id"
)

type AdminSessionAuthInput struct {
	SessionToken string
	DeviceID     string
	DeviceSecret string
	ClientIP     string
	UserAgent    string
}

type AdminSessionContext struct {
	AdminUserID  string
	DisplayName  string
	CredentialID string
	DeviceID     string
	SessionID    string
}

type AdminSessionAuthorizer interface {
	AuthorizeAdminSession(ctx context.Context, input AdminSessionAuthInput) (*AdminSessionContext, error)
}

func AdminSession(authorizer AdminSessionAuthorizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authorizer == nil {
			apires.RespondServiceUnavailable(c, "admin authentication unavailable")
			c.Abort()
			return
		}

		sessionToken, ok := adminCookie(c, AdminSessionCookieName)
		if !ok {
			apires.RespondUnauthorized(c, "unauthorized")
			c.Abort()
			return
		}
		deviceID, ok := adminCookie(c, AdminDeviceIDCookieName)
		if !ok {
			apires.RespondUnauthorized(c, "unauthorized")
			c.Abort()
			return
		}
		deviceSecret, ok := adminCookie(c, AdminDeviceSecretCookieName)
		if !ok {
			apires.RespondUnauthorized(c, "unauthorized")
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		authCtx, err := authorizer.AuthorizeAdminSession(ctx, AdminSessionAuthInput{
			SessionToken: sessionToken,
			DeviceID:     deviceID,
			DeviceSecret: deviceSecret,
			ClientIP:     c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
		})
		if err != nil || authCtx == nil {
			logger.HandlerWarn(c, "hypervisor.admin.session", err, "admin session rejected")
			if errors.Is(err, errorx.ErrUnauthorized) {
				apires.RespondUnauthorized(c, "unauthorized")
				c.Abort()
				return
			}
			apires.RespondServiceUnavailable(c, "admin authentication unavailable")
			c.Abort()
			return
		}

		c.Set(CtxKeyAdminUserID, authCtx.AdminUserID)
		c.Set(CtxKeyAdminDisplayName, authCtx.DisplayName)
		c.Set(CtxKeyAdminCredential, authCtx.CredentialID)
		c.Set(CtxKeyAdminDeviceID, authCtx.DeviceID)
		c.Set(CtxKeyAdminSessionID, authCtx.SessionID)
		if authCtx.AdminUserID != "" {
			c.Set(logger.KeyUserID, authCtx.AdminUserID)
		}
		c.Next()
	}
}

func adminCookie(c *gin.Context, name string) (string, bool) {
	value, err := c.Cookie(name)
	if err != nil || strings.TrimSpace(value) == "" {
		return "", false
	}
	return value, true
}

func GetAdminUserID(c *gin.Context) string {
	v, ok := c.Get(CtxKeyAdminUserID)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func GetAdminDisplayName(c *gin.Context) string {
	v, ok := c.Get(CtxKeyAdminDisplayName)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func ClearAdminAuthCookies(c *gin.Context) {
	for _, name := range []string{AdminSessionCookieName, AdminDeviceIDCookieName, AdminDeviceSecretCookieName} {
		http.SetCookie(c.Writer, &http.Cookie{Name: name, Path: "/", MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode})
	}
}
