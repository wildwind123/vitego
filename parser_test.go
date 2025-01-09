package manifestparser

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestManifest(t *testing.T) {
	b, err := os.ReadFile("test_data/manifest.json")
	if err != nil {
		t.Error(err)
		return
	}

	manifest, err := EncodeManifestReader(bytes.NewReader(b))
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	if len(manifest) != 13 {
		t.Error("wrong manifest count item")
	}
	m := manifest["entities-legacy.html"]
	if m.File != "assets/js/entities-legacy-3d69250e.js" {
		t.Error("wrong file")
	}
	if m.Src != "entities-legacy.html" {
		t.Error("wrong src")
	}
	if !m.IsEntry {
		t.Error("wrong IsEntry")
	}
	if len(m.Imports) != 3 {
		t.Error("wrong import counts")
	}

	for _, v := range []struct {
		ExpectedValue string
		Index         int
	}{
		{
			ExpectedValue: "_footer-legacy-f2943c72.js",
			Index:         0,
		},
		{
			ExpectedValue: "_pages-legacy-ed749491.js",
			Index:         1,
		},
		{
			ExpectedValue: "_status-legacy-003337eb.js",
			Index:         2,
		},
	} {
		if v.ExpectedValue != m.Imports[v.Index] {
			t.Error("wrong value")
		}
	}
}

func TestManifest2(t *testing.T) {
	b, err := os.ReadFile("test_data/manifest2.json")
	if err != nil {
		t.Error(err)
		return
	}
	manifest, err := EncodeManifestReader(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
		return
	}
	m := manifest["entities-legacy.html"]
	// check imports
	for _, v := range []struct {
		ExpectedValue string
		Index         int
	}{
		{
			ExpectedValue: "_footer-legacy-f2943c72.js",
			Index:         0,
		},
		{
			ExpectedValue: "_pages-legacy-ed749491.js",
			Index:         1,
		},
		{
			ExpectedValue: "_status-legacy-003337eb.js",
			Index:         2,
		},
	} {
		if v.ExpectedValue != m.Imports[v.Index] {
			t.Error("wrong value")
		}
	}
	// check css
	for _, v := range []struct {
		ExpectedValue string
		Index         int
	}{
		{
			ExpectedValue: "assets/css/entities-42c4defa.css",
			Index:         0,
		},
		{
			ExpectedValue: "assets/css/pages-eaa62bee.css",
			Index:         1,
		},
		{
			ExpectedValue: "assets/css/status-be5c9c9f.css",
			Index:         2,
		},
	} {
		if v.ExpectedValue != m.CSS[v.Index] {
			t.Error("wrong value")
		}
	}
	// check dynamicImports
	for _, v := range []struct {
		ExpectedValue string
		Index         int
	}{
		{
			ExpectedValue: "src/views/AboutView.vue",
			Index:         0,
		},
		{
			ExpectedValue: "src/components/pages/RbacPage.vue",
			Index:         1,
		},
	} {
		if v.ExpectedValue != m.DynamicImports[v.Index] {
			t.Error("wrong value")
		}
	}

	if m.Name != "AboutView" {
		t.Error("wrong name")
	}
	if m.File != "assets/js/entities-legacy-3d69250e.js" {
		t.Error("wrong file")
	}
	if m.Src != "entities-legacy.html" {
		t.Error("wrong src")
	}
	if !m.IsEntry {
		t.Error("wrong is entry")
	}
	if !m.IsDynamicEntry {
		t.Error("wrong IsDynamicEntry")
	}
}

func TestXxx(t *testing.T) {
	vg, err := New(&ViteGoParams{
		ManifestPath: "test_data/manifest4.json",
		ParamsGetHeads: &ParamsGetHeads{
			BasePath: "vite/",
		},
	})
	if err != nil {
		t.Error(err)
		return
	}
	vg.FillHeads()
	headStr, err := vg.GetHeadsString("src/components/entrypoints/admin/index.html")
	if err != nil {
		t.Error(err)
		return
	}
	if headStr != `<script type='module' crossorigin src='vite//assets/index-CnU_a7Ch.js'></script>
<link rel='modulepreload' crossorigin href='vite//assets/LinkCss.vue_vue_type_script_setup_true_lang-qyXjZeGd.js'>
<link rel='modulepreload' crossorigin href='vite//assets/_commonjsHelpers-Cpj98o6Y.js'>
<link rel='modulepreload' crossorigin href='vite//assets/_plugin-vue_export-helper-DlAUqK2U.js'>
<link rel='stylesheet' crossorigin href='vite//assets/index-DHXAGmwn.css>` {
		t.Error("wrong headStr")
	}
}

func TestManifestWath(t *testing.T) {
	t.Skip("manual test")
	vg, err := New(&ViteGoParams{
		ManifestPath: "test_data/manifest4.json",
		ParamsGetHeads: &ParamsGetHeads{
			BasePath: "vite/",
		},
		Logger: slog.Default(),
	})
	if err != nil {
		t.Error(err)
		return
	}
	err = vg.FillHeads()
	if err != nil {
		fmt.Println("err FillHeads", err)
	}
	go vg.WatchManifest()

	time.Sleep(time.Minute * 25)
}
