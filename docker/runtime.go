// Package docker contains Docker-related things that can be (re)used to programmatically spawn Docker containers.
package docker

import (
	"sync"

	"github.com/taxibeat/bake/docker/component"
)

// Runtime is the docker runtime, used to run and tear components down.
type Runtime struct {
	components []Component
}

// Component groups together several containers and can run in a runtime.
type Component interface {
	Start() error
	Teardown() []error
	GetContainer(name string) component.Container
}

// NewRuntime prepares a new Docker runtime.
func NewRuntime() *Runtime {
	return &Runtime{}
}

// WithComponent adds a component to the runtime.
func (r *Runtime) WithComponent(c Component) *Runtime {
	r.components = append(r.components, c)
	return r
}

// Start the components.
func (r *Runtime) Start() []error {
	chErr := make(chan error, len(r.components))

	var wg sync.WaitGroup
	wg.Add(len(r.components))
	for _, c := range r.components {
		go func(c Component) {
			err := c.Start()
			if err != nil {
				chErr <- err
			}
			wg.Done()
		}(c)
	}
	wg.Wait()
	close(chErr)

	errors := make([]error, 0)
	for err := range chErr {
		errors = append(errors, err)
	}
	return errors
}

// Teardown tears components down.
func (r *Runtime) Teardown() []error {
	chErrs := make(chan []error, len(r.components))

	var wg sync.WaitGroup
	wg.Add(len(r.components))
	for _, c := range r.components {
		go func(c Component) {
			errs := c.Teardown()
			if errs != nil {
				chErrs <- errs
			}
			wg.Done()
		}(c)
	}
	wg.Wait()
	close(chErrs)

	errors := make([]error, 0)
	for errs := range chErrs {
		errors = append(errors, errs...)
	}
	return errors
}
