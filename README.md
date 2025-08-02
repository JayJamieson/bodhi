# bodhi

Lightweight Javascript powered web framework. Inspired by [Jar](https://github.com/healeycodes/jar) and [Remix](https://github.com/remix-run/remix).

The core principle for bodhi is to reduce complexity as much as possible. Have a single executable that can be run without additional dependencies and complexities such such as PHP, NodeJS etc.

Go's `net/http` is a capable enough web/application server and no extras are needed (apache/httpd, nginx, php-fpm).

> [!NOTE]
> The current version of bodhi doesn't provide access methods to databases, http requests, host os (env vars, files) and third part libraries. The Javascript VM is
> also limited to ~ES6 features [sobek](github.com/grafana/sobek) provides.

## Usage

Running `bodhi ./page` will start bodhi on `localhost:8080`. Add the following `index.tmpl` and `index.js` file and away you go.

Required entrypoint **./pages/index.tmpl**

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Bodhi Framework</title>
</head>
<body>
    <div class="container">
        <ul>
        <li><a href="/">home</a></li>
        {{range .pages}}
                    <li><a href="{{.}}">{{.}}</a></li>
        {{end}}
        </ul>
        {{template "content"}}
    </div>
</body>
</html>
```

Required index page render script **./pages/index.js**

```js
function render() {
  return `
  <h1>Hello World from Bodhi!</h1>
  <p>This is a demo of using Javascript and Go <code>htmp/template</code> to create a webframework.</p>
  <p>All you need is the bodhi executable, <code>index.tmpl</code> and any number of <code>*.js</code> page files to get started.</p>`;
}
```

### Data Loader

Page files can have an optional `loader(req)` function defined that provides access to the current request object.

## Building

  go build

## Installing

  go get github.com/JayJamieson/bodhi

## Examples

See [examples directory](https://github.com/JayJamieson/bodhi/tree/master/examples)
