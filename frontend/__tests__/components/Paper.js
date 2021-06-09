import React from "react"
import {render} from "@testing-library/react"
import Paper from "components/Paper"

const random = x => Math.floor(Math.random() * x)
const randomString = (length = 8) => Math.random().toString(16).substr(2, length)

test("renders paper correctly in default mode", () => {
  const refs = Array.from({length: random(8)}, (_, i) => ({
    ID: (i + 1).toString(),
    Title: randomString(),
  }))

  const paper = {
    Abstract: {
      FirstLine: "first line",
      Rest: "rest of the abstract",
    },
    URL: "https://url.co",
    Title: "title",
    Author: "author",
    Refs: refs,
  }

  const {queryByText} = render(
    <Paper paper={paper} mode="default" />,
  )

  expect(queryByText(/title/)).toBeTruthy()
  expect(queryByText(/author/)).toBeTruthy()
  expect(queryByText(/first line/)).toBeTruthy()
  expect(queryByText(/rest of the abstract/)).toBeTruthy()
  refs.forEach(({Title}) => {
    expect(queryByText(new RegExp(Title))).toBeTruthy()
  })
})

test("renders paper correctly in compact mode", () => {
  const refs = Array.from({length: random(8)}, (_, i) => ({
    ID: (i + 1).toString(),
    Title: randomString(),
  }))

  const paper = {
    Abstract: {
      FirstLine: "first line",
      Rest: "rest of the abstract",
    },
    URL: "https://url.co",
    Title: "title",
    Author: "author",
    Refs: refs,
  }

  const {queryByText} = render(
    <Paper paper={paper} mode="compact" />,
  )

  expect(queryByText(/title/)).toBeTruthy()
  expect(queryByText(/author/)).toBeTruthy()
  expect(queryByText(/first line/)).toBeTruthy()
  expect(queryByText(/rest of the abstract/)).toBeTruthy()
  refs.forEach(({Title}) => {
    expect(queryByText(new RegExp(Title))).toBeTruthy()
  })
})
