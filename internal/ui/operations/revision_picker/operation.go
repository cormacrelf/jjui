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
	marker         string
	position       common.PickerPosition
	selected       *jj.Commit
	PreviousRevset string
}

func NewOperation(msg common.ShowRevisionPickerMsg) *Operation {
	marker := msg.Marker
	if marker == "" {
		marker = "select"
	}
	return &Operation{title: msg.Title, marker: marker, position: msg.Position}
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
	if o.selected == nil || commit.GetChangeId() != o.selected.GetChangeId() {
		return ""
	}

	var expectedPos operations.RenderPosition
	switch o.position {
	case common.PickerBefore:
		expectedPos = operations.RenderPositionAfter
	case common.PickerInto:
		expectedPos = operations.RenderBeforeChangeId
	default:
		expectedPos = operations.RenderPositionBefore
	}
	if pos != expectedPos {
		return ""
	}

	markerStyle := common.DefaultPalette.Get("rebase target_marker")
	markerText := "<< " + o.marker + " >>"
	if o.position == common.PickerInto {
		return markerStyle.Render(markerText + " ")
	}
	marker := markerStyle.Render(markerText)
	if o.title != "" {
		dimmedStyle := common.DefaultPalette.Get("rebase dimmed")
		return lipgloss.JoinHorizontal(lipgloss.Left, marker, " ", dimmedStyle.Render(o.title))
	}
	return marker
}

func Show(msg common.ShowRevisionPickerMsg) tea.Cmd {
	return func() tea.Msg { return msg }
}

func cmdMsg(msg tea.Msg) tea.Cmd {
	return func() tea.Msg { return msg }
}
