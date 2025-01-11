package vitego

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-faster/errors"
)

type ViteGo struct {
	heads        sync.Map
	ViteGoParams *ViteGoParams
}

type ViteGoParams struct {
	ManifestPath string
	*ParamsGetHeads
	Logger *slog.Logger
}

type Manifest map[string]ManifestItem

type ManifestItem struct {
	File           string   `json:"file"`
	Src            string   `json:"src"`
	IsEntry        bool     `json:"isEntry"`
	Imports        []string `json:"imports"`
	CSS            []string `json:"css"`
	DynamicImports []string `json:"dynamicImports"`
	IsDynamicEntry bool     `json:"isDynamicEntry"`
	Name           string   `json:"name"`
}

type ParamsGetHeads struct {
	// example 'vite/' last symbol should be '/'
	BasePath string
	DevMode  bool
	DevHost  string
}

func New(params *ViteGoParams) (*ViteGo, error) {

	vg := ViteGo{
		ViteGoParams: params,
	}

	return &vg, nil
}

func (vg *ViteGo) WatchManifest() error {
	logger := vg.ViteGoParams.Logger

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "cant fsnotify.NewWatcher()")
	}
	defer watcher.Close()

	// Start listening for events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("Event:", event)
				if event.Name != vg.ViteGoParams.ManifestPath {
					continue
				}

				// Check for specific events
				if event.Op&fsnotify.Create == fsnotify.Create {
					err = vg.FillHeads()
					if err != nil {
						logger.Error("cant FillHeads", slog.Any("err", err))
					}
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					err = vg.FillHeads()
					if err != nil {
						logger.Error("cant FillHeads", slog.Any("err", err))
					}
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					logger.Error("File Deleted:")
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					logger.Error("File Renamed:")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Error("watcher.Errors", slog.Any("err", err))
				log.Println("Error:", err)
			}
		}
	}()

	// Add a path to watch
	manifestDirPath := filepath.Dir(vg.ViteGoParams.ManifestPath)
	err = watcher.Add(manifestDirPath)
	if err != nil {
		logger.Error("watcher.Add, retry after 5 sec", slog.Any("err", err), slog.String("manifest", manifestDirPath))
		for range time.NewTicker(time.Second * 5).C {
			err = watcher.Add(manifestDirPath)
			if err != nil {
				logger.Error("watcher.Add, retry after 5 sec", slog.Any("err", err), slog.String("manifest", manifestDirPath))
				continue
			}
			break
		}
	}

	done := make(chan bool)

	<-done
	return nil
}

func (vg *ViteGo) FillHeads() error {
	b, err := os.ReadFile(vg.ViteGoParams.ManifestPath)
	if err != nil {
		return errors.Wrap(err, "cant ReadFile")
	}

	manifest, err := EncodeManifestReader(bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "cant EncodeManifestReader")
	}

	// exist keys
	forDelKeys := make(map[string]bool)

	// fill exist keys to forDelKeys
	vg.heads.Range(func(key, value interface{}) bool {
		if realKey, ok := key.(string); ok {
			forDelKeys[realKey] = true
		}
		return true
	})

	for key, v := range manifest {
		if _, ok := forDelKeys[key]; ok {
			forDelKeys[key] = false
		}
		if !v.IsEntry {
			continue
		}
		heads, err := manifest.GetHeads(key, vg.ViteGoParams.ParamsGetHeads)
		if err != nil {
			return errors.Wrap(err, "cant ")
		}
		vg.heads.Store(key, heads)
	}

	for key, v := range forDelKeys {
		if !v {
			continue
		}
		vg.heads.Delete(key)
		fmt.Println("heads key deleted", key)
	}

	// log
	// fmt.Println("------- head ---------")
	// vg.heads.Range(func(key, value interface{}) bool {
	// 	if realKey, ok := key.(string); ok {
	// 		fmt.Println("realKey", realKey)
	// 	}
	// 	return true
	// })
	// fmt.Println("------- end head ---------")
	return nil
}

func (vg *ViteGo) GetHeads(entryPoint string) ([]string, error) {

	if !vg.ViteGoParams.DevMode {
		v, ok := vg.heads.Load(entryPoint)
		if !ok {
			return nil, errors.Errorf("entryPoint does not exist = %s", entryPoint)
		}

		return v.([]string), nil
	}

	return []string{
		fmt.Sprintf("<script type='module' src='%s/@vite/client'></script>", vg.ViteGoParams.DevHost),
		fmt.Sprintf("<script type='module' src='%s/%s'></script>", vg.ViteGoParams.DevHost, entryPoint),
	}, nil
}

func (vg *ViteGo) GetHeadsString(entryPoint string) (string, error) {
	heads, err := vg.GetHeads(entryPoint)
	if err != nil {
		return "", errors.Wrap(err, "cant GetHeads")
	}
	str := ""

	for i := range heads {
		if i == 0 {
			str = heads[i]
			continue
		}
		str = fmt.Sprintf("%s\n%s", str, heads[i])
	}

	return str, nil
}

func EncodeManifestReader(reader io.Reader) (Manifest, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "cant io.ReadAll")
	}
	return EncodeManifest(b)
}

func EncodeManifest(b []byte) (Manifest, error) {

	v := make(Manifest)

	err := json.Unmarshal(b, &v)
	if err != nil {
		return nil, errors.Wrap(err, "cant json.Unmarshal")
	}

	return v, nil
}

func (m Manifest) GetHeads(entryPoint string, params *ParamsGetHeads) ([]string, error) {
	v, ok := m[entryPoint]
	if !ok {
		return nil, errors.Errorf("entryPoint not exist = %s", entryPoint)
	}
	heads := []string{}

	// entrypoint
	if strings.HasSuffix(v.File, ".js") {
		segment := fmt.Sprintf("%s%s", params.BasePath, v.File)
		if v.IsEntry {
			heads = append(heads, fmt.Sprintf(`<script type='module' crossorigin src='%s'></script>`, segment))
		} else {
			heads = append(heads, fmt.Sprintf(`<link rel='modulepreload' crossorigin href='%s'>`, segment))
		}
	}

	// import js
	for i := range v.Imports {
		if !strings.HasSuffix(v.Imports[i], ".js") {
			continue
		}
		sList, err := m.GetHeads(v.Imports[i], params)
		if err != nil {
			return nil, errors.Wrap(err, "cant GetHeads")
		}
		heads = append(heads, sList...)
	}

	// css
	for i := range v.CSS {
		heads = append(heads, fmt.Sprintf(`<link rel='stylesheet' crossorigin href='%s%s>`, params.BasePath, v.CSS[i]))
	}

	return heads, nil
}
