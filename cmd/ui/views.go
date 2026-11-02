package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if m.quitting {
		return "\n  Goodbye from Helios.\n\n"
	}

	var s string
	switch m.screen {
	case screenSplash:
		s = m.viewSplash()
	case screenInit:
		s = m.viewInit()
	case screenMainMenu:
		s = m.viewMainMenu()
	case screenClusterList:
		s = m.viewClusterList()
	case screenClusterDetail:
		s = m.viewClusterDetail()
	case screenDashboard:
		s = m.viewDashboard()
	case screenSubmitConfig:
		s = m.viewSubmitConfig()
	case screenConfirmDelete:
		s = m.viewConfirmDelete()
	}

	return s
}

func (m model) viewSplash() string {
	logo := `
   ██╗  ██╗███████╗██╗     ██╗ ██████╗ ███████╗
   ██║  ██║██╔════╝██║     ██║██╔═══██╗██╔════╝
   ███████║█████╗  ██║     ██║██║   ██║███████╗
   ██╔══██║██╔══╝  ██║     ██║██║   ██║╚════██║
   ██║  ██║███████╗███████╗██║╚██████╔╝███████║
   ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝ ╚═════╝ ╚══════╝`

	styled := lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(logo)
	sub := subtitleStyle.Render("An orchestrator for today's engineers")
	prompt := lipgloss.NewStyle().Foreground(colorMuted).MarginTop(2).Render("Press Enter to initialize Helios...")

	content := lipgloss.JoinVertical(lipgloss.Center, styled, "", sub, prompt)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m model) viewInit() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("⚡ Initializing Helios"))
	b.WriteString("\n\n")

	for i, step := range m.initSteps {
		if step.done {
			b.WriteString(statusRunning.Render("  ✓ "))
			b.WriteString(lipgloss.NewStyle().Foreground(colorWhite).Render(step.label))
		} else if i == m.initCurrent {
			b.WriteString(statusCreating.Render("  ◌ "))
			b.WriteString(lipgloss.NewStyle().Foreground(colorWarning).Render(step.label))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render("  · " + step.label))
		}
		b.WriteString("\n")
	}

	if m.initDone {
		b.WriteString("\n")
		b.WriteString(successMsg.Render("  Helios is ready!"))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  Press Enter to continue"))
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		boxStyle.Render(b.String()))
}

func (m model) viewMainMenu() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("☀ Helios Control Center"))
	b.WriteString("\n")

	if m.helios != nil {
		running := 0
		for _, svc := range m.helios.Services {
			if svc.Status == StatusRunning {
				running++
			}
		}
		b.WriteString(subtitleStyle.Render(fmt.Sprintf(
			"Helios cluster: %s | Services: %d/%d running | Managed clusters: %d",
			m.helios.Name, running, len(m.helios.Services), len(m.clusters))))
		b.WriteString("\n\n")
	}

	items := mainMenuItems()
	for i, item := range items {
		if i == m.cursor {
			b.WriteString(selectedMenuStyle.Render("▸ " + item))
		} else {
			b.WriteString(menuItemStyle.Render("  " + item))
		}
		b.WriteString("\n")
	}

	if m.hasFlash() {
		b.WriteString("\n")
		b.WriteString(successMsg.Render("  " + m.flashMsg))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  j/k: navigate • enter: select • q: quit"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		boxStyle.Render(b.String()))
}

func (m model) viewClusterList() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📋 Registered Clusters"))
	b.WriteString("\n")

	header := fmt.Sprintf("  %-20s %-12s %-10s %-20s", "NAME", "SERVICES", "STATUS", "CREATED")
	b.WriteString(tableHeaderStyle.Render(header))
	b.WriteString("\n")

	var allClusters []ClusterInfo
	if m.helios != nil {
		allClusters = append(allClusters, *m.helios)
	}
	allClusters = append(allClusters, m.clusters...)

	for i, c := range allClusters {
		running := 0
		for _, svc := range c.Services {
			if svc.Status == StatusRunning {
				running++
			}
		}
		status := statusRunning.Render("Healthy")
		if running < len(c.Services) {
			status = statusCreating.Render("Degraded")
		}
		if running == 0 {
			status = statusStopped.Render("Down")
		}

		name := c.Name
		if c.IsHelios {
			name += " ★"
		}

		row := fmt.Sprintf("%-20s %d/%-9d %-10s %s",
			name, running, len(c.Services), "", formatAge(c.CreatedAt))

		if i == m.cursor {
			b.WriteString(selectedRowStyle.Render("▸ " + row))
			b.WriteString("  " + status)
		} else {
			b.WriteString(tableRowStyle.Render("  " + row))
			b.WriteString("  " + status)
		}
		b.WriteString("\n")
	}

	if m.hasFlash() {
		b.WriteString("\n")
		b.WriteString(successMsg.Render("  " + m.flashMsg))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  j/k: navigate • enter: details • d: dashboard • esc: back"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		boxStyle.Render(b.String()))
}

