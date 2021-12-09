package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	log "github.com/sirupsen/logrus"

	"strconv"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewWebserver(defaultLogger *log.Logger, tokenManager *TokenManager) *fiber.App {
	app := fiber.New()
	metrics := NewMetricService("fiber", "http", "/metrics")
	metrics.Register(app)
	app.Use(logger.New(logger.Config{
		Format: "${status} @ ${latency} - ${ip} > ${method} ${path}\n",
		Output: defaultLogger.Writer(),
	}))

	app.Use(tokenManager.GetMiddleWare())

	return app
}

type MetricService struct {
	Namespace   string
	Subsystem   string
	MetricPath  string
	reqCount    *prometheus.CounterVec
	reqDuration *prometheus.HistogramVec
}

func (m *MetricService) PrometheusHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == m.MetricPath {
			c.Next()
			return nil
		}

		start := time.Now()

		c.Next()

		r := c.Route()

		statusCode := strconv.Itoa(c.Response().StatusCode())
		elapsed := float64(time.Since(start)) / float64(time.Second)

		m.reqCount.WithLabelValues(statusCode, c.Method(), r.Path).Inc()
		m.reqDuration.WithLabelValues(c.Method(), r.Path).Observe(elapsed)
		return nil
	}
}

func (m *MetricService) Register(app *fiber.App) {
	m.registerDefaultMetrics()
	m.SetupPath(app)
	app.Use(m.PrometheusHandler())
}

func (m *MetricService) SetupPath(app *fiber.App) {
	app.Get(m.MetricPath, adaptor.HTTPHandler(promhttp.Handler()))
}

func (m *MetricService) registerDefaultMetrics() {
	m.reqCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "requests_total",
			Namespace: m.Namespace,
			Subsystem: m.Subsystem,
			Help:      "Number of HTTP requests",
		},
		[]string{"status_code", "method", "path"},
	)

	m.reqDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "request_duration_seconds",
		Namespace: m.Namespace,
		Subsystem: m.Subsystem,
		Help:      "Duration of HTTP requests",
	}, []string{"method", "handler"})
}

func NewMetricService(namespace string, subsystem string, metricPath string) *MetricService {
	return &MetricService{
		Namespace:  namespace,
		Subsystem:  subsystem,
		MetricPath: metricPath,
	}
}
