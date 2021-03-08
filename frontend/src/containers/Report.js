import React, {useState} from "react"
import PropTypes from "prop-types"

import "containers/containers.css"
import Paper from "components/Paper"
import Loader from "components/Loader"
import Header from "components/ReportHeader"
import Switch from "components/Switch"
import {Either, Maybe} from "utils"

const Report = ({papers, stats, label, changeLabel, mode, toggleMode}) => {
  const [checked, setChecked] = useState(new Set())
  const [viewAll, setViewAll] = useState(true)

  return (
    <div data-testid="report">
      <Header
        changeLabel={changeLabel}
        label={label}
        stats={stats}
        papers={papers}
      />
      <h2>
        New papers
        <Switch label={mode} onClick={toggleMode} />
        <Maybe cond={checked.size > 0}>
          <Switch
            label={viewAll ? "hide selected" : "show all"}
            onClick={_ => setViewAll(!viewAll)}
          />
        </Maybe>
      </h2>
      <Either cond={papers.length > 0}>
        <ul className={`main__papers main__papers--${mode}`}>
          {
            papers.map((paper, i) => (
              <Maybe key={paper.Title} cond={viewAll || !checked.has(i)}>
                <li className="main__papers-paper">
                  <input
                    id={`paper-checkbox-${i}`}
                    type="checkbox"
                    checked={checked.has(i)}
                    onChange={_ => {
                      if (checked.has(i)) {
                        checked.delete(i)
                      } else {
                        checked.add(i)
                      }

                      setChecked(new Set(checked))
                    }}
                  />
                  <Paper paper={paper} mode={mode} />
                </li>
              </Maybe>
            ))
          }
        </ul>
        <Loader />
      </Either>
    </div>
  )
}

Report.propTypes = {
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

export default Report
