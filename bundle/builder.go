package bundle

import (
	"bytes"
	"context"
	"io/fs"
	"testing/fstest"

	layerfs "github.com/dschmidt/go-layerfs"
	opaBundle "github.com/open-policy-agent/opa/bundle"
	opaCompile "github.com/open-policy-agent/opa/compile"
)

type Bundle struct {
	fs fs.FS
}

func Build(bundle Bundle, ctx context.Context) ([]byte, error) {

	buf := bytes.NewBuffer(nil)

	b, err := createBundle(bundle)
	if err != nil {
		return nil, err
	}

	compiler := opaCompile.New().
		WithBundle(&b).
		WithOutput(buf)

	err = compiler.Build(ctx)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}

func createBundle(bundle Bundle) (opaBundle.Bundle, error) {
	loader, err := opaBundle.NewFSLoader(bundle.fs)
	if err != nil {
		return opaBundle.Bundle{}, err
	}

	reader := opaBundle.NewCustomReader(loader)

	b, err := reader.Read()
	if err != nil {
		return opaBundle.Bundle{}, err
	}

	return b, nil
}

func CISKubernetesBundle() Bundle {
	return Bundle{fs: layerfs.New(CommonEmbed, CISRulesEmbed)}
}

func CISEksBundle() Bundle {
	return Bundle{fs: layerfs.New(CommonEmbed, EKSRulesEmbed)}
}

func (b *Bundle) With(path string, content string) Bundle {
	tmpFS := fstest.MapFS{
		path: {
			Data: []byte(content),
		},
	}
	b.fs = layerfs.New(tmpFS, b.fs)
	return *b
}
