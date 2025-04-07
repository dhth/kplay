import gleam/dynamic/decode
import gleam/int
import lustre/effect
import lustre_http
import plinth/browser/window
import types.{
  type Msg, ConfigFetched, MessagesFetched, config_decoder,
  message_details_decoder,
}

pub fn fetch_config() -> effect.Effect(Msg) {
  let expect = lustre_http.expect_json(config_decoder(), ConfigFetched)

  lustre_http.get(window.location() <> "api/config", expect)
}

pub fn fetch_messages(num: Int) -> effect.Effect(Msg) {
  let expect =
    lustre_http.expect_json(
      decode.list(message_details_decoder()),
      MessagesFetched,
    )

  lustre_http.get(
    window.location() <> "api/fetch?num=" <> num |> int.to_string,
    expect,
  )
}
