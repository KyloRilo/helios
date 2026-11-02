package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenSplash screen = iota
	screenInit
	screenMainMenu
	screenClusterList
	screenClusterDetail
	screenDashboard
	screenSubmitConfig
	screenConfirmDelete
)

type initStep struct {
	label string
	done  bool
}

type model struct {
	screen       screen
	prevScreen   screen
	width        int
	height       int
	cursor       int
	serviceCursor int

	helios   *ClusterInfo
	clusters []ClusterInfo

	initSteps   []initStep
	initCurrent int
	initDone    bool

	selectedCluster int
	flashMsg        string
	flashTime       time.Time

	textarea textarea.Model
	editing  bool

	quitting bool
}

type tickMsg time.Time
type initTickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initTickCmd() tea.Cmd {
	return tea.Tick(400*time.Millisecond, func(t time.Time) tea.Msg {
		return initTickMsg{}
	})
}

func newModel() model {
	ta := textarea.New()
	ta.Placeholder = "Paste your HCL cluster config here..."
	ta.CharLimit = 4096
	ta.SetWidth(70)
	ta.SetHeight(15)
	ta.ShowLineNumbers = true

	return model{
		screen: screenSplash,
		initSteps: []initStep{
			{label: "Reading cluster config from bin/helios/"},
			{label: "Parsing HCL manifest"},
			{label: "Validating cluster definition"},
			{label: "Building dependency graph"},
			{label: "Creating compute nodes"},
			{label: "Starting consul service"},
			{label: "Starting localstack service"},
			{label: "Starting core service"},
			{label: "Starting worker service"},
			{label: "Starting api service"},
			{label: "Helios cluster initialized"},
		},
		textarea: ta,
	}
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(min(70, msg.Width-10))
		return m, nil

	case tea.KeyMsg:
		if m.editing {
			return m.updateTextarea(msg)
		}
		return m.handleKey(msg)

	case tickMsg:
		for i := range m.clusters {
			for j := range m.clusters[i].Services {
				randomizeMetrics(&m.clusters[i].Services[j])
			}
		}
		if m.helios != nil {
			for j := range m.helios.Services {
				randomizeMetrics(&m.helios.Services[j])
			}
		}
		if m.flashMsg != "" && time.Since(m.flashTime) > 3*time.Second {
			m.flashMsg = ""
		}
		return m, tickCmd()

	case initTickMsg:
		if m.initCurrent < len(m.initSteps) {
			m.initSteps[m.initCurrent].done = true
			m.initCurrent++
			if m.initCurrent >= len(m.initSteps) {
				m.initDone = true
				h := heliosCluster()
				m.helios = &h
				m.clusters = sampleUserClusters()
			}
			return m, initTickCmd()
		}
		return m, nil
	}

	return m, nil
}

func (m *model) hasFlash() bool {
	return m.flashMsg != "" && time.Since(m.flashTime) < 3*time.Second
}

func (m model) updateTextarea(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editing = false
		m.textarea.Blur()
		return m, nil
	case "ctrl+s":
		config := strings.TrimSpace(m.textarea.Value())
		if config == "" {
			m.flashMsg = "Config cannot be empty"
			m.flashTime = time.Now()
			return m, nil
		}
		name := extractClusterName(config)
		newCluster := ClusterInfo{
			Name:      name,
			CreatedAt: time.Now(),
			Services:  generateMockServices(config),
		}
		m.clusters = append(m.clusters, newCluster)
		m.editing = false
		m.textarea.Blur()
		m.textarea.Reset()
		m.screen = screenClusterList
		m.flashMsg = fmt.Sprintf("Cluster '%s' registered successfully", name)
		m.flashTime = time.Now()
		return m, nil
	}
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "ctrl+c", "q":
		if m.screen == screenSplash || m.screen == screenMainMenu {
			m.quitting = true
			return m, tea.Quit
		}
		if m.screen == screenInit && m.initDone {
			m.screen = screenMainMenu
			m.cursor = 0
			return m, nil
		}
		if m.screen != screenInit {
			m.screen = screenMainMenu
			m.cursor = 0
			return m, nil
		}
	case "esc":
		switch m.screen {
		case screenClusterDetail, screenDashboard:
			m.screen = screenClusterList
			m.cursor = 0
			m.serviceCursor = 0
		case screenClusterList, screenSubmitConfig:
			m.screen = screenMainMenu
			m.cursor = 0
		case screenConfirmDelete:
			m.screen = screenClusterDetail
			m.cursor = 0
		}
		return m, nil
	}

	switch m.screen {
	case screenSplash:
		if key == "enter" {
			m.screen = screenInit
			return m, initTickCmd()
		}
	case screenInit:
		if m.initDone && key == "enter" {
			m.screen = screenMainMenu
			m.cursor = 0
		}
	case screenMainMenu:
		return m.handleMainMenu(key)
	case screenClusterList:
		return m.handleClusterList(key)
	case screenClusterDetail:
		return m.handleClusterDetail(key)
	case screenDashboard:
		return m.handleDashboard(key)
	case screenSubmitConfig:
		if key == "enter" || key == "i" {
			m.editing = true
			m.textarea.Focus()
			return m, textarea.Blink
		}
	case screenConfirmDelete:
		return m.handleConfirmDelete(key)
	}

	return m, nil
}

