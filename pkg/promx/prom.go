package promx

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Enable         bool
	App            string
	ListenPort     int
	BasicUsername  string
	BasicPassword  string
	LogApi         map[string]struct{}
	LogMethod      map[string]struct{}
	Buckets        []float64
	Objectives     map[float64]float64
	DefaultCollect bool
}

type PrometheusWrapper struct {
	c                                  Config
	reg                                *prometheus.Registry
	gaugeState                         *prometheus.GaugeVec
	histogramLatency                   *prometheus.HistogramVec
	summaryLatency                     *prometheus.SummaryVec
	counterRequests, counterSendBytes  *prometheus.CounterVec
	counterRcvdBytes, counterException *prometheus.CounterVec
	counterEvent, counterSiteEvent     *prometheus.CounterVec
}

func (p *PrometheusWrapper) init() {
	p.counterRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_requests",
			Help: "number of module requests",
		},
		[]string{"app", "module", "api", "method", "code"},
	)

	p.counterSendBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_send_bytes",
			Help: "number of module send bytes",
		},
		[]string{"app", "module", "api", "method", "code"},
	)

	p.counterRcvdBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_received_bytes",
			Help: "number of module received bytes",
		},
		[]string{"app", "module", "api", "method", "code"},
	)

	p.histogramLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "histogram_latency",
			Help:    "histogram of module latency",
			Buckets: p.c.Buckets,
		},
		[]string{"app", "module", "api", "method"},
	)

	p.summaryLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "summary_latency",
			Help:       "summary of module latency",
			Objectives: p.c.Objectives,
		},
		[]string{"app", "module", "api", "method"},
	)

	p.gaugeState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gauge_state",
			Help: "gauge of app state",
		},
		[]string{"app", "module", "state"},
	)

	p.counterException = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_exception",
			Help: "number of module exception",
		},
		[]string{"app", "module", "exception"},
	)

	p.counterEvent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_event",
			Help: "number of module event",
		},
		[]string{"app", "module", "event"},
	)

	p.counterSiteEvent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_side_event",
			Help: "number of module site event",
		},
		[]string{"app", "module", "event", "site"},
	)

	p.reg.MustRegister(
		p.counterRequests,
		p.counterSendBytes,
		p.counterRcvdBytes,
		p.histogramLatency,
		p.summaryLatency,
		p.gaugeState,
		p.counterException,
		p.counterEvent,
		p.counterSiteEvent,
	)

	if p.c.DefaultCollect {
		p.reg.MustRegister(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
		)
	}
}

