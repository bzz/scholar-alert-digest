import React from "react"
import PropTypes from "prop-types"
import {init, changeLabel} from "effects"
import {views} from "constants"

import initLabels from "containers/Labels"
import initReport from "containers/Report"

const initApp = actions => {
  const {setView, setLabel, setLabels, setPapers, setMode} = actions
  const Labels = initLabels(actions)
  const Report = initReport({
    changeLabel: changeLabel({setView, setLabels, setLabel}),
    setPapers,
    setMode,
  })

  init(actions)

  const App = ({state}) => {
    if (state.view === views.report) {
      return (
        <Report
          loading={state.loading}
          papers={state.papers}
          label={state.currentLabel}
          mode={state.mode}
        />
      )
    }

    if (state.view === views.labels && state.labels.length > 0) {
      return (
        <Labels
          currentLabel={state.currentLabel}
          labels={state.labels}
        />
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
    state: PropTypes.shape({
      loading: PropTypes.bool.isRequired,
      currentLabel: PropTypes.string,
      labels: PropTypes.arrayOf(PropTypes.string).isRequired,
      view: PropTypes.string.isRequired,
      mode: PropTypes.string.isRequired,
      papers: PropTypes.shape({
        read: PropTypes.shape({
          papers: PropTypes.arrayOf(paperProps),
        }).isRequired,
        hidden: PropTypes.shape({
          papers: PropTypes.arrayOf(PropTypes.string),
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

  return App
}

export default initApp
