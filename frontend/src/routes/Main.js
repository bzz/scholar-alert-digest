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
        {papers.unread.stats.time}
      </li>
      <li>
        <b>Unread emails: </b>
        {papers.unread.stats.messages}
      </li>
      <li>
        <b>Paper titles: </b>
        {papers.unread.papers.length}
      </li>
      <li>
        <b>Unique paper titles: </b>
        {papers.unread.stats.papers}
      </li>
    </ul>
    <h2>New papers</h2>
    <ul>
      {papers.unread.papers.map(paper => (
        <Paper key={paper.Title} paper={paper} />
      ))}
    </ul>
  </div>
)

Main.propTypes = {
  papers: PropTypes.shape({
    read: PropTypes.shape({
      papers: PropTypes.arrayOf(PropTypes.object),
    }),
    unread: PropTypes.shape({
      papers: PropTypes.arrayOf(PropTypes.object),
      stats: PropTypes.shape({
        messages: PropTypes.number,
        papers: PropTypes.number,
        time: PropTypes.string,
      }),
    }),
  }).isRequired,
}

export default Main
