import gleam/int
import gleam/list
import gleam/option
import lustre/attribute
import lustre/element
import lustre/element/html
import lustre/event
import model.{type Model, display_model}
import types.{type Config, type MessageDetails, type Msg}
import utils.{http_error_to_string}

pub fn view(model: Model) -> element.Element(Msg) {
  html.div([attribute.class("bg-[#282828] text-[#ebdbb2] mx-4")], [
    html.div([], [
      html.div([attribute.class("flex flex-col h-screen")], [
        model_debug_section(model),
        messages_section(model),
        controls_section(model),
        error_section(model),
      ]),
    ]),
  ])
}

fn model_debug_section(model: Model) -> element.Element(Msg) {
  case model.debug {
    True ->
      html.div(
        [attribute.class("debug bg-gray-800 text-white p-4 overflow-auto mb-5")],
        [
          html.pre([attribute.class("whitespace-pre-wrap")], [
            model |> display_model |> element.text,
          ]),
        ],
      )
    False -> element.none()
  }
}

fn messages_section(model: Model) -> element.Element(Msg) {
  case model.messages {
    [] -> landing_section()
    [_, ..] -> messages_section_with_messages(model)
  }
}

fn landing_section() -> element.Element(Msg) {
  html.div(
    [
      attribute.class(
        "mt-4 flex-1 "
        <> " flex border-2 border-[#928374] border-opacity-20 items-center flex justify-center",
      ),
    ],
    [
      html.pre([attribute.class("text-[#928374]")], [
        element.text(
          "
kkkkkkkk                               lllllll                                         
k::::::k                               l:::::l                                         
k::::::k                               l:::::l                                         
k::::::k                               l:::::l                                         
 k:::::k    kkkkkkkppppp   ppppppppp    l::::l   aaaaaaaaaaaaayyyyyyy           yyyyyyy
 k:::::k   k:::::k p::::ppp:::::::::p   l::::l   a::::::::::::ay:::::y         y:::::y 
 k:::::k  k:::::k  p:::::::::::::::::p  l::::l   aaaaaaaaa:::::ay:::::y       y:::::y  
 k:::::k k:::::k   pp::::::ppppp::::::p l::::l            a::::a y:::::y     y:::::y   
 k::::::k:::::k     p:::::p     p:::::p l::::l     aaaaaaa:::::a  y:::::y   y:::::y    
 k:::::::::::k      p:::::p     p:::::p l::::l   aa::::::::::::a   y:::::y y:::::y     
 k:::::::::::k      p:::::p     p:::::p l::::l  a::::aaaa::::::a    y:::::y:::::y      
 k::::::k:::::k     p:::::p    p::::::p l::::l a::::a    a:::::a     y:::::::::y       
k::::::k k:::::k    p:::::ppppp:::::::pl::::::la::::a    a:::::a      y:::::::y        
k::::::k  k:::::k   p::::::::::::::::p l::::::la:::::aaaa::::::a       y:::::y         
k::::::k   k:::::k  p::::::::::::::pp  l::::::l a::::::::::aa:::a     y:::::y          
kkkkkkkk    kkkkkkk p::::::pppppppp    llllllll  aaaaaaaaaa  aaaa    y:::::y           
                    p:::::p                                         y:::::y            
                    p:::::p                                        y:::::y             
                   p:::::::p                                      y:::::y              
                   p:::::::p                                     y:::::y               
                   p:::::::p                                    yyyyyyy                
                   ppppppppp

kplay lets you inspect messages in a Kafka topic in a simple and deliberate manner

            Click on the buttons below to start fetching messages
",
        ),
      ]),
    ],
  )
}

fn messages_section_with_messages(model: Model) -> element.Element(Msg) {
  let current_index =
    model.current_message
    |> option.map(fn(a) {
      case a {
        #(i, _) -> i
      }
    })

  html.div(
    [
      attribute.class(
        "mt-4 flex-1 flex border-2 border-[#928374] border-opacity-20 overflow-y-auto",
      ),
    ],
    [
      html.div([attribute.class("w-2/5 overflow-auto")], [
        html.div([attribute.class("p-4")], [
          html.h2([attribute.class("text-[#fe8019] text-xl font-bold mb-4")], [
            html.text("Messages"),
          ]),
          html.div(
            [],
            model.messages
              |> list.index_map(fn(m, i) {
                message_list_item(
                  m,
                  i,
                  current_index,
                  model.behaviours.select_on_hover,
                )
              }),
          ),
        ]),
      ]),
      message_details_pane(model),
    ],
  )
}

