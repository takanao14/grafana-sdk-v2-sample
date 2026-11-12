package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/grafana/grafana-foundation-sdk/go/resource"
)

// uid は Grafana 上のダッシュボードの識別子で、出力ファイル名にも使う。
const uid = "dns-sample"

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: generate OUTPUT_DIRECTORY")
	}
	outputDir := os.Args[1]

	d, err := buildDnsSample()
	if err != nil {
		log.Fatal(err)
	}
	if err := validateDashboard(d); err != nil {
		log.Fatal(err)
	}

	// dashboardv2ではk8s manifest形式のjsonとして出力する
	manifest, err := resource.NewManifestBuilder().
		ApiVersion("dashboard.grafana.app/v2").
		Kind("Dashboard").
		Metadata(resource.Named(uid)).
		Spec(d).
		Build()
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, uid+".json"), data, 0o644); err != nil {
		log.Fatal(err)
	}
}
