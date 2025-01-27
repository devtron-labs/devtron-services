

## DEVTRON Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | ANALYTICS_DEBUG | bool |false |  |  | false |
 | APP | string |git-sensor |  |  | false |
 | CLI_CMD_TIMEOUT_GLOBAL_SECONDS | int |900 |  |  | false |
 | CLI_CMD_TIMEOUT_JSON | string | |  |  | false |
 | COMMIT_STATS_TIMEOUT_IN_SEC | int |2 |  |  | false |
 | CONSUMER_CONFIG_JSON | string | |  |  | false |
 | DEFAULT_LOG_TIME_LIMIT | int64 |1 |  |  | false |
 | ENABLE_FILE_STATS | bool |false |  |  | false |
 | ENABLE_STATSVIZ | bool |false |  |  | false |
 | GIT_HISTORY_COUNT | int |15 |  |  | false |
 | GOGIT_TIMEOUT_SECONDS | int |10 |  |  | false |
 | LOG_LEVEL | int |0 |  |  | false |
 | MIN_LIMIT_FOR_PVC | int |1 |  |  | false |
 | NATS_MSG_ACK_WAIT_IN_SECS | int |120 |  |  | false |
 | NATS_MSG_BUFFER_SIZE | int |-1 |  |  | false |
 | NATS_MSG_MAX_AGE | int |86400 |  |  | false |
 | NATS_MSG_PROCESSING_BATCH_SIZE | int |1 |  |  | false |
 | NATS_MSG_REPLICAS | int |0 |  |  | false |
 | NATS_SERVER_HOST | string |nats://devtron-nats.devtroncd:4222 |  |  | false |
 | PG_ADDR | string |127.0.0.1 |  |  | false |
 | PG_DATABASE | string |git_sensor |  |  | false |
 | PG_EXPORT_PROM_METRICS | bool |true |  |  | false |
 | PG_LOG_ALL_FAILURE_QUERIES | bool |true |  |  | false |
 | PG_LOG_ALL_QUERY | bool |false |  |  | false |
 | PG_LOG_SLOW_QUERY | bool |true |  |  | false |
 | PG_PASSWORD | string | |  |  | false |
 | PG_PORT | string |5432 |  |  | false |
 | PG_QUERY_DUR_THRESHOLD | int64 |5000 |  |  | false |
 | PG_USER | string | |  |  | false |
 | POLL_DURATION | int |2 |  |  | false |
 | POLL_WORKER | int |5 |  |  | false |
 | SERVER_GRPC_PORT | int |8081 |  |  | false |
 | SERVER_REST_PORT | int |8080 |  |  | false |
 | STREAM_CONFIG_JSON | string | |  |  | false |
 | USE_GIT_CLI | bool |false |  |  | false |
 | USE_GIT_CLI_ANALYTICS | bool |false |  |  | false |

