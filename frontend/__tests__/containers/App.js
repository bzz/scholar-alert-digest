/* eslint-disable react/jsx-props-no-spreading */

import React from "react"
import {render} from "@testing-library/react"
import initApp from "containers/App"

import {init, changeLabel} from "effects"
import {modes, views} from "constants"

jest.mock("effects", () => {
  const changeLabel = jest.fn()
  const selectLabel = jest.fn()
  const hidePapers = jest.fn()
  const restorePapers = jest.fn()

  changeLabel.mockReturnValue(jest.fn)
  selectLabel.mockReturnValue(jest.fn)

  return {
    init: jest.fn(),
    toggleMode: () => () => jest.fn(),
    changeLabel,
    selectLabel,
    hidePapers,
    restorePapers,
  }
})

afterEach(() => {
  jest.clearAllMocks()
})

test("renders app container", () => {
  const props = {
    state: {
      view: views.labels,
      mode: modes.default,
      loading: false,
      labels: [],
      papers: {
        read: {
          papers: [],
        },
        hidden: {
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
    },
  }

  const actions = {
    setLabels: jest.fn(),
    setLabel: jest.fn(),
    setPapers: jest.fn(),
    setMode: jest.fn(),
    setView: jest.fn(),
    setLoading: jest.fn(),
  }

  const App = initApp(actions)

  const {container} = render(
    <App {...props} />,
  )

  expect(init.mock.calls.length).toBe(1)
  expect(container.firstChild).toBe(null)
})

test("renders app container > labels", () => {
  const props = {
    state: {
      view: views.labels,
      mode: modes.default,
      loading: false,
      labels: ["label1", "label2"],
      papers: {
        read: {
          papers: [],
        },
        hidden: {
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
    },
  }

  const actions = {
    setLabels: jest.fn(),
    setLabel: jest.fn(),
    setPapers: jest.fn(),
    setMode: jest.fn(),
    setView: jest.fn(),
    setLoading: jest.fn(),
  }

  const App = initApp(actions)

  const {queryByTestId} = render(
    <App {...props} />,
  )

  expect(init.mock.calls.length).toBe(1)
  expect(queryByTestId("labels")).toBeTruthy()
})

test("renders app container > report", () => {
  const props = {
    state: {
      view: views.report,
      mode: modes.default,
      loading: false,
      currentLabel: "label1",
      labels: ["label1", "label2"],
      papers: {
        read: {
          papers: [],
        },
        hidden: {
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
    },
  }

  const actions = {
    setLabels: jest.fn(),
    setLabel: jest.fn(),
    setPapers: jest.fn(),
    setMode: jest.fn(),
    setView: jest.fn(),
    setLoading: jest.fn(),
  }

  const App = initApp(actions)

  const {queryByTestId} = render(
    <App {...props} />,
  )

  expect(init.mock.calls.length).toBe(1)
  expect(changeLabel.mock.calls.length).toBe(1)
  expect(queryByTestId("report")).toBeTruthy()
})
