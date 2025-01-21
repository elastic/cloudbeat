// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package k8s

import (
	"os"

	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type ClientGetterAPI interface {
	GetClient(log *clog.Logger, kubeConfig string, options kubernetes.KubeClientOptions) (k8s.Interface, error)
}

type ClientGetter struct{}

func (ClientGetter) GetClient(log *clog.Logger, kubeConfig string, options kubernetes.KubeClientOptions) (k8s.Interface, error) {
	// Prevent klog from writing anything other than errors to stderr
	if replacementLogger, err := zap.NewProduction(); err != nil {
		replacementLogger = replacementLogger.WithOptions(zap.IncreaseLevel(zap.ErrorLevel))
		klog.SetLogger(zapr.NewLogger(replacementLogger))
	} else {
		log.Warnf("could not suppress klog: %+v", err)
	}

	client, err := kubernetes.GetKubernetesClient(kubeConfig, options)
	if err != nil {
		if kubernetes.IsInCluster(kubeConfig) {
			log.Debugf("Could not create kubernetes client using in_cluster config: %+v", err)
		} else if kubeConfig == "" {
			log.Debugf("Could not create kubernetes client using config: %v: %+v", os.Getenv("KUBECONFIG"), err)
		} else {
			log.Debugf("Could not create kubernetes client using config: %v: %+v", kubeConfig, err)
		}
	}

	return client, err
}
