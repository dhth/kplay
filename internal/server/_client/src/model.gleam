import gleam/bool
import gleam/dict
import gleam/int
import gleam/list
import gleam/option
import lustre_http
import types.{
  type Behaviours, type Config, type MessageDetails, default_behaviours,
  display_config,
}

pub type Model {
  Model(
    config: option.Option(Config),
    behaviours: Behaviours,
    messages: List(MessageDetails),
    messages_cache: dict.Dict(Int, MessageDetails),
    http_error: option.Option(lustre_http.HttpError),
    current_message: option.Option(#(Int, MessageDetails)),
    fetching: Bool,
    debug: Bool,
  )
}

pub fn display_model(model: Model) -> String {
  let config = case model.config {
    option.None -> "empty"
    option.Some(c) -> c |> display_config
  }
  let current_message_index =
    model.current_message
    |> option.map(fn(a) {
      case a {
        #(i, _) -> i |> int.to_string
      }
    })
    |> option.unwrap("none")

  "- config: \n"
  <> config
  <> "\n"
  <> "- current_index: "
  <> current_message_index
  <> "\n"
  <> "- fetching: "
  <> bool.to_string(model.fetching)
}

pub fn test_init_model() -> Model {
  let messages =
    [1, 2, 3, 4, 5] |> list.flat_map(fn(_) { types.dummy_message() })
  let messages_cache =
    messages |> list.index_map(fn(m, i) { #(i, m) }) |> dict.from_list

  Model(
    config: option.None,
    behaviours: default_behaviours(),
    messages: messages,
    messages_cache: messages_cache,
    http_error: option.None,
    current_message: option.None,
    fetching: False,
    debug: True,
  )
}

pub fn init_model() -> Model {
  Model(
    config: option.None,
    behaviours: default_behaviours(),
    messages: [],
    messages_cache: dict.new(),
    http_error: option.None,
    current_message: option.None,
    fetching: False,
    debug: False,
  )
}
