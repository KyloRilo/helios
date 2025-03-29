package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/KyloRilo/helios/models"
	"github.com/KyloRilo/helios/pkg"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
	maxWidth     = 80
	INIT_CLUSTER = iota
	TEARDOWN_CLUSTER
)

var (
	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
)

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 4, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}

type state int

const (
	statusNormal state = iota
	stateDone
)

type Model struct {
	state  state
	form   *huh.Form
	lg     *lipgloss.Renderer
	styles *Styles
	width  int
}

func NewModel() Model {
	renderer := lipgloss.DefaultRenderer()
	return Model{
		width:  maxWidth,
		lg:     renderer,
		styles: NewStyles(renderer),
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Key("operation").
					Options(
						huh.NewOption("Init Cluster", INIT_CLUSTER),
						huh.NewOption("Teardown Cluster", TEARDOWN_CLUSTER),
					).
					Title("Welcome to Helios!").
					Description("The new cloud-agnostic docker orchestrator"),
				huh.NewConfirm().
					Title("Are you sure?").
					Affirmative("Comin' right up!").
					Negative("Whomp Whomp..."),
			),
		).WithWidth(45).WithShowErrors(true).WithShowHelp(true),
	}
}

func (m Model) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m Model) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}

func (m Model) View() string {
	style := m.styles
	switch m.form.State {
	case huh.StateCompleted:
		var b strings.Builder
		fmt.Fprintf(&b, "Test completed...")
		return style.Status.Margin(0, 1).Padding(1, 2).Width(48).Render(b.String()) + "\n\n"
	default:
		v := strings.TrimSuffix(m.form.View(), "\n\n")
		form := m.lg.NewStyle().Margin(1, 0).Render(v)
		errors := m.form.Errors()
		header := m.appBoundaryView("Test boundary")
		if len(errors) > 0 {
			header = m.appErrorBoundaryView(m.errorView())
		}
		footer := m.appBoundaryView("Test Footer")
		return style.Base.Render(header + "\n" + form + "\n\n" + footer)
	}
}

func (m Model) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}

	return s
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		cmds = append(cmds, tea.Quit)
	}

	return m, tea.Batch(cmds...)
}

func main() {
	const (
		DOCKER = "docker"
		CLOUD  = "cloud"
	)

	srv := flag.String("service", "none", "service name to run")
	flag.Parse()

	switch *srv {
	case DOCKER:
		pkg.InitDockerService()
	case CLOUD:
		pkg.InitCloudService(models.CloudConfig{
			Cloud:      pkg.AWS,
			Region:     "us-east-1",
			RoleArn:    "arn://test-role-arn",
			ExternalId: "some-uuid",
		})
	default:
		_, err := tea.NewProgram(NewModel()).Run()
		if err != nil {
			fmt.Println("Oh no:", err)
			os.Exit(1)
		}
	}
}
