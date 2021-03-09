import React, {useState, useEffect} from "react"
import PropTypes from "prop-types"

import "containers/containers.css"
import Paper from "components/Paper"
import Loader from "components/Loader"
import Header from "components/ReportHeader"
import Switch from "components/Switch"
import {Either, Maybe} from "utils"
import {modes} from "constants"
import {toggleMode, hidePapers, restorePapers} from "effects"

const Report = ({papers, label, changeLabel, mode, setMode, setPapers}) => {
  const hideSelectedPapers = hidePapers({setPapers, label, papers})
  const restoreSelectedPapers = restorePapers({label})
  const {stats} = papers.unread
  const unread = papers.unread.papers
  const hidden = papers.hidden.papers

  const [checked, setChecked] = useState(new Set())
  const [viewAll, setViewAll] = useState(hidden.length > 0)

  const ps = viewAll ?
    unread.filter(x => hidden.indexOf(x.Title) === -1) :
    unread

  useEffect(() => {
    setViewAll(hidden.length > 0)
  }, [hidden])

  return (
    <div data-testid="report">
      <Header
        changeLabel={changeLabel}
        label={label}
        stats={stats}
        papers={papers.unread.papers}
      />
      <h2>
        New papers
        <Switch
          label={mode}
          onClick={
            toggleMode({setMode})(mode === modes.default ? modes.compact : modes.default)
          }
        />
        <Maybe cond={checked.size > 0 || (viewAll && papers.hidden.papers.length > 0)}>
          <Switch
            label={checked.size > 0 ? "hide selected" : "show all"}
            onClick={_ => {
              if (checked.size > 0) {
                setChecked(new Set())
                hideSelectedPapers([...checked])
                setViewAll(false)
              } else {
                setViewAll(false)
                setChecked(new Set(papers.hidden.papers))
              }
            }}
          />
        </Maybe>
      </h2>
      <Either cond={ps.length > 0}>
        <ul className={`main__papers main__papers--${mode}`}>
          {
            ps.map((paper, i) => (
              <li key={paper.Title} className="main__papers-paper">
                <input
                  id={`paper-checkbox-${i}`}
                  type="checkbox"
                  checked={checked.has(paper.Title)}
                  onChange={_ => {
                    if (checked.has(paper.Title)) {
                      checked.delete(paper.Title)
                      restoreSelectedPapers(paper.Title)
                    } else {
                      checked.add(paper.Title)
                    }

                    setChecked(new Set(checked))
                  }}
                />
                <Paper paper={paper} mode={mode} />
              </li>
            ))
          }
        </ul>
        <Loader />
      </Either>
    </div>
  )
}

Report.propTypes = {
  setMode: PropTypes.func.isRequired,
  setPapers: PropTypes.func.isRequired,
  mode: PropTypes.string.isRequired,
  changeLabel: PropTypes.func.isRequired,
  label: PropTypes.string.isRequired,
  papers: PropTypes.shape({
    read: PropTypes.shape({papers: PropTypes.arrayOf(PropTypes.object)}),
    hidden: PropTypes.shape({papers: PropTypes.arrayOf(PropTypes.string)}),
    unread: PropTypes.shape({
      papers: PropTypes.arrayOf(PropTypes.object),
      stats: PropTypes.shape({
        messages: PropTypes.number,
        papers: PropTypes.number,
        time: PropTypes.string,
      }).isRequired,
    }),
  }).isRequired,
}

export default Report
