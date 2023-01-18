module github.com/elastic/cloudbeat

go 1.18

require (
	code.cloudfoundry.org/go-diodes v0.0.0-20190809170250-f77fb823c7ee // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20180905210152-236a6d29298a // indirect
	github.com/akavel/rsrc v0.10.2 // indirect
	github.com/aws/aws-sdk-go-v2 v1.17.2
	github.com/dlclark/regexp2 v1.7.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/gofrs/uuid v4.3.1+incompatible
	github.com/magefile/mage v1.14.0
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/gox v1.0.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2 // indirect
	github.com/pierrre/gotestcover v0.0.0-20160517101806-924dca7d15f0
	github.com/stretchr/testify v1.8.1
	github.com/tsg/go-daemon v0.0.0-20200207173439-e704b93fd89b
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/tools v0.4.0
	gopkg.in/jcmturner/aescts.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/dnsutils.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/goidentity.v3 v3.0.0 // indirect
	gopkg.in/jcmturner/rpc.v1 v1.1.0 // indirect
	gotest.tools/gotestsum v1.7.0
	k8s.io/apimachinery v0.25.3
	k8s.io/client-go v0.25.3
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.12.21
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.17
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.24.4
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.63.1
	github.com/aws/aws-sdk-go-v2/service/ecr v1.17.18
	github.com/aws/aws-sdk-go-v2/service/ecrpublic v1.13.8
	github.com/aws/aws-sdk-go-v2/service/eks v1.22.1
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.14.8
	github.com/aws/aws-sdk-go-v2/service/iam v1.18.23
	github.com/aws/aws-sdk-go-v2/service/s3 v1.26.12
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.19
	github.com/aws/smithy-go v1.13.5
	github.com/djherbis/times v1.5.0
	github.com/elastic/e2e-testing v1.99.2-0.20220117192005-d3365c99b9c4
	github.com/elastic/elastic-agent-autodiscover v0.5.0
	github.com/elastic/elastic-agent-libs v0.2.16
	github.com/elastic/go-licenser v0.4.1
	github.com/go-git/go-git/v5 v5.4.2
	github.com/gocarina/gocsv v0.0.0-20170324095351-ffef3ffc77be
	github.com/open-policy-agent/opa v0.44.1-0.20220927105354-00e835a7cc15
	github.com/pkg/errors v0.9.1
	go.elastic.co/go-licence-detector v0.5.0
	go.uber.org/goleak v1.2.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.25.3
)

