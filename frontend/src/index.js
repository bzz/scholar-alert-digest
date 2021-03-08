import React, {useReducer} from "react"
import {render} from "react-dom"

import "github-markdown-css/github-markdown.css"
import "style.css"
import reducer, {defaultState, actions} from "reducer"
import App from "containers/App"

const Container = () => {
  const [state, dispatch] = useReducer(reducer, defaultState)
  const {setView, setLabels, setLabel, setPapers, setMode} = actions(dispatch)

  return (
    <App
      state={state}
      setView={setView}
      setLabels={setLabels}
      setLabel={setLabel}
      setPapers={setPapers}
      setMode={setMode}
    />
  )
}

render(
  <Container />,
  document.querySelector(".container"),
)
