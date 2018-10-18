package api

type WeatherPodcastByPeriod struct {
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		// Main        string `json:"main"`
		Description string `json:"description"`
		// Icon        string `json:"icon"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Rain struct {
		TreeH float64 `json:"3h"`
	} `json:"rain"`
	Snow struct {
		TreeH float64 `json:"3h"`
	} `json:"snow"`
	DtTxt string `json:"dt_txt"`
}

type WeatherPodcast struct {
	Cod     string                   `json:"cod"`
	Message float64                  `json:"message"`
	Cnt     int                      `json:"cnt"`
	List    []WeatherPodcastByPeriod `json:"list"`
}

type WeatherData struct {
	Date        string  `json:"date"`
	Temp        float64 `json:"temp"`
	Humidity    int     `json:"humidity"`
	WeatherIcon string  `json:"weathericon"`
	WindSpeed   float64 `json:"windspeed"`
	Rain        float64 `json:"rain"`
	Snow        float64 `json:"snow"`
}

type Race struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Lat   string `json:"lat"`
	Lon   string `json:"lon"`
	Begin string `json:"begin"`
	End   string `json:"end"`
}

type Location struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}
