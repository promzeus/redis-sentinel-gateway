package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"redis-sentinel-failover/kube"
	"redis-sentinel-failover/sentinel"

	"github.com/redis/go-redis/v9"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// getEnv returns the value of the environment variable if set, or a fallback value otherwise
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// getNamespace reads the namespace in which the Pod is running from the Kubernetes service account secret
func getNamespace() string {
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Warning: Namespace file does not exist, using default namespace")
			return "default"
		} else {
			fmt.Printf("Error reading namespace file: %v\n", err)
			os.Exit(1)
		}
	}
	return string(data)
}

func main() {
	// Getting configuration from environment variables
	serviceName := getEnv("SERVICE_NAME", "redis-failover")
	sentinelAddr := getEnv("SENTINEL_ADDR", "rfs-redis-node:26379")
	masterName := getEnv("MASTER_NAME", "mymaster")
	namespace := getEnv("NAMESPACE", getNamespace())
	pollIntervalStr := getEnv("POLL_INTERVAL", "5")
	leaseName := getEnv("LEASE_NAME", "redis-failover-lease")
	portName := getEnv("PORT_NAME", "redis")
	portNumberStr := getEnv("PORT_NUMBER", "6379")
	tickIntervalStr := getEnv("TICK_INTERVAL", "5")
	leaderID := os.Getenv("HOSTNAME")
	password := os.Getenv("REDIS_PASSWORD")

	// Convert pollInterval, tickInterval, and portNumber from string to appropriate types
	pollInterval, err := time.ParseDuration(pollIntervalStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid poll interval: %v", err))
	}

	tickInterval, err := time.ParseDuration(tickIntervalStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid tick interval: %v", err))
	}

	portNumber, err := strconv.Atoi(portNumberStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid port number: %v", err))
	}

	// Context and signal handling
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	clientset, err := kube.GetKubernetesClient()
	if err != nil {
		panic(err.Error())
	}

	// Create Redis Sentinel client
	sentinelClient := redis.NewSentinelClient(&redis.Options{
		Addr:     sentinelAddr,
		Password: password,
	})

	err = kube.CreateKubeService(ctx, clientset, namespace, serviceName, portName, int32(portNumber))
	if err != nil {
		fmt.Printf("Error creating service: %v\n", err)
	}

	// Configure leader election via Kubernetes Lease
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      leaseName,
			Namespace: namespace,
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: leaderID,
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
			Lock:          lock,
			LeaseDuration: 15 * time.Second,
			RenewDeadline: 10 * time.Second,
			RetryPeriod:   2 * time.Second,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(ctx context.Context) {
					fmt.Println("Became the leader, starting to manage Endpoints")
					runEndpointUpdater(ctx, clientset, sentinelClient, namespace, serviceName, masterName, portName, int32(portNumber), pollInterval, tickInterval)
				},
				OnStoppedLeading: func() {
					fmt.Println("Lost leadership, stopping Endpoints management")
				},
			},
		})
	}()

	<-ctx.Done()

	wg.Wait()
}

func runEndpointUpdater(ctx context.Context, clientset *kubernetes.Clientset, sentinelClient *redis.SentinelClient, namespace, serviceName, masterName, portName string, portNumber int32, pollInterval, tickInterval time.Duration) {
	var lastMasterIP string
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			masterIP, err := sentinel.GetMasterIP(ctx, sentinelClient, masterName)
			if err != nil {
				fmt.Printf("Error getting master IP: %v\n", err)
				continue
			}

			if masterIP != lastMasterIP {
				err = kube.UpdateKubeEndpoint(ctx, clientset, namespace, serviceName, masterIP, portName, portNumber)
				if err != nil {
					fmt.Printf("Error updating Kubernetes Endpoint: %v\n", err)
				} else {
					fmt.Printf("Updated Kubernetes Endpoint with new master IP: %s\n", masterIP)
					lastMasterIP = masterIP
				}
			}
		case <-ctx.Done():
			fmt.Println("Stopping endpoint updater")
			return
		}
	}
}
