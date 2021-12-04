module github.com/martinisecurity/trusty

go 1.17

require (
	cloud.google.com/go v0.91.1
	cloud.google.com/go/storage v1.16.0
	github.com/aws/aws-sdk-go v1.40.8
	github.com/cloudflare/cfssl v1.6.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-logr/logr v0.4.0
	github.com/go-phorce/cov-report v1.1.1-0.20200622030546-3fb510c4b1ba
	github.com/go-phorce/dolly v0.8.5-0.20211126192753-f9e3e77988d9
	github.com/gogo/googleapis v1.4.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2
	github.com/googleapis/gax-go/v2 v2.0.5
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.4.0
	github.com/jinzhu/copier v0.3.2
	github.com/lib/pq v1.8.0
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/goreman v0.3.7
	github.com/mattn/goveralls v0.0.9
	github.com/miekg/dns v1.1.43
	github.com/oleiade/reflections v1.0.1
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/cors v1.8.0
	github.com/soheilhy/cmux v0.1.4
	github.com/sony/sonyflake v1.0.0
	github.com/stretchr/testify v1.7.0
	go.uber.org/config v1.4.0
	go.uber.org/dig v1.10.0
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.5
	google.golang.org/api v0.53.0
	google.golang.org/genproto v0.0.0-20210811021853-ddbe55d93216
	google.golang.org/grpc v1.40.0
	google.golang.org/grpc/examples v0.0.0-20200910201057-6591123024b3 // indirect
	google.golang.org/protobuf v1.27.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	sigs.k8s.io/controller-runtime v0.9.0
)

require (
	github.com/DataDog/datadog-go v4.8.0+incompatible // indirect
	github.com/GeertJohan/go.rice v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/daaku/go.zipexe v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-logr/zapr v0.4.0 // indirect
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/google/certificate-transparency-go v1.0.21 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmhodges/clock v0.0.0-20160418191101-880ee4c33548 // indirect
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/joho/godotenv v1.3.0 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/kisielk/sqlstruct v0.0.0-20150923205031-648daed35d49 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.7 // indirect
	github.com/mattn/go-sqlite3 v1.14.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/miekg/pkcs11 v1.0.3 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/ugorji/go/codec v1.2.6 // indirect
	github.com/weppos/publicsuffix-go v0.13.0 // indirect
	github.com/yuin/goldmark v1.4.0 // indirect
	github.com/zmap/zcrypto v0.0.0-20201128221613-3719af1573cf // indirect
	github.com/zmap/zlint/v3 v3.0.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/oauth2 v0.0.0-20210810183815-faf39c7919d5 // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/apiextensions-apiserver v0.21.1 // indirect
	k8s.io/component-base v0.21.1 // indirect
	k8s.io/klog/v2 v2.8.0 // indirect
	k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7 // indirect
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.0 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
