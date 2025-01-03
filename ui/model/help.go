package model

import "fmt"

var (
	helpText = fmt.Sprintf(`
  %s
%s
  %s

  %s
%s
  %s
%s
  %s
%s
`,
		helpHeaderStyle.Render("kplay Reference Manual"),
		helpSectionStyle.Render(`
  (scroll line by line with j/k/arrow keys or by half a page with <c-d>/<c-u>)

  kplay has 4 views:
  - Record List View
  - Record Metadata View
  - Record Value View
  - Help View (this one)
`),
		helpHeaderStyle.Render("Keyboard Shortcuts"),
		helpHeaderStyle.Render("General"),
		helpSectionStyle.Render(`
      <tab>                          Switch focus to next section
      <s-tab>                        Switch focus to previous section
      ?                              Show help view
`),
		helpHeaderStyle.Render("List View"),
		helpSectionStyle.Render(`
      j/<Up>                         Move cursor down
      k/<Down>                       Move cursor up
      J                              Scroll record value view down
      K                              Scroll record value view up
      n                              Fetch the next record from the topic
      N                              Fetch the next 10 records from the topic
      }                              Fetch the next 100 records from the topic
      s                              Toggle skipping mode; kplay will consume records,
                                         but not populate its internal list, effectively
                                         skipping over them
      p                              Toggle persist mode (kplay will start persisting
                                         records, at the location
                                         records/<topic>/<partition>/<offset>-<key>.md
`),
		helpHeaderStyle.Render("Message Metadata/Value View"),
		helpSectionStyle.Render(`
      j/<Up>                         Scroll down
      k/<Down>                       Scroll up
      f                              Toggle focussed section between full screen and
                                         regular mode
      1                              Maximize record metadata view
      2                              Maximize record value view
      q                              Minimize section, and return focus to list view
      [                              Show details for the previous entry in the list
      ]                              Show details for the next entry in the list
`),
	)
)
