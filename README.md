# Weather API


This is the RESTful API which can be used like the microservice for the web app. I created first version of this project when I get one technical task from one company. 
The API is used for fetching weather podcasts via the Open Weather API for races which are being held anytime, anywhere. Service automatically update forecasts for the races and automatically delete forecasts for races which are ended.


For now comments are on Croatian language.

## API endpoints

#### Get forecasts for a race
* Path: /race/:id/forecast
* Method: GET

#### Get the details about one race
* Path: /race/:id
* Method: GET

#### Get the details from the all races
* Path: /races
* Method: GET

#### Create race
* Path: /race
* Method: POST

#### Update one race
* Path: /race/:id
* Method: PUT

#### Delete a race
* Path: /race/:id
* Method: DELETE


## Prerequisites

For compile and running this project you need to have installed [GO](https://golang.org/dl/)(version 1.9 or newer) on your computer and [PostgreSQL](https://www.postgresql.org/).
Also is recommended that you install also [Dep](https://github.com/golang/dep) (Go dependency management tool), for fast installation of required library's.

## Built With

* [Gin](https://github.com/gin-gonic/gin) - Fast HTTP web framework written in Go.
* [Dep](https://github.com/golang/dep) - Dependency Management
* [PostgreSQL](https://www.postgresql.org/) - SQL database

## Project setup

1. Clone the repository in your `$GOPATH/src` directory, with `git clone https://github.com/leondominovic/weather_api`.
After that, just move to that folder with `cd weather_api`.

2. Then run `dep ensure` to get all required libraries.

3. Create a new blank database in Postgresql and a new user for that database. Then import a database from this folder with `psql -U yourNewUserName newDataBaseName < database.psql`

4. Then export all needed environment variables for database connection.
`
export DBNAME = newDataBaseName
export DBUSER = yourNewUserName
export DBPASS = password for Postgresql user
export DBHOST = host for Postgresql. If you run Postgresql on your computer then is "localhost"
export DBPORT = default is 5432`

5. Compile and run app with `go run main.go`