func (m model) viewClusterDetail() string {
	cluster := m.getSelectedCluster()
	if cluster == nil {
		return "No cluster selected"
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("🔧 Cluster: %s", cluster.Name)))
	if cluster.IsHelios {
		b.WriteString(subtitleStyle.Render("  (Helios System Cluster)"))
	}
	b.WriteString("\n\n")

	header := fmt.Sprintf("  %-15s %-25s %-12s %-10s %-8s %-12s",
		"SERVICE", "IMAGE", "STATUS", "CPU %", "MEM MB", "UPTIME")
	b.WriteString(tableHeaderStyle.Render(header))
	b.WriteString("\n")

	for i, svc := range cluster.Services {
		image := svc.Image
		if len(image) > 24 {
			image = image[:21] + "..."
		}

		prefix := "  "
		style := tableRowStyle
		if i == m.serviceCursor {
			prefix = "▸ "
			style = selectedRowStyle
		}

		left := fmt.Sprintf("%s%-15s %-25s ", prefix, svc.Name, image)
		right := fmt.Sprintf(" %5.1f%%   %6.0f   %s", svc.CPU, svc.MemoryMB, formatDuration(svc.Uptime))

		b.WriteString(style.Render(left))
		b.WriteString(styledStatus(svc.Status))
		b.WriteString(style.Render(right))
		b.WriteString("\n")
	}

	if m.hasFlash() {
		b.WriteString("\n")
		b.WriteString(successMsg.Render("  " + m.flashMsg))
	}

	b.WriteString("\n")
	actions := "  j/k: navigate services • s: start/stop • r: restart • d: dashboard"
	if !cluster.IsHelios {
		actions += " • x: delete cluster"
	}
	actions += " • esc: back"
	b.WriteString(helpStyle.Render(actions))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		boxStyle.Render(b.String()))
}

func (m model) viewDashboard() string {
	cluster := m.getSelectedCluster()
	if cluster == nil {
		return "No cluster selected"
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("📊 Dashboard: %s", cluster.Name)))
	b.WriteString("\n\n")

	var totalCPU, totalMem, totalNet float64
	running := 0
	for _, svc := range cluster.Services {
		totalCPU += svc.CPU
		totalMem += svc.MemoryMB
		totalNet += svc.NetworkMB
		if svc.Status == StatusRunning {
			running++
		}
	}

	summaryCards := lipgloss.JoinHorizontal(lipgloss.Top,
		dashboardCard.Render(fmt.Sprintf(
			"%s\n%s %d/%d",
			headerStyle.Render("Services"),
			metricValue.Render("Active:"),
			running, len(cluster.Services))),
		"  ",
		dashboardCard.Render(fmt.Sprintf(
			"%s\n%s %.1f%%",
			headerStyle.Render("Total CPU"),
			metricValue.Render("Usage:"),
			totalCPU)),
		"  ",
		dashboardCard.Render(fmt.Sprintf(
			"%s\n%s %.0f MB",
			headerStyle.Render("Total Memory"),
			metricValue.Render("Used:"),
			totalMem)),
		"  ",
		dashboardCard.Render(fmt.Sprintf(
			"%s\n%s %.1f MB/s",
			headerStyle.Render("Network I/O"),
			metricValue.Render("Traffic:"),
			totalNet)),
	)
	b.WriteString(summaryCards)
	b.WriteString("\n\n")

	b.WriteString(headerStyle.Render("Service Metrics"))
	b.WriteString("\n")

	for i, svc := range cluster.Services {
		prefix := "  "
		style := tableRowStyle
		if i == m.serviceCursor {
			prefix = "▸ "
			style = selectedRowStyle
		}

		nameStr := style.Render(fmt.Sprintf("%s%-14s", prefix, svc.Name))
		statusStr := styledStatus(svc.Status)

		cpuBar := renderBar(svc.CPU, 100, 20)
		memBar := renderBar(svc.MemoryMB, 4096, 20)
		netBar := renderBar(svc.NetworkMB, 50, 15)

		line := fmt.Sprintf("%s %s  CPU %s %5.1f%%  MEM %s %6.0fMB  NET %s %5.1fMB/s",
			nameStr, statusStr,
			cpuBar, svc.CPU,
			memBar, svc.MemoryMB,
			netBar, svc.NetworkMB)
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  j/k: navigate • esc: back   (metrics refresh every second)"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		boxStyle.Render(b.String()))
}

func (m model) viewSubmitConfig() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📝 Submit Cluster Configuration"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Paste an HCL cluster manifest in Helios format"))
	b.WriteString("\n\n")

	if m.editing {
		b.WriteString(inputStyle.Render(m.textarea.View()))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  ctrl+s: submit • esc: cancel editing"))
	} else {
		b.WriteString(inputStyle.Render(m.textarea.View()))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  enter/i: start editing • esc: back"))
	}

	example := subtitleStyle.Render(`
  Example format:
  cluster "my-app" {
      service "web" {
          image = "nginx:latest"
          ports = { "80":"80" }
      }
      service "api" {
          image = "myapp/api:v1"
          ports = { "3000":"3000" }
          depends_on = ["web"]
      }
  }`)
	b.WriteString("\n")
	b.WriteString(example)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		boxStyle.Render(b.String()))
}

func (m model) viewConfirmDelete() string {
	cluster := m.getSelectedCluster()
	name := "unknown"
	if cluster != nil {
		name = cluster.Name
	}

	content := fmt.Sprintf(
		"%s\n\n  Are you sure you want to delete cluster '%s'?\n  This will destroy all services.\n\n%s",
		titleStyle.Render("⚠ Confirm Delete"),
		name,
		helpStyle.Render("  y: confirm • n: cancel"),
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		boxStyle.BorderForeground(colorDanger).Render(content))
}

func renderBar(value, max float64, width int) string {
	if max <= 0 {
		max = 1
	}
	ratio := value / max
	if ratio > 1 {
		ratio = 1
	}
	if ratio < 0 {
		ratio = 0
	}
	filled := int(ratio * float64(width))
	empty := width - filled

	style := barFilled
	if ratio > 0.8 {
		style = barDanger
	} else if ratio > 0.6 {
		style = barWarning
	}

	return "[" + style.Render(strings.Repeat("█", filled)) +
		barEmpty.Render(strings.Repeat("░", empty)) + "]"
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 24 {
		return fmt.Sprintf("%dd %dh", h/24, h%24)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}
