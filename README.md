# Weather API


This is the RESTful API which can be used like the microservice for the web app. I created this project when I get one technical task from one company. 
The API is used for fetching weather podcasts via the Open Weather API for races which are being held during weekend, between 9am and 6pm, in the B&H. When one race was added, this API automatically fetch weather forecasts for that race. Service automatically update forecasts for the races and automatically delete forecasts for races which are ended.

## API endpoints

#### Get forecasts for a race
* Path: /utrka/:id/prognoza
* Method: GET

#### Get the deteils about one race
* Path: /utrka/:id
* Method: GET

#### Create race
* Path: /utrka
* Method: POST

#### Update one race
* Path: /utrka/:id
* Method: PUT

#### Delete a race
* Path: /utrka/:id
* Method: DELETE


## Prerequisites

For compile and running this project you need to have installed [GO](https://golang.org/dl/)(version 1.9 or newer) on your computer and [PostgreSQL](https://www.postgresql.org/).
Also is recommended that you install also [Dep](Go dependency management tool), for fast installation of required library's.

## Built With

* [Gin](https://github.com/gin-gonic/gin) - Fast HTTP web framework written in Go.
* [Dep](https://github.com/golang/dep) - Dependency Management
* [PostgreSQL][https://www.postgresql.org/] - SQL database