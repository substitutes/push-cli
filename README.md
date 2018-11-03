# Push CLI
This is the CLI of the substitutes/push project.

## Abstract
Push exposes an HTTP API endpoint, which receives GZIPPed POST data and uploads the data to a given FTP server.
Authentication is handled via HTTP Basic auth.
This is the command line interface, which listens for a file change and then TARS the files to push them to the backend. 
