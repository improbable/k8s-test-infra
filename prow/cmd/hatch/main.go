/*
Copyright 2016 The Kubernetes Authors.

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

package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"

	"google.golang.org/api/option"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/hatch"
	"k8s.io/test-infra/prow/logrusutil"
)

type options struct {
	port              int
	configPath        string
	webhookSecretFile string
	gcsSecretFile     string
}

func (o *options) Validate() error {
	return nil
}

func gatherOptions() options {
	o := options{}
	flag.IntVar(&o.port, "port", 8889, "Port to listen on.")

	flag.StringVar(&o.configPath, "config-path", "/etc/config/config", "Path to config.yaml.")

	flag.StringVar(&o.webhookSecretFile, "webhook-hmac-secret-path", "/etc/webhook/hmac", "Path to the file containing the GitHub HMAC secret.")
	flag.StringVar(&o.gcsSecretFile, "gcs-auth-path", "/etc/google/key.json", "Path to the file containing the GCS authentication to use (service account JSON).")
	flag.Parse()
	return o
}

func main() {
	o := gatherOptions()
	if err := o.Validate(); err != nil {
		logrus.Fatalf("Invalid options: %v", err)
	}
	logrus.SetFormatter(logrusutil.NewDefaultFieldsFormatter(nil, logrus.Fields{"component": "hatch"}))

	configAgent := &config.Agent{}
	if err := configAgent.Start(o.configPath, ""); err != nil {
		logrus.WithError(err).Fatal("Error starting config agent.")
	}

	// Ignore SIGTERM so that we don't drop hooks when the pod is removed.
	// We'll get SIGTERM first and then SIGKILL after our graceful termination
	// deadline.
	signal.Ignore(syscall.SIGTERM)

	webhookSecretRaw, err := ioutil.ReadFile(o.webhookSecretFile)
	if err != nil {
		logrus.WithError(err).Fatal("Could not read webhook secret file.")
	}
	webhookSecret := bytes.TrimSpace(webhookSecretRaw)

	server := &hatch.Server{
		HMACSecret:  webhookSecret,
		ConfigAgent: configAgent,
		GCSOptions:  option.WithCredentialsFile(o.gcsSecretFile),
	}

	// Return 200 on / for health checks.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	// For /hook, handle a webhook normally.
	http.Handle("/hook", server)

	logrus.Fatal(http.ListenAndServe(":"+strconv.Itoa(o.port), nil))
}
