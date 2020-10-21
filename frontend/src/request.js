/* eslint-disable */

const baseUrl = "http://localhost:8080"
const req = method => (endpoint = "", payload) =>
  fetch(new Request(
    `${baseUrl}/${endpoint}`,
    {
      method,
      body: JSON.stringify(payload),
      mode: "cors",
    },
  ))
  .then(r => {
    if (r.ok) {
      return Promise.resolve(r.json())
    }

    return Promise.reject(r.statusText)
  })

export const get = req("GET")
export const put = req("PUT")
export const post = req("POST")
