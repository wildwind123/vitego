package manifestparser

import (
	"encoding/json"
	"io"

	"github.com/go-faster/errors"
)

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

func EncodeManifestReader(reader io.Reader) (map[string]ManifestItem, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "cant io.ReadAll")
	}
	return EncodeManifest(b)
}

func EncodeManifest(b []byte) (map[string]ManifestItem, error) {

	v := make(map[string]ManifestItem)

	err := json.Unmarshal(b, &v)
	if err != nil {
		return nil, errors.Wrap(err, "cant json.Unmarshal")
	}

	return v, nil
}
