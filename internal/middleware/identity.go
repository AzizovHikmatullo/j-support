package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/AzizovHikmatullo/j-support/internal/channel"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var ErrUnauthorized = errors.New("unauthorized")

func ChannelIdentityMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, err := resolveIdentity(c, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("identity", identity)
		c.Next()
	}
}

func GetIdentity(c *gin.Context) (channel.Identity, bool) {
	val, exists := c.Get("identity")
	if !exists {
		return channel.Identity{}, false
	}
	identity, ok := val.(channel.Identity)
	return identity, ok
}

func resolveIdentity(c *gin.Context, jwtSecret string) (channel.Identity, error) {
	if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return resolveFromJWT(c, strings.TrimPrefix(auth, "Bearer "), jwtSecret)
	}

	if token := c.GetHeader("X-Session-Token"); token != "" {
		return resolveFromSessionToken(c, token)
	}

	if tgID := c.GetHeader("X-Telegram-ID"); tgID != "" {
		return resolveFromTelegramID(c, tgID)
	}

	return channel.Identity{}, ErrUnauthorized
}

func resolveFromJWT(c *gin.Context, tokenStr, secret string) (channel.Identity, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		},
	)
	if err != nil || !token.Valid {
		return channel.Identity{}, ErrUnauthorized
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return channel.Identity{}, ErrUnauthorized
	}

	c.Set("userID", claims.UserID)
	c.Set("role", claims.Role)
	return channel.Identity{
		ChannelType: channel.ChannelApp,
		ID:          strconv.Itoa(claims.UserID),
		Role:        claims.Role,
	}, nil
}

func resolveFromTelegramID(c *gin.Context, tgID string) (channel.Identity, error) {
	c.Set("role", "user")

	return channel.Identity{
		ChannelType: channel.ChannelTelegram,
		ID:          tgID,
		Role:        "user",
	}, nil
}

func resolveFromSessionToken(c *gin.Context, token string) (channel.Identity, error) {
	c.Set("role", "user")

	return channel.Identity{
		ChannelType: channel.ChannelWeb,
		ID:          token,
		Role:        "user",
	}, nil
}