// TODO:WAIT TO CHECK
func (p *PrometheusWrapper) run() {
	if p.c.ListenPort == 0 {
		return
	}

	mux := http.NewServeMux()
	mux.Handle("metrics", promhttp.Handler())
	handler := promhttp.InstrumentMetricHandler(p.reg, mux)
	http.HandleFunc("/metrics", func(writer http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		if !ok || !(username == p.c.BasicUsername && password == p.c.BasicPassword) {
			writer.WriteHeader(http.StatusUnauthorized)
			_, _ = writer.Write([]byte("401 Unauthorized"))
			return
		}
		handler.ServeHTTP(writer, req)
	})
	log.Printf("Prometheus listening on %d", p.c.ListenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", p.c.ListenPort), nil))

	// go func() {
	// 	handle := promhttp.HandlerFor(p.reg, promhttp.HandlerOpts{})
	// 	http.Handle("/metrics", promhttp.InstrumentMetricHandler(
	// 		p.reg,
	// 		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	// 			username, password, ok := req.BasicAuth()
	// 			if !ok || !(username == p.c.BasicUserName && password == p.c.BasicPassword) {
	// 				w.WriteHeader(http.StatusUnauthorized)
	// 				_, _ = w.Write([]byte("401 Unauthorized"))
	// 				return
	// 			}
	// 			handle.ServeHTTP(w, req)
	// 		}),
	// 	))
	// 	log.Printf("Prometheus listening on %d", p.c.ListenPort)
	// 	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", p.c.ListenPort), nil))
	// }()
}

func (p *PrometheusWrapper) Log(api, method, code string, sendBytes, rcvdBytes, latency float64) {
	if !p.c.Enable {
		return
	}
	if len(p.c.LogMethod) > 0 {
		if _, ok := p.c.LogMethod[method]; !ok {
			return
		}
	}
	if len(p.c.LogApi) > 0 {
		if _, ok := p.c.LogApi[api]; !ok {
			return
		}
	}
	p.counterRequests.WithLabelValues(p.c.App, "self", api, method, code).Inc()
	if sendBytes > 0 {
		p.counterSendBytes.WithLabelValues(p.c.App, "self", api, method, code).Add(sendBytes)
	}
	if rcvdBytes > 0 {
		p.counterRcvdBytes.WithLabelValues(p.c.App, "self", api, method, code).Add(rcvdBytes)
	}
	if len(p.c.Buckets) > 0 {
		p.histogramLatency.WithLabelValues(p.c.App, "self", api, method).Observe(latency)
	}
	if len(p.c.Objectives) > 0 {
		p.summaryLatency.WithLabelValues(p.c.App, "self", api, method).Observe(latency)
	}
}

func (p *PrometheusWrapper) RequestLog(module, api, method, code string) {
	if !p.c.Enable {
		return
	}
	p.counterRequests.WithLabelValues(p.c.App, module, api, method, code).Inc()
}

func (p *PrometheusWrapper) SendBytesLog(module, api, method, code string, byte float64) {
	if !p.c.Enable {
		return
	}
	p.counterSendBytes.WithLabelValues(p.c.App, module, api, method, code).Add(byte)
}

func (p *PrometheusWrapper) RcvdBytesLog(module, api, method, code string, byte float64) {
	if !p.c.Enable {
		return
	}
	p.counterRcvdBytes.WithLabelValues(module, api, method, code).Add(byte)
}

func (p *PrometheusWrapper) HistogramLatencyLog(module, api, method string, latency float64) {
	if !p.c.Enable {
		return
	}
	p.histogramLatency.WithLabelValues(p.c.App, module, api, method).Observe(latency)
}

func (p *PrometheusWrapper) SummaryLatencyLog(module, api, method string, latency float64) {
	if !p.c.Enable {
		return
	}
	p.summaryLatency.WithLabelValues(module, api, method).Observe(latency)
}

func (p *PrometheusWrapper) ExceptionLog(module string, exception string) {
	if !p.c.Enable {
		return
	}
	p.counterException.WithLabelValues(p.c.App, module, exception).Inc()
}

func (p *PrometheusWrapper) EventLog(module string, event string) {
	if !p.c.Enable {
		return
	}
	p.counterEvent.WithLabelValues(p.c.App, module, event).Inc()
}

func (p *PrometheusWrapper) SiteEventLog(module string, event string, site string) {
	if !p.c.Enable {
		return
	}
	p.counterSiteEvent.WithLabelValues(p.c.App, module, event, site).Inc()
}

func (p *PrometheusWrapper) StateLog(module string, state string, value float64) {
	if !p.c.Enable {
		return
	}
	p.gaugeState.WithLabelValues(p.c.App, module, state).Set(value)
}

func (p *PrometheusWrapper) ResetCounter() {
	if !p.c.Enable {
		return
	}
	p.counterRcvdBytes.Reset()
	p.counterSendBytes.Reset()
	p.counterException.Reset()
	p.counterEvent.Reset()
	p.counterSiteEvent.Reset()
}

func (p *PrometheusWrapper) RegCustomCollector(c prometheus.Collector) {
	p.reg.MustRegister(c)
}

func NewPrometheusWrapper(conf *Config) *PrometheusWrapper {
	if conf.App == "" {
		conf.App = "app"
	}
	if conf.Enable && conf.ListenPort == 0 {
		conf.ListenPort = 9100
	}

	wrapper := &PrometheusWrapper{
		c:   *conf,
		reg: prometheus.NewRegistry(),
	}
	if conf.Enable {
		wrapper.init()
		wrapper.run()
	}
	return wrapper
}
