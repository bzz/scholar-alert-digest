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
  .then(async r => {
    const json = await r.json()

    if (r.ok) {
      return Promise.resolve(json)
    }

    return Promise.reject(Object.assign(r, {payload: json.error}))
  })

export const get = req("GET")
export const put = req("PUT")
export const post = req("POST")
