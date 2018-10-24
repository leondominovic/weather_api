package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	// Dodatna biblioteka koja nam omogućuje
	// bolje parsiranje stringova u time varijable
	"github.com/araddon/dateparse"
)

// GetWeatherFromOpenWeather dohvaća prognoze za određenu lokaciju,
// filtrira ih preuzimajući samo one prognoze koje nam trebaju i to vraća.
func GetWeatherFromOpenWeather(lat, lon, start, end string) (data []WeatherData, err error) {

	// Ovdje unesite svoj API ključ za Open Weather API.
	apiKey := "e9ba147f034a8689b2fda25ac47794f5"
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?lat=%s&lon=%s&units=metric&lang=hr&APPID=%s", lat, lon, apiKey)

	// Incijaliziramo praznu listu strukture WeatherPodcastByPeriod.
	// U tu listu ćemo spremiti samo one prognoze koje nam odgovaraju
	filteredForecastData := []WeatherData{}

	// Izrada zahtjeva
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return filteredForecastData, err
	}

	// Kreiranje HTTP klijenta
	// Timeout postavljamo na maksimum 2 sekunde,
	// kako bi izbjegli zastoj servisa u slučaju nedostupnosti poslužitelja
	client := &http.Client{
		Timeout: time.Second * 2,
	}
	if err != nil {
		return filteredForecastData, err
	}

	// Slanje zahtijeva poslužitelju preko klijenta
	// "Do" šalje HTTP zahtjeva i vraća jedan
	res, err := client.Do(req)
	if err != nil {
		return filteredForecastData, err
	}

	// Potrebno je zatvoriti res.Body
	// nakon čitanja podataka iz njega.
	defer res.Body.Close()

	var forecastData WeatherPodcast

	// Koristimo json.Decode za čitanje strema JSON podataka.
	// Pri čitanju spremamo samo one podatke prognoze koje nam trebaju,
	// a definirani su u strukturi WeatherPodcast
	if err := json.NewDecoder(res.Body).Decode(&forecastData); err != nil {
		return filteredForecastData, err
	}

	// Preko Weather API-ja primili smo podatke za idućih pet dana,
	// ali nama trebaju samo podaci za viken i to u intervalu od 9:00 do 18:00.

	temp := WeatherData{}
	j := 0
	// Sami odgovor OpenWeather API-ja nam govori kollika je lista prognoza u atributu Cnt
	for i := 0; i < forecastData.Cnt; i++ {

		// Pretvaramo vrijeme prognoze u tip podataka Time prikladnim za rukovanje
		t, _ := dateparse.ParseLocal(strconv.Itoa(forecastData.List[i].Dt))
		var Start, End time.Time
		if start != "o" {
			Start, _ = dateparse.ParseLocal(start)
			End, _ = dateparse.ParseLocal(end)
		} else {
			Start = time.Now()
			year, _, _ := Start.Date()
			End = time.Date(year+1, 2, 15, 0, 0, 0, 0, time.Local)
		}
		// Preuzimamo samo one prognoze koje su
		// u intervalu utrke
		if (t.After(Start) || t.Equal(Start)) && (t.Before(End) || t.Equal(End)) && t.Hour() != 21 {
			temp.Date = t.Format(time.RFC3339)
			temp.Humidity = forecastData.List[i].Main.Humidity
			temp.Temp = forecastData.List[i].Main.Temp
			temp.Rain = forecastData.List[i].Rain.TreeH
			temp.WindSpeed = forecastData.List[i].Wind.Speed
			temp.WeatherIcon = forecastData.List[i].Weather[0].Description
			temp.Snow = forecastData.List[i].Snow.TreeH
			filteredForecastData = append(filteredForecastData, temp)
			j++
		}
	}

	return filteredForecastData, err
}

