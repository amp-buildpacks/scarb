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
		Cache:  true,
		Launch: true,
		Build:  true,
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
		// 解压到PATH
		if err := r.extractTarGz(artifact.Name(), filepath.Dir(file)); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to decompress %s to %s\n%w", artifact.Name(), file, err)
		}
		if err := os.Chmod(filepath.Dir(file), 0755); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to chmod %s\n%w", file, err)
		}
		return layer, nil
	})
}

func (r ScarbInit) Name() string {
	return r.LayerContributor.LayerName()
}
func (r ScarbInit) extractTarGz(src, dest string) error {
	// 打开 tar.gz 文件
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建 Gzip 读取器
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// 创建 Tar 读取器
	tarReader := tar.NewReader(gzipReader)

	// 逐个文件解压
	for {
		header, err := tarReader.Next()

		switch {
		case err == io.EOF:
			// 到达文件末尾
			return nil
		case err != nil:
			// 发生其他错误
			return err
		case header == nil:
			// 忽略空文件
			continue
		}

		// 获取目标文件路径
		targetFilePath := filepath.Join(dest, header.Name)

		// 确保目标路径存在
		if strings.HasSuffix(header.Name, "/") {
			if err := os.MkdirAll(targetFilePath, 0755); err != nil {
				return err
			}
			continue
		}
		// 创建或截断目标文件
		targetFile, err := os.Create(targetFilePath)
		if err != nil {
			log.Println("targetFilePath ======", targetFilePath)
			return err
		}
		defer targetFile.Close()

		os.Chmod(targetFile.Name(), os.FileMode(0111))
		if strings.HasSuffix(targetFile.Name(), "scarb") {
			log.Println("targetFile.Name() =============", filepath.Join(dest, "scarb"))
			if err := sherpa.CopyFile(targetFile, filepath.Join(dest, "scarb")); err != nil {
				return err
			}
		}
	}
}
