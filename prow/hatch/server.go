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

package hatch

import (
	"cloud.google.com/go/storage"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"net/http"
	"time"
)

const (
	defaultBucketObjectPrefix = "hatch-webhooks"
)

type Server struct {
	HMACSecret  []byte
	ConfigAgent *config.Agent
	GCSOptions  option.ClientOption

	c http.Client
}

type RepositoryRelatedPR struct {
	PullRequest struct {
		Number int `json:"number"`
	} `json:"pull_request"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// ServeHTTP validates an incoming webhook and puts it into the event channel.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bucketName := s.ConfigAgent.Config().Hatch.GCSBucketName

	if len(bucketName) == 0 {
		logrus.Error("No bucket name provided in config for Hatch")
		return
	}

	_, eventGUID, payload, ok := github.ValidateWebhook(w, r, s.HMACSecret)
	if !ok {
		fmt.Fprint(w, "Invalid webhook")
		return
	}

	webhookData := &RepositoryRelatedPR{}

	err := json.Unmarshal(payload, webhookData)
	if err != nil || len(webhookData.Repository.FullName) == 0 || webhookData.PullRequest.Number <= 0 {
		logrus.Infof("Ignoring webhook which doesn't have repository and pull request data attached")
		return
	}

	logrus.Infof("Storing a webhook about %v", webhookData.Repository.FullName)
	s.StoreWebhook(w, r, eventGUID, webhookData, payload)
}

func (s *Server) StoreWebhook(
	w http.ResponseWriter, r *http.Request, eventGUID string, webhookData RepositoryRelatedPR, payload []byte,
) {
	ctx := r.Context()

	client, err := storage.NewClient(ctx, s.GCSOptions)
	if err != nil {
		logrus.WithError(err).Error("Failed to get storage client")
		return
	}

	objectNamePrefix := s.ConfigAgent.Config().Hatch.GCSPrefix

	if len(objectNamePrefix) == 0 {
		objectNamePrefix = defaultBucketObjectPrefix
	}

	timeNow := time.Now().UTC()

	bkt := client.Bucket(bucketName)
	objectName := fmt.Sprintf("%s/%s/%d/%d/%d/%d/%s",
		objectNamePrefix, webhookData.Repository.FullName, webhookData.PullRequest.Number,
		timeNow.Year(), timeNow.Month(), timeNow.Day(), eventGUID)
	writer := bkt.Object(objectName).NewWriter(ctx)
	writer.Write(payload)
	writer.Close()

	logrus.Info("Logged to GCS")
}
