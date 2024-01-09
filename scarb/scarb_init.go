package scarb

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ScarbInit struct {
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewScarbInit(dependency libpak.BuildpackDependency, cache libpak.DependencyCache) ScarbInit {
	contributor := libpak.NewDependencyLayerContributor(dependency, cache, libcnb.LayerTypes{
		Cache: true,
	})
	return ScarbInit{
		LayerContributor: contributor,
	}
}

func (r ScarbInit) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	r.LayerContributor.Logger = r.Logger
	if err := os.Setenv("PATH", sherpa.AppendToEnvVar("PATH", ":", filepath.Join(layer.Path, "bin"))); err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to set $PATH\n%w", err)
	}

	return r.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		file := filepath.Join(layer.Path, "bin", filepath.Base(artifact.Name()))

		r.Logger.Bodyf("Copying to %s", filepath.Dir(file))

		if err := sherpa.CopyFile(artifact, file); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy %s to %s\n%w", artifact.Name(), file, err)
		}
		if err := r.DeCompress(filepath.Join(filepath.Dir(file), "scarb-v2.4.3-x86_64-unknown-linux-musl.tar.gz"), filepath.Dir(file)); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to decompress %s to %s\n%w", artifact.Name(), file, err)
		}
		if err := os.Chmod(file, 0755); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to chmod %s\n%w", file, err)
		}

		return layer, nil
	})
}

func (r ScarbInit) Name() string {
	return r.LayerContributor.LayerName()
}
func (r ScarbInit) DeCompress(tarFile, dest string) error {
	srcFile, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		filename := dest + hdr.Name
		file, err := createFile(filename)
		if err != nil {
			log.Println("create file err = ", err)
			return err
		}
		_, err = io.Copy(file, tr)
		if err != nil {
			log.Println("io copy err= ", err)
			return err
		}
	}
	return nil
}

func createFile(name string) (*os.File, error) {
	if ok := Exists(string([]rune(name)[0:strings.LastIndex(name, "/")])); !ok {
		err := os.MkdirAll(string([]rune(name)[0:strings.LastIndex(name, "/")]), 0755)
		if err != nil {
			return nil, err
		}
	}

	return os.Create(name)
}
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
