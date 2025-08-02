# bodhi

Lightweight Javascript powered web framework. Inspired by [Jar](https://github.com/healeycodes/jar).

## Specification

HTTP Server with a catchall route serving html template files from `rootDir`. Every `*.js` file is treated as a route with `index.tmpl` as the base template for all route files.

Given the following list of files in `rootDir`, `index.js`, `page1.js`, `page2.js` will yeild the following navigatible routes:

- `/`
- `/page1`
- `/page2`

### Page file

A page file has a required `render` function that takes in a data object and returns a html string.

To get data into the `render` function an optional `loader` function can be defined. The `loader` function has a optional HTTP request parameter giving access to the current request or can return plain JSON object.

When a page is requesed bodhi will look for `loader` function and if found, run it and pass the result to the `render` function. After execution of the render function the response is sent to the browser.

Execution is handled by [sobek](https://github.com/grafana/sobek) Javascript VM. Currently this is limited to using ~ES6 compatible Javascript features only.

The following are valid page file examples:

Page with render only function no loader.

```js
// index.js

function render() {
  return "<h1>Hello World</h1>"
}
```

Page with render and loader function.

```js
// page1.js
function render(data) {
  content = "<h1>Hello World</h1>"
  content += `<p>request path: ${data.request.path}</p>`

  return
}

// page1.js
function loader(req) {
  return {
        "request": JSON.stringify({
            "method": req.method,
            "path": req.path,
            "headers": req.headers,
            "body": req.body
        })
    }
}
```
