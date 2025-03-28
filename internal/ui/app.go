package ui

import (
	"fmt"
	"hhx/internal/config"
	"hhx/internal/models"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the UI model
type Model struct {
	Viewport      viewport.Model
	Spinner       spinner.Model
	IsLoading     bool
	StatusMessage string
	ErrorMessage  string
	Config        *config.Config
	RepoConfig    *config.RepoConfig
	Index         *models.Index
	Files         []*models.File
	Width         int
	Height        int
	Ready         bool
}

// NewModel creates a new UI model
func NewModel(cfg *config.Config, repoCfg *config.RepoConfig, index *models.Index) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		Spinner:       s,
		IsLoading:     false,
		StatusMessage: "Ready",
		Config:        cfg,
		RepoConfig:    repoCfg,
		Index:         index,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.Spinner.Tick, loadFiles(m.Index))
}

// Update handles UI updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			m.IsLoading = true
			m.StatusMessage = "Refreshing files..."
			return m, loadFiles(m.Index)
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		if !m.Ready {
			// First time initializing
			m.Viewport = viewport.New(msg.Width, msg.Height-6)
			m.Viewport.YPosition = 3
			m.Ready = true
		} else {
			m.Viewport.Width = msg.Width
			m.Viewport.Height = msg.Height - 6
		}

		return m, nil

	case spinner.TickMsg:
		var spinnerCmd tea.Cmd
		m.Spinner, spinnerCmd = m.Spinner.Update(msg)
		cmds = append(cmds, spinnerCmd)

	case filesLoadedMsg:
		m.IsLoading = false
		m.StatusMessage = fmt.Sprintf("Loaded %d files", len(msg))
		m.Files = msg
		m.Viewport.SetContent(renderFiles(msg, m.Width))
		return m, nil

	case errorMsg:
		m.IsLoading = false
		m.ErrorMessage = string(msg)
		m.StatusMessage = "Error"
		return m, nil
	}

	if m.Ready {
		var viewportCmd tea.Cmd
		m.Viewport, viewportCmd = m.Viewport.Update(msg)
		cmds = append(cmds, viewportCmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	if !m.Ready {
		return "Initializing..."
	}

	var status string
	if m.IsLoading {
		status = fmt.Sprintf("%s %s", m.Spinner.View(), m.StatusMessage)
	} else {
		status = m.StatusMessage
	}

	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(0, 1).
		Render(status)

	titleBar := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Padding(0, 1).
		Render(fmt.Sprintf("Headless Hawx - %s", m.RepoConfig.CurrentRemote))

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(0, 1).
		Render("Press q to quit, r to refresh")

	errorView := ""
	if m.ErrorMessage != "" {
		errorView = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")).
			Padding(0, 1).
			Render(m.ErrorMessage)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleBar,
		statusBar,
		m.Viewport.View(),
		errorView,
		help,
	)
}

// Messages
type filesLoadedMsg []*models.File
type errorMsg string

// Commands
func loadFiles(index *models.Index) tea.Cmd {
	return func() tea.Msg {
		newFiles, modifiedFiles, deletedFiles, err := index.ScanWorkingDirectory()
		if err != nil {
			return errorMsg(fmt.Sprintf("Error scanning working directory: %v", err))
		}

		// Get staged files
		stagedFiles := index.GetStagedFiles()

		// Combine all files
		allFiles := make([]*models.File, 0, len(newFiles)+len(modifiedFiles)+len(deletedFiles)+len(stagedFiles))
		allFiles = append(allFiles, newFiles...)
		allFiles = append(allFiles, modifiedFiles...)
		allFiles = append(allFiles, deletedFiles...)
		allFiles = append(allFiles, stagedFiles...)

		return filesLoadedMsg(allFiles)
	}
}

// Helper functions
func renderFiles(files []*models.File, width int) string {
	if len(files) == 0 {
		return "No files found. Use 'hhx stage <file>' to stage files."
	}

	var stagedFiles, modifiedFiles, untrackedFiles, syncedFiles []*models.File

	// Categorize files
	for _, file := range files {
		switch file.Status {
		case models.StatusStaged:
			stagedFiles = append(stagedFiles, file)
		case models.StatusModified:
			modifiedFiles = append(modifiedFiles, file)
		case models.StatusUntracked:
			untrackedFiles = append(untrackedFiles, file)
		case models.StatusSynced:
			syncedFiles = append(syncedFiles, file)
		}
	}

	// Render sections
	var sections []string

	// Staged files
	if len(stagedFiles) > 0 {
		var content string
		content += lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Render("Staged Files (ready to push):\n")

		for _, file := range stagedFiles {
			content += fmt.Sprintf("  %s (%s)\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(file.Path),
				formatSize(file.Size),
			)
		}
		sections = append(sections, content)
	}

	// Modified files
	if len(modifiedFiles) > 0 {
		var content string
		content += lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Render("Modified Files (not staged):\n")

		for _, file := range modifiedFiles {
			content += fmt.Sprintf("  %s (%s)\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(file.Path),
				formatSize(file.Size),
			)
		}
		sections = append(sections, content)
	}

	// Untracked files
	if len(untrackedFiles) > 0 {
		var content string
		content += lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Render("Untracked Files:\n")

		for _, file := range untrackedFiles {
			content += fmt.Sprintf("  %s (%s)\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(file.Path),
				formatSize(file.Size),
			)
		}
		sections = append(sections, content)
	}

	// Synced files
	if len(syncedFiles) > 0 {
		var content string
		content += lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Render("Synced Files:\n")

		for _, file := range syncedFiles {
			content += fmt.Sprintf("  %s (%s)\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Render(file.Path),
				formatSize(file.Size),
			)
		}
		sections = append(sections, content)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// formatSize formats a file size in bytes to a human-readable string
func formatSize(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	unitIndex := 0
	floatSize := float64(size)

	for floatSize >= 1024 && unitIndex < len(units)-1 {
		floatSize /= 1024
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%d %s", size, units[unitIndex])
	}

	return fmt.Sprintf("%.2f %s", floatSize, units[unitIndex])
}
