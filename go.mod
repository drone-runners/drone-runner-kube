module github.com/drone-runners/drone-runner-kube

go 1.16

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/bmatcuk/doublestar v1.1.1
	github.com/buildkite/yaml v2.1.0+incompatible
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/docker/distribution v2.8.0+incompatible
	github.com/docker/go-units v0.4.0
	github.com/drone/drone-go v1.7.1
	github.com/drone/envsubst v1.0.3
	github.com/drone/runner-go v1.12.0
	github.com/drone/signal v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/golang/mock v1.4.1
	github.com/google/go-cmp v0.5.5
	github.com/gosimple/slug v1.9.0
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-isatty v0.0.8
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.21.8
	k8s.io/apimachinery v0.21.8
	k8s.io/client-go v0.21.8
)
