package tui

import "fmt"

var helpText = fmt.Sprintf(`%s
%s
%s

%s
%s
%s
%s
`,
	helpHeaderStyle.Render("kplay Reference Manual"),
	helpSectionStyle.Render(`
(scroll with j/k/arrow/<c-d>/<c-u>)

kplay has 2 views:
  - Message List and Details View
  - Help View (this one)
`),
	helpHeaderStyle.Render("Keyboard Shortcuts"),
	helpHeaderStyle.Render("General"),
	helpSectionStyle.Render(`
    ?                              Show help view
    q/<esc>                        Go back/quit
    <ctrl+c>                       Quit immediately
`),
	helpHeaderStyle.Render("Message List/Details View"),
	helpSectionStyle.Render(`
    <tab>/<shift-tab>              Switch focus between panes
    j/<Down>                       Select next message/scroll details down
    k/<Up>                         Select previous message/scroll details up
    G                              Select last message/scroll details to bottom
    g                              Select first message/scroll details to top
    <ctrl+d>                       Scroll details half page down
    <ctrl+u>                       Scroll details half page up
    ]                              Select next message
    [                              Select previous message
    n                              Fetch the next message from the topic
    N                              Fetch the next 10 messages from the topic
    }                              Fetch the next 100 messages from the topic
    s                              Toggle skipping mode (if ON, kplay will consume messages,
                                       but not populate its internal list, effectively
                                       skipping over them)
    p                              Toggle persist mode (if ON, kplay will start persisting
                                       messages at the location
                                       messages/<topic>/partition-<partition>/offset-<offset>.txt)
    P                              Persist current message to local filesystem
    y                              Copy message details to clipboard
`),
)
