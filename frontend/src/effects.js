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

export const changeLabel = ({setView, setLabels}) => _ => {
  setView("labels")
  getLabels({setLabels})
}

export const init = ({setView, setLabels, setLabel, setPapers}) => {
  const maybeLabel = JSON.parse(localStorage.getItem("label"))

  if (maybeLabel) {
    const label = maybeLabel

    setView("report")
    setLabel(label)
    getMessages({label, setPapers})
  } else {
    getLabels({setLabels})
  }
}

export const selectLabel = ({setView, setLabel, setPapers}) => label => e => {
  e.preventDefault()

  setLabel(label)
  localStorage.setItem("label", JSON.stringify(label))

  post("json/messages", {label}).then(papers => {
    setPapers(papers)
    setView("report")
  })
}
