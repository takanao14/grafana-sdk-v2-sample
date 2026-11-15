package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// docs/structure-map.html の Go と JSON の抜粋は手書きなので、コードや生成結果とずれていないかを確かめる。
const structureMapPath = "../../docs/structure-map.html"

var (
	mapMarker   = regexp.MustCompile(`\s*⟪[^⟫]+⟫\s*$`)
	mapSpace    = regexp.MustCompile(`\s+`)
	mapJSONPair = regexp.MustCompile(`"[^"]+": (?:"[^"]*"|true|false|null|-?\d+(?:\.\d+)?)`)
	mapJSONStr  = regexp.MustCompile(`"[^"]*"`)
)

// structureMapSource は HTML 内の `const name = String.raw` + "`...`;" の中身を行ごとに返す。
func structureMapSource(t *testing.T, html, name string) []string {
	t.Helper()
	start := strings.Index(html, "const "+name+" = String.raw`")
	if start < 0 {
		t.Fatalf("%s not found in %s", name, structureMapPath)
	}
	body := html[start+len("const "+name+" = String.raw`"):]
	end := strings.Index(body, "`;\n")
	if end < 0 {
		t.Fatalf("end of %s not found", name)
	}
	body = body[:end]
	// Go のバッククォート文字列を埋め込むために String.raw を区切っている箇所を元に戻す。
	body = strings.ReplaceAll(body, "` + \"`", "`")
	body = strings.ReplaceAll(body, "`\" + String.raw`", "`")
	var lines []string
	for _, l := range strings.Split(body, "\n") {
		lines = append(lines, mapMarker.ReplaceAllString(l, ""))
	}
	return lines
}

func normalizeSpace(s string) string {
	return strings.TrimSpace(mapSpace.ReplaceAllString(s, " "))
}

func TestStructureMapExcerptsMatchCode(t *testing.T) {
	raw, err := os.ReadFile(structureMapPath)
	if err != nil {
		t.Fatal(err)
	}
	html := string(raw)

	d, err := buildDnsSample()
	if err != nil {
		t.Fatal(err)
	}
	generated, err := manifestJSON(d)
	if err != nil {
		t.Fatal(err)
	}
	jsonText := string(generated)
	for _, line := range structureMapSource(t, html, "JSON_SRC") {
		for _, pair := range mapJSONPair.FindAllString(line, -1) {
			if !strings.Contains(pair, "…") && !strings.Contains(jsonText, pair) {
				t.Errorf("JSON excerpt %s is not in the generated JSON", pair)
			}
		}
		for _, str := range mapJSONStr.FindAllString(line, -1) {
			if !strings.Contains(str, "…") && !strings.Contains(jsonText, str) {
				t.Errorf("JSON excerpt %s is not in the generated JSON", str)
			}
		}
	}

	// 抜粋の行は、ファイルごとにソースと同じ順で、間を空けずに現れる必要がある。
	// "// …" や空行で区切った箇所の後だけは離れていてよい。"const (" や "})." のような短い行は
	// ソースのあちこちに現れて位置の手がかりにならないので比べない。
	const maxGap, minLen = 200, 10
	var file, text string
	pos, gapOK := 0, true
	for _, line := range structureMapSource(t, html, "GO_SRC") {
		if name, ok := strings.CutPrefix(line, "# "); ok {
			b, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			file, text, pos, gapOK = name, stripGoComments(string(b)), 0, true
			continue
		}
		if strings.Contains(line, "// …") || strings.TrimSpace(line) == "" {
			gapOK = true
			continue
		}
		l := normalizeSpace(stripGoComments(line))
		if len(l) < minLen {
			continue
		}
		i := strings.Index(text[pos:], l)
		if i < 0 || (!gapOK && i > maxGap) {
			t.Errorf("Go excerpt %q does not follow the previous line in %s", l, file)
			gapOK = true
			continue
		}
		pos += i + len(l)
		gapOK = false
	}
}

// stripGoComments は行コメントを落とし、空白をまとめる。サンプルの Go 文字列には "//" を含めない。
func stripGoComments(src string) string {
	var b strings.Builder
	for _, l := range strings.Split(src, "\n") {
		if i := strings.Index(l, "//"); i >= 0 {
			l = l[:i]
		}
		b.WriteString(l)
		b.WriteString(" ")
	}
	return normalizeSpace(b.String())
}
