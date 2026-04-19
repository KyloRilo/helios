package compute

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/KyloRilo/helios/pkg/model/compute"

	"github.com/docker/go-connections/nat"
	"github.com/docker/go-sdk/client"
	"github.com/docker/go-sdk/image"

	"github.com/docker/go-sdk/container"
	apiClient "github.com/moby/moby/client"
)

type DockerCtrl struct {
	client    client.SDKClient
	platform  string
	authToken string
}

func buildContextArchive(contextPath string) (*os.File, error) {
	tmp, err := os.CreateTemp("", "helios-build-context-*.tar")
	if err != nil {
		return nil, err
	}

	tw := tar.NewWriter(tmp)
	walkErr := filepath.Walk(contextPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(contextPath, path)
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		return err
	})

	closeErr := tw.Close()
	if walkErr != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return nil, walkErr
	}
	if closeErr != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return nil, closeErr
	}

	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return nil, err
	}

	return tmp, nil
}

func (d DockerCtrl) buildImage(ctx context.Context, n *compute.Node) error {
	build := n.GetContext()
	tag := n.GetImage()
	if tag == "" {
		tag = fmt.Sprintf("helios/%s:local", n.GetName())
	}

	buildCtx, err := buildContextArchive(build.Path)
	if err != nil {
		return err
	}
	defer buildCtx.Close()
	defer os.Remove(buildCtx.Name())

	resp, err := d.client.ImageBuild(ctx, buildCtx, apiClient.ImageBuildOptions{
		Dockerfile: build.File,
		Tags:       []string{tag},
		Remove:     true,
	})

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (d DockerCtrl) pullImage(ctx context.Context, n *compute.Node) error {
	err := image.Pull(ctx, n.GetImage(), image.WithPullClient(d.client))
	if err != nil {
		return fmt.Errorf("Failed to pull Image => %s", err)
	}

	return nil
}

func parseCommand(command string) []string {
	if strings.TrimSpace(command) == "" {
		return []string{"sleep", "infinity"}
	}
	return strings.Fields(command)
}

func parsePorts(ports []string) (nat.PortSet, nat.PortMap, error) {
	if len(ports) == 0 {
		return nil, nil, nil
	}

	exposed, bindings, err := nat.ParsePortSpecs(ports)
	if err != nil {
		return nil, nil, err
	}

	return exposed, bindings, nil
}

func (d DockerCtrl) createNode(ctx context.Context, n *compute.Node) (string, error) {
	var ctr *container.Container
	var err error

	switch {
	case n.GetImage() != "":
		if err := d.pullImage(ctx, n); err != nil {
			return "", err
		}
	case n.GetContext() != nil:
		if err := d.buildImage(ctx, n); err != nil {
			return "", err
		}
	case n.GetImage() != "" && n.GetContext() != nil:
		return "", fmt.Errorf("Node '%s' has both Image and Build Context specified. Please specify only one.", n.GetName())
	default:
		return "", fmt.Errorf("Node '%s' must have either an Image or a Build Context specified.", n.GetName())
	}

	if ctr, err = container.Run(
		ctx,
		container.WithImage(n.GetImage()),
		container.WithImagePlatform("linux/amd64"),
		container.WithClient(d.client),
		container.WithExposedPorts(n.GetPorts().ToStringArray()...),
		container.WithCmd(parseCommand(n.GetCmd())...),
		container.WithEnv(n.GetEnv()),
		container.WithName(n.GetName()),
		container.WithBridgeNetwork(),
		container.WithNoStart(),
	); err != nil {
		return "", err
	}

	return ctr.ID(), nil
}

func (d DockerCtrl) startNode(ctx context.Context, n *compute.Node) error {
	var err error
	if _, err = d.client.ContainerStart(ctx, n.GetId(), apiClient.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Failed to start node => %s", err)
	}

	return nil
}

func (d DockerCtrl) listNodes(ctx context.Context) ([]*compute.Node, error) {
	var resp apiClient.ContainerListResult
	var err error
	if resp, err = d.client.ContainerList(ctx, apiClient.ContainerListOptions{}); err != nil {
		return nil, fmt.Errorf("Failed to list nodes => %s", err)
	}

	nodes := []*compute.Node{}
	for _, ctr := range resp.Items {
		ports := map[string]string{}
		for _, port := range ctr.Ports {
			ports[fmt.Sprintf("%d", port.PublicPort)] = fmt.Sprintf("%d", port.PrivatePort)
		}

		n := compute.NewNode(
			compute.WithId(ctr.ID),
			compute.WithName(ctr.Names[0]),
			compute.WithCmd(ctr.Command),
			compute.WithPorts(ports),
		)

		nodes = append(nodes, n)
	}

	return nodes, nil
}

func (d DockerCtrl) stopNode(ctx context.Context, n *compute.Node) error {
	var err error
	if _, err = d.client.ContainerStop(ctx, n.GetId(), apiClient.ContainerStopOptions{}); err != nil {
		return fmt.Errorf("Failed to stop container => %s", err)
	}

	return nil
}

func (d DockerCtrl) removeNode(ctx context.Context, n *compute.Node) error {
	var err error
	if _, err = d.client.ContainerRemove(ctx, n.GetId(), apiClient.ContainerRemoveOptions{}); err != nil {
		return fmt.Errorf("Failed to remove container => %s", err)
	}

	return nil
}

func newDockerCtrl(ctx context.Context) (CtrlShim, error) {
	client, err := client.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("Unable to create docker client => %s", err)
	}
	ctrl := DockerCtrl{
		client: client,
	}

	return ctrl, nil
}
