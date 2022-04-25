package env

import (
	"net/url"
	"sort"
	"strings"

	"github.com/taxibeat/bake/docker"
	"github.com/taxibeat/bake/docker/component/mongodb"
)

// ReplacementRule replaces in string to out string
type ReplacementRule interface {
	Replace(in string) (out string)
	Supports(envName, in string) bool
}

// ReplacementRuleList list of replacement rules
type ReplacementRuleList []ReplacementRule

// Replace replaces envs by list
func (l ReplacementRuleList) Replace(envs map[string]string) map[string]string {
	for envName, value := range envs {
		for _, r := range l {
			if r.Supports(envName, value) {
				envs[envName] = r.Replace(value)
			}
		}
	}
	return envs
}

// Merge two lists of rules
func (l ReplacementRuleList) Merge(extraRules []ReplacementRule) ReplacementRuleList {
	newList := l
	for _, r := range extraRules {
		newList = append(newList, r)
	}
	return newList
}

// FullReplacementRule replaces value by env name
type FullReplacementRule struct {
	envName string
	new     string
}

// NewFullReplacementRule creates full replacement rule
func NewFullReplacementRule(envName, value string) *FullReplacementRule {
	return &FullReplacementRule{envName: envName, new: value}
}

// Replace replaces old with new in the input string
func (f FullReplacementRule) Replace(_ string) (out string) {
	return f.new
}

// Supports by env name
func (f FullReplacementRule) Supports(name, _ string) bool {
	return name == f.envName
}

// SubstrReplacementRule string replacer
type SubstrReplacementRule struct {
	old string
	new string
}

// NewSubstrReplacement creates simple replacement
func NewSubstrReplacement(old, new string) *SubstrReplacementRule {
	return &SubstrReplacementRule{old: old, new: new}
}

// Replace replaces old with new in the input string
func (s SubstrReplacementRule) Replace(in string) (out string) {
	return strings.ReplaceAll(in, s.old, s.new)
}

// Supports by finding old substring
func (s SubstrReplacementRule) Supports(_, in string) bool {
	return strings.Contains(in, s.old)
}

// mongoURIReplacementRule for mongo uri
type mongoURIReplacementRule struct {
	SubstrReplacementRule
}

// newMongoURIReplacementRule creates mongo replacement rule
func newMongoURIReplacementRule(old, new string) *mongoURIReplacementRule {
	return &mongoURIReplacementRule{SubstrReplacementRule: SubstrReplacementRule{old: old, new: new}}
}

// Replace mongo uri with corresponding query params
// mongo uri must contain connect=direct query param in order to be able to connec to replica set
func (m mongoURIReplacementRule) Replace(in string) (out string) {
	u, err := url.Parse(in)
	if err != nil {
		return m.SubstrReplacementRule.Replace(in)
	}
	q := u.Query()
	q.Set("connect", "direct")
	u.RawQuery = q.Encode()
	u.Path = "/"

	return m.SubstrReplacementRule.Replace(u.String())
}

// newReplacementRulesList where key is docker related endpoint and new is corresponding localhost endpoint
func newReplacementRulesList(session *docker.Session, serviceName string) (ReplacementRuleList, error) {
	serviceNames := session.ServiceNames()
	sort.Strings(serviceNames)
	replacements := make(ReplacementRuleList, len(serviceNames))

	for i, svc := range serviceNames {
		dockerAddress, err := session.DockerToDockerServiceAddress(svc)
		if err != nil {
			return nil, err
		}
		localAddress, err := session.AutoServiceAddress(svc)
		if err != nil {
			return nil, err
		}
		switch svc {
		case serviceName:
			replacements[i] = NewFullReplacementRule("PATRON_HTTP_DEFAULT_PORT", strings.Split(localAddress, ":")[1])
		case mongodb.ServiceName:
			replacements[i] = newMongoURIReplacementRule(dockerAddress, localAddress)
		default:
			replacements[i] = NewSubstrReplacement(dockerAddress, localAddress)
		}
	}

	return replacements, nil
}
