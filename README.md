# GPX Downloader
Made for my Year 3 project, designed to download GPX files from a CSV file, parse them and save the data in a database for further analyical use.

Download and process thousands of files in batches with the power of goroutines, and upload said content to a database

# Usage
First ensure there is a "data-sources" directory within the project root, which should contain the CSV files that would contain the data you need to download

Ensure that the models' fields are correct

Ensure that in ./intenral/downloader/download.go, the DOWNLOAD_URL is updated to the destination API

Ensure that the configurations are correct, namely, Database Host, Port

Ensure that you have the secrets in the root directory, as specified in the docker-compose file

migrate database using command
```
make migrate
```

parse CSV, download all files, and upload to the database using
```
# to run directly
make test

# using docker
docker compose up
```
