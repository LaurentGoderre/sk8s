package sk8s

import (
	"io"
	"regexp"
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

func TestExec(t *testing.T) {
	cluster, err := GetCluster(t, t.Context())
	assert.NoError(t, err)

	t.Run("multiplexed", func(t *testing.T) {
		exit, outReader, err := cluster.Exec(t.Context(), []string{"sh", "-c", "ls /etc/passwd /foo/bar"}, WithMultiplexed())
		assert.NoError(t, err)
		assert.Equal(t, 2, exit)

		data, err := io.ReadAll(outReader)
		assert.Equal(t, "/etc/passwd\nls: cannot access '/foo/bar': No such file or directory\n", string(data))
	})
	t.Run("not multiplexed", func(t *testing.T) {
		exit, outReader, err := cluster.Exec(t.Context(), []string{"sh", "-c", "ls /etc/passwd /foo/bar"})
		assert.NoError(t, err)
		assert.Equal(t, 2, exit)

		data, err := io.ReadAll(outReader)
		reg := regexp.MustCompile("\x02\x00\x00\x00\x00\x00\x00[0-9]?ls: [0-9]?cannot access '\\/foo\\/bar': No such file or directory\n\x01\x00\x00\x00\x00\x00\x00\f\\/etc\\/passwd\n")
		assert.Regexp(t, reg, string(data))
	})
}
