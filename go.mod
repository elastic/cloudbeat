module github.com/elastic/cloudbeat

go 1.25.5

require (
	cloud.google.com/go/asset v1.22.0
	cloud.google.com/go/iam v1.5.3
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.13.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2 v2.3.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault v1.5.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor v0.11.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers v1.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql v1.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers v1.1.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph v0.9.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions v1.3.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity v0.14.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql v1.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage v1.8.1
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.30.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.54.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.54.0 // indirect
	github.com/aquasecurity/go-dep-parser v0.0.0-20240606050805-1de9a375c629
	github.com/aquasecurity/trivy v0.66.0
	github.com/aquasecurity/trivy-db v0.0.0-20250731052236-c7c831e2254d
	github.com/aws/aws-sdk-go-v2 v1.40.1
	github.com/aws/aws-sdk-go-v2/config v1.32.3
	github.com/aws/aws-sdk-go-v2/credentials v1.19.3
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.15
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.45.5
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.62.2
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.71.2
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.55.2
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.52.6
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.62.0
	github.com/aws/aws-sdk-go-v2/service/configservice v1.59.7
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.275.1
	github.com/aws/aws-sdk-go-v2/service/ecr v1.54.2
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.33.16
	github.com/aws/aws-sdk-go-v2/service/iam v1.53.0
	github.com/aws/aws-sdk-go-v2/service/kms v1.49.2
	github.com/aws/aws-sdk-go-v2/service/lambda v1.86.0
	github.com/aws/aws-sdk-go-v2/service/organizations v1.49.1
	github.com/aws/aws-sdk-go-v2/service/rds v1.112.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.93.0
	github.com/aws/aws-sdk-go-v2/service/s3control v1.67.0
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.67.0
	github.com/aws/aws-sdk-go-v2/service/sns v1.39.8
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.3
	github.com/aws/smithy-go v1.24.0
	github.com/dgraph-io/ristretto v1.0.0
	github.com/djherbis/times v1.6.0
	github.com/elastic/beats/v7 v7.0.0-alpha2.0.20251203075221-fdd2b2168d34
	github.com/elastic/e2e-testing v1.2.3
	github.com/elastic/elastic-agent-autodiscover v0.10.0
	github.com/elastic/elastic-agent-client/v7 v7.17.2
	github.com/elastic/elastic-agent-libs v0.26.2
	github.com/elastic/go-licenser v0.4.2
	github.com/elastic/go-ucfg v0.8.8
	github.com/go-logr/zapr v1.3.0
	github.com/go-viper/mapstructure/v2 v2.4.0
	github.com/gocarina/gocsv v0.0.0-20240520201108-78e41c74b4b1
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/googleapis/gax-go/v2 v2.15.0
	github.com/hashicorp/go-uuid v1.0.3
	github.com/huandu/xstrings v1.5.0
	github.com/magefile/mage v1.15.0
	github.com/masahiro331/go-xfs-filesystem v0.0.0-20250508093807-355a81f6ae2f
	github.com/microsoft/kiota-abstractions-go v1.9.3
	github.com/microsoftgraph/msgraph-sdk-go v1.85.0
	github.com/microsoftgraph/msgraph-sdk-go-core v1.4.0
	github.com/mikefarah/yq/v4 v4.47.2
	github.com/mitchellh/gox v1.0.1
	github.com/open-policy-agent/opa v1.7.1
	github.com/pierrre/gotestcover v0.0.0-20160517101806-924dca7d15f0
	github.com/samber/lo v1.51.0
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
	github.com/tsg/go-daemon v0.0.0-20200207173439-e704b93fd89b
	go.elastic.co/go-licence-detector v0.9.1
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/client v1.45.0 // indirect
	go.opentelemetry.io/collector/component v1.45.0 // indirect
	go.opentelemetry.io/collector/component/componenttest v0.139.0 // indirect
	go.opentelemetry.io/collector/consumer v1.45.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.139.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.139.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.45.0 // indirect
	go.opentelemetry.io/collector/pdata v1.45.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.139.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.45.0 // indirect
	go.opentelemetry.io/collector/receiver v1.45.0 // indirect
	go.opentelemetry.io/collector/receiver/receivertest v0.139.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.139.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.38.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.63.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0 // indirect
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0
	go.opentelemetry.io/otel/sdk/metric v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
	go.opentelemetry.io/proto/otlp v1.9.0
	go.uber.org/goleak v1.3.0
	go.uber.org/zap v1.27.1
	go.uber.org/zap/exp v0.3.0
	golang.org/x/lint v0.0.0-20241112194109-818c5a804067
	golang.org/x/oauth2 v0.34.0
	google.golang.org/api v0.256.0
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools/gotestsum v1.13.0
	k8s.io/api v0.34.1
	k8s.io/apimachinery v0.34.1
	k8s.io/client-go v0.34.1
	k8s.io/klog/v2 v2.130.1
	modernc.org/sqlite v1.39.0
)

