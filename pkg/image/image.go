package image

import (
	"deniable-im/network-simulation/pkg/client"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
)

type Image struct {
	client   client.Client
	buildCtx string
	buildOpt types.ImageBuildOptions
}

func NewImage(client client.Client, buildCtx string, buildOpt types.ImageBuildOptions) (*Image, error) {
	image := &Image{client, buildCtx, buildOpt}
	err := image.imageBuild()
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (image *Image) imageBuild() error {
	archive, err := archive.TarWithOptions(image.buildCtx, &archive.TarOptions{})
	if err != nil {
		return fmt.Errorf("Failed to get build context %w", err)
	}

	res, err := image.client.Cli.ImageBuild(image.client.Ctx, archive, image.buildOpt)
	if err != nil {
		return fmt.Errorf("Failed build image %w", err)
	}
	defer res.Body.Close()

	io.Copy(os.Stdout, res.Body)
	return nil
}
