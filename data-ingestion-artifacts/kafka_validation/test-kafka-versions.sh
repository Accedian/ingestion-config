#!/bin/bash
# Test Telegraf Kafka consumer across different Kafka versions
# This script launches docker-compose with different Kafka versions,
# pushes a test message, and verifies it appears in verification.log

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Kafka versions to test
KAFKA_VERSIONS=("latest" "3.9.0" "4.1.2")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test topic
TOPIC="pca_kpi_topic"

# Output directory
OUT_DIR="./out"
VERIFICATION_LOG="$OUT_DIR/verification.log"

cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    docker compose down -v 2>/dev/null || true
    rm -f "$VERIFICATION_LOG" 2>/dev/null || true
}

wait_for_kafka() {
    local max_attempts=30
    local attempt=1
    echo "Waiting for Kafka to be ready..."
    while [ $attempt -le $max_attempts ]; do
        if docker exec kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list &>/dev/null; then
            echo "Kafka is ready!"
            return 0
        fi
        echo "  Attempt $attempt/$max_attempts - waiting..."
        sleep 2
        ((attempt++))
    done
    echo -e "${RED}Kafka failed to become ready${NC}"
    return 1
}

create_topic() {
    echo "Creating topic $TOPIC..."
    docker exec kafka /opt/kafka/bin/kafka-topics.sh \
        --bootstrap-server localhost:9092 \
        --create \
        --topic "$TOPIC" \
        --partitions 1 \
        --replication-factor 1 \
        --if-not-exists
}

send_test_message() {
    local version=$1
    local timestamp=$(date +%s)
    local message="{\"kpi\":\"test_metric\",\"value\":42.5,\"timestamp\":$timestamp,\"device\":\"test-device\",\"node_id\":\"node-001\",\"schema\":\"test\",\"source_ip\":\"10.0.0.1\",\"node_ip\":\"10.0.0.2\",\"node_type\":\"test\",\"kafka_version_tested\":\"$version\"}"
    
    echo "Sending test message to $TOPIC..."
    echo "$message" | docker exec -i kafka /opt/kafka/bin/kafka-console-producer.sh \
        --bootstrap-server localhost:9092 \
        --topic "$TOPIC"
    
    echo "Message sent: $message"
}

verify_message() {
    local version=$1
    local max_attempts=30
    local attempt=1
    
    echo "Waiting for message to appear in verification.log..."
    while [ $attempt -le $max_attempts ]; do
        if [ -f "$VERIFICATION_LOG" ]; then
            if grep -q "test_metric" "$VERIFICATION_LOG" 2>/dev/null; then
                echo -e "${GREEN}SUCCESS: Message found in verification.log for Kafka $version${NC}"
                echo "Log contents:"
                cat "$VERIFICATION_LOG"
                return 0
            fi
        fi
        echo "  Attempt $attempt/$max_attempts - waiting for telegraf to process..."
        sleep 2
        ((attempt++))
    done
    
    echo -e "${RED}FAILED: Message not found in verification.log for Kafka $version${NC}"
    if [ -f "$VERIFICATION_LOG" ]; then
        echo "Log contents:"
        cat "$VERIFICATION_LOG"
    else
        echo "verification.log does not exist"
    fi
    return 1
}

test_kafka_version() {
    local version=$1
    echo ""
    echo "========================================"
    echo -e "${YELLOW}Testing Kafka version: $version${NC}"
    echo "========================================"
    
    # Clean up from previous run
    cleanup
    
    # Ensure output directory exists
    mkdir -p "$OUT_DIR"
    
    # Start services with specific Kafka version
    echo "Starting docker-compose with Kafka $version..."
    KAFKA_VERSION="$version" docker compose up -d
    
    # Wait for Kafka to be ready
    if ! wait_for_kafka; then
        echo -e "${RED}FAILED: Kafka $version did not start properly${NC}"
        docker compose logs
        return 1
    fi
    
    # Create topic
    create_topic
    
    # Give telegraf a moment to connect
    echo "Waiting for Telegraf to connect to Kafka..."
    sleep 10
    
    # Send test message
    send_test_message "$version"
    
    # Verify message was received
    if verify_message "$version"; then
        echo -e "${GREEN}✓ Kafka $version: PASSED${NC}"
        return 0
    else
        echo -e "${RED}✗ Kafka $version: FAILED${NC}"
        echo "Telegraf logs:"
        docker compose logs telegraf | tail -50
        return 1
    fi
}

# Main execution
main() {
    echo "============================================"
    echo "Kafka Version Compatibility Test Suite"
    echo "============================================"
    echo ""
    echo "Versions to test: ${KAFKA_VERSIONS[*]}"
    echo ""
    
    # Track results using parallel arrays (bash 3.x compatible)
    local passed_versions=""
    local failed_versions=""
    
    # Trap cleanup on exit
    trap cleanup EXIT
    
    for version in "${KAFKA_VERSIONS[@]}"; do
        if test_kafka_version "$version"; then
            passed_versions="$passed_versions $version"
        else
            failed_versions="$failed_versions $version"
        fi
        
        # Clean up between tests
        cleanup
        sleep 5
    done
    
    # Print summary
    echo ""
    echo "============================================"
    echo "TEST SUMMARY"
    echo "============================================"
    for version in "${KAFKA_VERSIONS[@]}"; do
        if echo "$passed_versions" | grep -q "$version"; then
            echo -e "Kafka $version: ${GREEN}PASSED${NC}"
        else
            echo -e "Kafka $version: ${RED}FAILED${NC}"
        fi
    done
    echo "============================================"
}

# Run single version test if argument provided, otherwise run all
if [ -n "$1" ]; then
    trap cleanup EXIT
    test_kafka_version "$1"
else
    main
fi
