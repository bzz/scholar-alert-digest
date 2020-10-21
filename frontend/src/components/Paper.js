import React from "react"
import PropTypes from "prop-types"

import "components/components.css"

const Paper = ({paper}) => (
  <li>
    <a href={paper.URL}>{paper.Title}</a>
    {`, ${paper.Author} `}
    ({paper.Refs.map(ref => (
      <a key={ref} href={`https://mail.google.com/mail/#inbox/${ref}`}>
        ?
      </a>
    ))})
    <details className="details">
      <summary>{paper.Abstract.FirstLine}</summary>
      <div>{paper.Abstract.Rest}</div>
    </details>
  </li>
)

Paper.propTypes = {
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

export default Paper
