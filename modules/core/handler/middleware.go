package handler

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func ProcessingTimeMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Set up a hook to add the header before the response is written
			c.Response().Before(func() {
				duration := time.Since(start).Microseconds()
				c.Response().Header().Set("X-Processing-Time-Micros", strconv.FormatInt(duration, 10))
			})

			err := next(c)

			return err
		}
	}
}

func PrometheusMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()
			method := c.Request().Method

			InFlightRequests.Inc()
			defer InFlightRequests.Dec()

			start := time.Now()

			err := next(c)

			duration := time.Since(start).Seconds()
			status := strconv.Itoa(c.Response().Status)

			RequestDuration.WithLabelValues(method, path, status).Observe(duration)
			RequestsTotal.WithLabelValues(method, path, status).Inc()

			if c.Response().Status >= 400 {
				ErrorsTotal.WithLabelValues(method, path, status).Inc()
			}

			return err
		}
	}
}
