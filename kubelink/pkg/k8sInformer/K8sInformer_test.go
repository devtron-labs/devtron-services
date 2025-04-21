package k8sInformer

import (
	"encoding/json"
	"fmt"
	util "github.com/devtron-labs/common-lib/utils"
	corev1 "k8s.io/api/core/v1"
	"os"
	"sigs.k8s.io/yaml"
	"testing"
)

const (
	// Path to a helm release YAML file
	filePath = "/Users/ashexp/Downloads/Debug/helm-release.yaml"
)

var (
	releaseJsonBytes []byte
)

func init() {
	sugaredLogger, err := util.NewSugardLogger()
	if err != nil {
		fmt.Println("failed to create logger: " + err.Error())
		return
	}
	// Create test data - a sample release object
	releaseYamlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return
	}
	// Convert to JSON
	secretJsonBytes, err := yaml.YAMLToJSON(releaseYamlBytes)
	if err != nil {
		sugaredLogger.Errorw("error in converting release object to JSON", "err", err)
		return
	}
	secretObject := corev1.Secret{}
	err = json.Unmarshal(secretJsonBytes, &secretObject)
	if err != nil {
		sugaredLogger.Errorw("error in unmarshalling release object", "err", err)
		return
	}
	releaseJsonBytes = secretObject.Data["release"]
}

func BenchmarkDecodeRelease(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			_, err := decodeRelease(string(releaseJsonBytes))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkDecodeReleaseWithJsonIterator(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			_, err := decodeReleaseWithJsonIterator(string(releaseJsonBytes))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkDecodeReleaseIntoCustomBean(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			_, err := decodeReleaseIntoCustomBean(string(releaseJsonBytes))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkDecodeReleaseIntoCustomBeanWithJsonIterator(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			_, err := decodeReleaseIntoCustomBeanWithJsonIterator(string(releaseJsonBytes))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkDecodeReleaseIntoMap(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			_, err := decodeReleaseIntoMap(string(releaseJsonBytes))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
