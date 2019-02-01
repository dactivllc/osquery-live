# Osquery Live [![CircleCI](https://circleci.com/gh/zwass/osquery-live.svg?style=svg)](https://circleci.com/gh/zwass/osquery-live)

Try osquery live in the browser. Runs a real osqueryi shell.

See it in action at [osquery.live](https://osquery.live).

![Screenshot of Osquery Live](./public/screenshot.png)

## Development

### Dependencies

- Go >= 1.11
- Node
- Yarn
- Osquery

### Local testing

```bash
yarn install
yarn build && ADDR=localhost:8080 go run ./server
```

Now open http://localhost:8080 in the web browser.

Restart the `yarn build && go run` commands when changes have been made.
