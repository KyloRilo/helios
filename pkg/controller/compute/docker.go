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
	dockerContainer "github.com/moby/moby/api/types/container"
	"github.com/docker/go-sdk/client"
	"github.com/docker/go-sdk/image"
	"github.com/docker/go-sdk/network"

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

func (d DockerCtrl) createNetwork(ctx context.Context, name string) error {
	_, err := network.New(ctx, network.WithClient(d.client), network.WithName(name))
	if err != nil {
		return fmt.Errorf("create network %s: %w", name, err)
	}
	return nil
}

func (d DockerCtrl) removeNetwork(ctx context.Context, name string) error {
	_, err := d.client.NetworkRemove(ctx, name, apiClient.NetworkRemoveOptions{})
	if err != nil {
		return fmt.Errorf("remove network %s: %w", name, err)
	}
	return nil
}

func (d DockerCtrl) buildImage(ctx context.Context, n *compute.Node) error {
	build := n.Context
	if n.Image == "" {
		n.Image = fmt.Sprintf("helios/%s:local", n.Name)
	}

	buildCtx, err := buildContextArchive(build.Path)
	if err != nil {
		return err
	}
	defer buildCtx.Close()
	defer os.Remove(buildCtx.Name())

	resp, err := d.client.ImageBuild(ctx, buildCtx, apiClient.ImageBuildOptions{
		Dockerfile: build.File,
		Tags:       []string{n.Image},
		Remove:     true,
	})

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return fmt.Errorf("reading build output: %w", err)
	}

	return nil
}

func (d DockerCtrl) pullImage(ctx context.Context, n *compute.Node) error {
	err := image.Pull(ctx, n.Image, image.WithPullClient(d.client))
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
	case n.Image != "":
		if err := d.pullImage(ctx, n); err != nil {
			return "", err
		}
	case n.Context != nil:
		if err := d.buildImage(ctx, n); err != nil {
			return "", err
		}
	case n.Image != "" && n.Context != nil:
		return "", fmt.Errorf("Node '%s' has both Image and Build Context specified. Please specify only one.", n.Name)
	default:
		return "", fmt.Errorf("Node '%s' must have either an Image or a Build Context specified.", n.Name)
	}

	opts := []container.ContainerCustomizer{
		container.WithImage(n.Image),
		container.WithImagePlatform("linux/amd64"),
		container.WithClient(d.client),
		container.WithExposedPorts(n.Ports.ToStringArray()...),
		container.WithCmd(parseCommand(n.Cmd)...),
		container.WithEnv(n.Env),
		container.WithName(n.Name),
		container.WithNoStart(),
	}

	if len(n.Networks) > 0 {
		for _, nw := range n.Networks {
			opts = append(opts, container.WithNetworkName([]string{n.Name}, nw))
		}
	} else {
		opts = append(opts, container.WithBridgeNetwork())
	}

	if n.Entrypoint != "" {
		opts = append(opts, container.WithEntrypoint(strings.Fields(n.Entrypoint)...))
	}

	if n.Hostname != "" {
		opts = append(opts, container.WithConfigModifier(func(cfg *dockerContainer.Config) {
			cfg.Hostname = n.Hostname
		}))
	}

	if len(n.Volumes) > 0 {
		binds := make([]string, 0, len(n.Volumes))
		for src, dst := range n.Volumes {
			abs, err := filepath.Abs(src)
			if err != nil {
				abs = src
			}
			binds = append(binds, abs+":"+dst)
		}
		opts = append(opts, container.WithHostConfigModifier(func(hc *dockerContainer.HostConfig) {
			hc.Binds = append(hc.Binds, binds...)
		}))
	}

	if ctr, err = container.Run(ctx, opts...); err != nil {
		return "", err
	}

	return ctr.ID(), nil
}

func (d DockerCtrl) startNode(ctx context.Context, n *compute.Node) error {
	var err error
	if _, err = d.client.ContainerStart(ctx, n.Id, apiClient.ContainerStartOptions{}); err != nil {
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
	if _, err = d.client.ContainerStop(ctx, n.Id, apiClient.ContainerStopOptions{}); err != nil {
		return fmt.Errorf("Failed to stop container => %s", err)
	}

	return nil
}

func (d DockerCtrl) removeNode(ctx context.Context, n *compute.Node) error {
	var err error
	if _, err = d.client.ContainerRemove(ctx, n.Id, apiClient.ContainerRemoveOptions{}); err != nil {
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
