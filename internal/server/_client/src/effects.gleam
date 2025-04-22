import gleam/dynamic/decode
import gleam/int
import lustre/effect
import lustre_http
import plinth/browser/window
import types.{behaviours_decoder, config_decoder, message_details_decoder}

const dev = False

fn base_url() -> String {
  case dev {
    False -> window.location()
    True -> "http://127.0.0.1:6500/"
  }
}

pub fn fetch_config() -> effect.Effect(types.Msg) {
  let expect = lustre_http.expect_json(config_decoder(), types.ConfigFetched)

  lustre_http.get(base_url() <> "api/config", expect)
}

pub fn fetch_behaviours() -> effect.Effect(types.Msg) {
  let expect =
    lustre_http.expect_json(behaviours_decoder(), types.BehavioursFetched)

  lustre_http.get(base_url() <> "api/behaviours", expect)
}

pub fn fetch_messages(num: Int, commit: Bool) -> effect.Effect(types.Msg) {
  let expect =
    lustre_http.expect_json(
      decode.list(message_details_decoder()),
      types.MessagesFetched,
    )

  let commit_query_param = case commit {
    False -> "false"
    True -> "true"
  }

  lustre_http.get(
    base_url()
      <> "api/fetch?num="
      <> num |> int.to_string
      <> "&commit="
      <> commit_query_param,
    expect,
  )
}
