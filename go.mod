module github.com/yanakipre/bot

go 1.23.3

replace (
	k8s.io/api => k8s.io/api v0.26.15
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.26.15
	k8s.io/apimachinery => k8s.io/apimachinery v0.26.15
	k8s.io/apiserver => k8s.io/apiserver v0.26.15
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.26.15
	k8s.io/client-go => k8s.io/client-go v0.26.15
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.26.15
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.26.15
	k8s.io/code-generator => k8s.io/code-generator v0.26.15
	k8s.io/component-base => k8s.io/component-base v0.26.15
	k8s.io/component-helpers => k8s.io/component-helpers v0.26.15
	k8s.io/controller-manager => k8s.io/controller-manager v0.26.15
	k8s.io/cri-api => k8s.io/cri-api v0.26.15
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.26.15
	k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.26.15
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.26.15
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.26.15
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.26.15
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.26.15
	k8s.io/kubectl => k8s.io/kubectl v0.26.15
	k8s.io/kubelet => k8s.io/kubelet v0.26.15
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.26.15
	k8s.io/metrics => k8s.io/metrics v0.26.15
	k8s.io/mount-utils => k8s.io/mount-utils v0.26.15
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.26.15
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.26.15
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.26.15
	k8s.io/sample-controller => k8s.io/sample-controller v0.26.15
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/ClickHouse/ch-go v0.58.2 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Microsoft/hcsshim v0.11.4 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.15 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.24.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.9 // indirect
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coder/websocket v1.8.13 // indirect
	github.com/containerd/containerd v1.7.12 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/distribution/reference v0.5.0 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/xor v1.0.0 // indirect
	github.com/go-faster/yaml v0.4.6 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-redis/redis/v7 v7.4.1 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/gotd/ige v0.2.2 // indirect
	github.com/gotd/neo v0.1.5 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/onsi/gomega v1.33.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc5 // indirect
	github.com/paulmach/orb v0.10.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/samber/slog-common v0.18.1 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shirou/gopsutil/v3 v3.24.5 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/exp v0.0.0-20240103183307-be819d1f06fc // indirect
	golang.org/x/mod v0.26.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/term v0.33.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231120223509-83a465c0220f // indirect
	google.golang.org/grpc v1.59.0 // indirect
	rsc.io/qr v0.2.0 // indirect
)

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.14.1
	github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs v1.0.0
	github.com/atrox/haikunatorgo v2.0.0+incompatible
	github.com/aws/aws-sdk-go-v2 v1.27.0
	github.com/aws/aws-sdk-go-v2/config v1.27.15
	github.com/brianvoe/gofakeit/v6 v6.11.0
	github.com/docker/docker v25.0.5+incompatible
	github.com/docker/go-connections v0.5.0
	github.com/docker/go-units v0.5.0
	github.com/dranikpg/gtrs v0.6.1
	github.com/getsentry/sentry-go v0.31.1
	github.com/go-faster/errors v0.7.1
	github.com/go-faster/jx v1.1.0
	github.com/google/go-cmp v0.7.0
	github.com/google/uuid v1.6.0
	github.com/gotd/contrib v0.21.0
	github.com/gotd/td v0.122.0
	github.com/gotd/td/examples v0.0.0-20250801131019-be2dd8a56f56
	github.com/heetch/confita v0.10.0
	github.com/jackc/pgerrcode v0.0.0-20240316143900-6e2875d9b438
	github.com/jackc/pgx/v4 v4.18.3
	github.com/jackc/pgx/v5 v5.5.5
	github.com/jackc/tern/v2 v2.2.1-0.20240605100943-963db3b9b302
	github.com/jellydator/ttlcache/v3 v3.3.0
	github.com/jmoiron/sqlx v1.3.5
	github.com/kamilsk/retry/v5 v5.0.0-rc8
	github.com/lithammer/shortuuid/v4 v4.0.0
	github.com/mitchellh/hashstructure/v2 v2.0.2
	github.com/ogen-go/ogen v1.14.0
	github.com/orlangure/gnomock v0.25.0
	github.com/pgvector/pgvector-go v0.2.0
	github.com/prometheus/client_golang v1.20.5
	github.com/redis/go-redis/v9 v9.0.5
	github.com/rekby/fixenv v0.7.0
	github.com/reugn/go-quartz v0.13.0
	github.com/samber/lo v1.47.0
	github.com/samber/slog-zap/v2 v2.6.2
	github.com/sashabaranov/go-openai v1.26.1
	github.com/shopspring/decimal v1.4.0
	github.com/sourcegraph/conc v0.3.0
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.10.0
	github.com/testcontainers/testcontainers-go v0.28.0
	github.com/testcontainers/testcontainers-go/modules/clickhouse v0.25.0
	github.com/testcontainers/testcontainers-go/modules/postgres v0.28.0
	github.com/tucnak/telebot v2.0.0+incompatible
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.45.0
	go.opentelemetry.io/otel v1.37.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.19.0
	go.opentelemetry.io/otel/sdk v1.36.0
	go.opentelemetry.io/otel/trace v1.37.0
	go.uber.org/zap v1.27.0
	golang.org/x/net v0.42.0
	golang.org/x/sync v0.16.0
	golang.org/x/time v0.8.0
	golang.org/x/tools v0.35.0
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
	google.golang.org/protobuf v1.34.2
	gopkg.in/dnaeon/go-vcr.v3 v3.1.2
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	moul.io/zapfilter v1.7.0
	sigs.k8s.io/yaml v1.3.0
)
