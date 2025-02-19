package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Read the .env file to get the EXTERNAL_IP value
	file, err := os.Open(".env")
	if err != nil {
		fmt.Printf("Error reading .env file: %v\n", err)
		return
	}
	defer file.Close()

	var externalIP string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "EXTERNAL_IP=") {
			externalIP = strings.TrimPrefix(line, "EXTERNAL_IP=")
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading .env file: %v\n", err)
		return
	}

	if externalIP == "" {
		fmt.Println("EXTERNAL_IP not found in .env file")
		return
	}

	// Define the Docker Compose template
	template := `
services:
    zookeeper:
        image: confluentinc/cp-zookeeper:7.5.0
        hostname: zookeeper
        container_name: zookeeper
        restart: always
        ports:
            - "2181:2181"
        environment:
            ZOOKEEPER_CLIENT_PORT: 2181
            ZOOKEEPER_TICK_TIME: 2000
        volumes:
            - zookeeper_data:/var/lib/zookeeper

    broker:
        image: confluentinc/cp-server:7.5.0
        hostname: broker
        container_name: broker
        restart: always
        depends_on:
            - zookeeper
        ports:
            - "9092:9092"
            - "9101:9101"
        environment:
            KAFKA_BROKER_ID: 1
            KAFKA_ZOOKEEPER_CONNECT: "zookeeper:2181"
            KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT
            KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://broker:29092,PLAINTEXT_HOST://%s:9092
            KAFKA_METRIC_REPORTERS: io.confluent.metrics.reporter.ConfluentMetricsReporter
            KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
            KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS: 0
            KAFKA_CONFLUENT_LICENSE_TOPIC_REPLICATION_FACTOR: 1
            KAFKA_CONFLUENT_BALANCER_TOPIC_REPLICATION_FACTOR: 1
            KAFKA_TRANSACTION_STATE_LOG_MIN_ISR: 1
            KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1
            KAFKA_JMX_PORT: 9101
            KAFKA_JMX_HOSTNAME: localhost
            KAFKA_CONFLUENT_SCHEMA_REGISTRY_URL: http://schema-registry:8081
            CONFLUENT_METRICS_REPORTER_BOOTSTRAP_SERVERS: broker:29092
            CONFLUENT_METRICS_REPORTER_TOPIC_REPLICAS: 1
            CONFLUENT_METRICS_ENABLE: "true"
            CONFLUENT_SUPPORT_CUSTOMER_ID: "anonymous"
        volumes:
            - broker_data:/var/lib/kafka

    kafka-ui:
        image: provectuslabs/kafka-ui:latest
        hostname: kafka-ui
        container_name: kafka-ui
        restart: always
        depends_on:
            - broker
        ports:
            - "8080:8080"
        environment:
            KAFKA_CLUSTERS_0_NAME: kafka
            KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS: "broker:29092"
            DYNAMIC_CONFIG_ENABLED: "true"
        volumes:
            - kafka_ui_data:/app/data

    control-center:
        image: confluentinc/cp-enterprise-control-center:7.5.0
        hostname: control-center
        container_name: control-center
        restart: always
        depends_on:
            - broker
        ports:
            - "9021:9021"
        environment:
            CONTROL_CENTER_BOOTSTRAP_SERVERS: "broker:29092"
            CONTROL_CENTER_REPLICATION_FACTOR: 1
            CONTROL_CENTER_INTERNAL_TOPICS_PARTITIONS: 1
            CONTROL_CENTER_MONITORING_INTERCEPTOR_TOPIC_PARTITIONS: 1
            CONFLUENT_METRICS_TOPIC_REPLICATION: 1
            PORT: 9021
        volumes:
            - control_center_data:/var/lib/control-center

    postgres:
        image: quay.io/debezium/example-postgres:2.1
        restart: always
        ports:
            - 5432:5432
        environment:
            - POSTGRES_USER=postgres
            - POSTGRES_PASSWORD=postgres
        volumes:
            - postgres_data:/var/lib/postgresql/data

    connect:
        image: quay.io/debezium/connect:2.1
        restart: always
        ports:
            - 8083:8083
        links:
            - broker
            - postgres
        environment:
            - BOOTSTRAP_SERVERS=broker:29092
            - GROUP_ID=1
            - CONFIG_STORAGE_TOPIC=my_connect_configs
            - OFFSET_STORAGE_TOPIC=my_connect_offsets
            - STATUS_STORAGE_TOPIC=my_connect_statuses
        volumes:
            - connect_data:/var/lib/kafka-connect

volumes:
    zookeeper_data:
    broker_data:
    kafka_ui_data:
    control_center_data:
    postgres_data:
    connect_data:
`

	// Replace ${EXTERNAL_IP} in the template
	content := fmt.Sprintf(template, externalIP)

	// Write the content to res.yml
	outputFile := "docker-compose.yaml"
	err = os.WriteFile(outputFile, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing to docker-compose.yaml: %v\n", err)
		return
	}

	fmt.Println("res.yml file generated successfully!")
}
