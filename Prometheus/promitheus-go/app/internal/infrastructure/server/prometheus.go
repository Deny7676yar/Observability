package server

import (
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	Namespace   = "metricsexample"
	LabelMethod = "method"
	LabelStatus = "status"
)

type App struct {
	latencyHistogram,
	lineLengthHistogram *prometheus.HistogramVec
	lineCounter         prometheus.Counter
	lastLineLengthGauge prometheus.Gauge
}

func (a *App) processHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	line := r.URL.Query().Get("line")
	defer func() {
		a.latencyHistogram.
			With(prometheus.Labels{LabelMethod: r.Method}).Observe(sinceInMilliseconds(startTime))
		a.lineLengthHistogram.
			With(prometheus.Labels{LabelStatus: "OK"}).Observe(float64(len(line)))
		a.lineCounter.Inc()
		a.lastLineLengthGauge.Set(float64(len(line)))
	}()
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond) // имитация работы
	writeResponse(w, http.StatusOK, strings.ToUpper(line))
}

func (a *App) Init() error {
	// prometheus type: histogram
	a.latencyHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: Namespace,
		Name:      "latency",
		Help:      "The distribution of the latencies",
		Buckets: []float64{0, 25, 50, 75, 100, 200, 400, 600, 800, 1000, 2000, 4000,
			6000},
	}, []string{LabelMethod})

	// prometheus type: histogram
	a.lineLengthHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: Namespace,
		Name:      "line_lengths",
		Help:      "Groups the lengths of keys in buckets",
		// длины: [>=0B, >=5B, >=10B, >=15B, >=20B, >=40B, >=60B, >=80, >=100B, >=200B, >=400, >=600, >=800, >=1000]
		Buckets: []float64{0, 5, 10, 15, 20, 40, 60, 80, 100, 200, 400, 600, 800,
			1000},
	}, []string{LabelStatus})

	// prometheus type: counter
	a.lineCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "lines_in",
		Help:      "The number of lines from standard input",
	})

	//prometheus type: gauge
	a.lastLineLengthGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "last_line_length",
			Help:      "The length of the last received line",
		})
	prometheus.MustRegister(a.latencyHistogram)
	prometheus.MustRegister(a.lineLengthHistogram)
	prometheus.MustRegister(a.lineCounter)
	prometheus.MustRegister(a.lastLineLengthGauge)
	return nil
}

func sinceInMilliseconds(startTime time.Time) float64 {
	return float64(time.Since(startTime).Nanoseconds()) / 1e6
}
func writeResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_, _ = w.Write([]byte(message))
	_, _ = w.Write([]byte("\n"))
}
