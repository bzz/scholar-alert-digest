import React from "react"
import {render} from "react-dom"

import "style.css"
import {get} from "request"
import Main from "routes/Main"

get("?json").then(papers => {
  render(
    <Main papers={papers} />,
    document.querySelector(".container"),
  )
})
