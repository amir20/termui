package ui

import (
	"context"
	"dtop/internal/docker"
	"dtop/internal/ui/components/table"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	teaTable "github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

func NewModel(ctx context.Context, client *docker.Client) model {
	containerWatcher, err := client.WatchContainers(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	stats, err := client.WatchContainerStats(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	tbl := table.New(
		table.WithColumns([]table.Column{
			{Title: "", Width: 2, DisableStyle: true},
			{Title: "NAME", Width: 10, DisableStyle: true},
			{Title: "ID", Width: 13},
			{Title: "CPU", Width: 10, DisableStyle: true},
			{Title: "MEMORY", Width: 10, DisableStyle: true},
			{Title: "Status", Width: 10},
		}),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	tbl.SetStyles(teaTable.DefaultStyles())

	help := help.New()

	help.Styles.ShortKey = lipgloss.NewStyle().Bold(true)
	help.Styles.ShortDesc = lipgloss.NewStyle()

	return model{
		rows:             make(map[string]*row),
		orderedRows:      make([]*row, 0),
		table:            tbl,
		containerWatcher: containerWatcher,
		stats:            stats,
		keyMap:           defaultKeyMap,
		help:             help,
	}
}

func waitForContainerUpdate(ch <-chan []*docker.Container) tea.Cmd {
	return func() tea.Msg {
		c := <-ch
		return containers(c)
	}
}

func waitForStatsUpdate(ch <-chan docker.ContainerStat) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tick(),
		waitForContainerUpdate(m.containerWatcher),
		waitForStatsUpdate(m.stats),
	)
}
