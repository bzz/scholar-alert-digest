/* eslint-disable react/jsx-props-no-spreading */

import React from "react"
import {fireEvent, render} from "@testing-library/react"
import initLabels from "containers/Labels"

import {selectLabel} from "effects"

jest.mock("effects", () => {
  const selectLabel = jest.fn().mockReturnValue(jest.fn().mockReturnValue(jest.fn))

  return {
    init: jest.fn(),
    selectLabel,
  }
})

afterEach(() => {
  jest.clearAllMocks()
})

test("renders labels container w/o currentLabel", () => {
  const props = {
    currentLabel: null,
    labels: ["label1", "label2"],
  }

  const actions = {
    setLabel: jest.fn(),
    setPapers: jest.fn(),
    setView: jest.fn(),
    setLoading: jest.fn(),
  }

  const Labels = initLabels(actions)

  const {getByLabelText, getByTestId} = render(
    <Labels {...props} />,
  )

  expect(getByLabelText("label1").checked).toBe(true)
  expect(getByLabelText("label2").checked).toBe(false)

  fireEvent.click(getByLabelText("label2"))

  expect(getByLabelText("label1").checked).toBe(false)
  expect(getByLabelText("label2").checked).toBe(true)
  expect(selectLabel.mock.calls.length).toBe(2)

  fireEvent.submit(getByTestId("labels"))

  setTimeout(() => {
    expect(selectLabel.mock.calls.length).toBe(3)
  }, 0)
})

test("renders labels container w/ currentLabel", () => {
  const props = {
    currentLabel: "label2",
    labels: ["label1", "label2"],
  }

  const actions = {
    setLabel: jest.fn(),
    setPapers: jest.fn(),
    setView: jest.fn(),
    setLoading: jest.fn(),
  }

  const Labels = initLabels(actions)

  const {getByLabelText, getByTestId} = render(
    <Labels {...props} />,
  )

  expect(getByLabelText("label1").checked).toBe(false)
  expect(getByLabelText("label2").checked).toBe(true)

  fireEvent.click(getByLabelText("label1"))

  expect(getByLabelText("label1").checked).toBe(true)
  expect(getByLabelText("label2").checked).toBe(false)
  expect(selectLabel.mock.calls.length).toBe(2)

  fireEvent.submit(getByTestId("labels"))

  setTimeout(() => {
    expect(selectLabel.mock.calls.length).toBe(3)
  }, 0)
})