fn message_list_item(
  message: MessageDetails,
  index: Int,
  current_index: option.Option(Int),
  select_on_hover: Bool,
) -> element.Element(Msg) {
  let border_class = case current_index {
    option.Some(i) if i == index -> " text-[#fe8019] border-l-[#fe8019]"
    _ -> " text-[#d5c4a1] border-l-[#282828]"
  }
  let event_handler = case select_on_hover {
    False -> event.on_click(types.MessageChosen(index))
    True -> event.on_mouse_over(types.MessageChosen(index))
  }

  html.div(
    [
      attribute.class(
        "py-2 px-4 border-l-2 hover:border-l-[#b8bb26]"
        <> " hover:text-[#b8bb26] hover:border-l-2 cursor-pointer transition duration-100"
        <> " ease-in-out"
        <> border_class,
      ),
      event_handler,
    ],
    [
      html.p([attribute.class("text-base font-semibold")], [
        html.text(message.key),
      ]),
      html.div([attribute.class("flex space-x-2 text-sm")], [
        html.p([], [
          html.text("offset: " <> message.offset |> int.to_string),
        ]),
        html.p([], [
          html.text("partition: " <> message.partition |> int.to_string),
        ]),
        case message.decode_error {
          option.None -> html.p([], [])
          option.Some(_) ->
            html.p([], [
              html.text("decode error"),
            ])
        },
        case message.value {
          option.None -> html.p([], [html.text("🪦")])
          option.Some(_) -> element.none()
        },
      ]),
    ],
  )
}

