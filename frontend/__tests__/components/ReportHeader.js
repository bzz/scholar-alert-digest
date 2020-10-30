/* eslint-disable react/jsx-props-no-spreading */

import React from "react"
import {fireEvent, render} from "@testing-library/react"
import Header from "components/ReportHeader"

test("renders report header correctly", () => {
  const random = x => Math.floor(Math.random() * x)
  const changeLabel = jest.fn()
  const time = Date.now().toString()
  const messages = random(10)
  const papers = random(100)
  const paperList = Array.from({length: random(200)})

  const props = {
    changeLabel,
    label: "test-label",
    stats: {
      time,
      messages,
      papers,
    },
    papers: paperList,
  }

  const {container, queryByText, getByText} = render(
    <Header {...props} />,
  )

  const find = (r, el) =>
    [...container.querySelectorAll(el)].find(el => new RegExp(r).test(el.innerHTML))

  expect(queryByText(/test-label/)).toBeTruthy()
  fireEvent.click(getByText(/test-label/))
  expect(changeLabel.mock.calls.length).toBe(1)

  expect(find(`Date:.+${time}`, "li")).toBeTruthy()
  expect(find(`Unread emails:.+${messages}`, "li")).toBeTruthy()
  expect(find(`Paper titles:.+${paperList.length}`, "li")).toBeTruthy()
  expect(find(`Unique paper titles:.+${papers}`, "li")).toBeTruthy()
})
