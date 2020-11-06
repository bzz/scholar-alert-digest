/* eslint-disable react/no-array-index-key */

import React from "react"
import PropTypes from "prop-types"
import {modes} from "constants"
import {Maybe} from "utils"

import "components/components.css"

const PaperTitle = ({paper}) => {
  const refs = paper.Refs.sort((x, y) => x.Title.length > y.Title.length ? -1 : 1)

  return (
    <>
      <a className="paper__title" href={paper.URL}>{paper.Title}</a>
      <span className="paper__author">{`, ${paper.Author} `}</span>
      ({`${refs.length}: `} {refs.map((ref, i, refs) => (
        <a
          className="paper__ref"
          key={`${i}-${ref}`}
          href={`https://mail.google.com/mail/#inbox/${ref}`}
        >
          {ref.Title || i + 1}
          <Maybe cond={refs.length > (i + 1)}>
            {", "}
          </Maybe>
        </a>
      ))})
    </>
  )
}

const PaperCompact = ({paper}) => (
  <li>
    <details className="details">
      <summary>
        <PaperTitle paper={paper} />
      </summary>
      <div>{`${paper.Abstract.FirstLine} ${paper.Abstract.Rest}`}</div>
    </details>
  </li>
)

const PaperDefault = ({paper}) => (
  <li>
    <PaperTitle paper={paper} />
    <details className="details">
      <summary>{paper.Abstract.FirstLine}</summary>
      <div>{paper.Abstract.Rest}</div>
    </details>
  </li>
)

const Paper = ({paper, mode}) => {
  if (mode === modes.compact) {
    return <PaperCompact paper={paper} />
  }

  return <PaperDefault paper={paper} />
}

const paperTypes = {
  paper: PropTypes.shape({
    URL: PropTypes.string.isRequired,
    Title: PropTypes.string.isRequired,
    Author: PropTypes.string.isRequired,
    Refs: PropTypes.arrayOf(
      PropTypes.shape({ID: PropTypes.string, Title: PropTypes.string}),
    ).isRequired,
    Abstract: PropTypes.shape({
      FirstLine: PropTypes.string.isRequired,
      Rest: PropTypes.string.isRequired,
    }).isRequired,
  }).isRequired,
}

Paper.propTypes = {
  mode: PropTypes.string.isRequired,
  ...paperTypes,
}

PaperCompact.propTypes = paperTypes
PaperDefault.propTypes = paperTypes
PaperTitle.propTypes = paperTypes

export default Paper
