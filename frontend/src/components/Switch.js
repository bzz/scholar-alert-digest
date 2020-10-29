import React from "react"
import PropTypes from "prop-types"

import "components/components.css"

const Switch = ({label, onClick}) => (
  <button className="switch" type="button" onClick={onClick}>
    {label}
  </button>
)

Switch.propTypes = {
  label: PropTypes.string.isRequired,
  onClick: PropTypes.func.isRequired,
}

export default Switch
