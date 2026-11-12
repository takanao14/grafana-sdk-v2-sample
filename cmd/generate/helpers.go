package main

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/prometheus"
)

// 共有オブジェクトは状態を共有しないよう毎回新しいビルダーのインスタンスを返す

// データソース名。promQueryは先頭に"$"を付けて名前でデータソースを参照する
const datasourceVariable = "datasource"

// クエリ作成用のヘルパー
func promQuery() *prometheus.QueryV2Builder {
	return prometheus.NewQueryV2Builder().
		Datasource(dashboardv2.NewDashboardv2DataQueryKindDatasourceBuilder().Name("$" + datasourceVariable))
}

// パネルのクエリに名前を振る
func queries(qs ...cog.Builder[dashboardv2.DataQueryKind]) *dashboardv2.QueryGroupBuilder {
	group := dashboardv2.NewQueryGroupBuilder()
	for i, q := range qs {
		group.Target(dashboardv2.NewTargetBuilder().RefId(string(rune('A' + i))).Query(q))
	}
	return group
}

// デフォルトのツールチップで全系列の値を表示する
func defaultTooltip() *common.VizTooltipOptionsBuilder {
	return common.NewVizTooltipOptionsBuilder().Mode(common.TooltipDisplayModeMulti)
}

// デフォルトのLegendを表示する。Legendの表示位置は下、表示形式はリスト。
func defaultLegend() *common.VizLegendOptionsBuilder {
	return common.NewVizLegendOptionsBuilder().
		ShowLegend(true).
		DisplayMode(common.LegendDisplayModeList).
		Placement(common.LegendPlacementBottom)
}
