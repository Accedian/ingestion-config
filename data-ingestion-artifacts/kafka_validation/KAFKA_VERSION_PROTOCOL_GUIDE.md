# Telegraf kafka_version Configuration Guide

## What is `kafka_version`?

The `kafka_version` setting in Telegraf's Kafka consumer plugin tells Telegraf which **Kafka protocol version** to use when communicating with brokers. It does NOT limit which Kafka server versions you can connect to.

```toml
[[inputs.kafka_consumer]]
  kafka_version = "2.6.0"  # or "3.9.0"
```

### Key Concepts

1. **It's a client-side protocol setting** - Telegraf uses the [Sarama Go library](https://github.com/IBM/sarama) for Kafka communication
2. **Minimum protocol version** - Setting "2.6.0" means "use protocol features available up to Kafka 2.6.0"
3. **Backward compatible** - A lower version works with newer brokers; a higher version may fail with older brokers

---

## Protocol Version Comparison: 2.6.0 vs 3.9.0

### Features Available in Both (2.6.0+)

| Feature | Since Version | Description |
|---------|---------------|-------------|
| Message timestamps | 0.10.0 | Inner/outer message timestamps |
| Message headers | 0.11.0 | Key-value metadata on messages |
| Idempotent producer | 0.11.0 | Exactly-once semantics |
| Transactions | 0.11.0 | Atomic multi-partition writes |
| Consumer group improvements | 1.0.0 | Better rebalancing |
| SASL/OAUTHBEARER | 2.0.0 | OAuth authentication |
| Static membership | 2.3.0 | KIP-345 - reduced rebalances |
| Client quotas | 2.6.0 | Describe/Alter client quotas |

### Features Added After 2.6.0 (Available with 3.9.0)

| Feature | Since Version | KIP | Description |
|---------|---------------|-----|-------------|
| Flexible versions | 2.4.0 | KIP-482 | More efficient wire protocol |
| Partition reassignment | 2.4.0 | KIP-455 | AlterPartitionReassignments API |
| Leader election | 2.4.0 | KIP-460 | Forced leader election |
| Delete offsets | 2.4.0 | KIP-496 | DeleteOffsets API |
| Incremental fetch sessions | 3.0.0 | KIP-227 | Fetch performance improvements |
| KRaft (no ZooKeeper) | 3.0.0+ | KIP-500 | ZooKeeper-free cluster management |
| Tiered storage | 3.6.0 | KIP-405 | Remote storage support |
| New consumer protocol | 3.7.0+ | KIP-848 | Improved consumer rebalancing |

---

## Impact on Your Use Case

For **Telegraf consuming from Kafka**, the main differences are:

### Using `kafka_version = "2.6.0"`

**Pros:**
- ✅ Maximum compatibility - works with Kafka 2.6.0 and all newer versions
- ✅ Stable, well-tested protocol
- ✅ All essential consumer features available (headers, timestamps, consumer groups)

**Cons:**
- ❌ May not use optimal fetch session improvements
- ❌ Cannot use newer consumer group protocol features

### Using `kafka_version = "3.9.0"`

**Pros:**
- ✅ Uses latest protocol optimizations
- ✅ Better fetch session management for high-throughput scenarios
- ✅ Improved consumer group rebalancing

**Cons:**
- ❌ Will NOT work with Kafka brokers older than 3.9.0
- ❌ May cause connection issues if broker doesn't support requested API versions

---

## Practical Recommendation

### For Testing/Validation (your current use case)

Use **`kafka_version = "2.6.0"`** when testing across multiple Kafka versions:

```toml
# Broad compatibility for multi-version testing
kafka_version = "2.6.0"
```

This ensures your tests work with:
- Kafka 3.7.x
- Kafka 3.8.x
- Kafka 3.9.x
- Kafka 4.x (KRaft-only)

### For Production (known Kafka version)

Match or stay below your **minimum broker version**:

```toml
# If all brokers are Kafka 3.9.0+
kafka_version = "3.9.0"

# If you have mixed versions (3.7.x - 3.9.x)
kafka_version = "3.7.0"
```

---

## API Version Negotiation

When Telegraf connects to Kafka:

1. Client sends `ApiVersionsRequest` to broker
2. Broker responds with supported API versions for each request type
3. Client uses the highest mutually supported version

```
Client (kafka_version=3.9.0) ←→ Broker (actual=3.8.1)
     │                              │
     │  ApiVersionsRequest          │
     │─────────────────────────────>│
     │                              │
     │  ApiVersionsResponse         │
     │  (here's what I support)     │
     │<─────────────────────────────│
     │                              │
     │  Uses min(client, broker)    │
     │  for each API                │
```

If client requests a version the broker doesn't support → **Connection/API error**

---

## Supported Versions in Sarama (Telegraf's Kafka library)

From [Sarama v1.47.0](https://pkg.go.dev/github.com/IBM/sarama#pkg-variables):

```go
// Supported versions range
MinVersion     = V0_8_2_0
MaxVersion     = V4_1_1_0
DefaultVersion = V2_1_0_0

// Currently supported:
V2_6_0_0, V2_6_1_0, V2_6_2_0, V2_6_3_0,
V2_7_0_0, V2_7_1_0, V2_7_2_0,
V2_8_0_0, V2_8_1_0, V2_8_2_0,
V3_0_0_0, V3_0_1_0, V3_0_2_0,
// ... through ...
V3_9_0_0, V3_9_1_0,
V4_0_0_0, V4_0_1_0,
V4_1_0_0, V4_1_1_0
```

---

## Version Format

The version string format differs between Kafka 0.x and 1.0+:

| Kafka Version | Format | Example |
|---------------|--------|---------|
| 0.x releases | 4 digits | `"0.10.2.0"` |
| 1.0+ releases | 3 digits | `"2.6.0"`, `"3.9.0"` |

---

## Troubleshooting

### Error: "kafka: client has run out of available brokers"

Possible cause: `kafka_version` set higher than broker supports

```toml
# Try lowering the version
kafka_version = "2.6.0"  # Instead of "3.9.0"
```

### Error: "kafka: unsupported version"

The broker doesn't support the API version requested. Lower `kafka_version`.

### Message headers not appearing

Ensure `kafka_version` is at least `"0.11.0"` - headers were introduced in Kafka 0.11.

---

## Summary

| Scenario | Recommended `kafka_version` |
|----------|----------------------------|
| Multi-version testing | `"2.6.0"` |
| Production (known 3.9+) | `"3.9.0"` |
| Maximum compatibility | `"2.1.0"` (Sarama default) |
| Legacy systems (0.10+) | `"0.10.2.0"` |

**Rule of thumb**: Use the lowest version that provides the features you need for maximum broker compatibility.
