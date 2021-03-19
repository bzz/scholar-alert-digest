import React, {useReducer, useMemo} from "react"
import {render} from "react-dom"

import "github-markdown-css/github-markdown.css"
import "style.css"
import reducer, {defaultState, actions} from "reducer"
import initApp from "containers/App"

const Container = () => {
  const [state, dispatch] = useReducer(reducer, defaultState)
  const App = useMemo(() => initApp(actions(dispatch)), [])

  return (
    <App state={state} />
  )
}

render(
  <Container />,
  document.querySelector(".container"),
)
