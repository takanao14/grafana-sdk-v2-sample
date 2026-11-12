package main

import (
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/stat"
	"github.com/grafana/grafana-foundation-sdk/go/timeseries"
)

// ダッシュボード内のパネル識別子
const (
	dnsSamplePanelResolverQPS           = 1
	dnsSamplePanelResolverServfailRate  = 2
	dnsSamplePanelPdnsAuthQPS           = 3
	dnsSamplePanelPdnsAuthServfailRate  = 4
	dnsSamplePanelResolverQueryRate     = 5
	dnsSamplePanelResolverResponseCodes = 6
	dnsSamplePanelPdnsAuthQueryRate     = 7
	dnsSamplePanelPdnsAuthResponseCodes = 8
)

// 最小構成のダッシュボードを作る
func buildDnsSample() (*dashboardv2.Dashboard, error) {
	// PromQL のラベルセレクター。環境依存なので外に出しておく。
	const (
		resolver = `job="scrapeConfig/monitoring/node-exporter-external",service="knot-resolver"`
		pdns     = `job="scrapeConfig/monitoring/pdns-auth-external"`
	)

	// しきい値は昇順に並べる。Value: nil は「下限なし」を表す。
	// フルサービスリゾルバと権威ではSERVER FAILの意味合いが違うとして、しきい値を変えている
	resolverServfailRateThresholds := dashboardv2.NewThresholdsConfigBuilder().
		Mode(dashboardv2.ThresholdsModeAbsolute).
		Steps([]dashboardv2.Threshold{
			{Value: nil, Color: "green"},
			{Value: new(0.1), Color: "yellow"},
			{Value: new(1.0), Color: "red"},
		})
	pdnsAuthServfailRateThresholds := dashboardv2.NewThresholdsConfigBuilder().
		Mode(dashboardv2.ThresholdsModeAbsolute).
		Steps([]dashboardv2.Threshold{
			{Value: nil, Color: "green"},
			{Value: new(0.001), Color: "red"},
		})

	// サマリ行のパネル定義

	resolverQPS := dashboardv2.NewPanelBuilder().
		Id(dnsSamplePanelResolverQPS).
		Title("Resolver QPS").
		Visualization(stat.NewVisualizationV2Builder().
			Unit("reqps").
			// 中立的な量なので色は1色だけを定義。しきい値を省くと80以上で赤となる
			Thresholds(dashboardv2.NewThresholdsConfigBuilder().
				Mode(dashboardv2.ThresholdsModeAbsolute).
				Steps([]dashboardv2.Threshold{
					{Value: nil, Color: "blue"},
				})).
			Decimals(1)).
		Data(queries(promQuery().
			Expr(`sum(rate(resolver_request_total{` + resolver + `}[$__rate_interval]))`).
			Instant(true).Range(false).
			LegendFormat("QPS")))

	resolverServfailRate := dashboardv2.NewPanelBuilder().
		Id(dnsSamplePanelResolverServfailRate).
		Title("Resolver SERVFAIL Rate").
		Description("Share of resolver answers that were SERVFAIL.").
		Visualization(stat.NewVisualizationV2Builder().
			Unit("percent").
			Thresholds(resolverServfailRateThresholds).
			ColorMode(common.BigValueColorModeBackground).
			Decimals(2)).
		Data(queries(promQuery().
			Expr(`100 * sum(rate(resolver_answer_rcode_servfail_total{` + resolver + `}[$__rate_interval])) / sum(rate(resolver_answer_total{` + resolver + `}[$__rate_interval]))`).
			Instant(true).Range(false).
			LegendFormat("SERVFAIL")))

	pdnsAuthQPS := dashboardv2.NewPanelBuilder().
		Id(dnsSamplePanelPdnsAuthQPS).
		Title("pdns-auth QPS").
		Description("Authoritative queries per second across both transports.").
		Visualization(stat.NewVisualizationV2Builder().
			Unit("reqps").
			// 中立的な量なので色は1色だけを定義。しきい値を省くと80以上で赤となる
			Thresholds(dashboardv2.NewThresholdsConfigBuilder().
				Mode(dashboardv2.ThresholdsModeAbsolute).
				Steps([]dashboardv2.Threshold{
					{Value: nil, Color: "blue"},
				})).
			Decimals(1)).
		Data(queries(promQuery().
			Expr(`sum(rate(pdns_auth_udp_queries{` + pdns + `}[$__rate_interval]))` +
				` + sum(rate(pdns_auth_tcp_queries{` + pdns + `}[$__rate_interval]))`).
			Instant(true).Range(false).
			LegendFormat("QPS")))

	pdnsAuthServfailRate := dashboardv2.NewPanelBuilder().
		Id(dnsSamplePanelPdnsAuthServfailRate).
		Title("pdns-auth SERVFAIL Rate").
		Description("Share of authoritative queries answered with SERVFAIL.").
		Visualization(stat.NewVisualizationV2Builder().
			Unit("percent").
			Thresholds(pdnsAuthServfailRateThresholds).
			ColorMode(common.BigValueColorModeBackground).
			Decimals(2)).
		Data(queries(promQuery().
			Expr(`100 * sum(rate(pdns_auth_servfail_packets{` + pdns + `}[$__rate_interval]))` +
				` / (sum(rate(pdns_auth_udp_queries{` + pdns + `}[$__rate_interval]))` +
				` + sum(rate(pdns_auth_tcp_queries{` + pdns + `}[$__rate_interval])))`).
			Instant(true).Range(false).
			LegendFormat("SERVFAIL")))

	// Knot Resolverの時系列パネル

	resolverQueryRate := dashboardv2.NewPanelBuilder().
		Id(dnsSamplePanelResolverQueryRate).
		Title("Resolver Query Rate").
		Visualization(timeseries.NewVisualizationV2Builder().
			Unit("reqps").
			Tooltip(defaultTooltip()).
			Legend(defaultLegend())).
		Data(queries(promQuery().
			Expr(`sum by (instance) (rate(resolver_request_total{` + resolver + `}[$__rate_interval]))`).
			LegendFormat("{{instance}}")))

	resolverResponseCodes := dashboardv2.NewPanelBuilder().
		Id(dnsSamplePanelResolverResponseCodes).
		Title("Resolver Response Codes").
		Visualization(timeseries.NewVisualizationV2Builder().
			Unit("reqps").
			Tooltip(defaultTooltip()).
			Legend(defaultLegend()).
			FillOpacity(10).
			Stacking(common.NewStackingConfigBuilder().Mode(common.StackingModeNormal))).
		Data(queries(
			promQuery().
				Expr(`sum(rate(resolver_answer_rcode_noerror_total{`+resolver+`}[$__rate_interval]))`).
				LegendFormat("NOERROR"),
			promQuery().
				Expr(`sum(rate(resolver_answer_rcode_nodata_total{`+resolver+`}[$__rate_interval]))`).
				LegendFormat("NODATA"),
			promQuery().
				Expr(`sum(rate(resolver_answer_rcode_nxdomain_total{`+resolver+`}[$__rate_interval]))`).
				LegendFormat("NXDOMAIN"),
			promQuery().
				Expr(`sum(rate(resolver_answer_rcode_servfail_total{`+resolver+`}[$__rate_interval]))`).
				LegendFormat("SERVFAIL")))

	// PowerDNSの時系列パネル

	pdnsAuthQueryRate := dashboardv2.NewPanelBuilder().
		Id(dnsSamplePanelPdnsAuthQueryRate).
		Title("pdns-auth Query Rate").
		Visualization(timeseries.NewVisualizationV2Builder().
			Unit("reqps").
			Tooltip(defaultTooltip()).
			Legend(defaultLegend())).
		Data(queries(
			promQuery().
				Expr(`rate(pdns_auth_udp_queries{`+pdns+`}[$__rate_interval])`).
				LegendFormat("{{instance}} UDP"),
			promQuery().
				Expr(`rate(pdns_auth_tcp_queries{`+pdns+`}[$__rate_interval])`).
				LegendFormat("{{instance}} TCP")))

	pdnsAuthResponseCodes := dashboardv2.NewPanelBuilder().
		Id(dnsSamplePanelPdnsAuthResponseCodes).
		Title("pdns-auth Response Codes").
		Visualization(timeseries.NewVisualizationV2Builder().
			Unit("reqps").
			Tooltip(defaultTooltip()).
			Legend(defaultLegend()).
			FillOpacity(10).
			Stacking(common.NewStackingConfigBuilder().Mode(common.StackingModeNormal)).
			// 凡例名ごとに色を固定する。
			Override(dashboardv2.MatcherConfig{Id: "byName", Options: "NOERROR"}, []dashboardv2.DynamicConfigValue{
				{Id: "color", Value: map[string]any{"mode": "fixed", "fixedColor": "green"}},
			}).
			Override(dashboardv2.MatcherConfig{Id: "byName", Options: "NXDOMAIN"}, []dashboardv2.DynamicConfigValue{
				{Id: "color", Value: map[string]any{"mode": "fixed", "fixedColor": "yellow"}},
			}).
			Override(dashboardv2.MatcherConfig{Id: "byName", Options: "SERVFAIL"}, []dashboardv2.DynamicConfigValue{
				{Id: "color", Value: map[string]any{"mode": "fixed", "fixedColor": "red"}},
			}).
			Override(dashboardv2.MatcherConfig{Id: "byName", Options: "REFUSED"}, []dashboardv2.DynamicConfigValue{
				{Id: "color", Value: map[string]any{"mode": "fixed", "fixedColor": "orange"}},
			})).
		Data(queries(
			promQuery().
				Expr(`sum(rate(pdns_auth_noerror_packets{`+pdns+`}[$__rate_interval]))`).
				LegendFormat("NOERROR"),
			promQuery().
				Expr(`sum(rate(pdns_auth_nxdomain_packets{`+pdns+`}[$__rate_interval]))`).
				LegendFormat("NXDOMAIN"),
			promQuery().
				Expr(`sum(rate(pdns_auth_servfail_packets{`+pdns+`}[$__rate_interval]))`).
				LegendFormat("SERVFAIL"),
			promQuery().
				Expr(`sum(rate(pdns_auth_refused_packets{`+pdns+`}[$__rate_interval]))`).
				LegendFormat("REFUSED")))

	// ダッシュボードを組み立てる
	d, err := dashboardv2.NewDashboardBuilder("DNS Sample").
		Tags([]string{"dns", "sample"}).
		TimeSettings(dashboardv2.NewTimeSettingsBuilder().Timezone("browser").From("now-6h").To("now").AutoRefresh("30s")).
		CursorSync(dashboardv2.DashboardCursorSyncCrosshair).
		// データソース変数を指定
		DatasourceVariable(dashboardv2.NewDatasourceVariableBuilder(datasourceVariable).
			Label("Datasource").
			PluginId("prometheus")).
		// パネルを定義
		Panel("resolver-qps", resolverQPS).
		Panel("resolver-servfail-rate", resolverServfailRate).
		Panel("pdns-auth-qps", pdnsAuthQPS).
		Panel("pdns-auth-servfail-rate", pdnsAuthServfailRate).
		Panel("resolver-query-rate", resolverQueryRate).
		Panel("resolver-response-codes", resolverResponseCodes).
		Panel("pdns-auth-query-rate", pdnsAuthQueryRate).
		Panel("pdns-auth-response-codes", pdnsAuthResponseCodes).
		// レイアウトを定義
		RowsLayout(dashboardv2.NewRowsBuilder().
			Row(dashboardv2.NewRowBuilder().
				Title("Summary").
				AutoGridLayout(dashboardv2.AutoGrid().
					MaxColumnCount(4).
					ColumnWidthMode(dashboardv2.AutoGridLayoutSpecColumnWidthModeNarrow).
					RowHeightMode(dashboardv2.AutoGridLayoutSpecRowHeightModeShort).
					WithItem("resolver-qps").
					WithItem("resolver-servfail-rate").
					WithItem("pdns-auth-qps").
					WithItem("pdns-auth-servfail-rate"))).
			Row(dashboardv2.NewRowBuilder().
				Title("Knot Resolver").
				AutoGridLayout(dashboardv2.AutoGrid().
					MaxColumnCount(1).
					ColumnWidthMode(dashboardv2.AutoGridLayoutSpecColumnWidthModeStandard).
					RowHeightMode(dashboardv2.AutoGridLayoutSpecRowHeightModeStandard).
					WithItem("resolver-query-rate").
					WithItem("resolver-response-codes"))).
			Row(dashboardv2.NewRowBuilder().
				Title("pdns-auth").
				AutoGridLayout(dashboardv2.AutoGrid().
					MaxColumnCount(1).
					ColumnWidthMode(dashboardv2.AutoGridLayoutSpecColumnWidthModeStandard).
					RowHeightMode(dashboardv2.AutoGridLayoutSpecRowHeightModeStandard).
					WithItem("pdns-auth-query-rate").
					WithItem("pdns-auth-response-codes")))).
		Build()
	if err != nil {
		return nil, err
	}
	return &d, nil
}
