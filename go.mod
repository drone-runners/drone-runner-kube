module github.com/drone-runners/drone-runner-kube

go 1.12

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/bmatcuk/doublestar v1.1.1
	github.com/buildkite/yaml v2.1.0+incompatible
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/go-units v0.4.0
	github.com/drone/drone-go v1.2.1-0.20200326064413-195394da1018
	github.com/drone/envsubst v1.0.2
	github.com/drone/runner-go v1.6.1-0.20200813033918-b849bd35b2eb
	github.com/drone/signal v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/golang/mock v1.4.3
	github.com/google/go-cmp v0.4.1
	github.com/google/go-containerregistry v0.1.2
	github.com/gosimple/slug v1.9.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/golang-lru v0.5.3
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-isatty v0.0.12
	github.com/natessilva/dag v0.0.0-20180124060714-7194b8dcc5c4
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.4
)
