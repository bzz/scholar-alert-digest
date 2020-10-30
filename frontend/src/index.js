import React, {useReducer} from "react"
import {render} from "react-dom"

import "github-markdown-css/github-markdown.css"
import "style.css"
import reducer, {defaultState, actions} from "reducer"
import App from "containers/App"

const Container = () => {
  const [state, dispatch] = useReducer(reducer, defaultState)
  const {setLabels, setLabel, setPapers, toggleMode} = actions(dispatch)

  return (
    <App
      state={state}
      setLabels={setLabels}
      setLabel={setLabel}
      setPapers={setPapers}
      toggleMode={toggleMode}
    />
  )
}

render(
  <Container />,
  document.querySelector(".container"),
)
