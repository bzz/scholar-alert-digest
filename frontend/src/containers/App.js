import React, {useEffect} from "react"
import PropTypes from "prop-types"
import {init, changeLabel} from "effects"

import Labels from "containers/Labels"
import Report from "containers/Report"

const App = ({state, setLabels, setLabel, setPapers, toggleMode}) => {
  useEffect(() => init({setLabels, setLabel, setPapers}), [])

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

const paperProps = PropTypes.shape({
  paper: PropTypes.shape({
    URL: PropTypes.string.isRequired,
    Title: PropTypes.string.isRequired,
    Author: PropTypes.string.isRequired,
    Refs: PropTypes.arrayOf(PropTypes.string).isRequired,
    Abstract: PropTypes.shape({
      FirstLine: PropTypes.string.isRequired,
      Rest: PropTypes.string.isRequired,
    }).isRequired,
  }),
})

App.propTypes = {
  toggleMode: PropTypes.func.isRequired,
  setPapers: PropTypes.func.isRequired,
  setLabel: PropTypes.func.isRequired,
  setLabels: PropTypes.func.isRequired,
  state: PropTypes.shape({
    currentLabel: PropTypes.string,
    labels: PropTypes.arrayOf(PropTypes.string).isRequired,
    mode: PropTypes.string.isRequired,
    papers: PropTypes.shape({
      read: PropTypes.shape({
        papers: PropTypes.arrayOf(paperProps),
      }).isRequired,
      unread: PropTypes.shape({
        papers: PropTypes.arrayOf(paperProps),
        stats: PropTypes.shape({
          messages: PropTypes.number,
          papers: PropTypes.number,
          time: PropTypes.string,
        }),
      }).isRequired,
    }).isRequired,
  }).isRequired,
}

export default App
