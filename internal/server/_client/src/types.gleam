import gleam/dynamic/decode
import gleam/option
import gleam/string
import lustre_http

pub type Config {
  Config(profile_name: String, brokers: List(String), topic: String)
}

pub fn config_decoder() -> decode.Decoder(Config) {
  use profile_name <- decode.field("profile_name", decode.string)
  use brokers <- decode.field("brokers", decode.list(decode.string))
  use topic <- decode.field("topic", decode.string)
  decode.success(Config(profile_name:, brokers:, topic:))
}

pub fn display_config(config: Config) -> String {
  [
    " -> profile_name: " <> config.profile_name,
    " -> brokers: " <> config.brokers |> string.join(", "),
    " -> topic: " <> config.topic,
  ]
  |> string.join("\n")
}

pub type Behaviours {
  Behaviours(select_on_hover: Bool)
}

pub fn default_behaviours() -> Behaviours {
  Behaviours(select_on_hover: False)
}

pub fn behaviours_decoder() -> decode.Decoder(Behaviours) {
  use select_on_hover <- decode.field("select_on_hover", decode.bool)
  decode.success(Behaviours(select_on_hover:))
}

pub type MessageOffset =
  Int

pub type MessageDetails {
  MessageDetails(
    key: String,
    offset: Int,
    partition: Int,
    metadata: String,
    value: option.Option(String),
    decode_error: option.Option(String),
    decode_error_fallback: option.Option(String),
  )
}

pub fn message_details_decoder() -> decode.Decoder(MessageDetails) {
  use key <- decode.field("key", decode.string)
  use offset <- decode.field("offset", decode.int)
  use partition <- decode.field("partition", decode.int)
  use metadata <- decode.field("metadata", decode.string)
  use value <- decode.field("value", decode.optional(decode.string))
  use decode_error <- decode.field(
    "decode_error",
    decode.optional(decode.string),
  )
  use decode_error_fallback <- decode.optional_field(
    "decode_error_fallback",
    option.None,
    decode.optional(decode.string),
  )
  decode.success(MessageDetails(
    key:,
    offset:,
    partition:,
    metadata:,
    value:,
    decode_error:,
    decode_error_fallback:,
  ))
}

pub type Msg {
  ConfigFetched(Result(Config, lustre_http.HttpError))
  BehavioursFetched(Result(Behaviours, lustre_http.HttpError))
  FetchMessages(Int)
  ClearMessages
  HoverSettingsChanged(Bool)
  MessageChosen(Int)
  MessagesFetched(Result(List(MessageDetails), lustre_http.HttpError))
  GoToStart
  GoToEnd
}

pub fn dummy_message() -> List(MessageDetails) {
  let key = "20693f56-b784-4594-b79e-38c6d1756035"
  let offset = 1
  let partition = 0
  let metadata =
    "
- key                  41bb43d3-8589-4819-8785-048dc2dd4c8a
- timestamp            2025-04-06 11:18:03.801 +0200 CEST
- partition            0
- offset               0"

  let value =
    "
{
  \"id\": \"41bb43d3-8589-4819-8785-048dc2dd4c8a\",
  \"colorTheme\": \"#43feb0\",
  \"backgroundImageUrl\": \"https://mxrlttzt.nghffmsk.com\",
  \"customDomain\": \"xoubtzfn.com\"
}"
  [
    MessageDetails(
      key: key,
      offset: offset,
      partition: partition,
      metadata: metadata,
      value: option.Some(value),
      decode_error: option.None,
      decode_error_fallback: option.None,
    ),
  ]
}
