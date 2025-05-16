# IMAGESCANER CONFIGMAP


| Variable Name       | Value                                  | Description                   |
|---------------------|----------------------------------------|-------------------------------|
| CLAIR_ADDR          | clair-dcd.devtroncd:6060               | For connecting to Clair if it's enabled |
| ENABLE_PROGRESSING_SCAN_CHECK | "true"                                | Flag to enable/disable checking for progressing scans (set to "false" to disable recovery) |
| RECOVERY_BATCH_SIZE | "10"                                  | Number of scans to process in each batch during recovery |
| RECOVERY_BATCH_DELAY_SECONDS | "5"                                   | Delay between processing batches in seconds |
| RECOVERY_MAX_WORKERS | "3"                                   | Maximum number of concurrent workers for recovery |
| RECOVERY_START_DELAY_SECONDS | "10"                                  | Delay before starting recovery process after startup |
| CLIENT_ID           | client-2                               | Client ID                        |
| NATS_SERVER_HOST    | nats://devtron-nats.devtroncd:4222    | For connecting to NATS         |
| PG_LOG_QUERY        | "false"                                | PostgreSQL Query Logging (false to disable) |
| PG_ADDR             | postgresql-postgresql.devtroncd        | PostgreSQL Server Address       |
| PG_DATABASE         | orchestrator                           | PostgreSQL Database Name       |
| PG_PORT             | "5432"                                 | PostgreSQL Port Number         |
| PG_USER             | postgres                               | PostgreSQL User Name           |

