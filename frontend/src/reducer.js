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
}

export const actions = getActions(handlers)
export default createReducer(handlers)
