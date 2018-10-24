package api

// WeatherPodcastByPeriod struktura
type WeatherPodcastByPeriod struct {
	Dt   int `json:"dt"`
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
}

// WeatherPodcast struktura
type WeatherPodcast struct {
	Cod     string                   `json:"cod"`
	Message float64                  `json:"message"`
	Cnt     int                      `json:"cnt"`
	List    []WeatherPodcastByPeriod `json:"list"`
}

// WeatherData struktura
type WeatherData struct {
	Date        string  `json:"date"`
	Temp        float64 `json:"temp"`
	Humidity    int     `json:"humidity"`
	WeatherIcon string  `json:"weathericon"`
	WindSpeed   float64 `json:"windspeed"`
	Rain        float64 `json:"rain"`
	Snow        float64 `json:"snow"`
}

// Race struktura
type Race struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Lat   string `json:"lat"`
	Lon   string `json:"lon"`
	Begin string `json:"begin"`
	End   string `json:"end"`
}

// NotFinishedRace struktura
type NotFinishedRace struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Lat   string `json:"lat"`
	Lon   string `json:"lon"`
	Begin string `json:"begin"`
	End   string `json:"end"`
	LocID int    `json:"loc"`
}

// Location struktura
type Location struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// AllData struktura je za sve
// prognoze na određenoj lokaciji
type AllData struct {
	Loc  int
	Data []WeatherData
}
