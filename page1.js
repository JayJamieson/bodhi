function render(data) {
  let content = "<h1>Page 1</h1>";
  if (data && data.request) {
    content += `<p>Request path: ${data.request.path}</p>`;
    content += `<p>Request method: ${data.request.method}</p>`;
    content += `<button onclick="handleClick1" >click me</button>`;
    content += `<script>function handleClick1(e) {console.log(e)}</script>`

  }
  return content;
}

function loader(req) {
  return {
    "request": {
      "method": "<script>(function handleClick1() {alert('aaaaaa')})()</script>",
      "path": req.Path,
      "headers": req.Headers,
      "body": req.Body
    }
  };
}
