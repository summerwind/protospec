package spec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(filepath.Join(wd, "testdata", "valid.yaml"))
	assert.Nil(t, err)

	_, err = Load(filepath.Join(wd, "testdata", "not-exist.yaml"))
	assert.NotNil(t, err)

	_, err = Load(filepath.Join(wd, "testdata", "error.yaml"))
	assert.NotNil(t, err)
}

func TestSpecValidate(t *testing.T) {
	tests := []struct {
		spec Spec
		err  bool
	}{
		// Valid
		{
			spec: Spec{
				Version: "2023-01-14",
				ID:      "TestSpecValidate-valid",
				Name:    "TestSpecValidate-valid",
				Endpoint: Endpoint{
					Type:      "client",
					Transport: "tcp",
				},
				Tests: []Test{
					{
						Name: "TestSpecValidate-valid-test1",
						Steps: []TestStep{
							{
								Action: "test",
							},
						},
					},
				},
			},
			err: false,
		},
		// No endpoint
		{
			spec: Spec{
				Version: "2023-01-14",
				ID:      "TestSpecValidate-no-endpoint",
				Name:    "TestSpecValidate-no-endpoint",
				Tests: []Test{
					{
						Name: "TestSpecValidate-no-endpoint-test1",
						Steps: []TestStep{
							{
								Action: "test",
							},
						},
					},
				},
			},
			err: true,
		},
		// No tests
		{
			spec: Spec{
				Version: "2023-01-14",
				ID:      "TestSpecValidate-valid",
				Name:    "TestSpecValidate-valid",
				Endpoint: Endpoint{
					Type:      "client",
					Transport: "tcp",
				},
			},
			err: true,
		},
	}

	for _, tt := range tests {
		err := tt.spec.Validate()
		if tt.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestEndpointValidate(t *testing.T) {
	tests := []struct {
		endpoint Endpoint
		err      bool
	}{
		// Valid
		{
			endpoint: Endpoint{
				Type:      "client",
				Transport: "tcp",
			},
			err: false,
		},
		// No type
		{
			endpoint: Endpoint{
				Transport: "tcp",
			},
			err: true,
		},
		// No transport
		{
			endpoint: Endpoint{
				Type: "client",
			},
			err: true,
		},
	}

	for _, tt := range tests {
		err := tt.endpoint.Validate()
		if tt.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestTestValidate(t *testing.T) {
	tests := []struct {
		test Test
		err  bool
	}{
		// Valid
		{
			test: Test{
				Name: "TestTestValidate-valid-test1",
				Steps: []TestStep{
					{
						Action: "test",
					},
				},
			},
			err: false,
		},
		// No name
		{
			test: Test{
				Steps: []TestStep{
					{
						Action: "test",
					},
				},
			},
			err: true,
		},
		// No steps
		{
			test: Test{
				Name: "TestTestValidate-no-steps-test1",
			},
			err: true,
		},
	}

	for _, tt := range tests {
		err := tt.test.Validate()
		if tt.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestTestStepValidate(t *testing.T) {
	tests := []struct {
		step TestStep
		err  bool
	}{
		// Valid
		{
			step: TestStep{
				Action: "test",
			},
			err: false,
		},
		// No action
		{
			step: TestStep{},
			err:  true,
		},
	}

	for _, tt := range tests {
		err := tt.step.Validate()
		if tt.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}
