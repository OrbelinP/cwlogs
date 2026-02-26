package cwlogs

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#438f39")).
			Padding(0, 1)
)

type listType int

const (
	fetchedList listType = iota
	historyList
)

type selectModel struct {
	configDir string

	// select view
	listType listType
	list     list.Model
	keys     *listKeyMap
	selected *list.Item

	fetchedGroups []list.Item
	historyGroups []list.Item

	err error
}

func newSelectModel(logGroups []LogGroupDetails) selectModel {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

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
			listKeys.toggleHistory,
		}
	}
	logGroupList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleHistory,
		}
	}
	logGroupList.KeyMap.ShowFullHelp = listKeys.showFullHelp
	logGroupList.KeyMap.CloseFullHelp = listKeys.showShortHelp
	logGroupList.KeyMap.Filter = listKeys.startFiltering

	logGroupList.Paginator.PerPage = 20

	return selectModel{
		configDir:     configDir,
		listType:      fetchedList,
		list:          logGroupList,
		keys:          listKeys,
		fetchedGroups: items,
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

		if key.Matches(msg, m.keys.choose) {
			m.selected = new(m.list.SelectedItem())
			return m, tea.Quit
		}

		if key.Matches(msg, m.keys.toggleHistory) {
			if m.listType == fetchedList {
				if m.historyGroups == nil {
					var err error
					m.historyGroups, err = getHistoryLogGroups(m.configDir)
					if err != nil {
						m.err = err
						return m, tea.Quit
					}
				}

				m.list.Title = "History"
				m.list.ResetFilter()
				m.list.SetItems(m.historyGroups)
				m.listType = historyList
			} else {
				m.list.ResetFilter()
				m.list.Title = "Log Groups"
				m.list.SetItems(m.fetchedGroups)
				m.listType = fetchedList
			}
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m selectModel) View() tea.View {
	v := tea.NewView(appStyle.Render(m.list.View()))
	v.AltScreen = true
	return v
}

type listKeyMap struct {
	choose         key.Binding
	toggleHistory  key.Binding
	showFullHelp   key.Binding
	showShortHelp  key.Binding
	startFiltering key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
		toggleHistory: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "toggle history"),
		),
		showFullHelp: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "more"),
		),
		showShortHelp: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "close help"),
		),
		startFiltering: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "filter"),
		),
	}
}

func getHistoryLogGroups(basePath string) ([]list.Item, error) {
	history, err := LoadHistory(basePath)
	if err != nil {
		return nil, fmt.Errorf("loading history: %w", err)
	}

	items := make([]list.Item, len(history.LogGroups))
	for i, logGroup := range history.LogGroups {
		items[i] = logGroup
	}

	return items, nil
}
