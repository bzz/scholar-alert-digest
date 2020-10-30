import React from "react"
import {render} from "@testing-library/react"
import Paper from "components/Paper"

test("renders paper correctly in default mode", () => {
  const random = x => Math.floor(Math.random() * x)
  const refs = ["1"].concat(Array.from({length: random(8)}, (_, i) => (i + 2).toString()))

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
  expect(queryByText(new RegExp(`(${refs.join("*")})`))).toBeTruthy()
  expect(queryByText(/first line/)).toBeTruthy()
  expect(queryByText(/rest of the abstract/)).toBeTruthy()
})

test("renders paper correctly in compact mode", () => {
  const random = x => Math.floor(Math.random() * x)
  const refs = ["1"].concat(Array.from({length: random(8)}, (_, i) => (i + 2).toString()))

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
  expect(queryByText(new RegExp(`(${refs.join("*")})`))).toBeTruthy()
  expect(queryByText(/first line/)).toBeTruthy()
  expect(queryByText(/rest of the abstract/)).toBeTruthy()
})
