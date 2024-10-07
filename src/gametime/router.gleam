import simplifile
import wisp.{type Request, type Response}
import gleam/string_builder
import gleam/http.{Get}
import gametime/web

pub fn handle_request(req: Request) -> Response {
  use req <- web.middleware(req)

  // Wisp doesn't have a special router abstraction, instead we recommend using
  // regular old pattern matching. This is faster than a router, is type safe,
  // and means you don't have to learn or be limited by a special DSL.
  //
  case wisp.path_segments(req) {
    // This matches `/`.
    [] -> home_page(req)

    // This matches `/yuh`.
    ["yuh"] -> yuh(req)

    // This matches `/comments/:id`.
    // The `id` segment is bound to a variable and passed to the handler.
    ["comments", id] -> show_comment(req, id)

    // This matches all other paths.
    _ -> wisp.not_found()
  }
}


fn home_page(req: Request) -> Response {
  // The home page can only be accessed via GET requests, so this middleware is
  // used to return a 405: Method Not Allowed response for all other methods.
  use <- wisp.require_method(req, Get)

  let html = string_builder.from_string("
<!DOCTYPE html>
<html>
    <head>
        <script src='https://unpkg.com/htmx.org@2.0.3' integrity='sha384-0895/pl2MU10Hqc6jd4RvrthNlDiE9U1tWmX7WRESftEDRosgxNsQG/Ze9YMRzHq' crossorigin='anonymous'></script>
    </head>
    <body>
        <button hx-get='/yuh' hx-target='#content'>
            Hello!
        </button>
        <div id='content'></div>
    </body>
</html>
  ")

  wisp.ok()
  |> wisp.html_body(html)
}

fn yuh(req: Request) -> Response {
  use <- wisp.require_method(req, Get)

  let html = string_builder.from_string("yuh")
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
