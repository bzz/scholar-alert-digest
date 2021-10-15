# Google Scholar alert digest front-end app

## Node version

Node version is set in `.nvmrc` (read more about [nvm](https://github.com/nvm-sh/nvm)). It is not necessary to use nvm, but for compatibility reasons, it is better if the node version matches.

## Run in dev mode

1. Make sure `nodejs` and `npm` are installed ([direct download](https://nodejs.org/en/download/) or [package manager](https://nodejs.org/en/download/package-manager/))
2. Run server with `go run ./cmd/server -dev` (add `-test` for mock data)
3. `cd frontend`
4. `npm install`
5. `npm run dev`
6. open `localhost:9000` in browser
