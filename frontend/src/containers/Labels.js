import React, {useState} from "react"
import PropTypes from "prop-types"

import {post} from "request"

import "containers/containers.css"

const handleSubmit = ({setLabel, setPapers}) => label => e => {
  e.preventDefault()

  setLabel(label)
  localStorage.setItem("label", JSON.stringify(label))

  post("json/messages", {label}).then(setPapers)
}

const Labels = ({labels, setLabel, setPapers}) => {
  const [checked, check] = useState(labels[0])

  return (
    <form onSubmit={handleSubmit({setLabel, setPapers})(checked)}>
      <h1>Select a gmail label to aggregate</h1>
      <ul className="labels__list">
        {labels.map(label => (
          <li key={label}>
            <label htmlFor={label} className="labels__label">
              <input
                id={label}
                className="labels__radio"
                type="radio"
                name={label}
                value={label}
                checked={checked === label}
                onChange={_ => check(label)}
              />
              {label}
            </label>
          </li>
        ))}
      </ul>
      <button className="labels__submit" type="submit">
        Save and view report
      </button>
    </form>
  )
}

Labels.propTypes = {
  labels: PropTypes.arrayOf(PropTypes.string).isRequired,
  setLabel: PropTypes.func.isRequired,
  setPapers: PropTypes.func.isRequired,
}

export default Labels
