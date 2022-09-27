import React from "react"
import PropTypes from "prop-types"
import Switch from "./Switch"

import "components/components.css"

const Download = ({label, filetype, filename, content, disabled = false}) => {
  if (disabled) {
    return (
      <Switch label={label} disabled />
    )
  }

  const dataURI = `data:${filetype};base64,${content}`

  return (
    <a className="clickable-label clickable-label--link" download={filename} href={dataURI}>
      {label}
    </a>
  )
}

Download.defaultProps = {
  disabled: false,
}

Download.propTypes = {
  disabled: PropTypes.bool,
  label: PropTypes.string.isRequired,
  filetype: PropTypes.string.isRequired,
  filename: PropTypes.string.isRequired,
  content: PropTypes.string.isRequired,
}

export default Download
