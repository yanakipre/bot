module github.com/yanakipe/bot

go 1.22.2

require (
	github.com/gorilla/csrf v1.7.2
	github.com/gwatts/gin-adapter v0.0.0-20170508204228-c44433c485ad
	golang.org/x/crypto v0.25.0
)

replace (
	github.com/yanakipe/bot/pkg/lint/errlint => ./pkg/lint/errlint
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
	4d63.com/gocheckcompilerdirectives v1.2.1 // indirect
	4d63.com/gochecknoglobals v0.2.1 // indirect
	cloud.google.com/go/compute v1.23.3 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	dario.cat/mergo v1.0.0 // indirect
	github.com/4meepo/tagalign v1.3.4 // indirect
	github.com/Abirdcfly/dupword v0.0.14 // indirect
	github.com/Antonboom/errname v0.1.13 // indirect
	github.com/Antonboom/nilnil v0.1.9 // indirect
	github.com/Antonboom/testifylint v1.3.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/internal v0.7.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.0.0 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/ClickHouse/ch-go v0.58.2 // indirect
	github.com/Crocmagnon/fatcontext v0.2.2 // indirect
	github.com/Djarvur/go-err113 v0.0.0-20210108212216-aea10b59be24 // indirect
	github.com/GaijinEntertainment/go-exhaustruct/v3 v3.2.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Microsoft/hcsshim v0.11.4 // indirect
	github.com/OpenPeeDeeP/depguard/v2 v2.2.0 // indirect
	github.com/a8m/envsubst v1.4.2 // indirect
	github.com/alecthomas/go-check-sumtype v0.1.4 // indirect
	github.com/alecthomas/participle/v2 v2.0.0 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alexkohler/nakedret/v2 v2.0.4 // indirect
	github.com/alexkohler/prealloc v1.0.0 // indirect
	github.com/alingse/asasalint v0.0.11 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/ashanbrown/forbidigo v1.6.0 // indirect
	github.com/ashanbrown/makezero v1.1.1 // indirect
	github.com/asyncapi/converter-go v0.3.0 // indirect
	github.com/asyncapi/parser-go v0.5.0 // indirect
	github.com/asyncapi/spec-json-schemas/v4 v4.3.1 // indirect
	github.com/aws/aws-sdk-go v1.44.314 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.15 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/kms v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.24.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.9 // indirect
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bkielbasa/cyclop v1.2.1 // indirect
	github.com/blizzy78/varnamelen v0.8.0 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/bombsimon/wsl/v4 v4.2.1 // indirect
	github.com/breml/bidichk v0.2.7 // indirect
	github.com/breml/errchkjson v0.3.6 // indirect
	github.com/butuzov/ireturn v0.3.0 // indirect
	github.com/butuzov/mirror v1.2.0 // indirect
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/catenacyber/perfsprint v0.7.1 // indirect
	github.com/ccojocar/zxcvbn-go v1.0.2 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/charithe/durationcheck v0.0.10 // indirect
	github.com/chavacava/garif v0.1.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/ckaznocha/intrange v0.1.2 // indirect
	github.com/containerd/containerd v1.7.12 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/curioswitch/go-reassign v0.2.0 // indirect
	github.com/daixiang0/gci v0.13.4 // indirect
	github.com/dave/dst v0.27.0 // indirect
	github.com/dave/jennifer v1.6.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/denis-tingaikin/go-header v0.5.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/distribution/reference v0.5.0 // indirect
	github.com/dlclark/regexp2 v1.11.0 // indirect
	github.com/dnephin/pflag v1.0.7 // indirect
	github.com/elliotchance/orderedmap v1.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.10.1 // indirect
	github.com/ettle/strcase v0.2.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/firefart/nonamedreturns v1.0.5 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/fzipp/gocyclo v0.6.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/getkin/kin-openapi v0.94.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/ghostiam/protogetter v0.3.6 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-critic/go-critic v0.11.4 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/xor v1.0.0 // indirect
	github.com/go-faster/yaml v0.4.6 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/analysis v0.21.2 // indirect
	github.com/go-openapi/inflect v0.21.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/loads v0.21.1 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.0 // indirect
	github.com/go-redis/redis/v7 v7.4.1 // indirect
	github.com/go-resty/resty/v2 v2.7.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-toolsmith/astcast v1.1.0 // indirect
	github.com/go-toolsmith/astcopy v1.1.0 // indirect
	github.com/go-toolsmith/astequal v1.2.0 // indirect
	github.com/go-toolsmith/astfmt v1.1.0 // indirect
	github.com/go-toolsmith/astp v1.1.0 // indirect
	github.com/go-toolsmith/strparse v1.1.0 // indirect
	github.com/go-toolsmith/typep v1.1.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0 // indirect
	github.com/go-xmlfmt/xmlfmt v1.1.2 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/goccy/go-yaml v1.11.0 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/gddo v0.0.0-20210115222349-20d68f94ee1f // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golangci/dupl v0.0.0-20180902072040-3e9179ac440a // indirect
	github.com/golangci/gofmt v0.0.0-20231018234816-f50ced29576e // indirect
	github.com/golangci/lint-1 v0.0.0-20191013205115-297bf364a8e0 // indirect
	github.com/golangci/misspell v0.6.0 // indirect
	github.com/golangci/modinfo v0.3.4 // indirect
	github.com/golangci/plugin-module-register v0.1.1 // indirect
	github.com/golangci/revgrep v0.5.3 // indirect
	github.com/golangci/unconvert v0.0.0-20240309020433-c5143eacb3ed // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20240424215950-a892ee059fd6 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/wire v0.5.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/gordonklaus/ineffassign v0.1.0 // indirect
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/gostaticanalysis/analysisutil v0.7.1 // indirect
	github.com/gostaticanalysis/comment v1.4.2 // indirect
	github.com/gostaticanalysis/forcetypeassert v0.1.0 // indirect
	github.com/gostaticanalysis/nilerr v0.1.1 // indirect
	github.com/gotd/ige v0.2.2 // indirect
	github.com/gotd/neo v0.1.5 // indirect
	github.com/gotd/td v0.108.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hexops/gotextdiff v1.0.3 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jellydator/ttlcache/v3 v3.3.0 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/jgautheron/goconst v1.7.1 // indirect
	github.com/jingyugao/rowserrcheck v1.1.1 // indirect
	github.com/jinzhu/copier v0.3.5 // indirect
	github.com/jirfag/go-printf-func-name v0.0.0-20200119135958-7558a9eaa5af // indirect
	github.com/jjti/go-spancheck v0.6.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/julz/importas v0.1.0 // indirect
	github.com/karamaru-alpha/copyloopvar v1.1.0 // indirect
	github.com/kisielk/errcheck v1.7.0 // indirect
	github.com/kkHAIKE/contextcheck v1.1.5 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/kulti/thelper v0.6.3 // indirect
	github.com/kunwardeep/paralleltest v1.0.10 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/kyoh86/exportloopref v0.1.11 // indirect
	github.com/labstack/echo/v4 v4.9.0 // indirect
	github.com/labstack/gommon v0.3.1 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lasiar/canonicalheader v1.1.1 // indirect
	github.com/ldez/gomoddirectives v0.2.4 // indirect
	github.com/ldez/tagliatelle v0.5.0 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/leonklingele/grouper v1.1.2 // indirect
	github.com/lufeee/execinquery v1.2.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/macabu/inamedparam v0.1.3 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/maratori/testableexamples v1.0.0 // indirect
	github.com/maratori/testpackage v1.1.1 // indirect
	github.com/matoous/godox v0.0.0-20230222163458-006bad1f9d26 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mgechev/revive v1.3.7 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/moricho/tparallel v0.3.1 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nakabonne/nestif v0.3.1 // indirect
	github.com/nishanths/exhaustive v0.12.0 // indirect
	github.com/nishanths/predeclared v0.2.2 // indirect
	github.com/nunnatsa/ginkgolinter v0.16.2 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc5 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/paulmach/orb v0.10.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pgvector/pgvector-go v0.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/polyfloyd/go-errorlint v1.5.2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/quasilyte/go-ruleguard v0.4.2 // indirect
	github.com/quasilyte/go-ruleguard/dsl v0.3.22 // indirect
	github.com/quasilyte/gogrep v0.5.0 // indirect
	github.com/quasilyte/regex/syntax v0.0.0-20210819130434-b3f0c404a727 // indirect
	github.com/quasilyte/stdinfo v0.0.0-20220114132959-f7386bf02567 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/ryancurrah/gomodguard v1.3.2 // indirect
	github.com/ryanrolds/sqlclosecheck v0.5.1 // indirect
	github.com/sanposhiho/wastedassign/v2 v2.0.7 // indirect
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 // indirect
	github.com/sashabaranov/go-openai v1.26.1 // indirect
	github.com/sashamelentyev/interfacebloat v1.1.0 // indirect
	github.com/sashamelentyev/usestdlibvars v1.26.0 // indirect
	github.com/securego/gosec/v2 v2.20.1-0.20240525090044-5f0084eb01a9 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/segmentio/backo-go v1.0.0 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/shazow/go-diff v0.0.0-20160112020656-b6b7b6733b8c // indirect
	github.com/shirou/gopsutil/v3 v3.24.5 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sivchari/containedctx v1.0.3 // indirect
	github.com/sivchari/tenv v1.7.1 // indirect
	github.com/sonatard/noctx v0.0.2 // indirect
	github.com/sourcegraph/go-diff v0.7.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.13.0 // indirect
	github.com/ssgreg/nlreturn/v2 v2.2.1 // indirect
	github.com/stbenjam/no-sprintf-host-port v0.1.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.4.1 // indirect
	github.com/t-yuki/gocover-cobertura v0.0.0-20180217150009-aaee18c8195c // indirect
	github.com/tdakkota/asciicheck v0.2.0 // indirect
	github.com/tetafro/godot v1.4.16 // indirect
	github.com/tidwall/gjson v1.17.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/timakin/bodyclose v0.0.0-20230421092635-574207250966 // indirect
	github.com/timonwong/loggercheck v0.9.4 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/tomarrell/wrapcheck/v2 v2.8.3 // indirect
	github.com/tommy-muehle/go-mnd/v2 v2.5.1 // indirect
	github.com/toqueteos/webbrowser v1.2.0 // indirect
	github.com/tucnak/telebot v2.0.0+incompatible // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/tychoish/fun v0.8.5 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	github.com/ultraware/funlen v0.1.0 // indirect
	github.com/ultraware/whitespace v0.1.1 // indirect
	github.com/uudashr/gocognit v1.1.2 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	github.com/vaughan0/go-ini v0.0.0-20130923145212-a98ad7ee00ec // indirect
	github.com/x-cray/logrus-prefixed-formatter v0.5.2 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xen0n/gosmopolitan v1.2.2 // indirect
	github.com/yagipy/maintidx v1.0.0 // indirect
	github.com/yeya24/promlinter v0.3.0 // indirect
	github.com/ykadowak/zerologlint v0.1.5 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	gitlab.com/bosi/decorder v0.4.2 // indirect
	go-simpler.org/musttag v0.12.2 // indirect
	go-simpler.org/sloglint v0.7.1 // indirect
	go.mongodb.org/mongo-driver v1.12.1 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/exp/typeparams v0.0.0-20240314144324-c7f7c6466f7f // indirect
	golang.org/x/mod v0.19.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/term v0.22.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231120223509-83a465c0220f // indirect
	google.golang.org/grpc v1.59.0 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/op/go-logging.v1 v1.0.0-20160211212156-b2cb9fa56473 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/telebot.v3 v3.3.8 // indirect
	honnef.co/go/tools v0.4.7 // indirect
	k8s.io/apiextensions-apiserver v0.26.15 // indirect
	k8s.io/component-base v0.26.15 // indirect
	k8s.io/kube-openapi v0.0.0-20221012153701-172d655c2280 // indirect
	mvdan.cc/gofumpt v0.6.0 // indirect
	mvdan.cc/unparam v0.0.0-20240528143540-8a5130ca722f // indirect
	nhooyr.io/websocket v1.8.11 // indirect
	rsc.io/qr v0.2.0 // indirect
	sigs.k8s.io/controller-runtime v0.14.5 // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.7.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.3.0
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/azkeys v0.10.0
	github.com/ClickHouse/clickhouse-go/v2 v2.14.1
	github.com/Daniel-ef/go-hubspot v0.1.3
	github.com/Masterminds/squirrel v1.5.2
	github.com/Nerzal/gocloak/v13 v13.8.0
	github.com/alecthomas/units v0.0.0-20231202071711-9a357b53e9c9
	github.com/atrox/haikunatorgo v2.0.0+incompatible
	github.com/aws/aws-sdk-go-v2 v1.27.0
	github.com/aws/aws-sdk-go-v2/config v1.27.15
	github.com/aws/aws-sdk-go-v2/service/s3 v1.54.2
	github.com/aws/aws-sdk-go-v2/service/sqs v1.32.2
	github.com/brianvoe/gofakeit/v6 v6.11.0
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/deepmap/oapi-codegen v1.10.1
	github.com/docker/docker v25.0.5+incompatible
	github.com/docker/go-connections v0.5.0
	github.com/docker/go-units v0.5.0
	github.com/dranikpg/gtrs v0.6.1
	github.com/dustin/go-humanize v1.0.1
	github.com/evanphx/json-patch/v5 v5.6.0
	github.com/fatih/structs v1.1.0
	github.com/getsentry/sentry-go v0.16.0
	github.com/gin-contrib/static v0.0.1
	github.com/gin-gonic/gin v1.9.1
	github.com/go-chi/chi/v5 v5.0.7
	github.com/go-faster/errors v0.7.1
	github.com/go-faster/jx v1.1.0
	github.com/go-faster/yamlx v0.3.2
	github.com/go-logr/logr v1.4.2
	github.com/go-openapi/errors v0.20.2
	github.com/go-openapi/runtime v0.23.1
	github.com/go-openapi/strfmt v0.21.2
	github.com/go-openapi/swag v0.21.1
	github.com/go-openapi/validate v0.21.0
	github.com/go-swagger/go-swagger v0.28.0
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/golang/mock v1.6.0
	github.com/golangci/golangci-lint v1.59.1
	github.com/google/go-cmp v0.6.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v61 v61.0.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/sessions v1.2.1
	github.com/gorilla/websocket v1.4.2
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/hashicorp/golang-lru/v2 v2.0.1
	github.com/heetch/confita v0.10.0
	github.com/jackc/pgconn v1.14.3
	github.com/jackc/pgerrcode v0.0.0-20240316143900-6e2875d9b438
	github.com/jackc/pgx/v4 v4.18.3
	github.com/jackc/pgx/v5 v5.5.5
	github.com/jackc/tern/v2 v2.2.1-0.20240605100943-963db3b9b302
	github.com/jefflinse/githubsecret v1.0.2
	github.com/jmattheis/goverter v0.17.1
	github.com/jmoiron/sqlx v1.3.5
	github.com/kamilsk/retry/v5 v5.0.0-rc8
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lerenn/asyncapi-codegen v0.31.0
	github.com/lib/pq v1.10.9
	github.com/lithammer/shortuuid/v4 v4.0.0
	github.com/lpar/gzipped v1.1.0
	github.com/mikefarah/yq/v4 v4.34.2
	github.com/mitchellh/hashstructure/v2 v2.0.2
	github.com/mitchellh/mapstructure v1.5.0
	github.com/mroth/weightedrand/v2 v2.0.1
	github.com/neondatabase/autoscaling v0.31.0
	github.com/yanakipe/bot/pkg/lint/errlint v0.0.0-00010101000000-000000000000
	github.com/nukosuke/go-zendesk v0.15.0
	github.com/ogen-go/ogen v1.1.0
	github.com/onsi/ginkgo/v2 v2.17.3
	github.com/orbcorp/orb-go v0.32.0
	github.com/orlangure/gnomock v0.25.0
	github.com/ory/hydra-client-go v1.11.8
	github.com/pganalyze/pg_query_go/v4 v4.2.3
	github.com/posthog/posthog-go v0.0.0-20240327112532-87b23fe11103
	github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/common v0.37.0
	github.com/redis/go-redis/v9 v9.0.5
	github.com/rekby/fixenv v0.3.1
	github.com/reugn/go-quartz v0.7.1-0.20230722144436-0a4c5734c373
	github.com/rs/cors v1.8.2
	github.com/rsniezynski/go-asset-helper v0.0.0-20150405181857-38e753e5e853
	github.com/samber/lo v1.39.0
	github.com/segmentio/analytics-go/v3 v3.2.1
	github.com/segmentio/golines v0.11.0
	github.com/sethvargo/go-password v0.2.0
	github.com/shopspring/decimal v1.4.0
	github.com/sourcegraph/conc v0.3.0
	github.com/spf13/cobra v1.8.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.9.0
	github.com/stripe/stripe-go/v74 v74.6.0
	github.com/sugerio/suger-sdk-go v0.0.0-20231023160603-f99a259fec07
	github.com/testcontainers/testcontainers-go v0.28.0
	github.com/testcontainers/testcontainers-go/modules/clickhouse v0.25.0
	github.com/testcontainers/testcontainers-go/modules/postgres v0.28.0
	github.com/wadey/gocovmerge v0.0.0-20160331181800-b5bfa59ec0ad
	github.com/wagslane/go-password-validator v0.3.0
	github.com/xdg-go/scram v1.1.2
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.44.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.45.0
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.19.0
	go.opentelemetry.io/otel/metric v1.28.0
	go.opentelemetry.io/otel/sdk v1.25.0
	go.opentelemetry.io/otel/trace v1.28.0
	go.uber.org/automaxprocs v1.5.3
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	gocloud.dev v0.34.0
	golang.org/x/exp v0.0.0-20240103183307-be819d1f06fc
	golang.org/x/net v0.27.0
	golang.org/x/oauth2 v0.15.0
	golang.org/x/sync v0.8.0
	golang.org/x/text v0.16.0
	golang.org/x/time v0.5.0
	golang.org/x/tools v0.23.0
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
	google.golang.org/api v0.152.0
	google.golang.org/protobuf v1.33.0
	gopkg.in/dnaeon/go-vcr.v3 v3.1.2
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools/gotestsum v1.8.0
	k8s.io/api v0.29.1
	k8s.io/apimachinery v0.29.1
	k8s.io/client-go v0.26.15
	k8s.io/klog/v2 v2.90.1
	k8s.io/utils v0.0.0-20230220204549-a5ecb0141aa5
	moul.io/zapfilter v1.7.0
	sigs.k8s.io/yaml v1.3.0
)
