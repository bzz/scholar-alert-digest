import React, {useEffect, useReducer} from "react"
import reducer, {defaultState, actions} from "reducer"
import {init, changeLabel} from "effects"

import Labels from "containers/Labels"
import Report from "containers/Report"

const App = () => {
  const [state, dispatch] = useReducer(reducer, defaultState)
  const {setLabels, setLabel, setPapers, toggleMode} = actions(dispatch)

  useEffect(() => {
    init({setLabels, setLabel, setPapers})
  }, [])

  if (state.currentLabel) {
    const {stats, papers} = state.papers.unread

    return (
      <Report
        stats={stats}
        papers={papers}
        label={state.currentLabel}
        changeLabel={changeLabel({setLabels, setLabel})}
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
