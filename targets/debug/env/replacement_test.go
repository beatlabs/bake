package env

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplacement_SimpleRule(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		envName string
		input   string
		source  string
		target  string
		output  string
	}{
		"url": {
			envName: "TEST_HTTP_URL",
			input:   "http://000-mockserver:1080",
			source:  "000-mockserver:1080",
			target:  "localhost:64952",
			output:  "http://localhost:64952",
		},
		"address": {
			envName: "TEST_KAFKA_BROKER",
			input:   "000-kafka:9092",
			source:  "000-kafka:9092",
			target:  "localhost:64952",
			output:  "localhost:64952",
		},
		"other address": {
			envName: "TEST_MONGO",
			input:   "000-kafka:9092",
			source:  "000-mongo:27017",
			target:  "localhost:64952",
			output:  "000-kafka:9092",
		},
		"other value": {
			envName: "TEST_VALUE",
			input:   "the_queue_name",
			source:  "000-mongo:27017",
			target:  "localhost:64952",
			output:  "the_queue_name",
		},
		"empty": {
			envName: "TEST_EMPTY",
			input:   "",
			source:  "000-mongo:27017",
			target:  "localhost:64952",
			output:  "",
		},
	}

	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule := NewSimpleReplacement(tt.source, tt.target)
			res := rule.Replace(tt.input)
			assert.Equal(t, tt.output, res)
			assert.Equal(t, tt.source, rule.Name())
		})
	}
}

func TestReplacement_MongoUriRule(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		envName string
		input   string
		source  string
		target  string
		output  string
	}{
		"simple string": {
			envName: "TEST_DATA",
			input:   "000-mongo:27017",
			source:  "000-mongo:27017",
			target:  "localhost:64952",
			output:  "localhost:64952",
		},
		"simple mongo uri": {
			envName: "TEST_URI",
			input:   "mongodb://root:password@000-mongo:27017",
			source:  "000-mongo:27017",
			target:  "localhost:64952",
			output:  "mongodb://root:password@localhost:64952/?connect=direct",
		},
		"simple mongo uri with ending /": {
			envName: "TEST_URI",
			input:   "mongodb://root:password@000-mongo:27017/",
			source:  "000-mongo:27017",
			target:  "localhost:64952",
			output:  "mongodb://root:password@localhost:64952/?connect=direct",
		},
		"mongo uri with query params": {
			envName: "TEST_MONGO_URI",
			input:   "mongodb://root:password@000-mongo:27017?retryWrites=true&w=majority",
			source:  "000-mongo:27017",
			target:  "localhost:64952",
			output:  "mongodb://root:password@localhost:64952/?connect=direct&retryWrites=true&w=majority",
		},
		"mongo uri with query params with ending /": {
			envName: "TEST_MONGO_URI",
			input:   "mongodb://root:password@000-mongo:27017/?retryWrites=true&w=majority",
			source:  "000-mongo:27017",
			target:  "localhost:64952",
			output:  "mongodb://root:password@localhost:64952/?connect=direct&retryWrites=true&w=majority",
		},
		"other url": {
			envName: "TEST_HTTP_URL",
			input:   "http://000-mockserver:1080",
			source:  "000-mongo:27017",
			target:  "localhost:64952",
			output:  "http://000-mockserver:1080",
		},
		"other address": {
			envName: "TEST_KAFKA_BROKER",
			input:   "000-kafka:9092",
			source:  "000-mongo:27017",
			target:  "localhost:64952",
			output:  "000-kafka:9092",
		},
	}

	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule := mongoURIReplacementRule{
				SimpleReplacementRule{source: tt.source, target: tt.target},
			}
			res := rule.Replace(tt.input)
			assert.Equal(t, tt.output, res)
			assert.Equal(t, tt.source, rule.Name())
		})
	}
}

func TestNewReplacementList(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		sessionFile string
		expList     ReplacementRuleList
		expErr      string
	}{
		"ok": {
			sessionFile: "./testdata/ok.json",
			expList: ReplacementRuleList{
				&SimpleReplacementRule{source: "000-kafka:9092", target: "localhost:64949"},
				&SimpleReplacementRule{source: "000-localstack:4566", target: "localhost:64950"},
				&SimpleReplacementRule{source: "000-mockserver:1080", target: "localhost:64953"},
				&mongoURIReplacementRule{SimpleReplacementRule{source: "000-mongo:27017", target: "localhost:64952"}},
				&SimpleReplacementRule{source: "000-test-service:8080", target: "localhost:65071"},
				&SimpleReplacementRule{source: "000-zookeeper:2181", target: "localhost:64951"},
			},
		},
		"empty": {
			sessionFile: "./testdata/empty.json",
			expList:     ReplacementRuleList{},
		},
		"missed localhost": {
			sessionFile: "./testdata/no_localhost.json",
			expErr:      `external service address not registered for "kafka"`,
		},
	}

	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			session := loadTestSessionFromFile(t, tt.sessionFile)
			replacementList, err := newReplacementRulesList(session)
			sort.SliceStable(replacementList, func(i, j int) bool {
				return replacementList[i].Name() < replacementList[j].Name()
			})

			if tt.expErr != "" {
				assert.Empty(t, replacementList)
				assert.EqualError(t, err, tt.expErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expList, replacementList)
			}
		})
	}
}
