import React from "react"
import PropTypes from "prop-types"

import "components/components.css"

const Switch = ({label, onClick, disabled = false}) => (
  <button className="clickable-label" type="button" onClick={onClick} disabled={disabled}>
    {label}
  </button>
)

Switch.defaultProps = {
  disabled: false,
}

Switch.propTypes = {
  disabled: PropTypes.bool,
  label: PropTypes.string.isRequired,
  onClick: PropTypes.func.isRequired,
}

export default Switch
