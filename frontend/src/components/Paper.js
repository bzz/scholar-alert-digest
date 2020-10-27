import React from "react"
import PropTypes from "prop-types"

import "components/components.css"

const PaperCompact = ({paper}) => (
  <li>
    <details className="details">
      <summary>
        <a href={paper.URL}>{paper.Title}</a>
        {`, ${paper.Author} `}
        ({paper.Refs.map((ref, i) => (
          <a key={ref} href={`https://mail.google.com/mail/#inbox/${ref}`}>
            {i + 1}
          </a>
        ))})
      </summary>
      <div>{`${paper.Abstract.FirstLine} ${paper.Abstract.Rest}`}</div>
    </details>
  </li>
)

const PaperDefault = ({paper}) => (
  <li>
    <a href={paper.URL}>{paper.Title}</a>
    {`, ${paper.Author} `}
    ({paper.Refs.map((ref, i) => (
      <a key={ref} href={`https://mail.google.com/mail/#inbox/${ref}`}>
        {i + 1}
      </a>
    ))})
    <details className="details">
      <summary>{paper.Abstract.FirstLine}</summary>
      <div>{paper.Abstract.Rest}</div>
    </details>
  </li>
)

const Paper = ({paper, mode}) => {
  if (mode === "compact") {
    return <PaperCompact paper={paper} />
  }

  return <PaperDefault paper={paper} />
}

const paperTypes = {
  paper: PropTypes.shape({
    URL: PropTypes.string.isRequired,
    Title: PropTypes.string.isRequired,
    Author: PropTypes.string.isRequired,
    Refs: PropTypes.arrayOf(PropTypes.string).isRequired,
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

export default Paper
