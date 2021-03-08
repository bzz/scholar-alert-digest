import {get, post} from "request"
import {views} from "constants"

export const login = url => {
  window.location = url
}

export const handleError = e => {
  if (e.status === 401) {
    login(e.payload.Redirect)
  }
}

export const getLabels = ({setLabels}) => {
  get("labels")
    .then(setLabels)
    .catch(handleError)
}

export const getMessages = ({label, setPapers}) => {
  post("messages", {label})
    .then(setPapers)
    .catch(handleError)
}

export const changeLabel = ({setView, setLabels}) => _ => {
  setView(views.labels)
  getLabels({setLabels})
}

export const init = ({setView, setLabels, setLabel, setPapers, setMode}) => {
  const maybeLabel = JSON.parse(localStorage.getItem("label"))
  const mode = JSON.parse(localStorage.getItem("mode"))

  if (maybeLabel) {
    const label = maybeLabel

    setView(views.report)
    setLabel(label)
    getMessages({label, setPapers})
  } else {
    getLabels({setLabels})
  }

  setMode(mode)
}

export const selectLabel = ({setView, setLabel, setPapers}) => label => e => {
  e.preventDefault()

  setLabel(label)
  localStorage.setItem("label", JSON.stringify(label))

  post("messages", {label}).then(papers => {
    setPapers(papers)
    setView(views.report)
  })
}

export const toggleMode = ({setMode}) => mode => _ => {
  localStorage.setItem("mode", JSON.stringify(mode))

  setMode(mode)
}
