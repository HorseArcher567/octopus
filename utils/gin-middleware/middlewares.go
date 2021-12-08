package ginmiddlewares

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k8s-practice/octopus/xlog"
)

func GinLogger(logger xlog.Log) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			start   = time.Now()
			path    = c.Request.URL.Path
			query   = c.Request.URL.RawQuery
			body, _ = ioutil.ReadAll(c.Request.Body)
		)

		c.Request.Body.Close()
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		c.Next()

		cost := time.Since(start)

		// at most print 1000 bytes

		// gin access log
		logger.Infof("[path]:{%s}, [status]:{%d}, [method]:{%s}, [query]:{%s}, [body]:{%q}, [ip]:{%s}, [user-agent]:{%s}, [errors]:{%s}, [cost]:{%dms}",
			path, c.Writer.Status(), c.Request.Method, query, body, c.ClientIP(), c.Request.UserAgent(), c.Errors.ByType(gin.ErrorTypePrivate).String(), cost.Milliseconds())
	}
}
