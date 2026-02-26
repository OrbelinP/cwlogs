package cwlogs

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
)

type itemDelegate struct {
	list.DefaultDelegate
}

func newItemDelegate() itemDelegate {
	d := itemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
	}

	d.SetSpacing(0)

	keys := newDelegatedKeyMap()

	help := []key.Binding{keys.choose}
	d.ShortHelpFunc = func() []key.Binding { return help }
	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}
	return d
}

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#438f39"))
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2)
)

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(LogGroupDetails)
	if !ok {
		return
	}
	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	_, err := fmt.Fprint(w, fn(i.Title()))
	if err != nil {
		panic(err)
	}
}

type delegateKeyMap struct {
	choose key.Binding
}

func newDelegatedKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
	}
}
