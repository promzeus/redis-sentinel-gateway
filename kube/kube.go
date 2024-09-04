package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetKubernetesClient creates a Kubernetes client for interacting with the API
func GetKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to construct new client from config: %w", err)
	}

	return clientset, nil
}

// CreateKubeService creates a Service in Kubernetes with specified parameters
func CreateKubeService(ctx context.Context, clientset *kubernetes.Clientset, namespace, serviceName, portName string, portNumber int32) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     portName,
					Port:     portNumber,
					Protocol: corev1.ProtocolTCP,
				},
			},
			ClusterIP: "None",
		},
	}

	_, err := clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// If the service already exists, log a message and return without an error
			fmt.Printf("Service %s already exists in namespace %s\n", serviceName, namespace)
			return nil
		}
		// Wrap the error with additional context and return
		return fmt.Errorf("failed to create service %s in namespace %s: %w", serviceName, namespace, err)
	}

	return nil
}

// UpdateKubeEndpoint updates or creates Endpoints in Kubernetes with specified parameters
func UpdateKubeEndpoint(ctx context.Context, clientset *kubernetes.Clientset, namespace, endpointName, masterIP, portName string, portNumber int32) error {
	endpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: endpointName,
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{IP: masterIP},
				},
				Ports: []corev1.EndpointPort{
					{
						Name:     portName,
						Port:     portNumber,
						Protocol: corev1.ProtocolTCP,
					},
				},
			},
		},
	}

	_, err := clientset.CoreV1().Endpoints(namespace).Update(ctx, endpoint, metav1.UpdateOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// If the endpoint does not exist, try to create it
			_, err = clientset.CoreV1().Endpoints(namespace).Create(ctx, endpoint, metav1.CreateOptions{})
			if err != nil {
				// Wrap the error with additional context and return
				return fmt.Errorf("failed to create endpoint %s in namespace %s: %w", endpointName, namespace, err)
			}
		} else {
			// Wrap the error with additional context and return
			return fmt.Errorf("failed to update endpoint %s in namespace %s: %w", endpointName, namespace, err)
		}
	}
	return nil
}
