# Grafana Dashboards

[grafana-foundation-sdk](https://github.com/grafana/grafana-foundation-sdk) を使って
Grafana ダッシュボードを Go で定義し、その定義から dashboardv2 形式の JSON を生成する。
定義と生成物を常に一致させるための構成である。

コードから Grafana に表示されるまでの処理は [docs/walkthrough.md](docs/walkthrough.md) で順に解説している。

## ダッシュボード

| 名前 | 対象 |
|---|---|
| `dns-sample` | SDK を学ぶためのサンプル。Knot Resolver と pdns-auth の QPS・SERVFAIL 率のサマリーと、クエリレート・応答コードのグラフ |

## 構成

- `cmd/generate/main.go`: ダッシュボードの生成・検証と JSON 出力
- `cmd/generate/helpers.go`: 何度も使うクエリ・ツールチップ・凡例の部品
- `cmd/generate/validate*.go`: 生成したダッシュボードの不変条件
- `cmd/generate/*.go`: 各ダッシュボードの定義
- `dashboards/`: 生成された JSON(gitignore 済み)
- `docs/walkthrough.md`: 処理の解説
- `provisioning/`: ローカル Grafana のデータソースとファイルプロバイダー
- `docker-compose.yml`: ローカル Grafana
- `mise.toml`: Go、uv、podman-compose のバージョン
- `setup-linux.sh`: Ubuntu / Rocky 向け開発環境セットアップ

`main.go` は `buildDnsSample()` だけを呼ぶ。ダッシュボードを増やすときは、ビルダーを作り
`main.go` から呼び出す。

## ゆるいルール

SDKのビルダーはポインタを共有するため、オブジェクトは使い回さない。helpers.goの関数は、
呼び出すたびに新しいビルダーを返す。

`helpers.go` には、何度も繰り返す定型の処理だけを置く。
ヘルパーに切り出すのは、すべての呼び出し箇所を 一緒に変えるべき場合に限る。

qpsなどの中立的な量には、1色のしきい値を設定する。指定しないとデフォルトでは80以上が赤になる。

### クエリの書き方

- 取得間隔が 30 秒より長いジョブでは、`$__rate_interval` に 2 サンプルが入るように
  パネルの Min interval を大きくする。1 分間隔なら `2m` にする。

- 再起動ごとに変わる `instance` や `pod` ラベルは、join で持ち込まれたものも含めて
  集約で落として、系列が分割されないようにする。

### ダッシュボードの流れとレイアウト

ダッシュボードは、異常の検知から原因の診断へと進む順に並べる。

1. 変数で対象を絞り込む。
2. サマリーやステータスのパネルで、今起きている問題を見つける。
3. 診断用の行で、関連する推移をまとめて見せる。
4. テーブル、内訳、ログで詳細を示す。

dashboardv2 ではパネル本体と配置を分けて書く。`DashboardBuilder.Panel(name, panel)`
でパネルを要素名で登録し、`RowsLayout` の各行の AutoGrid から `WithItem(name)` で
その名前を参照する。要素名はパネルタイトルを小文字・ハイフン区切りにしたものにそろえる。

基本的にAutoGrid をつかう。

## 開発

Grafana が参照する Prometheus の URL を `.env` に書く。

```bash
PROMETHEUS_URL=http://prometheus.example:9090
```

`dns_sample.go` の `resolver` と `pdns` のセレクターは作者の環境の job 名なので、
自分の Prometheus の job 名に書き換える。そのままではグラフにデータが出ない。

Go と podman-compose は [mise](https://mise.jdx.dev/) で `mise.toml` のバージョンを入れる。
podman-compose は uv で独立した環境にインストールされる。

```bash
mise install
```

ダッシュボードを `dashboards/` に生成し、Grafana を起動する。

```bash
make dev
```

Grafana は <http://localhost:3000> で開く。`make dev` は Docker Compose か
podman-compose を自動で選ぶ。`COMPOSE=...` で上書きできる。ファイルプロバイダーは
`dashboards/` を 10 秒ごとに読み直すので、ビルダーを変えたあとは `make generate` を
実行すれば反映される。データソースや `.env` を変えたときはコンテナを再起動する。

rootless Podman では、コンテナの UID 472 がバインドマウントをたどって読めなければ
ならず、ユーザーには `/etc/subuid` の範囲が必要。確認には `setup-linux.sh` を使う。

Grafana を止める。

```bash
make dev-stop
```