require (
	cel.dev/expr v0.24.0 // indirect
	cloud.google.com/go/auth v0.17.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/monitoring v1.24.3 // indirect
	dario.cat/mergo v1.0.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/containers/azcontainerregistry v0.2.3 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4 v4.8.0 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/Intevation/gval v1.3.0 // indirect
	github.com/Intevation/jsonpath v0.2.1 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/aquasecurity/iamgo v0.0.10 // indirect
	github.com/aquasecurity/jfather v0.0.8 // indirect
	github.com/aquasecurity/trivy-checks v1.11.3-0.20250604022615-9a7efa7c9169 // indirect
	github.com/aws/aws-sdk-go v1.55.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.53.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.11.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.42.18 // indirect
	github.com/bitfield/gotestdox v0.2.2 // indirect
	github.com/bitnami/go-version v0.0.0-20251209151629-b46c9d8f1f27 // indirect
	github.com/blakesmith/ar v0.0.0-20190502131153-809d4375e1fb // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cncf/xds/go v0.0.0-20251210132809-ee656c7534f5 // indirect
	github.com/containerd/cgroups/v3 v3.0.5 // indirect
	github.com/containerd/containerd/api v1.9.0 // indirect
	github.com/containerd/containerd/v2 v2.1.5 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v1.0.0-rc.1 // indirect
	github.com/containerd/plugin v1.0.0 // indirect
	github.com/cyberphone/json-canonicalization v0.0.0-20241213102144-19d51d7fe467 // indirect
	github.com/dgraph-io/badger/v4 v4.8.0 // indirect
	github.com/dgraph-io/ristretto/v2 v2.2.0 // indirect
	github.com/digitorus/pkcs7 v0.0.0-20250730155240-ffadbf3f398c // indirect
	github.com/digitorus/timestamp v0.0.0-20250524132541-c45532741eea // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/ebitengine/purego v0.9.1 // indirect
	github.com/elastic/gokrb5/v8 v8.0.0-20251105095404-23cc45e6a102 // indirect
	github.com/elastic/sarama v1.19.1-0.20250603175145-7672917f26b6 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.35.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fvbommel/sortorder v1.1.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.5 // indirect
	github.com/go-chi/chi v4.1.2+incompatible // indirect
	github.com/go-git/go-git/v5 v5.16.4 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/go-json-experiment/json v0.0.0-20251027170946-4849db3c2f7e // indirect
	github.com/go-ldap/ldap/v3 v3.4.6 // indirect
	github.com/go-openapi/swag/cmdutils v0.25.4 // indirect
	github.com/go-openapi/swag/conv v0.25.4 // indirect
	github.com/go-openapi/swag/fileutils v0.25.4 // indirect
	github.com/go-openapi/swag/jsonname v0.25.4 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.4 // indirect
	github.com/go-openapi/swag/loading v0.25.4 // indirect
	github.com/go-openapi/swag/mangling v0.25.4 // indirect
	github.com/go-openapi/swag/netutils v0.25.4 // indirect
	github.com/go-openapi/swag/stringutils v0.25.4 // indirect
	github.com/go-openapi/swag/typeutils v0.25.4 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.4 // indirect
	github.com/gocsaf/csaf/v3 v3.3.0 // indirect
	github.com/gofrs/uuid/v5 v5.4.0 // indirect
	github.com/gohugoio/hashstructure v0.6.0 // indirect
	github.com/google/certificate-transparency-go v1.1.8 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-github/v62 v62.0.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/jcmturner/goidentity/v6 v6.0.1 // indirect
	github.com/jedisct1/go-minisign v0.0.0-20241212093149-d2f9f49435c7 // indirect
	github.com/letsencrypt/boulder v0.0.0-20240620165639-de9c06129bec // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/microsoft/kiota-authentication-azure-go v1.3.1 // indirect
	github.com/microsoft/kiota-http-go v1.5.4 // indirect
	github.com/microsoft/kiota-serialization-form-go v1.1.2 // indirect
	github.com/microsoft/kiota-serialization-json-go v1.1.2 // indirect
	github.com/microsoft/kiota-serialization-multipart-go v1.1.2 // indirect
	github.com/microsoft/kiota-serialization-text-go v1.1.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/sys/atomicwriter v0.1.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/nozzle/throttler v0.0.0-20180817012639-2ea982251481 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	github.com/openvex/discovery v0.1.1-0.20240802171711-7c54efc57553 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20250313105119-ba97887b0a25 // indirect
	github.com/rust-secure-code/go-rustaudit v0.0.0-20250226111315-e20ec32e963c // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/samber/oops v1.18.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/sassoftware/go-rpmutils v0.4.0 // indirect
	github.com/sassoftware/relic v7.2.1+incompatible // indirect
	github.com/shirou/gopsutil/v4 v4.25.11 // indirect
	github.com/sigstore/cosign/v2 v2.2.4 // indirect
	github.com/sigstore/protobuf-specs v0.5.0 // indirect
	github.com/sigstore/sigstore v1.9.5 // indirect
	github.com/sigstore/timestamp-authority v1.2.2 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spiffe/go-spiffe/v2 v2.6.0 // indirect
	github.com/std-uritemplate/std-uritemplate/go/v2 v2.0.5 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	github.com/testcontainers/testcontainers-go v0.39.0 // indirect
	github.com/theupdateframework/go-tuf v0.7.0 // indirect
	github.com/titanous/rocacheck v0.0.0-20171023193734-afe73141d399 // indirect
	github.com/tonistiigi/go-csvvalue v0.0.0-20240814133006-030d3b2625d0 // indirect
	github.com/transparency-dev/merkle v0.0.2 // indirect
	github.com/vektah/gqlparser/v2 v2.5.30 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/exp v0.0.0-20250911091902-df9299821621 // indirect
	golang.org/x/tools v0.40.0 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	kernel.org/pub/linux/libs/security/libcap/cap v1.2.76 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.76 // indirect
	mvdan.cc/sh/v3 v3.11.0 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
)

