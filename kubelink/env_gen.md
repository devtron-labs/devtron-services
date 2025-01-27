

## DEVTRON Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | APP | string |kubelink |  |  | false |
 | BUILD_NODES_BATCH_SIZE | int |2 |  |  | false |
 | CHART_WORKING_DIRECTORY | string |/home/devtron/devtroncd/charts/ |  |  | false |
 | CONSUMER_CONFIG_JSON | string | |  |  | false |
 | DEFAULT_LOG_TIME_LIMIT | int64 |1 |  |  | false |
 | ENABLE_HELM_RELEASE_CACHE | bool |true |  |  | false |
 | ENABLE_STATSVIZ | bool |false |  |  | false |
 | K8s_CLIENT_MAX_IDLE_CONNS_PER_HOST | int |25 |  |  | false |
 | K8s_TCP_IDLE_CONN_TIMEOUT | int |300 |  |  | false |
 | K8s_TCP_KEEPALIVE | int |30 |  |  | false |
 | K8s_TCP_TIMEOUT | int |30 |  |  | false |
 | K8s_TLS_HANDSHAKE_TIMEOUT | int |10 |  |  | false |
 | KUBELINK_GRPC_MAX_RECEIVE_MSG_SIZE | int |20 |  |  | false |
 | KUBELINK_GRPC_MAX_SEND_MSG_SIZE | int |4 |  |  | false |
 | LOG_LEVEL | int |-1 |  |  | false |
 | MANIFEST_FETCH_BATCH_SIZE | int |2 |  |  | false |
 | MAX_COUNT_FOR_HELM_RELEASE | int |20 |  |  | false |
 | NATS_MSG_ACK_WAIT_IN_SECS | int |120 |  |  | false |
 | NATS_MSG_BUFFER_SIZE | int |-1 |  |  | false |
 | NATS_MSG_MAX_AGE | int |86400 |  |  | false |
 | NATS_MSG_PROCESSING_BATCH_SIZE | int |1 |  |  | false |
 | NATS_MSG_REPLICAS | int |0 |  |  | false |
 | NATS_SERVER_HOST | string |nats://devtron-nats.devtroncd:4222 |  |  | false |
 | PARENT_CHILD_GVK_MAPPING | string | |  |  | false |
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
 | RUN_HELM_INSTALL_IN_ASYNC_MODE | bool |false |  |  | false |
 | STREAM_CONFIG_JSON | string | |  |  | false |
 | USE_CUSTOM_HTTP_TRANSPORT | bool |false |  |  | false |

