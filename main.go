package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Gauge
var onlineUsers = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "goapp_online_users", // nome da métrica
	Help: "Online users",       // descrição da métrica
	ConstLabels: map[string]string{
		"course": "fullcycle", // podemos colocar nome que quiser
	},
})

// Counter
var httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "goapp_http_requests_total",
	Help: "Count of all HTTP requests for goapp",
}, []string{})

// Histogram
var httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "goapp_http_requests_duration",
	Help: "Duration in seconds of all HTTP requests",
}, []string{"handler"})

func main() {
	r := prometheus.NewRegistry()
	r.MustRegister(onlineUsers)
	r.MustRegister(httpRequestsTotal)
	r.MustRegister(httpDuration)

	go func() {
		for {
			onlineUsers.Set(float64(rand.Intn(2000)))
		}
	}()

	home := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rand.Seed(time.Now().UnixNano())
		n := rand.Intn(1000)
		time.Sleep(time.Duration(n) * time.Millisecond)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello FullCycle"))
	})

	contact := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rand.Seed(time.Now().UnixNano())
		n := rand.Intn(5000)
		time.Sleep(time.Duration(n) * time.Millisecond)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Contato"))
	})

	// duration
	d := promhttp.InstrumentHandlerDuration(
		httpDuration.MustCurryWith(prometheus.Labels{"handler": "home"}),
		promhttp.InstrumentHandlerCounter(httpRequestsTotal, home),
	)

	d2 := promhttp.InstrumentHandlerDuration(
		httpDuration.MustCurryWith(prometheus.Labels{"handler": "contact"}),
		promhttp.InstrumentHandlerCounter(httpRequestsTotal, contact),
	)

	// acessando "/" o contador será incrementado e executado o handler da home
	// http.Handle("/", promhttp.InstrumentHandlerCounter(httpRequestsTotal, home))

	// agora a home com a duration:
	http.Handle("/", d)
	http.Handle("/contact", d2)

	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":8081", nil))
}
