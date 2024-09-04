package sentinel

import (
	"context"
	"net"

	"github.com/redis/go-redis/v9"
)

// GetMasterIP retrieves the IP address of the Redis master from Sentinel using the provided Sentinel client.
func GetMasterIP(ctx context.Context, sentinel *redis.SentinelClient, masterName string) (string, error) {
	addr, err := sentinel.GetMasterAddrByName(ctx, masterName).Result()
	if err != nil {
		return "", err
	}

	ips, err := net.LookupIP(addr[0])
	if err != nil {
		return "", err
	}
	return ips[0].String(), nil
}
