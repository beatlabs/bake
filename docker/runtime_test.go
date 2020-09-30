package docker

import (
	"errors"
	"testing"

	"github.com/taxibeat/bake/docker/component"

	"github.com/stretchr/testify/assert"
)

type mockComponent struct {
	startErr     error
	teardownErrs []error
}

func (c *mockComponent) Start() error {
	return c.startErr
}

func (c *mockComponent) Teardown() []error {
	return c.teardownErrs
}

func (c *mockComponent) GetContainer(name string) component.Container {
	return nil
}

func Test_Runtime_Start(t *testing.T) {
	errs := []error{
		errors.New("foo"),
		errors.New("bar"),
	}
	testCases := map[string]struct {
		components   []Component
		expectedErrs []error
	}{
		"no components": {},
		"all components successful": {
			components: []Component{&mockComponent{}, &mockComponent{}, &mockComponent{}},
		},
		"several components fail": {
			components: []Component{
				&mockComponent{},
				&mockComponent{
					startErr: errs[0],
				},
				&mockComponent{
					startErr: errs[1],
				},
				&mockComponent{},
			},
			expectedErrs: []error{errs[0], errs[1]},
		},
	}
	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			r := NewRuntime()
			for _, c := range tt.components {
				r.WithComponent(c)
			}
			errs := r.Start()
			assert.ElementsMatch(t, tt.expectedErrs, errs)
		})
	}
}

func Test_Runtime_Teardown(t *testing.T) {
	errs := []error{
		errors.New("foo"),
		errors.New("bar"),
		errors.New("baz"),
	}
	testCases := map[string]struct {
		components   []Component
		expectedErrs []error
	}{
		"no components": {},
		"all components successful": {
			components: []Component{&mockComponent{}, &mockComponent{}, &mockComponent{}},
		},
		"several components fail": {
			components: []Component{
				&mockComponent{},
				&mockComponent{
					teardownErrs: []error{errs[0], errs[1]},
				},
				&mockComponent{
					teardownErrs: []error{errs[2]},
				},
				&mockComponent{},
			},
			expectedErrs: []error{errs[0], errs[1], errs[2]},
		},
	}
	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			r := NewRuntime()
			for _, c := range tt.components {
				r.WithComponent(c)
			}
			errs := r.Teardown()
			assert.ElementsMatch(t, tt.expectedErrs, errs)
		})
	}
}
