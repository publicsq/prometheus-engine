// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/prometheus-engine/e2e/deploy"
	"github.com/GoogleCloudPlatform/prometheus-engine/e2e/kube"
	monitoringv1 "github.com/GoogleCloudPlatform/prometheus-engine/pkg/operator/apis/monitoring/v1"
	"github.com/GoogleCloudPlatform/prometheus-engine/pkg/operator/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func TestTLSPodMonitoring(t *testing.T) {
	ctx := context.Background()
	kubeClient, opClient, err := setupCluster(ctx, t)
	if err != nil {
		t.Fatalf("error instantiating clients. err: %s", err)
	}

	t.Run("collector-deployed", testCollectorDeployed(ctx, kubeClient))
	t.Run("enable-target-status", testEnableTargetStatus(ctx, opClient))
	t.Run("patch-example-app-args", testPatchExampleAppArgs(ctx, kubeClient, []string{"--tls-create-self-signed=true"}))

	pm := &monitoringv1.PodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls-ready",
			Namespace: "default",
		},
		Spec: monitoringv1.PodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Scheme:   "https",
					Port:     intstr.FromString("web"),
					Interval: "5s",
					HTTPClientConfig: monitoringv1.HTTPClientConfig{
						TLS: &monitoringv1.TLS{
							InsecureSkipVerify: true,
						},
					},
				},
			},
		},
	}
	t.Run("tls-podmonitoring-ready", testEnsurePodMonitoringReady(ctx, opClient, pm))

	pmFail := &monitoringv1.PodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls-fail",
			Namespace: "default",
		},
		Spec: monitoringv1.PodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Scheme:   "https",
					Port:     intstr.FromString("web"),
					Interval: "5s",
				},
			},
		},
	}
	errMsg := "x509: certificate signed by unknown authority"
	t.Run("tls-podmonitoring-failure", testEnsurePodMonitoringFailure(ctx, opClient, pmFail, errMsg))
}

func TestTLSClusterPodMonitoring(t *testing.T) {
	ctx := context.Background()
	kubeClient, opClient, err := setupCluster(ctx, t)
	if err != nil {
		t.Fatalf("error instantiating clients. err: %s", err)
	}

	t.Run("collector-deployed", testCollectorDeployed(ctx, kubeClient))
	t.Run("enable-target-status", testEnableTargetStatus(ctx, opClient))
	t.Run("patch-example-app-args", testPatchExampleAppArgs(ctx, kubeClient, []string{"--tls-create-self-signed=true"}))

	cpm := &monitoringv1.ClusterPodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls-ready",
			Namespace: "default",
		},
		Spec: monitoringv1.ClusterPodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Scheme:   "https",
					Port:     intstr.FromString("web"),
					Interval: "5s",
					HTTPClientConfig: monitoringv1.HTTPClientConfig{
						TLS: &monitoringv1.TLS{
							InsecureSkipVerify: true,
						},
					},
				},
			},
		},
	}
	t.Run("tls-clusterpodmonitoring-ready", testEnsureClusterPodMonitoringReady(ctx, opClient, cpm))

	cpmFail := &monitoringv1.ClusterPodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls-fail",
			Namespace: "default",
		},
		Spec: monitoringv1.ClusterPodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Scheme:   "https",
					Port:     intstr.FromString("web"),
					Interval: "5s",
				},
			},
		},
	}
	errMsg := "x509: certificate signed by unknown authority"
	t.Run("tls-clusterpodmonitoring-failure", testEnsureClusterPodMonitoringFailure(ctx, opClient, cpmFail, errMsg))
}

