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
  const hpo = JSON.parse(localStorage.getItem("hiddenPapers")) || {}
  const hps = hpo[label] || []

  return post("messages", {label})
    .then(papers => setPapers({...papers, hidden: {papers: hps}}))
    .catch(handleError)
}

export const changeLabel = ({setView, setLabels}) => _ => {
  setView(views.labels)
  getLabels({setLabels})
}

export const init = ({setView, setLabels, setLabel, setPapers, setMode, setLoading}) => {
  const maybeLabel = JSON.parse(localStorage.getItem("label"))
  const mode = JSON.parse(localStorage.getItem("mode"))

  setLoading(true)

  if (maybeLabel) {
    const label = maybeLabel

    setView(views.report)
    setLabel(label)
    getMessages({label, setPapers}).then(_ => setLoading(false))
  } else {
    getLabels({setLabels})
  }

  setMode(mode)
}

export const selectLabel = ({setView, setLabel, setPapers, setLoading}) => label => e => {
  e.preventDefault()

  setLabel(label)
  localStorage.setItem("label", JSON.stringify(label))

  getMessages({label, setPapers})
    .then(_ => setView(views.report))
    .then(_ => setLoading(false))
}

export const toggleMode = ({setMode}) => mode => _ => {
  localStorage.setItem("mode", JSON.stringify(mode))

  setMode(mode)
}

export const hidePapers = ({setPapers, label, papers}) => papersToHide => {
  const hpo = JSON.parse(localStorage.getItem("hiddenPapers")) || {}

  if (hpo[label]) {
    hpo[label] = hpo[label].concat(papersToHide)
  } else {
    hpo[label] = papersToHide
  }

  localStorage.setItem("hiddenPapers", JSON.stringify(hpo))

  setPapers({
    ...papers,
    hidden: {...papers.hidden, papers: hpo[label]},
  })
}

export const restorePapers = ({label}) => paper => {
  const hpo = JSON.parse(localStorage.getItem("hiddenPapers")) || {}

  if (hpo[label]) {
    hpo[label] = hpo[label].filter(x => x !== paper)
    localStorage.setItem("hiddenPapers", JSON.stringify(hpo))
  }
}
