# gh-rel

GH Rel is a small dashboard to keep track of releases in several github projects at once.

It includes a web server to render the dashboard and a cli application to add new projects to the dashboard.

## Howto

The program is built as a single binary with `go build .`.

To add a new project to the dashboard, run `gh-rel add org/name`.
The server needs to be ofline for this because boltdb doesn't support multiple processes accessing the same database.

To start the server, run `gh-rel serve`.

The server takes the releases information from GitHub's releases API described in https://developer.github.com/v3/repos/releases.

## Disclaimer

I wrote this code in a few hours, use it at your own discretion. Pull Requests are very welcome.

## License

MIT
