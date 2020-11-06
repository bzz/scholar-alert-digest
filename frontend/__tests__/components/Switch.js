import React from "react"
import {fireEvent, render} from "@testing-library/react"
import Switch from "components/Switch"

test("renders switch", () => {
  const onClick = jest.fn()
  const {queryByText, getByText} = render(
    <Switch label="test-label" onClick={onClick} />,
  )

  expect(queryByText(/test-label/)).toBeTruthy()
  fireEvent.click(getByText(/test-label/))
  expect(onClick.mock.calls.length).toBe(1)
})
