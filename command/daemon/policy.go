package daemon

import (
	"bytes"
	"io"

	"github.com/buildkite/yaml"
)

const Kind = "Policy"

type Policy struct {
	Kind string
	Name string

	Metadata       Metadata          `json:"metadata,omitempty"`
	ServiceAccount string            `json:"service_account,omitempty" envconfig:"DRONE_SERVICE_ACCOUNT_DEFAULT"`
	Images         Images            `json:"images,omitempty"`
	Resources      Resources         `json:"resources,omitempty"`
	NodeSelector   map[string]string `json:"node_selector,omitempty" envconfig:"DRONE_NODE_SELECTOR_DEFAULT"`
	Tolerations    []Toleration      `json:"tolerations,omitempty"`

	Match []string `json:"match,omitempty" default:"**"`
}

type Metadata struct {
	Namespace   string            `json:"namespace,omitempty" envconfig:"DRONE_NAMESPACE_DEFAULT" default:"default"`
	Labels      map[string]string `json:"labels,omitempty" envconfig:"DRONE_LABELS_DEFAULT"`
	Annotations map[string]string `json:"annotations,omitempty" envconfig:"DRONE_ANNOTATIONS_DEFAULT"`
}

type Resources struct {
	LimitCPU      int64     `envconfig:"DRONE_RESOURCE_LIMIT_CPU"`
	LimitMemory   BytesSize `envconfig:"DRONE_RESOURCE_LIMIT_MEMORY"`
	RequestCPU    int64     `envconfig:"DRONE_RESOURCE_REQUEST_CPU"`
	RequestMemory BytesSize `envconfig:"DRONE_RESOURCE_REQUEST_MEMORY"`
}

type Toleration struct {
	Effect            string `json:"effect,omitempty"`
	Key               string `json:"key,omitempty"`
	Operator          string `json:"operator,omitempty"`
	TolerationSeconds *int   `json:"toleration_seconds,omitempty"`
	Value             string `json:"value,omitempty"`
}

type Images struct {
	Clone       string `envconfig:"DRONE_IMAGE_CLONE"`
	Placeholder string `envconfig:"DRONE_IMAGE_PLACEHOLDER"`
}

func parsePolicies(b []byte) ([]Policy, error) {
	buf := bytes.NewBuffer(b)
	res := []Policy{}
	dec := yaml.NewDecoder(buf)
	for {
		out := new(Policy)
		err := dec.Decode(out)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if out.Kind == Kind {
			res = append(res, *out)
		}
	}
	return res, nil
}
