import simplifile
import cx
import gametime/web.{type Context}
import wisp.{type Request, type Response}
import gleam/string_builder
import gleam/http.{Get}
import tagg

pub fn handle_request(req: Request, web_context: Context) -> Response {
  use req <- web.middleware(req)

  // Wisp doesn't have a special router abstraction, instead we recommend using
  // regular old pattern matching. This is faster than a router, is type safe,
  // and means you don't have to learn or be limited by a special DSL.
  //
  case wisp.path_segments(req) {
    // This matches `/`.
    [] -> home_page(req, web_context)

    // This matches `/yuh`.
    ["yuh"] -> yuh(req)

    // This matches `/comments/:id`.
    // The `id` segment is bound to a variable and passed to the handler.
    ["comments", id] -> show_comment(req, id)

    // This matches all other paths.
    _ -> wisp.not_found()
  }
}


fn home_page(req: Request, web_context: Context) -> Response {
  // The home page can only be accessed via GET requests, so this middleware is
  // used to return a 405: Method Not Allowed response for all other methods.
  use <- wisp.require_method(req, Get)

  let context = cx.dict()

  case tagg.render(web_context.tagg, "dashboard.html", context) {
    Ok(html) -> {
      wisp.ok()
      |> wisp.html_body(string_builder.from_string(html))
    }
    Error(_) -> wisp.internal_server_error()
  }
}

fn yuh(req: Request) -> Response {
  use <- wisp.require_method(req, Get)

  let html = string_builder.from_string("
  <h1 class='text-lg font-medium'>
    yuh
  </h1>
  ")
  wisp.ok()
  |> wisp.html_body(html)
}


fn show_comment(req: Request, id: String) -> Response {
  use <- wisp.require_method(req, Get)

  // The `id` path parameter has been passed to this function, so we could use
  // it to look up a comment in a database.
  // For now we'll just include in the response body.
  let html = string_builder.from_string("Comment with id " <> id)
  wisp.ok()
  |> wisp.html_body(html)
}
