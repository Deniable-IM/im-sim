package image

import (
	"deniable-im/im-sim/pkg/client"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	dockerImage "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/archive"
)

type PullOptions struct {
	RefStr string
	Opt    *dockerImage.PullOptions
}

type Options struct {
	BuildOpt *types.ImageBuildOptions
	PullOpt  *PullOptions
}

type Image struct {
	client   *client.Client
	buildCtx string
	Options  *Options
}

func NewImage(client *client.Client, buildCtx string, options *Options) (*Image, error) {
	image := &Image{client, buildCtx, options}
	if options.BuildOpt != nil {
		err := image.imageBuild(*options.BuildOpt)
		if err != nil {
			return nil, err
		}
		return image, nil
	} else if options.PullOpt != nil {
		err := image.imagePull(options.PullOpt.RefStr, options.PullOpt.Opt)
		if err != nil {
			return nil, err
		}
		return image, nil
	}

	return nil, fmt.Errorf("Failed to create new image.")
}

func (image *Image) imageBuild(buildOpt types.ImageBuildOptions) error {
	archive, err := archive.TarWithOptions(image.buildCtx, &archive.TarOptions{})
	if err != nil {
		return fmt.Errorf("Failed to get build context: %w.", err)
	}

	res, err := image.client.Cli.ImageBuild(image.client.Ctx, archive, buildOpt)
	if err != nil {
		return fmt.Errorf("Failed build image: %w.", err)
	}
	defer res.Body.Close()

	io.Copy(os.Stdout, res.Body)
	return nil
}

func (image *Image) imagePull(refStr string, pullOpt *dockerImage.PullOptions) error {
	if pullOpt == nil {
		pullOpt = &dockerImage.PullOptions{}
	}

	_, err := image.client.Cli.ImagePull(image.client.Ctx, refStr, *pullOpt)
	if err != nil {
		return fmt.Errorf("Faild pull of image: %w.", err)
	}

	return nil
}
