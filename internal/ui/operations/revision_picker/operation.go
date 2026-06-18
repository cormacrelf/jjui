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

type MultiSelectedMsg struct {
	ChangeIDs []string
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
	markerMulti    string
	multi          bool
	position       common.PickerPosition
	selected       *jj.Commit
	checked        map[string]bool
	checkedOrder   []string
	PreviousRevset string
}

func NewOperation(msg common.ShowRevisionPickerMsg) *Operation {
	marker := msg.Marker
	if marker == "" {
		marker = "select"
	}
	markerMulti := msg.MarkerMulti
	if markerMulti == "" {
		markerMulti = marker
	}
	return &Operation{
		title:       msg.Title,
		marker:      marker,
		markerMulti: markerMulti,
		multi:       msg.Multi,
		position:    msg.Position,
		checked:     make(map[string]bool),
	}
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
		if o.multi {
			if len(o.checked) == 0 {
				return cmdMsg(CancelledMsg{}), true
			}
			return cmdMsg(MultiSelectedMsg{ChangeIDs: append([]string(nil), o.checkedOrder...)}), true
		}
		if o.selected == nil {
			return cmdMsg(CancelledMsg{}), true
		}
		return cmdMsg(SelectedMsg{ChangeID: o.selected.GetChangeId()}), true
	case intents.RevisionPickerToggleSelect:
		if o.multi && o.selected != nil {
			id := o.selected.GetChangeId()
			if o.checked[id] {
				delete(o.checked, id)
				for i, v := range o.checkedOrder {
					if v == id {
						o.checkedOrder = append(o.checkedOrder[:i], o.checkedOrder[i+1:]...)
						break
					}
				}
			} else {
				o.checked[id] = true
				o.checkedOrder = append(o.checkedOrder, id)
			}
		}
		return nil, true
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
	changeId := commit.GetChangeId()

	if o.multi && o.checked[changeId] && pos == operations.RenderBeforeChangeId {
		markerStyle := common.DefaultPalette.Get("rebase source_marker")
		return markerStyle.Render("<< " + o.markerMulti + " >> ")
	}

	if o.selected == nil || changeId != o.selected.GetChangeId() {
		return ""
	}

	var expectedPos operations.RenderPosition
	switch o.position {
	case common.PickerBefore:
		expectedPos = operations.RenderPositionAfter
	case common.PickerInto:
		if o.multi {
			expectedPos = operations.RenderPositionBefore
		} else {
			expectedPos = operations.RenderBeforeChangeId
		}
	default:
		expectedPos = operations.RenderPositionBefore
	}
	if pos != expectedPos {
		return ""
	}

	markerStyle := common.DefaultPalette.Get("rebase target_marker")
	markerText := "<< " + o.marker + " >>"
	if o.position == common.PickerInto && !o.multi {
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
