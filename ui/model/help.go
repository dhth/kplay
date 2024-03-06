package model

var (
	helpText = `
           kplay has three sections:
           - Record List View
           - Record Metadata View
           - Record Value View

           Keyboard Shortcuts:

           General
           <tab>       Switch focus to next section
           <s-tab>     Switch focus to previous section

           List View
           h/<Up>      Move cursor up
           k/<Down>    Move cursor down
           n           Fetch the next record from the topic
           N           Fetch the next 10 records from the topic
           }           Fetch the next 100 records from the topic
           s           Toggle skipping mode; kplay will consume records,
                           but not populate its internal list, effectively
                           skipping over them
           p           Toggle persist mode (kplay will start persisting
                           records, at the location
                           records/<topic>/<partition>/<offset>-<key>.md


           Message Metadata/Value View   
           f           Toggle focussed section between full screen and
                           regular mode
           1           Maximize record metadata view
           2           Maximize record value view
           q           Minimize section, and return focus to list view
           [           Show details for the previous entry in the list
           ]           Show details for the next entry in the list
`
)
