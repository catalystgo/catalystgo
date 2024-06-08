package kafka_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/catalystgo/tracerok/logger"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	brokers []string
)

func TestMain(m *testing.M) {
	// Run the tests at the beginning to skip tets
	code := m.Run()
	if code == 0 {
		return
	}

	// TODO: Fix initialization of Kafka and Zookeeper containers

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	zookeeperEnv := map[string]string{
		"ZOOKEEPER_CLIENT_PORT": "2181",
	}

	// Define Zookeeper container
	zookeeperContainerReq := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-zookeeper:7.2.0",
		ExposedPorts: []string{"2181/tcp"},
		WaitingFor:   wait.ForLog("binding to port").WithStartupTimeout(60 * time.Second),
		Env:          zookeeperEnv,
	}

	// Start Zookeeper container
	zookeeperContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: zookeeperContainerReq,
		Started:          true,
	})

	if err != nil {
		logger.Fatalf(ctx, "create generic container: %v", err)
	}

	defer zookeeperContainer.Terminate(ctx)

	// Get Zookeeper container's IP address and port
	zookeeperHost, err := zookeeperContainer.Host(ctx)
	if err != nil {
		logger.Fatalf(ctx, "get container host", err)
	}

	zookeeperPort, err := zookeeperContainer.MappedPort(ctx, "2181")
	if err != nil {
		logger.Fatalf(ctx, "get container port", err)
	}

	kafkaEnv := map[string]string{
		"KAFKA_ADVERTISED_LISTENERS":      "PLAINTEXT://localhost:9092",
		"KAFKA_AUTO_CREATE_TOPICS_ENABLE": "true",
		"KAFKA_ZOOKEEPER_CONNECT":         fmt.Sprintf("%s:%s", zookeeperHost, zookeeperPort.Port()),
	}

	// Define Kafka container
	kafkaContainerReq := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-kafka:7.2.0",
		ExposedPorts: []string{"9092/tcp"},
		WaitingFor:   wait.ForLog("started (kafka.server.KafkaServer)").WithStartupTimeout(60 * time.Second),
		Env:          kafkaEnv,
	}

	// Start Kafka container
	kafkaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: kafkaContainerReq,
		Started:          true,
	})
	if err != nil {
		logger.Fatalf(ctx, "create generic container: %v", err)
	}

	defer kafkaContainer.Terminate(ctx)

	// Get Kafka container's IP address and port
	kafkaHost, err := kafkaContainer.Host(ctx)
	if err != nil {
		logger.Fatalf(ctx, "get container host", err)
	}

	kafkaPort, err := kafkaContainer.MappedPort(ctx, "9092")
	if err != nil {
		logger.Fatalf(ctx, "get container port", err)
	}

	// Set Kafka brokers
	brokers = []string{kafkaHost + ":" + kafkaPort.Port()}

	os.Exit(m.Run())
}
