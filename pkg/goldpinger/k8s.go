// Copyright 2018 Bloomberg Finance L.P.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goldpinger

import (
	"io/ioutil"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodNamespace is the auto-detected namespace for this goldpinger pod
var PodNamespace = getPodNamespace()

// GoldpingerPod contains just the basic info needed to ping and keep track of a given goldpinger pod
type GoldpingerPod struct {
	Name   string // Name is the name of the pod
	PodIP  string // PodIP is the IP address of the pod
	HostIP string // HostIP is the IP address of the host where the pod lives
}

func getPodNamespace() string {
	b, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		zap.L().Warn("Unable to determine namespace", zap.Error(err))
		return ""
	}
	namespace := string(b)
	return namespace
}

// GetAllPods returns a mapping from a pod name to a pointer to a GoldpingerPod(s)
func GetAllPods() map[string]*GoldpingerPod {
	timer := GetLabeledKubernetesCallsTimer()
	listOpts := metav1.ListOptions{
		LabelSelector: GoldpingerConfig.LabelSelector,
		FieldSelector: "status.phase=Running", // only select Running pods, otherwise we will get them before they have IPs
	}
	pods, err := GoldpingerConfig.KubernetesClient.CoreV1().Pods(*GoldpingerConfig.Namespace).List(listOpts)
	if err != nil {
		zap.L().Error("Error getting pods for selector", zap.String("selector", GoldpingerConfig.LabelSelector), zap.Error(err))
		CountError("kubernetes_api")
	} else {
		timer.ObserveDuration()
	}

	podMap := make(map[string]*GoldpingerPod)
	for _, pod := range pods.Items {
		podMap[pod.Name] = &GoldpingerPod{
			Name:   pod.Name,
			PodIP:  pod.Status.PodIP,
			HostIP: pod.Status.HostIP,
		}
	}
	return podMap
}