func TestBasicAuthPodMonitoring(t *testing.T) {
	ctx := context.Background()
	kubeClient, opClient, err := setupCluster(ctx, t)
	if err != nil {
		t.Fatalf("error instantiating clients. err: %s", err)
	}

	t.Run("collector-deployed", testCollectorDeployed(ctx, kubeClient))
	t.Run("enable-target-status", testEnableTargetStatus(ctx, opClient))
	t.Run("patch-example-app-args", testPatchExampleAppArgs(ctx, kubeClient, []string{"--basic-auth-username=user"}))

	pm := &monitoringv1.PodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "basic-auth-ready",
			Namespace: "default",
		},
		Spec: monitoringv1.PodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Port:     intstr.FromString("web"),
					Interval: "5s",
					HTTPClientConfig: monitoringv1.HTTPClientConfig{
						BasicAuth: &monitoringv1.BasicAuth{
							Username: "user",
						},
					},
				},
			},
		},
	}
	t.Run("basic-auth-podmonitoring-ready", testEnsurePodMonitoringReady(ctx, opClient, pm))

	pmFail := &monitoringv1.PodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "basic-auth-fail",
			Namespace: "default",
		},
		Spec: monitoringv1.PodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Port:     intstr.FromString("web"),
					Interval: "5s",
				},
			},
		},
	}
	errMsg := "server returned HTTP status 401 Unauthorized"
	t.Run("basic-auth-podmonitoring-failure", testEnsurePodMonitoringFailure(ctx, opClient, pmFail, errMsg))
}

func TestBasicAuthClusterPodMonitoring(t *testing.T) {
	ctx := context.Background()
	kubeClient, opClient, err := setupCluster(ctx, t)
	if err != nil {
		t.Fatalf("error instantiating clients. err: %s", err)
	}

	t.Run("collector-deployed", testCollectorDeployed(ctx, kubeClient))
	t.Run("enable-target-status", testEnableTargetStatus(ctx, opClient))
	t.Run("patch-example-app-args", testPatchExampleAppArgs(ctx, kubeClient, []string{"--basic-auth-username=user"}))

	cpm := &monitoringv1.ClusterPodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "basic-auth-ready",
			Namespace: "default",
		},
		Spec: monitoringv1.ClusterPodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Port:     intstr.FromString("web"),
					Interval: "5s",
					HTTPClientConfig: monitoringv1.HTTPClientConfig{
						BasicAuth: &monitoringv1.BasicAuth{
							Username: "user",
						},
					},
				},
			},
		},
	}
	t.Run("basic-auth-clusterpodmonitoring-ready", testEnsureClusterPodMonitoringReady(ctx, opClient, cpm))

	cpmFail := &monitoringv1.ClusterPodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "basic-auth-fail",
			Namespace: "default",
		},
		Spec: monitoringv1.ClusterPodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Port:     intstr.FromString("web"),
					Interval: "5s",
				},
			},
		},
	}
	errMsg := "server returned HTTP status 401 Unauthorized"
	t.Run("basic-auth-clusterpodmonitoring-failure", testEnsureClusterPodMonitoringFailure(ctx, opClient, cpmFail, errMsg))
}

func TestAuthorizationPodMonitoring(t *testing.T) {
	ctx := context.Background()
	kubeClient, opClient, err := setupCluster(ctx, t)
	if err != nil {
		t.Fatalf("error instantiating clients. err: %s", err)
	}

	t.Run("collector-deployed", testCollectorDeployed(ctx, kubeClient))
	t.Run("enable-target-status", testEnableTargetStatus(ctx, opClient))
	t.Run("patch-example-app-args", testPatchExampleAppArgs(ctx, kubeClient, []string{"--auth-scheme=Bearer"}))

	pm := &monitoringv1.PodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "auth-ready",
			Namespace: "default",
		},
		Spec: monitoringv1.PodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Port:     intstr.FromString("web"),
					Interval: "5s",
					HTTPClientConfig: monitoringv1.HTTPClientConfig{
						Authorization: &monitoringv1.Auth{
							Type: "Bearer",
						},
					},
				},
			},
		},
	}
	t.Run("auth-podmonitoring-ready", testEnsurePodMonitoringReady(ctx, opClient, pm))

	pmFail := &monitoringv1.PodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "auth-fail",
			Namespace: "default",
		},
		Spec: monitoringv1.PodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Port:     intstr.FromString("web"),
					Interval: "5s",
				},
			},
		},
	}
	errMsg := "server returned HTTP status 401 Unauthorized"
	t.Run("auth-podmonitoring-failure", testEnsurePodMonitoringFailure(ctx, opClient, pmFail, errMsg))
}

