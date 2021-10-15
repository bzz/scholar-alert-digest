import React, {useState} from "react"
import PropTypes from "prop-types"

import {selectLabel} from "effects"

import "containers/containers.css"

const initLabels = actions => {
  const Labels = ({currentLabel, labels}) => {
    const [checked, check] = useState(currentLabel || labels[0])

    return (
      <form
        data-testid="labels"
        onSubmit={selectLabel(actions)(checked)}
      >
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

  Labels.defaultProps = {
    currentLabel: "",
  }

  Labels.propTypes = {
    currentLabel: PropTypes.string,
    labels: PropTypes.arrayOf(PropTypes.string).isRequired,
  }

  return Labels
}

export default initLabels
