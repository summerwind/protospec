package endpoint

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientConfigValidate(t *testing.T) {
	f, err := os.CreateTemp("", "protospec.yaml")
	if err != nil {
		t.Fail()
	}
	defer os.Remove(f.Name())

	c1 := ClientConfig{}
	c1.Path = f.Name()
	assert.Nil(t, c1.Validate())

	c2 := ClientConfig{}
	c2.Path = "non-exist.yaml"
	assert.NotNil(t, c2.Validate())
}
