import React, {useState, useEffect} from "react"
import PropTypes from "prop-types"

import "containers/containers.css"
import Paper from "components/Paper"
import Loader from "components/Loader"
import Header from "components/ReportHeader"
import Switch from "components/Switch"
import Download from "components/Download"
import {Either, Maybe, fromMaybe} from "utils"
import {modes as m} from "constants"
import {toggleMode, hidePapers, restorePapers} from "effects"

const getEncodedPapers = hiddenPapers => fromMaybe(null)(_ => {
  // unicode breaks btoa
  const formattedPapers = hiddenPapers.map(paper => ({
    ...paper,
    Abstract: {
      ...paper.Abstract,
      Rest: paper.Abstract.Rest.replace("…", "..."),
    },
  }))

  return btoa(JSON.stringify({hiddenPapers: formattedPapers}))
})

const initReport = actions => {
  const {changeLabel, setPapers} = actions

  const Report = ({loading, papers, label, mode}) => {
    const hideSelectedPapers = hidePapers({setPapers, label, papers})
    const restoreSelectedPapers = restorePapers({label})
    const {stats} = papers.unread
    const unread = papers.unread.papers
    const hidden = papers.hidden.papers

    const [checked, setChecked] = useState(new Set())
    const [papersHidden, setPapersHidden] = useState(hidden.length > 0)

    useEffect(() => {
      setPapersHidden(hidden.length > 0)
    }, [hidden])

    const [visiblePapers, selectedPapers] =
      unread.reduce(([visible, selected], paper) => {
        if (!hidden.includes(paper.Title)) {
          visible.push(paper)
        }

        if (checked.has(paper.Title)) {
          selected.push(paper)
        }

        return [visible, selected]
      }, [[], []])

    const papersToRender = papersHidden ? visiblePapers : unread

    return (
      <div data-testid="report">
        <Header
          changeLabel={changeLabel}
          label={label}
          stats={stats}
          papers={unread}
        />
        <h2>
          New papers
          <Switch
            label={mode}
            onClick={_ => {
              const nextMode = mode === m.default ? m.compact : m.default
              toggleMode(actions)(nextMode)
            }}
          />
          <Maybe cond={!papersHidden && !loading}>
            <Download
              label="download selected"
              disabled={checked.size === 0}
              filename="scholar-alert-digest-hidden-papers.json"
              filetype="application/json"
              content={getEncodedPapers(selectedPapers)}
            />
          </Maybe>
          <Switch
            label="hide selected"
            disabled={checked.size === 0}
            onClick={_ => {
              hideSelectedPapers([...checked])
              setChecked(new Set())
              setPapersHidden(true)
            }}
          />
          <Maybe cond={papersHidden}>
            <Switch
              label="show all"
              onClick={_ => {
                setChecked(new Set([...checked, ...hidden]))
                setPapersHidden(false)
              }}
            />
          </Maybe>
        </h2>
        <Either cond={loading}>
          <Loader />
          <ul className={`main__papers main__papers--${mode}`}>
            {
              papersToRender.map((paper, i) => (
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
        </Either>
      </div>
    )
  }

  Report.propTypes = {
    mode: PropTypes.string.isRequired,
    loading: PropTypes.bool.isRequired,
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

  return Report
}

export default initReport