// AutomaticUpdate vrši automatsko ažuriranje podataka u bazi,
func AutomaticUpdate() {

	// Dohvaćamo podatke utrke koje još nisu završile,
	races, err := GetNotFinishedRaces()
	if err != nil {
		log.Printf("Zaustavljamo automatsko ažuriranje. Greska:%v", err)
		return
	}

	// Kako ne bi za različite utrke na istim lokacijama
	// dohvaćali iste vremenske podatke, dohvaćene lokacije spremamu
	// u listu i svaki put provjeravamo.
	var alredyFetched []Location
	alredyFetched = append(alredyFetched, Location{})
	numCalls := 1
	var allData []AllData
	for _, race := range races {

		if areAlredyFetched(alredyFetched, Location{race.Lat, race.Lon}) {

			// Besplatna verzija Open Weather API-ja ima ograničene od 60 zahtjeva
			// po minuti, ako se pređe ta granica, blokira se pristup api-ju.
			// Kako bi to izbjegli pratimo broj poziva API-ja. Ako se pređe granica
			// od 40, rutina čeka minutu kako bi nastavila posao. Ovim nije blokirana cijela
			// aplikacija, nego samo jedna go rutina.
			// Ostavljamo mogućnost da pored ovih poziva, u isto vrijeme, bude još drugih,
			// zato ne riskiramo s brojem od 60, nego uzimamo sigurniji broj od 45.
			if numCalls > 45 {
				fmt.Println("Velik broj zahtjeva. Pričekajte minutu za nastavak.")
				time.Sleep(1 * time.Minute)
				numCalls = 1
			}

			data, err := GetWeatherFromOpenWeather(race.Lat, race.Lon, "o", "o")
			if err != nil {
				log.Printf(`Zaustavljamo pokušaj automatsko ažuriranje. 
								Neuspješan dohvat vremenske prognoze. Greska:%v`, err)
				return
			}

			allData = append(allData, AllData{race.LocID, data})
			if numCalls == 1 {
				alredyFetched[0] = Location{race.Lat, race.Lon}
			} else {
				alredyFetched = append(alredyFetched, Location{race.Lat, race.Lon})
			}
			numCalls++
		}

	}
	err = UpdateWeather(allData)
	if err != nil {
		log.Printf(`Zaustavljamo pokušaj automatsko ažuriranje.
					Neuspješno ažuriranje podataka. Greska:%v`, err)
		return
	}

	return
}

// Funkcija koja provjeravamo da li smo za određenu lokaciju već dohvatili podatke
func areAlredyFetched(alredyFetched []Location, location Location) bool {
	for _, loc := range alredyFetched {
		if loc.Lat == location.Lat && loc.Lon == location.Lon {
			return false
		}
	}
	return true
}

// CheckData provjerava podataka za CreateRace and UpdateRace
func CheckData(naziv, lat, lon, pocetak, kraj string) (start, end time.Time, err error) {
	// Provjera da li ima praznih varijabli
	if naziv == "" || lat == "" || lon == "" || pocetak == "" || kraj == "" {
		err = errors.New("nisu poslani svi podatci")
		return start, end, err
	}

	// Pretvaramo vrijeme početka i kraja utrke
	// u tip Time prikladnim za rukovanje
	start, err = dateparse.ParseLocal(pocetak)
	end, err = dateparse.ParseLocal(kraj)
	if err != nil {
		err = errors.New("format vremena održavanja utrke je nevaljan")
		return start, end, err
	}

	// Provjera mogućih grešaka
	// if int(start.Weekday()) != int(end.Weekday()) {
	// 	err = errors.New("utrka mora započeti i završiti isti dan")
	// 	return start, end, err
	// } else if int(start.Weekday()) != 6 && int(start.Weekday()) != 0 {
	// 	err = errors.New("utrka se mora održavati vikendom")
	// 	return start, end, err
	// } else
	if start.Before(time.Now()) || end.Before(time.Now()) {
		err = errors.New("utrka se ne može održavati u prošlosti")
		return start, end, err
	} else if start.After(end) {
		err = errors.New("utrka se ne može prije završiti nego što je počela")
	} else if start.Equal(end) {
		err = errors.New("utrka ne može završiti i započeti u isto vrijeme")
	}
	//  else if (start.Hour() < 9 || start.Hour() > 18) || (end.Hour() < 9 || end.Hour() > 18) {
	// 	err = errors.New("utrka se mora održavati između 9 i 18 sati")
	// }
	return start, end, err
}
