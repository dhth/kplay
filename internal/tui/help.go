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
  - Help Pane (this one)
`),
	helpHeaderStyle.Render("Keyboard Shortcuts"),
	helpHeaderStyle.Render("General"),
	helpSectionStyle.Render(`
    ?                              Show help view
    q                              Go back/quit
    Q                              Quit from anywhere
`),
	helpHeaderStyle.Render("Message List and Details View"),
	helpSectionStyle.Render(`
    <tab>/<shift-tab>              Switch focus between panes
    j/<Down>                       Move cursor/details pane down
    k/<Up>                         Move cursor/details pane up
    n                              Fetch the next message from the topic
    N                              Fetch the next 10 messages from the topic
    }                              Fetch the next 100 messages from the topic
    s                              Toggle skipping mode; kplay will consume messages,
                                       but not populate its internal list, effectively
                                       skipping over them
    p                              Toggle persist mode (kplay will start persisting
                                       messages at the location
                                       messages/<topic>/<partition>/<offset>-<key>.txt
    c                              Toggle commit mode (whether to commit messages back
                                        to Kafka or not)
    y                              Copy message details to clipboard
    [                              Move to previous item in list
    ]                              Move to next item in list
`),
)
