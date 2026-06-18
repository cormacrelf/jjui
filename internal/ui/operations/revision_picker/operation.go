package revision_picker

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/ui/actions"
	"github.com/idursun/jjui/internal/ui/common"
	"github.com/idursun/jjui/internal/ui/intents"
	"github.com/idursun/jjui/internal/ui/layout"
	"github.com/idursun/jjui/internal/ui/operations"
	"github.com/idursun/jjui/internal/ui/render"
)

type SelectedMsg struct {
	ChangeID string
}

type CancelledMsg struct{}

var (
	_ operations.Operation              = (*Operation)(nil)
	_ operations.TracksSelectedRevision = (*Operation)(nil)
	_ common.ScopeProvider              = (*Operation)(nil)
	_ common.ScopeHandler               = (*Operation)(nil)
)

type Operation struct {
	title          string
	selected       *jj.Commit
	PreviousRevset string
}

func NewOperation(title string) *Operation {
	return &Operation{title: title}
}

func (o *Operation) Name() string { return "revision_picker" }

func (o *Operation) Init() tea.Cmd { return nil }

func (o *Operation) Update(msg tea.Msg) tea.Cmd {
	if intent, ok := msg.(intents.Intent); ok {
		cmd, _ := o.HandleIntent(intent)
		return cmd
	}
	return nil
}

func (o *Operation) ViewRect(_ *render.DisplayContext, _ layout.Box) {}

func (o *Operation) Scopes() []common.Scope {
	return []common.Scope{
		{
			Name:    actions.ScopeRevisionPicker,
			Leak:    common.LeakAll,
			Handler: o,
		},
	}
}

func (o *Operation) HandleIntent(intent intents.Intent) (tea.Cmd, bool) {
	switch intent.(type) {
	case intents.Apply:
		if o.selected == nil {
			return cmdMsg(CancelledMsg{}), true
		}
		return cmdMsg(SelectedMsg{ChangeID: o.selected.GetChangeId()}), true
	case intents.Cancel:
		return cmdMsg(CancelledMsg{}), true
	}
	return nil, false
}

func (o *Operation) SetSelectedRevision(commit *jj.Commit) tea.Cmd {
	o.selected = commit
	return nil
}

func (o *Operation) Render(commit *jj.Commit, pos operations.RenderPosition) string {
	if pos != operations.RenderPositionBefore || o.selected == nil {
		return ""
	}
	if commit.GetChangeId() != o.selected.GetChangeId() {
		return ""
	}

	markerStyle := common.DefaultPalette.Get("rebase target_marker")
	marker := markerStyle.Render("<< select >>")
	if o.title != "" {
		dimmedStyle := common.DefaultPalette.Get("rebase dimmed")
		return lipgloss.JoinHorizontal(lipgloss.Left, marker, " ", dimmedStyle.Render(o.title))
	}
	return marker
}

func Show(title, revset string) tea.Cmd {
	return func() tea.Msg {
		return common.ShowRevisionPickerMsg{Title: title, Revset: revset}
	}
}

func cmdMsg(msg tea.Msg) tea.Cmd {
	return func() tea.Msg { return msg }
}
