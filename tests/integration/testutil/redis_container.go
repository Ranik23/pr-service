package testutil

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type (
	
	RedisContainer struct {
		testcontainers.Container
		Config RedisContainerConfig
	}

	RedisContainerOption func(c *RedisContainerConfig)

	RedisContainerConfig struct {
		ImageTag   string
		Password   string
		MappedPort string
		Host       string
	}
)

func (c *RedisContainer) GetDSN() string {
	if c.Config.Password != "" {
		return fmt.Sprintf("redis://:%s@%s:%s", c.Config.Password, c.Config.Host, c.Config.MappedPort)
	}
	return fmt.Sprintf("redis://%s:%s", c.Config.Host, c.Config.MappedPort)
}

func NewRedisContainer(ctx context.Context, opts ...RedisContainerOption) (*RedisContainer, error) {
	const (
		redisImage = "redis"
		redisPort  = "6379"
	)

	config := RedisContainerConfig{
		ImageTag: "latest",
		Password: "",
	}

	for _, opt := range opts {
		opt(&config)
	}

	containerPort := redisPort + "/tcp"

	env := map[string]string{}
	if config.Password != "" {
		env["REDIS_PASSWORD"] = config.Password
	}

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env:          env,
			ExposedPorts: []string{containerPort},
			Image:        fmt.Sprintf("%s:%s", redisImage, config.ImageTag),
			WaitingFor:   wait.ForListeningPort(nat.Port(containerPort)),
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start Redis container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(containerPort))
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port for (%s): %w", containerPort, err)
	}
	config.MappedPort = mappedPort.Port()
	config.Host = host

	fmt.Println("Redis Host:", config.Host, config.MappedPort)

	return &RedisContainer{
		Container: container,
		Config:    config,
	}, nil
}
