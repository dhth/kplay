import gleam/int
import gleam/list
import gleam/option
import lustre/attribute
import lustre/element
import lustre/element/html
import lustre/event
import model.{type Model, display_model}
import types.{type MessageDetails, type Msg}
import utils.{http_error_to_string}

pub fn view(model: Model) -> element.Element(Msg) {
  html.div([attribute.class("bg-[#282828] text-[#ebdbb2] mt-4 mx-4")], [
    html.div([], [
      html.div([], [
        model_debug_div(model),
        messages_section(model),
        controls_div(model),
        error_section(model),
      ]),
    ]),
  ])
}

fn model_debug_div(model: Model) -> element.Element(Msg) {
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
  let height = case model.fetch_err {
    option.None -> "h-[91vh]"
    option.Some(_) -> "h-[80vh]"
  }
  let current_index =
    model.current_message
    |> option.map(fn(a) {
      case a {
        #(i, _) -> i
      }
    })
  let content = case model.messages {
    [] -> [
      html.p([attribute.class("text-[#928374] p-4")], [
        element.text(
          "kplay lets you inspect messages in a Kafka topic in a simple and deliberate manner. "
          <> "Click on the buttons below to start fetching messages.",
        ),
      ]),
    ]
    [_, ..] -> [
      html.div([attribute.class("w-2/5 overflow-auto")], [
        html.div([attribute.class("p-4")], [
          html.h2([attribute.class("text-[#fe8019] text-xl font-bold mb-4")], [
            html.text("Messages"),
          ]),
          html.div(
            [],
            model.messages
              |> list.index_map(fn(m, i) {
                message_list_item(m, i, current_index, model.select_on_hover)
              }),
          ),
        ]),
      ]),
      message_details_pane(model),
    ]
  }
  html.div(
    [
      attribute.class(
        "mt-4 " <> height <> " flex border-2 border-[#928374] border-opacity-20",
      ),
    ],
    content,
  )
}

fn message_list_item(
  message: MessageDetails,
  index: Int,
  current_index: option.Option(Int),
  select_on_hover: Bool,
) -> element.Element(Msg) {
  let border_class = case current_index {
    option.Some(i) if i == index -> " text-[#b8bb26] border-l-[#b8bb26]"
    _ -> " text-[#d5c4a1] border-l-[#282828]"
  }
  let event_handler = case select_on_hover {
    False -> event.on_click(types.MessageChosen(index))
    True -> event.on_mouse_over(types.MessageChosen(index))
  }

  html.div(
    [
      attribute.class(
        "py-2 px-4 border-l-2 hover:border-l-[#fe8019]"
        <> " hover:text-[#fe8019] hover:border-l-2 cursor-pointer transition duration-100"
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
        html.p([], [html.text("offset: " <> message.offset |> int.to_string)]),
        html.p([], [
          html.text("partition: " <> message.partition |> int.to_string),
        ]),
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
          case model.select_on_hover {
            True -> "Hover on"
            False -> "Select"
          }
          <> " an entry in the left pane to view details here.",
        ),
      ])
    option.Some(#(_, msg)) ->
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
  }

  html.div([attribute.class("w-3/5 p-6 overflow-auto")], [
    html.h2([attribute.class("text-[#fe8019] text-xl font-bold mb-4")], [
      html.text("Details"),
    ]),
    message_details,
  ])
}

fn controls_div(model: Model) -> element.Element(Msg) {
  html.div([attribute.class("flex items-center space-x-2 mt-4")], [
    html.button(
      [
        attribute.class(
          "font-bold px-4 py-1 bg-[#b8bb26] text-[#282828] hover:bg-[#fe8019]",
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
    html.button(
      [
        attribute.class(
          "font-semibold px-4 py-1 bg-[#d3869b] text-[#282828] hover:bg-[#fabd2f]",
        ),
        attribute.disabled(model.fetching),
        event.on_click(types.FetchMessages(1)),
      ],
      [element.text("Fetch next")],
    ),
    html.button(
      [
        attribute.class(
          "font-semibold px-4 py-1 bg-[#d3869b] text-[#282828] hover:bg-[#fabd2f]",
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
      [attribute.class("font-semibold px-4 py-1 flex items-center space-x-2")],
      [
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
          attribute.checked(model.select_on_hover),
        ]),
      ],
    ),
  ])
}

fn error_section(model: Model) -> element.Element(Msg) {
  case model.fetch_err {
    option.None -> element.none()
    option.Some(err) ->
      html.div(
        [
          attribute.role("alert"),
          attribute.class(
            "text-[#fb4934] border-2 border-[#fb4934] border-opacity-50 px-4 py-4 mt-8",
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
