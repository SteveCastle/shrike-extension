## Command Line Extension Server

A webserver to pass URLs and arguments to a CLI program. Accepts a POST request containing json in the form `{"Command":"echo", Arguments: ["arg1", "http://example.com"]}`.

## Dependencies

- Install GO. https://go.dev/

## Usage

- Clone the repository.
- Change directory into the project directory.
- Build the server by running `go build .`
- Run the server with `./server"`.
