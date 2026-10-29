package sk8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithImage(t *testing.T) {
	opts := ClusterOptions{image: defaultImage}

	err := WithImage("rancher/k3s:v1.33.5-k3s1").Customize(&opts)
	assert.NoError(t, err)
	assert.Equal(t, "rancher/k3s:v1.33.5-k3s1", opts.image)

	err = WithImage("Not A Valid Image!!").Customize(&opts)
	assert.Error(t, err)
	assert.Equal(t, "rancher/k3s:v1.33.5-k3s1", opts.image, "invalid image must not overwrite the previous value")
}

func TestOutputVolumeName(t *testing.T) {
	assert.Equal(t, getOutputVolumeName("abc/def%blah😀"), "sk8s-out-abc_def_blah_")
}
