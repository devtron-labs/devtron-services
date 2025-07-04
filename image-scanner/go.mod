module github.com/devtron-labs/image-scanner

go 1.24.0

toolchain go1.24.3

require (
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible
	github.com/antihax/optional v1.0.0
	github.com/aws/aws-sdk-go v1.55.7
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/caarlos0/env/v6 v6.10.1
	github.com/devtron-labs/common-lib v0.19.0
	github.com/go-pg/pg v6.15.1+incompatible
	github.com/google/go-containerregistry v0.6.0
	github.com/google/wire v0.6.0
	github.com/gorilla/mux v1.8.1
	github.com/juju/errors v1.0.0
	github.com/optiopay/klar v2.4.0+incompatible
	github.com/prometheus/client_golang v1.22.0
	github.com/quay/claircore v1.3.1
	github.com/tidwall/gjson v1.14.4
	go.uber.org/zap v1.27.0
	golang.org/x/oauth2 v0.30.0
)

require (
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	github.com/arl/statsviz v0.6.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coreos/clair v2.0.1-0.20171220021131-30bd568d8361+incompatible // indirect
	github.com/docker/cli v28.1.1+incompatible // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.7.0 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nats.go v1.42.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.64.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	google.golang.org/genproto v0.0.0-20250519155744-55703ea1f237 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250519155744-55703ea1f237 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250519155744-55703ea1f237 // indirect
	google.golang.org/grpc v1.72.2 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	mellium.im/sasl v0.3.2 // indirect
)

replace github.com/devtron-labs/common-lib => github.com/devtron-labs/devtron-services/common-lib v0.0.0-20250704064738-51fefccfd963
