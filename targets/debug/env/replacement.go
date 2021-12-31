package env

import (
	"net/url"
	"strings"

	"github.com/taxibeat/bake/docker"
	"github.com/taxibeat/bake/docker/component/mongodb"
)

// ReplacementRule replaces in string to out string
type ReplacementRule interface {
	Name() string
	Replace(in string) (out string)
}

// ReplacementRuleList list of replacement rules
type ReplacementRuleList []ReplacementRule

// Replace replaces envs by list
func (list ReplacementRuleList) Replace(envs map[string]string) map[string]string {
	for name := range envs {
		for _, r := range list {
			envs[name] = r.Replace(envs[name])
		}
	}
	return envs
}

// SimpleReplacementRule string replacer
type SimpleReplacementRule struct {
	source string
	target string
}

// NewSimpleReplacement creates simple replacement
func NewSimpleReplacement(source, target string) *SimpleReplacementRule {
	return &SimpleReplacementRule{source: source, target: target}
}

// Name returns name of rule
func (s SimpleReplacementRule) Name() string {
	return s.source
}

// Replace replaces source with target in the input string
func (s SimpleReplacementRule) Replace(in string) (out string) {
	return strings.ReplaceAll(in, s.source, s.target)
}

// mongoURIReplacementRule for mongo uri
type mongoURIReplacementRule struct {
	SimpleReplacementRule
}

// Replace mongo uri with corresponding query params
// mongo uri must contain connect=direct query param in order to be able to connec to replica set
func (m mongoURIReplacementRule) Replace(in string) (out string) {
	if !strings.Contains(in, m.source) {
		return in
	}
	u, err := url.Parse(in)
	if err != nil {
		return m.SimpleReplacementRule.Replace(in)
	}
	q := u.Query()
	q.Set("connect", "direct")
	u.RawQuery = q.Encode()
	u.Path = "/"

	return m.SimpleReplacementRule.Replace(u.String())
}

// newServiceReplacementRule creates replacement for service
func newServiceReplacementRule(name, source, target string) ReplacementRule {
	replacement := NewSimpleReplacement(source, target)
	if name == mongodb.ServiceName {
		return &mongoURIReplacementRule{*replacement}
	}
	return replacement
}

// newReplacementRulesList where key is docker related endpoint and value is corresponding localhost endpoint
func newReplacementRulesList(session *docker.Session) (ReplacementRuleList, error) {
	replacements := make(ReplacementRuleList, len(session.ServiceNames()))

	for i, svc := range session.ServiceNames() {
		dockerAddress, err := session.DockerToDockerServiceAddress(svc)
		if err != nil {
			return nil, err
		}
		localAddress, err := session.AutoServiceAddress(svc)
		if err != nil {
			return nil, err
		}
		replacements[i] = newServiceReplacementRule(svc, dockerAddress, localAddress)
	}

	return replacements, nil
}
