import React from "react"
import PropTypes from "prop-types"

const Header = ({changeLabel, label, stats, papers}) => (
  <>
    <h1>
      Google Scholar Alert Digest
      <button className="main__switch" type="button" onClick={changeLabel}>
        {label}
      </button>
    </h1>
    <ul className="metadata">
      <li>
        <b>Date: </b>
        {stats.time}
      </li>
      <li>
        <b>Unread emails: </b>
        {stats.messages}
      </li>
      <li>
        <b>Paper titles: </b>
        {papers.length}
      </li>
      <li>
        <b>Unique paper titles: </b>
        {stats.papers}
      </li>
    </ul>
  </>
)

Header.propTypes = {
  changeLabel: PropTypes.func.isRequired,
  label: PropTypes.string.isRequired,
  papers: PropTypes.arrayOf(PropTypes.object).isRequired,
  stats: PropTypes.shape({
    messages: PropTypes.number,
    papers: PropTypes.number,
    time: PropTypes.string,
  }).isRequired,
}

export default Header
