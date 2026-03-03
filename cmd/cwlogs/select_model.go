package cwlogs

import (
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

func (m selectModel) View() tea.View {
	v := tea.NewView(appStyle.Render(m.list.View()))
	v.AltScreen = true
	return v
}
