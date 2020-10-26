import React from "react"
import {render} from "react-dom"

import "github-markdown-css/github-markdown.css"
import "style.css"
import App from "containers/App"

render(
  <App />,
  document.querySelector(".container"),
)