func TestAuthorizationClusterPodMonitoring(t *testing.T) {
	ctx := context.Background()
	kubeClient, opClient, err := setupCluster(ctx, t)
	if err != nil {
		t.Fatalf("error instantiating clients. err: %s", err)
	}

	t.Run("collector-deployed", testCollectorDeployed(ctx, kubeClient))
	t.Run("enable-target-status", testEnableTargetStatus(ctx, opClient))
	t.Run("patch-example-app-args", testPatchExampleAppArgs(ctx, kubeClient, []string{"--auth-scheme=Bearer"}))

	cpm := &monitoringv1.ClusterPodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "auth-ready",
			Namespace: "default",
		},
		Spec: monitoringv1.ClusterPodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Port:     intstr.FromString("web"),
					Interval: "5s",
					HTTPClientConfig: monitoringv1.HTTPClientConfig{
						Authorization: &monitoringv1.Auth{
							Type: "Bearer",
						},
					},
				},
			},
		},
	}
	t.Run("auth-clusterpodmonitoring-ready", testEnsureClusterPodMonitoringReady(ctx, opClient, cpm))

	cpmFail := &monitoringv1.ClusterPodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "auth-fail",
			Namespace: "default",
		},
		Spec: monitoringv1.ClusterPodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Port:     intstr.FromString("web"),
					Interval: "5s",
				},
			},
		},
	}
	errMsg := "server returned HTTP status 401 Unauthorized"
	t.Run("auth-clusterpodmonitoring-failure", testEnsureClusterPodMonitoringFailure(ctx, opClient, cpmFail, errMsg))
}

func TestOAuth2PodMonitoring(t *testing.T) {
	ctx := context.Background()
	kubeClient, opClient, err := setupCluster(ctx, t)
	if err != nil {
		t.Fatalf("error instantiating clients. err: %s", err)
	}
	var (
		clientID    = "gmp-user-client-id-no-client-secret"
		clientScope = "read"
		accessToken = "abc123"
	)

	t.Run("collector-deployed", testCollectorDeployed(ctx, kubeClient))
	t.Run("enable-target-status", testEnableTargetStatus(ctx, opClient))
	t.Run("patch-example-app-args", testPatchExampleAppArgs(ctx, kubeClient,
		[]string{
			fmt.Sprintf("--oauth2-client-id=%s", clientID),
			fmt.Sprintf("--oauth2-scopes=%s", clientScope),
			fmt.Sprintf("--oauth2-access-token=%s", accessToken)}))

	pm := &monitoringv1.PodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "oauth-ready",
			Namespace: "default",
		},
		Spec: monitoringv1.PodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Port:     intstr.FromString("web"),
					Interval: "5s",
					HTTPClientConfig: monitoringv1.HTTPClientConfig{
						OAuth2: &monitoringv1.OAuth2{
							ClientID: clientID,
							Scopes:   []string{clientScope},
							TokenURL: "http://go-synthetic.default.svc.cluster.local:8080/token",
						},
					},
				},
			},
		},
	}
	t.Run("oauth2-podmonitoring-ready", testEnsurePodMonitoringReady(ctx, opClient, pm))

	pmFail := &monitoringv1.PodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "oauth-fail",
			Namespace: "default",
		},
		Spec: monitoringv1.PodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Port:     intstr.FromString("web"),
					Interval: "5s",
				},
			},
		},
	}
	t.Run("oauth2-podmonitoring-failure", testEnsurePodMonitoringFailure(ctx, opClient, pmFail, "server returned HTTP status 401 Unauthorized"))
}

