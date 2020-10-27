import React, {useEffect, useReducer} from "react"
import {get, post} from "request"
import reducer, {actions} from "reducer"

import Labels from "containers/Labels"
import Main from "containers/Main"

const App = () => {
  const [state, dispatch] = useReducer(reducer, {labels: []})
  const {setLabels, setLabel, setPapers} = actions(dispatch)

  useEffect(() => {
    const maybeLabel = localStorage.getItem("label")

    const login = url => {
      window.location = url
    }

    const handleError = e => {
      if (e.status === 401) {
        login(e.payload.Redirect)
      }
    }

    if (maybeLabel) {
      setLabel(maybeLabel)
      post("json/messages", {label: maybeLabel})
        .then(setPapers)
        .catch(handleError)
    } else {
      get("json/labels")
        .then(({labels}) => {
          setLabels(labels)
          localStorage.setItem("labels", JSON.stringify(labels))
        })
        .catch(handleError)
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
