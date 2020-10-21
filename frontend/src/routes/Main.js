import React from "react"
import PropTypes from "prop-types"

import "routes/routes.css"
import Paper from "components/Paper"

const Main = ({papers}) => (
  <div>
    <h1>Google Scholar Alert Digest</h1>
    <ul className="metadata">
      <li>
        <b>Date: </b>
        {new Date().toUTCString()}
      </li>
      <li>
        <b>Unread emails: </b>
        ?
      </li>
      <li>
        <b>Paper titles: </b>
        {papers.unread.length}
      </li>
      <li>
        <b>Unique paper titles: </b>
        ?
      </li>
    </ul>
    <h2>New papers</h2>
    <ul>
      {papers.unread.map(paper => (
        <Paper key={paper.Title} paper={paper} />
      ))}
    </ul>
  </div>
)

Main.propTypes = {
  papers: PropTypes.shape({
    read: PropTypes.arrayOf(PropTypes.object),
    unread: PropTypes.arrayOf(PropTypes.object),
  }).isRequired,
}

export default Main
