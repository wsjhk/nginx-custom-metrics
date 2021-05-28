/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metric

import (
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/sets"
	"../metric/collectors"
)

// Collector defines the interface for a metric collector
type Collector interface {
	SetHosts(sets.String)

	Start()
	Stop()
}

type collector struct {
	nginxStatus  collectors.NGINXStatusCollector
	nginxProcess collectors.NGINXProcessCollector

	socket *collectors.SocketCollector

	registry *prometheus.Registry
}

// NewCollector creates a new metric collector the for ingress controller
func NewCollector(metricsPerHost bool, registry *prometheus.Registry) (Collector, error) {
	podNamespace := os.Getenv("POD_NAMESPACE")
	if podNamespace == "" {
		podNamespace = "default"
	}

	podName := os.Getenv("POD_NAME")
    nginxClass := os.Getenv("NGINX_CLASS")

	nc, err := collectors.NewNGINXStatus(podName, podNamespace, nginxClass)
	if err != nil {
		return nil, err
	}

	pc, err := collectors.NewNGINXProcess(podName, podNamespace, nginxClass)
	if err != nil {
		return nil, err
	}

	s, err := collectors.NewSocketCollector(podName, podNamespace, nginxClass, metricsPerHost)
	if err != nil {
		return nil, err
	}

	return Collector(&collector{
		nginxStatus:  nc,
		nginxProcess: pc,
		socket: s,
		registry: registry,
	}), nil
}

func (c *collector) Start() {
	c.registry.MustRegister(c.nginxStatus)
	c.registry.MustRegister(c.nginxProcess)
	c.registry.MustRegister(c.socket)

	// the default nginx-tmpl.conf does not contains
	// a server section with the status port
	go func() {
		time.Sleep(5 * time.Second)
		c.nginxStatus.Start()
	}()
	go c.nginxProcess.Start()
	go c.socket.Start()
}

func (c *collector) Stop() {
	c.registry.Unregister(c.nginxStatus)
	c.registry.Unregister(c.nginxProcess)
	c.registry.Unregister(c.socket)

	c.nginxStatus.Stop()
	c.nginxProcess.Stop()
	c.socket.Stop()
}

func (c *collector) SetHosts(hosts sets.String) {
	c.socket.SetHosts(hosts)
}

