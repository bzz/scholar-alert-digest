/* eslint-disable arrow-body-style */

import {getActions, createReducer} from "utils"
import {modes, views} from "constants"

const handlers = {
  setLabels: (s, {payload: {labels}}) => {
    return {
      ...s,
      labels,
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
  setMode: (s, {payload}) => {
    return {
      ...s,
      mode: payload || modes.default,
    }
  },
  setView: (s, {payload}) => {
    return {
      ...s,
      view: payload,
    }
  },
}

export const actions = getActions(handlers)
export default createReducer(handlers)
export const defaultState = {
  labels: [],
  view: views.labels,
  mode: modes.default,
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
