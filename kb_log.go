package main

// Based on github.com/stephenmuss/ginerus but adds more fields.

import (
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func KBLog(logger *logrus.Logger, timeFormat string, utc bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		c.Next()

		end := time.Now()
		task, _ := c.Get("task")
		data, _ := c.Get("data")
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}

		entry := logger.WithFields(logrus.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"name":       task.(Task).Name,
			"procs":      strings.Join(task.(Task).Procs, ","),
			"data":       data,
			"latency":    latency,
			"user-agent": c.Request.UserAgent(),
			"at":         start.Format(timeFormat),
		})

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			entry.Error(c.Errors.String())
		} else {
			entry.Info()
		}
	}
}
