package components

import (
	"fmt"
	"hhx/internal/models"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FileItem represents a file item in the list
type FileItem struct {
	File *models.File
}

// FilterValue returns the filter value for the file item
func (i FileItem) FilterValue() string {
	return i.File.Path
}

// Title returns the title for the file item
func (i FileItem) Title() string {
	return i.File.Path
}

// Description returns the description for the file item
func (i FileItem) Description() string {
	var status string
	switch i.File.Status {
	case models.StatusStaged:
		status = "Staged"
	case models.StatusModified:
		status = "Modified"
	case models.StatusUntracked:
		status = "Untracked"
	case models.StatusSynced:
		status = "Synced"
	}

	return fmt.Sprintf("%s - %d bytes", status, i.File.Size)
}

// FileListModel represents the file list model
type FileListModel struct {
	List     list.Model
	Files    []*models.File
	Selected *models.File
}

// NewFileListModel creates a new file list model
func NewFileListModel(width, height int) FileListModel {
	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), width, height)
	listModel.Title = "Files"
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true)
	listModel.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		MarginLeft(2)

	return FileListModel{
		List:  listModel,
		Files: []*models.File{},
	}
}

// SetFiles sets the files in the list
func (m *FileListModel) SetFiles(files []*models.File) {
	m.Files = files

	// Sort files by status and name
	sort.Slice(files, func(i, j int) bool {
		if files[i].Status != files[j].Status {
			// Order: Staged, Modified, Untracked, Synced
			statusOrder := map[models.FileStatus]int{
				models.StatusStaged:    0,
				models.StatusModified:  1,
				models.StatusUntracked: 2,
				models.StatusSynced:    3,
			}
			return statusOrder[files[i].Status] < statusOrder[files[j].Status]
		}
		return files[i].Path < files[j].Path
	})

	// Create items
	items := make([]list.Item, len(files))
	for i, file := range files {
		items[i] = FileItem{File: file}
	}

	m.List.SetItems(items)
}

// Update handles file list updates
func (m FileListModel) Update(msg tea.Msg) (FileListModel, tea.Cmd) {
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)

	// Update selected file
	if item, ok := m.List.SelectedItem().(FileItem); ok {
		m.Selected = item.File
	} else {
		m.Selected = nil
	}

	return m, cmd
}

// View renders the file list
func (m FileListModel) View() string {
	return m.List.View()
}