func (m model) handleMainMenu(key string) (tea.Model, tea.Cmd) {
	items := mainMenuItems()
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
		}
	case "enter":
		switch m.cursor {
		case 0:
			m.screen = screenClusterList
			m.cursor = 0
		case 1:
			m.screen = screenSubmitConfig
			m.cursor = 0
		case 2:
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func mainMenuItems() []string {
	return []string{
		"View Clusters",
		"Submit New Cluster Config",
		"Quit",
	}
}

func (m model) handleClusterList(key string) (tea.Model, tea.Cmd) {
	total := len(m.clusters)
	if m.helios != nil {
		total++
	}
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < total-1 {
			m.cursor++
		}
	case "enter":
		m.selectedCluster = m.cursor
		m.screen = screenClusterDetail
		m.cursor = 0
		m.serviceCursor = 0
	case "d":
		m.selectedCluster = m.cursor
		m.screen = screenDashboard
		m.serviceCursor = 0
	}
	return m, nil
}

func (m model) handleClusterDetail(key string) (tea.Model, tea.Cmd) {
	cluster := m.getSelectedCluster()
	if cluster == nil {
		return m, nil
	}
	switch key {
	case "up", "k":
		if m.serviceCursor > 0 {
			m.serviceCursor--
		}
	case "down", "j":
		if m.serviceCursor < len(cluster.Services)-1 {
			m.serviceCursor++
		}
	case "s":
		svc := &cluster.Services[m.serviceCursor]
		if svc.Status == StatusRunning {
			svc.Status = StatusStopped
			svc.CPU = 0
			svc.MemoryMB = 0
			svc.NetworkMB = 0
			svc.Uptime = 0
			m.flashMsg = fmt.Sprintf("Service '%s' stopped", svc.Name)
		} else {
			svc.Status = StatusRunning
			svc.CPU = 2.0
			svc.MemoryMB = 128
			svc.NetworkMB = 1.0
			svc.Uptime = 1 * time.Second
			m.flashMsg = fmt.Sprintf("Service '%s' started", svc.Name)
		}
		m.flashTime = time.Now()
	case "r":
		svc := &cluster.Services[m.serviceCursor]
		svc.Status = StatusCreating
		svc.CPU = 0
		svc.MemoryMB = 0
		svc.Uptime = 0
		m.flashMsg = fmt.Sprintf("Service '%s' restarting...", svc.Name)
		m.flashTime = time.Now()
	case "d":
		m.screen = screenDashboard
	case "x":
		if !cluster.IsHelios {
			m.screen = screenConfirmDelete
			m.cursor = 0
		}
	}
	return m, nil
}

func (m model) handleDashboard(key string) (tea.Model, tea.Cmd) {
	cluster := m.getSelectedCluster()
	if cluster == nil {
		return m, nil
	}
	switch key {
	case "up", "k":
		if m.serviceCursor > 0 {
			m.serviceCursor--
		}
	case "down", "j":
		if m.serviceCursor < len(cluster.Services)-1 {
			m.serviceCursor++
		}
	}
	return m, nil
}

func (m model) handleConfirmDelete(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y":
		idx := m.clusterIndex()
		if idx >= 0 && idx < len(m.clusters) {
			name := m.clusters[idx].Name
			m.clusters = append(m.clusters[:idx], m.clusters[idx+1:]...)
			m.flashMsg = fmt.Sprintf("Cluster '%s' deleted", name)
			m.flashTime = time.Now()
		}
		m.screen = screenClusterList
		m.cursor = 0
	case "n", "esc":
		m.screen = screenClusterDetail
		m.cursor = 0
	}
	return m, nil
}

func (m *model) getSelectedCluster() *ClusterInfo {
	if m.helios != nil {
		if m.selectedCluster == 0 {
			return m.helios
		}
		idx := m.selectedCluster - 1
		if idx < len(m.clusters) {
			return &m.clusters[idx]
		}
		return nil
	}
	if m.selectedCluster < len(m.clusters) {
		return &m.clusters[m.selectedCluster]
	}
	return nil
}

func (m *model) clusterIndex() int {
	if m.helios != nil {
		return m.selectedCluster - 1
	}
	return m.selectedCluster
}

func extractClusterName(config string) string {
	for _, line := range strings.Split(config, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "cluster") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return strings.Trim(parts[1], `"`)
			}
		}
	}
	return fmt.Sprintf("cluster-%d", time.Now().Unix())
}

func generateMockServices(config string) []ServiceInfo {
	var services []ServiceInfo
	for _, line := range strings.Split(config, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "service") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				name := strings.Trim(parts[1], `"`)
				services = append(services, ServiceInfo{
					Name:     name,
					Image:    name + ":latest",
					Status:   StatusCreating,
					Replicas: 1,
					Ports:    "-",
				})
			}
		}
	}
	if len(services) == 0 {
		services = append(services, ServiceInfo{
			Name:     "default",
			Image:    "default:latest",
			Status:   StatusCreating,
			Replicas: 1,
			Ports:    "-",
		})
	}
	return services
}
