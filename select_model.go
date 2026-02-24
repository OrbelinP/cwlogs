package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#438f39")).
			Padding(0, 1)
)

type selectModel struct {
	// select view
	list list.Model
	keys *listKeyMap
}

func newSelectModel(logGroups []LogGroupDetails) selectModel {
	items := make([]list.Item, len(logGroups))
	for i, logGroup := range logGroups {
		items[i] = logGroup
	}

	listKeys := newListKeyMap()

	logGroupList := list.New(items, newItemDelegate(), 0, 0)
	logGroupList.Title = "Log Groups"
	logGroupList.Styles.Title = titleStyle
	logGroupList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.choose,
		}
	}
	logGroupList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.choose,
		}
	}

	logGroupList.Paginator.PerPage = 20

	return selectModel{
		list: logGroupList,
		keys: listKeys,
	}
}

func (m selectModel) Init() tea.Cmd {
	return nil
}

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.choose):
			return m, tea.Quit
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m selectModel) View() string {
	return appStyle.Render(m.list.View())
}

type listKeyMap struct {
	choose key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
	}
}
