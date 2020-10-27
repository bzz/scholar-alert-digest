import React, {useEffect, useReducer} from "react"
import {get, post} from "request"
import reducer, {defaultState, actions} from "reducer"

import Labels from "containers/Labels"
import Main from "containers/Main"

const App = () => {
  const [state, dispatch] = useReducer(reducer, defaultState)
  const {setLabels, setLabel, setPapers, toggleMode} = actions(dispatch)

  const login = url => {
    window.location = url
  }

  const handleError = e => {
    if (e.status === 401) {
      login(e.payload.Redirect)
    }
  }

  useEffect(() => {
    const maybeLabel = localStorage.getItem("label")

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

  const changeLabel = _ => {
    setLabel(null)
    get("json/labels")
      .then(({labels}) => {
        setLabels(labels)
        localStorage.setItem("labels", JSON.stringify(labels))
      })
      .catch(handleError)
  }

  if (state.currentLabel) {
    return (
      <Main
        papers={state.papers}
        label={state.currentLabel}
        changeLabel={changeLabel}
        mode={state.mode}
        toggleMode={toggleMode}
      />
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