func TestOAuth2ClusterPodMonitoring(t *testing.T) {
	ctx := context.Background()
	kubeClient, opClient, err := setupCluster(ctx, t)
	if err != nil {
		t.Fatalf("error instantiating clients. err: %s", err)
	}
	var (
		clientID    = "gmp-user-client-id-no-client-secret"
		clientScope = "read"
		accessToken = "abc123"
	)

	t.Run("collector-deployed", testCollectorDeployed(ctx, kubeClient))
	t.Run("enable-target-status", testEnableTargetStatus(ctx, opClient))
	t.Run("patch-example-app-args", testPatchExampleAppArgs(ctx, kubeClient,
		[]string{
			fmt.Sprintf("--oauth2-client-id=%s", clientID),
			fmt.Sprintf("--oauth2-scopes=%s", clientScope),
			fmt.Sprintf("--oauth2-access-token=%s", accessToken)}))

	cpm := &monitoringv1.ClusterPodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name: "oauth-ready",
		},
		Spec: monitoringv1.ClusterPodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Port:     intstr.FromString("web"),
					Interval: "5s",
					HTTPClientConfig: monitoringv1.HTTPClientConfig{
						OAuth2: &monitoringv1.OAuth2{
							ClientID: clientID,
							Scopes:   []string{clientScope},
							TokenURL: "http://go-synthetic.default.svc.cluster.local:8080/token",
						},
					},
				},
			},
		},
	}
	t.Run("oauth2-clusterpodmonitoring-ready", testEnsureClusterPodMonitoringReady(ctx, opClient, cpm))

	cpmFail := &monitoringv1.ClusterPodMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name: "oauth-fail",
		},
		Spec: monitoringv1.ClusterPodMonitoringSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "go-synthetic",
				},
			},
			Endpoints: []monitoringv1.ScrapeEndpoint{
				{
					Port:     intstr.FromString("web"),
					Interval: "5s",
				},
			},
		},
	}
	t.Run("oauth2-clusterpodmonitoring-failure", testEnsureClusterPodMonitoringFailure(ctx, opClient, cpmFail, "server returned HTTP status 401 Unauthorized"))
}

func testPatchExampleAppArgs(ctx context.Context, kubeClient kubernetes.Interface, args []string) func(*testing.T) {
	return func(t *testing.T) {
		scheme, err := newScheme()
		if err != nil {
			t.Errorf("create scheme: %s", err)
		}

		deployment, service, err := deploy.SyntheticAppResources(scheme)
		if err != nil {
			t.Errorf("get synthetic app resources: %s", err)
		}

		container, err := kube.DeploymentContainer(deployment, deploy.SyntheticAppContainerName)
		if err != nil {
			t.Errorf("find synthetic app container: %s", err)
		}
		container.Args = append(container.Args, args...)
		_, err = kubeClient.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{})
		if err != nil {
			t.Errorf("create deployment: %s", err)
		}

		_, err = kubeClient.CoreV1().Services("default").Create(ctx, service, metav1.CreateOptions{})
		if err != nil {
			t.Errorf("create service: %s", err)
		}
	}
}

func isPodMonitoringScrapeEndpointFailure(status *monitoringv1.ScrapeEndpointStatus, errMsg string) error {
	if status.UnhealthyTargets == 0 {
		return errors.New("expected no healthy targets")
	}
	if status.CollectorsFraction == "0" {
		return fmt.Errorf("expected collectors fraction to be 0 but found: %s", status.CollectorsFraction)
	}
	if len(status.SampleGroups) == 0 {
		return errors.New("missing sample groups")
	}
	for i, group := range status.SampleGroups {
		if len(group.SampleTargets) == 0 {
			return fmt.Errorf("missing sample targets for group %d", i)
		}
		for _, target := range group.SampleTargets {
			if target.Health == "up" {
				return fmt.Errorf("healthy target %q at group %d", target.Health, i)
			}
			if target.LastError == nil {
				return fmt.Errorf("missing error for target at group %d", i)
			}
			if !strings.Contains(*target.LastError, errMsg) {
				return fmt.Errorf("expected error message %q at group %d: got %s", errMsg, i, *target.LastError)
			}
		}
	}
	return nil
}

func testEnsurePodMonitoringFailure(ctx context.Context, opClient versioned.Interface, pm *monitoringv1.PodMonitoring, errMsg string) func(*testing.T) {
	return testEnsurePodMonitoringStatus(ctx, opClient, pm,
		func(status *monitoringv1.ScrapeEndpointStatus) error {
			return isPodMonitoringScrapeEndpointFailure(status, errMsg)
		})
}

func testEnsureClusterPodMonitoringFailure(ctx context.Context, opClient versioned.Interface, cpm *monitoringv1.ClusterPodMonitoring, errMsg string) func(*testing.T) {
	return testEnsureClusterPodMonitoringStatus(ctx, opClient, cpm,
		func(status *monitoringv1.ScrapeEndpointStatus) error {
			return isPodMonitoringScrapeEndpointFailure(status, errMsg)
		})
}
