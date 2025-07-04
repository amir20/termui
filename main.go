package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"dtop/config"
	"dtop/internal/docker"
	"dtop/internal/ui"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	version = "dev"
	commit  = "n/a"
	date    = "n/a"
)

func main() {
	var cfg config.Cli
	kong.Parse(&cfg, kong.Configuration(kongyaml.Loader, "./config.yaml", "./config.yml", "~/.config/dtop/config.yaml", "~/.config/dtop/config.yml", "~/.dtop.yml", "~/.dtop.yaml"))

	if cfg.Version {
		fmt.Printf("dtop version: %s\nCommit: %s\nBuilt on: %s\n", version, commit, date)
		os.Exit(0)
	}

	var hosts []docker.Host
	for _, hc := range cfg.Hosts {
		if hc.Host == "local" {
			cli, err := config.NewLocalClient()
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			host := docker.Host{
				Client:     cli,
				HostConfig: hc,
			}
			hosts = append(hosts, host)
		} else if strings.HasPrefix(hc.Host, "ssh://") {
			cli, err := config.NewSSHClient(hc.Host)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			host := docker.Host{
				Client:     cli,
				HostConfig: hc,
			}
			hosts = append(hosts, host)
		} else if strings.HasPrefix(hc.Host, "tcp://") {
			cli, err := config.NewRemoteClient(hc.Host)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			host := docker.Host{
				Client:     cli,
				HostConfig: hc,
			}
			hosts = append(hosts, host)
		} else {
			fmt.Println("Unsupported host type:", hc.Host)
			os.Exit(1)
		}
	}

	client := docker.NewMultiClient(hosts...)

	p := tea.NewProgram(ui.NewModel(context.Background(), client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
