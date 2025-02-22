package analyzer

import (
	"context"
	"strings"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/magiconair/properties/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestIngressAnalyzer(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		})
	ingressAnalyzer := IngressAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := ingressAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
}

func TestIngressAnalyzerWithMultipleIngresses(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		},
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example-2",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		},
	)
	ingressAnalyzer := IngressAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	analysisResults, err := ingressAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 2)
}

func TestIngressAnalyzerWithoutIngressClassAnnotation(t *testing.T) {

	clientset := fake.NewSimpleClientset(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		})
	ingressAnalyzer := IngressAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	analysisResults, err := ingressAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}

	var errorFound bool
	for _, analysis := range analysisResults {
		for _, err := range analysis.Error {
			if strings.Contains(err.Text, "does not specify an Ingress class") {
				errorFound = true
				break
			}
		}
		if errorFound {
			break
		}
	}
	if !errorFound {
		t.Error("expected error 'does not specify an Ingress class' not found in analysis results")
	}
}
