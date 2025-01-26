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
    n                              Fetch the next record from the topic
    N                              Fetch the next 10 records from the topic
    }                              Fetch the next 100 records from the topic
    s                              Toggle skipping mode; kplay will consume records,
                                       but not populate its internal list, effectively
                                       skipping over them
    p                              Toggle persist mode (kplay will start persisting
                                       records, at the location
                                       records/<topic>/<partition>/<offset>-<key>.md
    y                              Copy message details to clipboard
    [                              Move to previous item in list
    ]                              Move to next item in list
`),
)
