package main

import (
	"fmt"
	"math/rand" // #nosec
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"../metric"
)

var (
	//EnableMetrics, _ = strconv.ParseBool(os.Getenv("EnableMetrics"))
	//MetricsPerHost, _ = strconv.ParseBool(os.Getenv("MetricsPerHost"))
	ListenPorts, _ = strconv.ParseInt(os.Getenv("ListenPorts"), 10, 64)
	//err error
)


func main() {
	rand.Seed(time.Now().UnixNano())

	reg := prometheus.NewRegistry()

	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{
		PidFn:        func() (int, error) { return os.Getpid(), nil },
		ReportErrors: true,
	}))

	mc, err := metric.NewCollector(reg)
	if err != nil {
		println(time.Now().Format(time.UnixDate),": ","Error creating prometheus collector:  %v", err)
	}

	mc.Start()

	mux := http.NewServeMux()
	registerMetrics(reg, mux)
	registerProfiler(mux)

	go startHTTPServer(int(ListenPorts), mux)

	for {
		time.Sleep(time.Second * 60)
		//Println(time.Now().Format(time.UnixDate), "the NumGoroutine done is: ", runtime.NumGoroutine())
	}
}

func registerMetrics(reg *prometheus.Registry, mux *http.ServeMux) {
	mux.Handle(
		"/metrics",
		promhttp.InstrumentMetricHandler(
			reg,
			promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
		),
	)
}

func registerProfiler(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/heap", pprof.Index)
	mux.HandleFunc("/debug/pprof/mutex", pprof.Index)
	mux.HandleFunc("/debug/pprof/goroutine", pprof.Index)
	mux.HandleFunc("/debug/pprof/threadcreate", pprof.Index)
	mux.HandleFunc("/debug/pprof/block", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func startHTTPServer(port int, mux *http.ServeMux) {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%v", port),
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      300 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	println(time.Now().Format(time.UnixDate),": ", server.ListenAndServe())
}
