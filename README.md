# Redis Sentinel Gateway 🚀

**Redis Sentinel Gateway** is a lightweight Go application designed to monitor Redis Sentinel for master node changes and automatically update Kubernetes Endpoints based on these changes. It simplifies the failover process by serving as a single point of entry for Redis, allowing clients to avoid dealing with Sentinel logic directly.

![Go](https://img.shields.io/github/go-mod/go-version/promzeus/redis-sentinel-gateway)
![License](https://img.shields.io/github/license/promzeus/redis-sentinel-gateway)
![Stars](https://img.shields.io/github/stars/promzeus/redis-sentinel-gateway)

## ✨ Features

- **Automatic Master Node Detection**: Monitors Redis Sentinel for master node changes in real-time.
- **Kubernetes Integration**: Automatically updates Kubernetes Endpoints based on the detected changes.
- **Leader Election**: Ensures high availability using Kubernetes leases for leader election.
- **Scalable**: Configurable polling and tick intervals for optimized performance and responsiveness.
- **Simple to Use**: Single point of entry for Redis clients, avoiding the need to implement Sentinel logic in your applications.

## 🛠 Configuration

The application is fully configurable via environment variables:

| Environment Variable | Description                                          | Default Value         |
|----------------------|------------------------------------------------------|-----------------------|
| `SERVICE_NAME`        | Name of the Kubernetes service                      | `redis-failover`      |
| `SENTINEL_ADDR`       | Address of the Redis Sentinel                        | `rfs-redis-node:26379`|
| `MASTER_NAME`         | Name of the Redis master in Sentinel                 | `mymaster`            |
| `NAMESPACE`           | Kubernetes namespace (auto-detected if not set)      | Auto-detected         |
| `POLL_INTERVAL`       | Interval between Kubernetes Endpoint updates         | `10s`                 |
| `TICK_INTERVAL`       | Interval for checking master node changes            | `1s`                  |
| `LEASE_NAME`          | Name of the Kubernetes lease for leader election     | `redis-failover-lease`|
| `PORT_NAME`           | Name of the port for the service                     | `redis`               |
| `PORT_NUMBER`         | Port number for the service                          | `6379`                |
| `REDIS_PASSWORD`      | Password for connecting to Redis Sentinel            | None                  |
| `HOSTNAME`            | Pod's hostname, used as a unique identifier          | Auto-detected         |

## ⚙️ How It Works

1. **Monitoring**: The gateway continuously monitors Redis Sentinel for changes in the master node status.
2. **Automatic Updates**: When a master node change is detected, the Kubernetes Endpoints are automatically updated.
3. **Leader Election**: Using Kubernetes leases, only one instance of the gateway is responsible for updating the Endpoints at any time, ensuring high availability.
4. **Polling & Tick Intervals**: Configurable intervals control how frequently the application checks the Redis Sentinel and how often Kubernetes Endpoints are updated.

## 📦 Installation & Usage

### Helm Chart

To install the Redis Sentinel Gateway via Helm, follow these steps:

.....>>> next time 
