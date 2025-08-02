function render(data) {
  let content = "<h1>Page 1</h1>";

  if (data && data.request) {
    content += `<p>Request: </p>`;
    content += `<pre><code>${JSON.stringify(data, null, 4)}</code></pre>`;
  }

  return content;
}

function loader(req) {
  return {
    "request": {
      "method": req.Method,
      "path": req.Path,
      "headers": req.Headers,
      "body": req.Body
    }
  };
}