require (
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/accesscontextmanager v1.9.7 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/longrunning v0.7.0 // indirect
	cloud.google.com/go/orgpolicy v1.15.1 // indirect
	cloud.google.com/go/osconfig v1.15.1 // indirect
	cloud.google.com/go/storage v1.58.0 // indirect
	code.cloudfoundry.org/go-diodes v0.0.0-20251217145408-467f62baff7d // indirect
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20201103192249-000122071b78 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.20.0
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.2 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.6.0 // indirect
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/CycloneDX/cyclonedx-go v0.9.2 // indirect
	github.com/GoogleCloudPlatform/docker-credential-gcr v2.0.5+incompatible // indirect
	github.com/Jeffail/gabs/v2 v2.6.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/Masterminds/squirrel v1.5.4 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hcsshim v0.13.0 // indirect
	github.com/ProtonMail/go-crypto v1.3.0 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/a8m/envsubst v1.4.3 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/akavel/rsrc v0.10.2 // indirect
	github.com/alecthomas/chroma v0.10.0 // indirect
	github.com/alecthomas/participle/v2 v2.1.4 // indirect
	github.com/anchore/go-struct-converter v0.0.0-20221118182256-c68fdcfa2092 // indirect
	github.com/apparentlymart/go-cidr v1.1.0 // indirect
	github.com/aquasecurity/go-gem-version v0.0.0-20201115065557-8eed6fe000ce // indirect
	github.com/aquasecurity/go-npm-version v0.0.2 // indirect
	github.com/aquasecurity/go-pep440-version v0.0.1 // indirect
	github.com/aquasecurity/go-version v0.0.1 // indirect
	github.com/aquasecurity/table v1.11.0 // indirect
	github.com/aquasecurity/tml v0.6.1 // indirect
	github.com/aquasecurity/trivy-java-db v0.0.0-20250912094916-94bd09842933 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/ebs v1.33.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.54.3
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.11 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/bmatcuk/doublestar/v4 v4.9.1 // indirect
	github.com/briandowns/spinner v1.23.0 // indirect
	github.com/bytecodealliance/wasmtime-go/v3 v3.0.2 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/cheggaaa/pb/v3 v3.1.7 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/cloudfoundry-community/go-cfclient v0.0.0-20220930021109-9c4e6c59ccf1 // indirect
	github.com/cloudfoundry/noaa v2.1.0+incompatible // indirect
	github.com/cloudfoundry/sonde-go v0.0.0-20251217143644-d1670a435f2a // indirect
	github.com/containerd/containerd v1.7.29 // indirect
	github.com/containerd/continuity v0.4.5 // indirect
	github.com/containerd/fifo v1.1.0 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.16.3 // indirect
	github.com/containerd/ttrpc v1.2.7 // indirect
	github.com/containerd/typeurl/v2 v2.2.3 // indirect
	github.com/cyphar/filepath-securejoin v0.5.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.11.0 // indirect
	github.com/dnephin/pflag v1.0.7 // indirect
	github.com/docker/cli v28.3.3+incompatible // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker v28.4.0+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.9.3 // indirect
	github.com/docker/go-connections v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dop251/goja v0.0.0-20251201205617-2bb4c724c0f9 // indirect
	github.com/dop251/goja_nodejs v0.0.0-20171011081505-adff31b136e6 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/elastic/elastic-agent-system-metrics v0.13.4 // indirect
	github.com/elastic/go-concert v0.3.0 // indirect
	github.com/elastic/go-lumber v0.1.2-0.20220819171948-335fde24ea0f // indirect
	github.com/elastic/go-seccomp-bpf v1.5.0 // indirect
	github.com/elastic/go-structform v0.0.12 // indirect
	github.com/elastic/go-sysinfo v1.15.4 // indirect
	github.com/elastic/go-windows v1.0.2 // indirect
	github.com/elastic/gosigar v0.14.3 // indirect
	github.com/elliotchance/orderedmap v1.8.0 // indirect
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/evanphx/json-patch v5.9.11+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.4.3
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/analysis v0.24.2 // indirect
	github.com/go-openapi/errors v0.22.6 // indirect
	github.com/go-openapi/jsonpointer v0.22.4 // indirect
	github.com/go-openapi/jsonreference v0.21.4 // indirect
	github.com/go-openapi/loads v0.23.2 // indirect
	github.com/go-openapi/runtime v0.29.2 // indirect
	github.com/go-openapi/spec v0.22.3 // indirect
	github.com/go-openapi/strfmt v0.25.0 // indirect
	github.com/go-openapi/swag v0.25.4 // indirect
	github.com/go-openapi/validate v0.25.1 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/gofrs/flock v0.12.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/gomodule/redigo v1.9.2 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-containerregistry v0.20.6
	github.com/google/licenseclassifier v0.0.0-20251014224845-584483407d21 // indirect
	github.com/google/licenseclassifier/v2 v2.0.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0
	github.com/google/wire v0.7.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.7 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/h2non/filetype v1.1.1 // indirect
	github.com/hashicorp/cronexpr v1.1.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-getter v1.7.9 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-safetemp v1.0.0 // indirect
	github.com/hashicorp/go-version v1.8.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/hcl/v2 v2.24.0 // indirect
	github.com/hashicorp/nomad/api v0.0.0-20251229083939-6c0bb0a00afe // indirect
	github.com/in-toto/in-toto-golang v0.9.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	github.com/jmespath/go-jmespath v0.4.1-0.20220621161143-b0104c826a24 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/josephspurrier/goversioninfo v1.5.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/knqyf263/go-apk-version v0.0.0-20200609155635-041fdbb8563f // indirect
	github.com/knqyf263/go-deb-version v0.0.0-20241115132648-6f4aee6ccd23 // indirect
	github.com/knqyf263/go-rpm-version v0.0.0-20240918084003-2afd7dc6a38f // indirect
	github.com/knqyf263/go-rpmdb v0.1.1 // indirect
	github.com/knqyf263/nested v0.0.1 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/liamg/memoryfs v1.6.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lufia/plan9stats v0.0.0-20251013123823-9fd1530e3ec3 // indirect
	github.com/lunixbochs/struc v0.0.0-20241101090106-8d528fa2c543 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/masahiro331/go-disk v0.0.0-20240625071113-56c933208fee // indirect
	github.com/masahiro331/go-ebs-file v0.0.0-20240917043618-e6d2bea5c32e // indirect
	github.com/masahiro331/go-ext4-filesystem v0.0.0-20240620024024-ca14e6327bbd // indirect
	github.com/masahiro331/go-mvn-version v0.0.0-20250131095131-f4974fa13b8a // indirect
	github.com/masahiro331/go-vmdk-parser v0.0.0-20221225061455-612096e4bbbd // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/miekg/dns v1.1.68 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/iochan v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/buildkit v0.23.2 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/signal v0.7.1 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runtime-spec v1.2.1 // indirect
	github.com/opencontainers/selinux v1.13.1 // indirect
	github.com/openvex/go-vex v0.2.5 // indirect
	github.com/owenrumney/go-sarif/v2 v2.3.3 // indirect
	github.com/owenrumney/squealer v1.2.11 // indirect
	github.com/package-url/packageurl-go v0.1.3 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.1 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rubenv/sql-migrate v1.8.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.9.1 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/shibumi/go-pathspec v1.3.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sigstore/rekor v1.4.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/spdx/tools-golang v0.5.5 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/cobra v1.10.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tchap/go-patricia/v2 v2.3.3 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/twitchtv/twirp v8.1.3+incompatible // indirect
	github.com/ugorji/go/codec v1.1.8 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	github.com/vbatts/tar-split v0.12.1 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/yashtewari/glob-intersection v0.2.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zclconf/go-cty v1.16.4 // indirect
	github.com/zclconf/go-cty-yaml v1.1.0 // indirect
	go.elastic.co/apm/module/apmelasticsearch/v2 v2.7.1 // indirect
	go.elastic.co/apm/module/apmhttp/v2 v2.7.2 // indirect
	go.elastic.co/apm/v2 v2.7.2 // indirect
	go.elastic.co/ecszap v1.0.3 // indirect
	go.elastic.co/fastjson v1.5.1 // indirect
	go.etcd.io/bbolt v1.4.3 // indirect
	go.mongodb.org/mongo-driver v1.17.6 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/term v0.38.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	golang.org/x/time v0.14.0
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/genproto v0.0.0-20251213004720-97cd9d5aeac2 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251213004720-97cd9d5aeac2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251213004720-97cd9d5aeac2 // indirect
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.11
	gopkg.in/cheggaaa/pb.v1 v1.0.28 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/op/go-logging.v1 v1.0.0-20160211212156-b2cb9fa56473 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	helm.sh/helm/v3 v3.18.6 // indirect
	howett.net/plist v1.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.33.3 // indirect
	k8s.io/apiserver v0.33.3 // indirect
	k8s.io/cli-runtime v0.33.3 // indirect
	k8s.io/component-base v0.33.3 // indirect
	k8s.io/kube-openapi v0.0.0-20251125145642-4e65d59e963e // indirect
	k8s.io/kubectl v0.33.3 // indirect
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397 // indirect
	modernc.org/libc v1.66.8 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	oras.land/oras-go/v2 v2.6.0 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/kustomize/api v0.19.0 // indirect
	sigs.k8s.io/kustomize/kyaml v0.19.0 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

replace (
	github.com/apoydence/eachers => github.com/poy/eachers v0.0.0-20181020210610-23942921fe77 //indirect, see https://github.com/elastic/beats/pull/29780 for details.
	github.com/dop251/goja => github.com/elastic/goja v0.0.0-20190128172624-dd2ac4456e20
	github.com/eclipse/paho.mqtt.golang v1.3.5 => github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/google/gopacket => github.com/elastic/gopacket v1.1.20-0.20241002174017-e8c5fda595e6
)
