import React, {useEffect, useReducer} from "react"
import {get, post} from "request"
import reducer, {actions} from "reducer"

import Labels from "routes/Labels"
import Main from "routes/Main"

const App = () => {
  const [state, dispatch] = useReducer(reducer, {labels: []})
  const {setLabels, setLabel, setPapers} = actions(dispatch)

  useEffect(() => {
    const maybeLabel = localStorage.getItem("label")

    if (maybeLabel) {
      setLabel(maybeLabel)
      post("json/messages", {label: maybeLabel}).then(setPapers)
    } else {
      get("json/labels").then(({labels}) => {
        setLabels(labels)
        localStorage.setItem("labels", JSON.stringify(labels))
      })
    }
  }, [])

  if (state.papers != null) {
    return (
      <Main papers={state.papers} />
    )
  }

  if (state.labels.length > 0) {
    return (
      <Labels labels={state.labels} setLabel={setLabel} setPapers={setPapers} />
    )
  }

  return null
}

export default App
