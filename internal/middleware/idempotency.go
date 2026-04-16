package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"net/http"
)

func IdempotencyMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		var cached struct {
			Response   []byte `db:"response"`
			StatusCode int    `db:"status_code"`
		}

		err := db.GetContext(c.Request.Context(), &cached, `SELECT response, status_code FROM idempotency_keys WHERE key = $1`, key)

		if err == nil {
			c.Data(cached.StatusCode, "application/json", cached.Response)
			c.Abort()
			return
		}

		rw := &responseWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = rw
		c.Next()

		db.ExecContext(c.Request.Context(),
			`INSERT INTO idempotency_keys (key, response, status_code)
             VALUES ($1, $2, $3)
             ON CONFLICT (key) DO NOTHING`,
			key, rw.body.Bytes(), rw.status)
	}
}

type responseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
