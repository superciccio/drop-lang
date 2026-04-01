# Drop

**Write a `.drop` file. Run it. Get a working prototype.**

Drop is a zero-ceremony prototyping language. No boilerplate, no config, no dependencies. If you're typing anything that isn't your idea, Drop has failed.

## Install

```sh
go install github.com/superciccio/drop-lang/cmd/drop@latest
```

Or download a binary from [Releases](https://github.com/superciccio/drop-lang/releases).

## Hello World

```
name = "World"
print "Hello, {name}!"
```

```sh
drop hello.drop
```

## Todo App (Full CRUD + UI in 24 lines)

```
serve 8080
store "todos"

get "/"
  page "My Todos"
    each todo in load "todos"
      row
        text todo.name
        button "Delete" -> delete "/todos/{todo.id}"
    form "New Todo"
      input "name"
      submit "Add" -> post "/todos"

post "/todos"
  save "todos" body
  respond body 201

delete "/todos/:id"
  remove "todos" id
  respond "Deleted"

get "/api/todos"
  respond load "todos"
```

That's it. Run `drop todo.drop`, open `localhost:8080`. You get a styled dark-theme UI, JSON API, and persistent storage.

## Features

### Language

Variables, strings with interpolation, lists, maps, functions, if/else, for/in, comments. No type declarations, no semicolons, no ceremony.

```
colors = ["red", "green", "blue"]
user = {name: "Andrea", role: "builder"}

do greet(who)
  return "Hey, {who}!"

for color in colors
  print color
```

### Web Server

`serve` starts an HTTP server. Define routes with `get`, `post`, `put`, `delete`. Return data with `respond`. Access request body with `body` and route params by name.

```
serve 3000

get "/users/:id"
  respond {id: id, name: "Andrea"}
```

Strings respond as `text/html`. Lists and maps respond as `application/json`. Routes under `/api/*` always return JSON.

### Data Persistence

`store` declares a JSON-backed data store. `save`, `load`, `remove` do what you'd expect. IDs are auto-generated.

```
store "notes"
save "notes" {title: "First note"}
notes = load "notes"
```

### UI Rendering

`page` blocks build styled HTML with zero CSS. Use `text`, `row`, `each`, `button`, `form`, `input`, `submit`, `link`, and `image`. The `->` operator connects UI actions to routes.

```
page "Dashboard"
  text "Welcome back"
  button "Refresh" -> get "/data"
```

Hit any route from a browser and lists/maps auto-render as styled HTML tables and cards.

### Fetch

Pull data from external APIs. JSON responses are parsed automatically.

```
data = fetch "https://api.example.com/items"
```

### Hot Reload

Server apps auto-restart when you save the `.drop` file. No manual restarts during development.

## Editor Support

Install syntax highlighting for your editor:

```sh
drop --editor vscode
drop --editor vim
drop --editor sublime
drop --editor zed
```

## Docs

Full language reference and guides at [superciccio.github.io/drop-lang](https://superciccio.github.io/drop-lang/).

## License

[MIT](LICENSE)
