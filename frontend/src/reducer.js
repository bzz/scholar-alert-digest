/* eslint-disable arrow-body-style */

import {getActions, createReducer} from "utils"

const handlers = {
  setLabels: (s, {payload}) => {
    return {
      ...s,
      labels: payload,
    }
  },
  setLabel: (s, {payload}) => {
    return {
      ...s,
      currentLabel: payload,
    }
  },
  setPapers: (s, {payload}) => {
    return {
      ...s,
      papers: payload,
    }
  },
  toggleMode: s => {
    return {
      ...s,
      mode: s.mode === "default" ? "compact" : "default",
    }
  },
}

export const actions = getActions(handlers)
export default createReducer(handlers)
export const defaultState = {
  labels: [],
  mode: "default",
  papers: {
    read: {
      papers: [],
    },
    unread: {
      papers: [],
      stats: {
        messages: 0,
        papers: 0,
        time: "?",
      },
    },
  },
}
