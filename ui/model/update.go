package model

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dhth/kplay/ui/model/generated"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const useHighPerformanceRenderer = false

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	m.msg = ""
	m.errorMsg = ""

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "n", " ":
			return m, FetchNextKMsg(m.kCl, 1)
		case "N":
			return m, FetchNextKMsg(m.kCl, 10)
		case "}":
			return m, FetchNextKMsg(m.kCl, 100)
		case "f":
			switch m.activeView {
			case kMsgMetadataView:
				switch m.vpFullScreen {
				case false:
					m.msgMetadataVP.Height = m.terminalHeight - 8
					m.vpFullScreen = true
				case true:
					m.msgMetadataVP.Height = 12
					m.vpFullScreen = false
				}
			case kMsgValueView:
				switch m.vpFullScreen {
				case false:
					m.msgValueVP.Height = m.terminalHeight - 8
					m.vpFullScreen = true
				case true:
					m.msgValueVP.Height = 12
					m.vpFullScreen = false
				}
				return m, tea.Batch(cmds...)
			}
		case "tab":
			if m.activeView == kMsgsListView {
				m.activeView = kMsgMetadataView
			} else if m.activeView == kMsgMetadataView {
				m.activeView = kMsgValueView
			} else if m.activeView == kMsgValueView {
				m.activeView = kMsgsListView
			}
		case "shift+tab":
			if m.activeView == kMsgsListView {
				m.activeView = kMsgValueView
			} else if m.activeView == kMsgMetadataView {
				m.activeView = kMsgsListView
			} else if m.activeView == kMsgValueView {
				m.activeView = kMsgMetadataView
			}
		}

	case tea.WindowSizeMsg:
		_, h := stackListStyle.GetFrameSize()
		m.terminalHeight = msg.Height
		m.terminalWidth = msg.Width
		m.kMsgsList.SetHeight(msg.Height - h - 2)

		if !m.msgMetadataVPReady {
			m.msgMetadataVP = viewport.New(120, m.terminalHeight/2-8)
			m.msgMetadataVP.HighPerformanceRendering = useHighPerformanceRenderer
			m.msgMetadataVPReady = true
		} else {
			m.msgMetadataVP.Width = 120
			m.msgMetadataVP.Height = 12
		}

		if !m.msgValueVPReady {
			m.msgValueVP = viewport.New(120, m.terminalHeight/2-8)
			m.msgValueVP.HighPerformanceRendering = useHighPerformanceRenderer
			m.msgValueVPReady = true
		} else {
			m.msgValueVP.Width = 120
			m.msgValueVP.Height = 12
		}

	case KMsgChosenMsg:
		// value
		if len(msg.item.record.Value) == 0 {
			m.msgValueVP.SetContent("Tombstone" + " " + string(msg.item.record.Key))
		} else {
			message := &generated.ApplicationState{}
			if err := proto.Unmarshal(msg.item.record.Value, message); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to deserialize message")
			} else {
				jsonData, err := protojson.Marshal(message)
				if err != nil {
					m.errorMsg = "Failed to marshal message to JSON"
				} else {
					var cont bytes.Buffer
					err = json.Indent(&cont, jsonData, "", "    ")
					c := cont.String()
					m.msgValueVP.SetContent(c)
				}
			}
		}
		// metadata
		var headers string
		var other string
		other += fmt.Sprintf("%s: %s\n", RightPadTrim("timestamp", 20), msg.item.record.Timestamp)
		other += fmt.Sprintf("%s: %d\n", RightPadTrim("partition", 20), msg.item.record.Partition)
		other += fmt.Sprintf("%s: %d\n", RightPadTrim("offset", 20), msg.item.record.Offset)
		for _, h := range msg.item.record.Headers {
			headers += fmt.Sprintf("%s: %s\n", RightPadTrim(h.Key, 20), string(h.Value))
		}
		metadata := fmt.Sprintf("%s\nHeaders:\n%s", other, headers)
		m.msgMetadataVP.SetContent(metadata)

		return m, tea.Batch(cmds...)

	case KMsgFetchedMsg:
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		} else {
			if len(msg.records) == 0 {
				m.msg = "No new messages found"
			} else {
				for _, rec := range msg.records {
					m.kMsgsList.InsertItem(len(m.kMsgsList.Items()), KMsgItem{record: *rec})
				}
				m.msg = fmt.Sprintf("%d message(s) fetched", len(msg.records))

			}
		}

	}

	switch m.activeView {
	case kMsgsListView:
		m.kMsgsList, cmd = m.kMsgsList.Update(msg)
		cmds = append(cmds, cmd)
	case kMsgMetadataView:
		m.msgMetadataVP, cmd = m.msgMetadataVP.Update(msg)
		cmds = append(cmds, cmd)
	case kMsgValueView:
		m.msgValueVP, cmd = m.msgValueVP.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
