package sk8s

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestApplyRemoteYAMLs(t *testing.T) {
	c, err := GetClusterWithProvider(
		t, t.Context(),
		&K3sClusterProvider{
			Opts: []CustomizeClusterOption{
				WithCustomizers(
					tc.WithHostPortAccess(8080),
				),
			},
		},
	)
	require.NoError(t, err)

	http.HandleFunc("/", handler)
	go func() {
		err = http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	t.Run("valid url", func(t *testing.T) {
		err = c.ApplyRemoteYAMLs(t.Context(), []string{"http://host.testcontainers.internal:8080/configmap.yaml"})
		require.NoError(t, err)

		config, err := c.Client().CoreV1().ConfigMaps("default").Get(t.Context(), "foo", metav1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, "bar", config.Data["foo"])
	})

	t.Run("invalid url", func(t *testing.T) {
		err = c.ApplyRemoteYAMLs(t.Context(), []string{"http://host.testcontainers.internal:8080/notfound.yaml"})
		require.Error(t, err)
	})

	t.Run("invalid data", func(t *testing.T) {
		err = c.ApplyRemoteYAMLs(t.Context(), []string{"http://host.testcontainers.internal:8080/invalid.yaml"})
		require.Error(t, err)
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/configmap.yaml" {
		config := corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
			Data: map[string]string{
				"foo": "bar",
			},
		}

		out, err := yaml.Marshal(config)
		if err != nil {
			http.Error(w, "Failed to marshall resource", http.StatusInternalServerError)
		}

		fmt.Fprint(w, string(out))
		return
	}
	if r.URL.Path == "/invalid.yaml" {
		fmt.Fprint(w, "foo\nbar")
	}
	http.NotFound(w, r)
}