fn message_details_pane(model: Model) -> element.Element(Msg) {
  let message_details = case model.current_message {
    option.None ->
      html.p([attribute.class("text-[#928374]")], [
        html.text(
          case model.behaviours.select_on_hover {
            True -> "Hover on"
            False -> "Select"
          }
          <> " an entry in the left pane to view details here.",
        ),
      ])
    option.Some(#(_, msg)) ->
      case msg.decode_error {
        option.None ->
          html.div([], [
            html.p([attribute.class("text-[#fabd2f] text-lg mb-4")], [
              html.text("Metadata"),
            ]),
            html.pre([attribute.class("text-[#d5c4a1] text-base mb-8")], [
              html.text(msg.metadata),
            ]),
            html.p([attribute.class("text-[#fabd2f] text-lg mb-4")], [
              html.text("Value"),
            ]),
            case msg.value {
              option.None -> html.p([], [html.text("tombstone 🪦")])
              option.Some(v) ->
                html.pre([attribute.class("text-[#d5c4a1] text-base mb-4")], [
                  html.text(v),
                ])
            },
          ])
        option.Some(e) ->
          html.pre([attribute.class("text-[#fb4934] text-base mb-4")], [
            html.text(e),
          ])
      }
  }

  html.div([attribute.class("w-3/5 p-6 overflow-auto")], [
    html.h2([attribute.class("text-[#fe8019] text-xl font-bold mb-4")], [
      html.text("Details"),
    ]),
    message_details,
  ])
}

fn controls_section(model: Model) -> element.Element(Msg) {
  case model.config {
    option.Some(c) -> controls_div_with_config(model, c)
    option.None -> controls_div_when_no_config()
  }
}

fn controls_div_when_no_config() -> element.Element(Msg) {
  html.div([attribute.class("flex items-center space-x-2 mt-4")], [
    html.button(
      [
        attribute.class(
          "font-bold px-4 py-1 bg-[#fe8019] text-[#282828] hover:bg-[#b8bb26] cursor-pointer",
        ),
        attribute.disabled(True),
      ],
      [
        html.a(
          [
            attribute.href("https://github.com/dhth/kplay"),
            attribute.target("_blank"),
          ],
          [element.text("kplay")],
        ),
      ],
    ),
    html.p([attribute.class("text-[#bdae93]")], [
      element.text("couldn't load config; make sure "),
      html.code([attribute.class("text-[#fe8019]")], [
        element.text("kplay serve"),
      ]),
      element.text(" is running. If it still doesn't work let "),
      html.a(
        [
          attribute.class("text-[#fabd2f]"),
          attribute.href("https://github.com/dhth"),
          attribute.target("_blank"),
        ],
        [element.text("@dhth")],
      ),
      element.text(" know about this error via "),
      html.a(
        [
          attribute.class("text-[#fabd2f]"),
          attribute.href("https://github.com/dhth/kplay/issues"),
          attribute.target("_blank"),
        ],
        [element.text("https://github.com/dhth/kplay/issues")],
      ),
    ]),
  ])
}

fn controls_div_with_config(
  model: Model,
  config: types.Config,
) -> element.Element(Msg) {
  html.div([attribute.class("flex items-center space-x-2 mt-4")], [
    html.button(
      [
        attribute.class(
          "font-bold px-4 py-1 bg-[#fe8019] text-[#282828] hover:bg-[#b8bb26] cursor-pointer",
        ),
        attribute.disabled(True),
      ],
      [
        html.a(
          [
            attribute.href("https://github.com/dhth/kplay"),
            attribute.target("_blank"),
          ],
          [element.text("kplay")],
        ),
      ],
    ),
    consumer_info(config),
    html.button(
      [
        attribute.class(
          "font-semibold px-4 py-1 bg-[#b8bb26] text-[#282828] hover:bg-[#fabd2f] disabled:bg-[#bdae93]",
        ),
        attribute.disabled(model.fetching),
        event.on_click(types.FetchMessages(1)),
      ],
      [element.text("Fetch next")],
    ),
    html.button(
      [
        attribute.class(
          "font-semibold px-4 py-1 bg-[#b8bb26] text-[#282828] hover:bg-[#fabd2f] disabled:bg-[#bdae93]",
        ),
        attribute.disabled(model.fetching),
        event.on_click(types.FetchMessages(10)),
      ],
      [element.text("Fetch next 10")],
    ),
    html.button(
      [
        attribute.class(
          "font-semibold px-4 py-1 bg-[#bdae93] text-[#282828] hover:bg-[#fabd2f]",
        ),
        attribute.disabled(model.fetching),
        event.on_click(types.ClearMessages),
      ],
      [element.text("Clear Messages")],
    ),
    html.div(
      [
        attribute.class(
          "border-2 border-[#928374] border-opacity-40 border-dashed font-semibold px-4 py-1 flex items-center space-x-4",
        ),
      ],
      [
        html.div([attribute.class("flex items-center space-x-2")], [
          html.label(
            [
              attribute.class("cursor-pointer"),
              attribute.for("hover-control-input"),
            ],
            [element.text("select on hover")],
          ),
          html.input([
            attribute.class(
              "w-4 h-4 text-[#fabd2f] bg-[#282828] focus:ring-[#fabd2f] cursor-pointer",
            ),
            attribute.id("hover-control-input"),
            attribute.type_("checkbox"),
            event.on_check(types.HoverSettingsChanged),
            attribute.checked(model.behaviours.select_on_hover),
          ]),
        ]),
      ],
    ),
  ])
}

fn consumer_info(config: Config) -> element.Element(Msg) {
  html.div(
    [attribute.class("font-bold px-4 py-1 flex items-center space-x-2")],
    [
      html.div([attribute.class("relative group")], [
        html.p(
          [
            attribute.class("text-nowrap w-48 text-[#fabd2f] overflow-clip"),
          ],
          [element.text(config.topic)],
        ),
        html.div(
          [
            attribute.class(
              "absolute left-1/2 -translate-x-1/2 bottom-full mb-2 hidden group-hover:block bg-[#a89984] text-[#282828] text-xs px-2 py-1 text-center",
            ),
          ],
          [html.text(config.topic)],
        ),
      ]),
    ],
  )
}

fn error_section(model: Model) -> element.Element(Msg) {
  case model.http_error {
    option.None -> element.none()
    option.Some(err) ->
      html.div(
        [
          attribute.role("alert"),
          attribute.class(
            "text-[#fb4934] border-2 border-[#fb4934] border-opacity-50 px-4 py-4 my-4",
          ),
        ],
        [
          html.strong([attribute.class("font-bold")], [html.text("Error: ")]),
          html.span([attribute.class("block sm:inline")], [
            html.text(err |> http_error_to_string),
          ]),
        ],
      )
  }
}
