

## DEVTRON Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | APP | string |image-scanner |  |  | false |
 | CLAIR_ADDR | string |http://localhost:6060 |  |  | false |
 | CLAIR_TIMEOUT | int |30 |  |  | false |
 | CONSUMER_CONFIG_JSON | string | |  |  | false |
 | DEFAULT_LOG_TIME_LIMIT | int64 |1 |  |  | false |
 | ENABLE_STATSVIZ | bool |false |  |  | false |
 | IMAGE_SCAN_ASYNC_TIMEOUT | int |3 |  |  | false |
 | IMAGE_SCAN_TIMEOUT | int |10 |  |  | false |
 | IMAGE_SCAN_TRY_COUNT | int |1 |  |  | false |
 | JSON_OUTPUT | bool |true |  |  | false |
 | LOG_LEVEL | int |0 |  |  | false |
 | NATS_MSG_ACK_WAIT_IN_SECS | int |120 |  |  | false |
 | NATS_MSG_BUFFER_SIZE | int |-1 |  |  | false |
 | NATS_MSG_MAX_AGE | int |86400 |  |  | false |
 | NATS_MSG_PROCESSING_BATCH_SIZE | int |1 |  |  | false |
 | NATS_MSG_REPLICAS | int |0 |  |  | false |
 | NATS_SERVER_HOST | string |nats://devtron-nats.devtroncd:4222 |  |  | false |
 | PG_ADDR | string |127.0.0.1 |  |  | false |
 | PG_DATABASE | string |orchestrator |  |  | false |
 | PG_EXPORT_PROM_METRICS | bool |true |  |  | false |
 | PG_LOG_ALL_FAILURE_QUERIES | bool |true |  |  | false |
 | PG_LOG_ALL_QUERY | bool |false |  |  | false |
 | PG_LOG_SLOW_QUERY | bool |true |  |  | false |
 | PG_PASSWORD | string | |  |  | false |
 | PG_PORT | string |5432 |  |  | false |
 | PG_QUERY_DUR_THRESHOLD | int64 |5000 |  |  | false |
 | PG_USER | string | |  |  | false |
 | PROJECT_ID | string |projects/devtron-project-id |  |  | false |
 | SCANNER_TYPE | string | |  |  | false |
 | SERVER_HTTP_PORT | int |8080 |  |  | false |
 | STREAM_CONFIG_JSON | string | |  |  | false |