require (
	github.com/ProtonMail/go-crypto v0.0.0-20211221144345-a4f6767435ab // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.3 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.17.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.26 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.20 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.24 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.18.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.13.6 // indirect
	github.com/elastic/elastic-agent-shipper-client v0.4.0 // indirect
	github.com/emicklei/go-restful/v3 v3.8.0 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351 // indirect
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/tchap/go-patricia/v2 v2.3.1 // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)

require (
	code.cloudfoundry.org/gofileutils v0.0.0-20170111115228-4d0c80011a0f // indirect
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/bytecodealliance/wasmtime-go v1.0.0 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/containerd/containerd v1.6.12 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/docker/distribution v2.8.1+incompatible
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/elastic/beats/v7 v7.0.0-alpha2.0.20221222172013-15d9a8760574
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/cronexpr v1.1.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/iochan v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/santhosh-tekuri/jsonschema v1.2.4 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/urso/diag v0.0.0-20200210123136-21b3cc8eb797 // indirect
	github.com/xdg/stringprep v1.0.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/yashtewari/glob-intersection v0.1.0 // indirect
	go.elastic.co/fastjson v1.1.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/otel v1.7.0 // indirect
	go.opentelemetry.io/otel/trace v1.7.0 // indirect
	golang.org/x/term v0.3.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.70.1
	k8s.io/kube-openapi v0.0.0-20220803162953-67bda5d908f1 // indirect
	k8s.io/utils v0.0.0-20220728103510-ee6ede2d64ed // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

require (
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Jeffail/gabs/v2 v2.6.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/Shopify/sarama v1.27.0 // indirect
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cloudfoundry-community/go-cfclient v0.0.0-20190808214049-35bcce23fc5f // indirect
	github.com/cloudfoundry/noaa v2.1.0+incompatible // indirect
	github.com/cloudfoundry/sonde-go v0.0.0-20171206171820-b33733203bb4 // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/dgraph-io/badger/v3 v3.2103.2 // indirect
	github.com/dnephin/pflag v1.0.7 // indirect
	github.com/docker/cli v20.10.20+incompatible // indirect
	github.com/docker/docker v20.10.20+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.7.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dop251/goja v0.0.0-20221106173738-3b8a68ca89b4 // indirect
	github.com/dop251/goja_nodejs v0.0.0-20221009164102-3aa5028e57f6 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/elastic/elastic-agent-client/v7 v7.0.3
	github.com/elastic/elastic-agent-system-metrics v0.4.5-0.20220927192933-25a985b07d51 // indirect
	github.com/elastic/go-concert v0.2.0 // indirect
	github.com/elastic/go-lumber v0.1.2-0.20220819171948-335fde24ea0f // indirect
	github.com/elastic/go-seccomp-bpf v1.3.0 // indirect
	github.com/elastic/go-structform v0.0.10 // indirect
	github.com/elastic/go-sysinfo v1.9.0 // indirect
	github.com/elastic/go-ucfg v0.8.6
	github.com/elastic/go-windows v1.0.1 // indirect
	github.com/elastic/gosigar v0.14.2 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gobuffalo/here v0.6.0 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gomodule/redigo v1.8.3 // indirect
	github.com/google/flatbuffers v1.12.1 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/licenseclassifier v0.0.0-20210325184830-bb04aff29e72 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/h2non/filetype v1.1.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/nomad/api v0.0.0-20201203164818-6318a8ac7bf8 // indirect
	github.com/jcchavezs/porto v0.4.0 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/jonboulle/clockwork v0.3.0 // indirect
	github.com/josephspurrier/goversioninfo v1.4.0 // indirect
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/markbates/pkger v0.17.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/miekg/dns v1.1.50 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/shirou/gopsutil/v3 v3.21.12 // indirect
	github.com/spf13/cobra v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	github.com/ugorji/go/codec v1.1.8 // indirect
	github.com/urso/sderr v0.0.0-20210525210834-52b04e8f5c71 // indirect
	github.com/xdg/scram v1.0.3 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.elastic.co/apm v1.13.0 // indirect
	go.elastic.co/apm/module/apmelasticsearch/v2 v2.0.0 // indirect
	go.elastic.co/apm/module/apmhttp/v2 v2.2.0 // indirect
	go.elastic.co/apm/v2 v2.2.0 // indirect
	go.elastic.co/ecszap v1.0.1 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.4.0 // indirect
	golang.org/x/mod v0.7.0 // indirect
	golang.org/x/net v0.4.0 // indirect
	golang.org/x/oauth2 v0.1.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	golang.org/x/time v0.0.0-20220722155302-e5dcc9cfc0b9 // indirect
	google.golang.org/genproto v0.0.0-20221207170731-23e4bf6bdc37 // indirect
	google.golang.org/grpc v1.51.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/jcmturner/gokrb5.v7 v7.5.0 // indirect; indirectresources/manager/data.go
	howett.net/plist v1.0.0 // indirect
	oras.land/oras-go v1.2.0 // indirect
)

replace (
	github.com/Microsoft/go-winio => github.com/bi-zone/go-winio v0.4.15
	github.com/Shopify/sarama => github.com/elastic/sarama v1.19.1-0.20210823122811-11c3ef800752
	github.com/apoydence/eachers => github.com/poy/eachers v0.0.0-20181020210610-23942921fe77 //indirect, see https://github.com/elastic/beats/pull/29780 for details.
	github.com/dop251/goja => github.com/andrewkroh/goja v0.0.0-20190128172624-dd2ac4456e20
	github.com/dop251/goja_nodejs => github.com/dop251/goja_nodejs v0.0.0-20171011081505-adff31b136e6
	github.com/fsnotify/fsnotify => github.com/adriansr/fsnotify v1.4.8-0.20211018144411-a81f2b630e7c
	github.com/golang/glog => github.com/elastic/glog v1.0.1-0.20210831205241-7d8b5c89dfc4
	github.com/google/gopacket => github.com/elastic/gopacket v1.1.20-0.20211202005954-d412fca7f83a

)

// exclude github.com/elastic/elastic-agent v0.0.0-20220831162706-5f1e54f40d3e
