import React from "react"
import PropTypes from "prop-types"

import "containers/containers.css"
import Paper from "components/Paper"
import Loader from "components/Loader"
import Header from "components/ReportHeader"
import Switch from "components/Switch"
import {Either} from "utils"

const Main = ({papers, stats, label, changeLabel, mode, toggleMode}) => (
  <div>
    <Header
      changeLabel={changeLabel}
      label={label}
      stats={stats}
      papers={papers}
    />
    <h2>
      New papers
      <Switch label={mode} onClick={toggleMode} />
    </h2>
    <Either cond={papers.length > 0}>
      <ul className={`main__papers main__papers--${mode}`}>
        {
          papers.map(paper => (
            <Paper key={paper.Title} paper={paper} mode={mode} />
          ))
        }
      </ul>
      <Loader />
    </Either>
  </div>
)

Main.propTypes = {
  toggleMode: PropTypes.func.isRequired,
  mode: PropTypes.string.isRequired,
  changeLabel: PropTypes.func.isRequired,
  label: PropTypes.string.isRequired,
  papers: PropTypes.arrayOf(PropTypes.object).isRequired,
  stats: PropTypes.shape({
    messages: PropTypes.number,
    papers: PropTypes.number,
    time: PropTypes.string,
  }).isRequired,
}

export default Main
