s# filestore-client
Includes artefacts that  deploy a simple filestore-client application that connects to the server app deployed using https://github.com/anithapriyanatarajan/filestore-server

# pre-requisite that the server is running on http://127.0.0.1:8080

# Notes:
- This being a simple WIP demo version, the http server is hardcoded to be exposed on `localhost` and port `8080`. Hence the client is also configured to interact with server on http://127.0.0.1:8080

# Steps to initialize the filestore client in your local 
1. clone this repo and make sure that the working directory has main.go.
2. run `$./store` on windows and `$store` (This is the ready compiled version )
NOTE: To run the client in dev mode `$go run main.go` or to generate another version of executable run `$go build -o <desired_executable_name> main.go`

# Supported CLI commands
1. add file/s to the server. `$./store add test.txt test2.txt`
2. update file. `$./store update test.txt`
3. list files. `$./store ls`
4. remove file `$./store rm test.txt`
5. wordcount of all files `$./store wc`