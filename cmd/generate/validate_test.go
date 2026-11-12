package main

import (
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
)

func TestDashboardValidates(t *testing.T) {
	d, err := buildDnsSample()
	if err != nil {
		t.Fatal(err)
	}
	if err := validateDashboard(d); err != nil {
		t.Error(err)
	}
}

func TestValidationRejectsBrokenDashboard(t *testing.T) {
	cases := map[string]func(*dashboardv2.Dashboard){
		"empty": func(d *dashboardv2.Dashboard) { d.Elements = nil },
		"missing reference": func(d *dashboardv2.Dashboard) {
			d.Layout.RowsLayoutKind.Spec.Rows[0].Spec.Layout.AutoGridLayoutKind.Spec.Items[0].Spec.Element.Name = "missing"
		},
		"duplicate ID": func(d *dashboardv2.Dashboard) {
			d.Elements["pdns-auth-servfail-rate"].PanelKind.Spec.Id = d.Elements["resolver-servfail-rate"].PanelKind.Spec.Id
		},
		"missing thresholds": func(d *dashboardv2.Dashboard) {
			d.Elements["resolver-servfail-rate"].PanelKind.Spec.VizConfig.Spec.FieldConfig.Defaults.Thresholds = nil
		},
		"missing query reference": func(d *dashboardv2.Dashboard) {
			d.Elements["resolver-servfail-rate"].PanelKind.Spec.Data.Spec.Queries[0].Spec.RefId = ""
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			d, err := buildDnsSample()
			if err != nil {
				t.Fatal(err)
			}
			mutate(d)
			if validateDashboard(d) == nil {
				t.Fatal("invalid dashboard accepted")
			}
		})
	}
}
