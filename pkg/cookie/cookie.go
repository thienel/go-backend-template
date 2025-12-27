package cookie

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/thienel/go-backend-template/pkg/config"
)

// SetAuthCookie sets the access token cookie
func SetAuthCookie(c *gin.Context, token string, expiry time.Time, cfg *config.CookieConfig) {
	sameSite := getSameSite(cfg.SameSite)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cfg.Name,
		Value:    token,
		Path:     cfg.Path,
		Domain:   cfg.Domain,
		Expires:  expiry,
		Secure:   cfg.Secure,
		HttpOnly: true,
		SameSite: sameSite,
	})
}

// SetRefreshCookie sets the refresh token cookie
func SetRefreshCookie(c *gin.Context, token string, expiry time.Time, cfg *config.CookieConfig) {
	sameSite := getSameSite(cfg.SameSite)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cfg.RefreshName,
		Value:    token,
		Path:     cfg.Path,
		Domain:   cfg.Domain,
		Expires:  expiry,
		Secure:   cfg.Secure,
		HttpOnly: true,
		SameSite: sameSite,
	})
}

// GetAuthCookie retrieves the access token from cookie
func GetAuthCookie(c *gin.Context, name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetRefreshCookie retrieves the refresh token from cookie
func GetRefreshCookie(c *gin.Context, name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// ClearAuthCookie clears the access token cookie
func ClearAuthCookie(c *gin.Context, cfg *config.CookieConfig) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cfg.Name,
		Value:    "",
		Path:     cfg.Path,
		Domain:   cfg.Domain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	})
}

// ClearRefreshCookie clears the refresh token cookie
func ClearRefreshCookie(c *gin.Context, cfg *config.CookieConfig) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cfg.RefreshName,
		Value:    "",
		Path:     cfg.Path,
		Domain:   cfg.Domain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func getSameSite(value string) http.SameSite {
	switch value {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
