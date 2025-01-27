

## DEVTRON Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | ACD_INFORMER | bool |true |  |  | false |
 | ACD_NAMESPACE | string |devtroncd |  |  | false |
 | APP | string |kubewatch |  |  | false |
 | CD_DEFAULT_NAMESPACE | string |devtron-cd |  |  | false |
 | CD_EXTERNAL_LISTENER_URL | string |http://devtroncd-orchestrator-service-prod.devtroncd:80 |  |  | false |
 | CD_EXTERNAL_NAMESPACE | string | |  |  | false |
 | CD_EXTERNAL_ORCHESTRATOR_TOKEN | string | |  |  | false |
 | CD_EXTERNAL_REST_LISTENER | bool |false |  |  | false |
 | CD_INFORMER | bool |true |  |  | false |
 | CI_INFORMER | bool |true |  |  | false |
 | CLUSTER_TYPE | string |IN_CLUSTER |  |  | false |
 | CONSUMER_CONFIG_JSON | string | |  |  | false |
 | DEFAULT_LOG_TIME_LIMIT | int64 |1 |  |  | false |
 | DEFAULT_NAMESPACE | string |devtron-ci |  |  | false |
 | ENABLE_STATSVIZ | bool |false |  |  | false |
 | K8s_CLIENT_MAX_IDLE_CONNS_PER_HOST | int |25 |  |  | false |
 | K8s_TCP_IDLE_CONN_TIMEOUT | int |300 |  |  | false |
 | K8s_TCP_KEEPALIVE | int |30 |  |  | false |
 | K8s_TCP_TIMEOUT | int |30 |  |  | false |
 | K8s_TLS_HANDSHAKE_TIMEOUT | int |10 |  |  | false |
 | LOG_LEVEL | int |-1 |  |  | false |
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
 | RUNTIME_CONFIG_LOCAL_DEV | LocalDevMode |false |  |  | false |
 | SLEEP_TIMEOUT | int |5 |  |  | false |
 | STREAM_CONFIG_JSON | string | |  |  | false |
 | USE_CUSTOM_HTTP_TRANSPORT | bool |false |  |  | false |

