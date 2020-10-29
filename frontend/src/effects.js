import {get, post} from "request"

export const login = url => {
  window.location = url
}

export const handleError = e => {
  if (e.status === 401) {
    login(e.payload.Redirect)
  }
}

export const getLabels = ({setLabels}) => {
  get("json/labels")
    .then(({labels}) => {
      setLabels(labels)
      localStorage.setItem("labels", JSON.stringify(labels))
    })
    .catch(handleError)
}

export const getMessages = ({label, setPapers}) => {
  post("json/messages", {label})
    .then(setPapers)
    .catch(handleError)
}

export const changeLabel = ({setLabels, setLabel}) => _ => {
  setLabel(null)
  getLabels({setLabels})
}

export const init = ({setLabels, setLabel, setPapers}) => {
  const maybeLabel = localStorage.getItem("label")

  if (maybeLabel) {
    const label = maybeLabel

    setLabel(label)
    getMessages({label, setPapers})
  } else {
    getLabels({setLabels})
  }
}
