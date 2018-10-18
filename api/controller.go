package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	// Dodatna biblioteka koja nam omogućuje
	// bolje parsiranje stringova u time varijable
	"github.com/araddon/dateparse"
)

// Ova funkcija dohvaća prognoze za određenu lokaciju,
// filtrira ih preuzimajući samo one prognoze koje nam trebaju i to vraća.
func GetWeatherFromOpenWeather(lat, lon, start, end string) (err error, data []WeatherData) {

	// Ovdje unesite svoj API ključ za Open Weather API.
	apiKey := "?????"
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?lat=%s&lon=%s&units=metric&lang=hr&APPID=%s", lat, lon, apiKey)

	// Incijaliziramo praznu listu strukture WeatherPodcastByPeriod.
	// U tu listu ćemo spremiti samo one prognoze koje nam odgovaraju
	filteredForecastData := []WeatherData{}

	// Izrada zahtjeva
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err, filteredForecastData
	}

	// Kreiranje HTTP klijenta
	// Timeout postavljamo na maksimum 2 sekunde,
	// kako bi izbjegli zastoj servisa u slučaju nedostupnosti poslužitelja
	client := &http.Client{
		Timeout: time.Second * 2,
	}
	if err != nil {
		return err, filteredForecastData
	}

	// Slanje zahtijeva poslužitelju preko klijenta
	// "Do" šalje HTTP zahtjeva i vraća jedan
	res, err := client.Do(req)
	if err != nil {
		return err, filteredForecastData
	}

	// Potrebno je zatvoriti res.Body
	// nakon čitanja podataka iz njega.
	defer res.Body.Close()

	var forecastData WeatherPodcast

	// Koristimo json.Decode za čitanje strema JSON podataka.
	// Pri čitanju spremamo samo one podatke prognoze koje nam trebaju,
	// a definirani su u strukturi WeatherPodcast
	if err := json.NewDecoder(res.Body).Decode(&forecastData); err != nil {
		return err, filteredForecastData
	}

	// Preko Weather API-ja primili smo podatke za idućih pet dana,
	// ali nama trebaju samo podaci za viken i to u intervalu od 9:00 do 18:00.

	temp := WeatherData{}
	j := 0
	// Sami odgovor OpenWeather API-ja nam govori kollika je lista prognoza u atributu Cnt
	for i := 0; i < forecastData.Cnt; i++ {

		// Pretvaramo vrijeme prognoze u tip podataka Time prikladnim za rukovanje
		t, _ := dateparse.ParseLocal(forecastData.List[i].DtTxt)
		Start, _ := dateparse.ParseLocal(start)
		End, _ := dateparse.ParseLocal(end)

		// Preuzimamo samo one prognoze koje su
		// u intervalu utrke
		if (t.After(Start) || t.Equal(Start)) && (t.Before(End) || t.Equal(End)) && t.Hour() != 21 {
			temp.Date = forecastData.List[i].DtTxt
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

	return err, filteredForecastData
}

// Ova funkcija vrši automatsko ažuriranje podataka u bazi,
func AutomaticUpdate() {

	// Dohvaćamo podatke utrke koje još nisu završile,
	// u suprotnom vraćamo grešku
	races, err := GetRaces()
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
			startOfDay, endOfDay := PrepareDateTime(race.Begin, race.End)

			err, data := GetWeatherFromOpenWeather(race.Lat, race.Lon, startOfDay, endOfDay)
			if err != nil {
				log.Printf(`Zaustavljamo pokušaj automatsko ažuriranje. 
								Neuspješan dohvat vremenske prognoze. Greska:%v`, err)
				return
			}

			// Ova funkcija će izvršiti ažuriranje postojećih prognoza
			err = UpdateWeather(data, race.Lat, race.Lon)
			if err != nil {
				log.Printf(`Zaustavljamo pokušaj automatsko ažuriranje. 
								Neuspješno ažuriranje podataka. Greska:%v`, err)
				return
			}

			// Ako postoji utrka koja još nema prognoza, a nije prošla, ova će
			// funkcija dodati prognoze za nju. Ako takve postoje.
			err = FirstTime(data, race.Lat, race.Lon)
			if err != nil {
				log.Printf(`Zaustavljamo pokušaj automatsko ažuriranje. 
								Neuspješno ažuriranje podataka. Greska:%v`, err)
				return
			}
			if numCalls == 1 {
				alredyFetched[0] = Location{race.Lat, race.Lon}
			} else {
				alredyFetched = append(alredyFetched, Location{race.Lat, race.Lon})
			}
			numCalls++
		}

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

// Provjera podataka za CreateRace and UpdateRace
func CheckData(naziv, lat, lon, pocetak, kraj string) (start, end time.Time, err error) {
	// Provjera da li ima praznih varijabli
	if naziv == "" || lat == "" || lon == "" || pocetak == "" || kraj == "" {
		err = errors.New("Nisu poslani svi podatci!")
		return start, end, err
	}

	// Pretvaramo vrijeme početka i kraja utrke
	// u tip Time prikladnim za rukovanje
	start, err = dateparse.ParseLocal(pocetak)
	end, err = dateparse.ParseLocal(kraj)
	if err != nil {
		err = errors.New("Format vremena održavanja utrke je nevaljan!")
		return start, end, err
	}

	// Provjera mogućih grešaka
	if int(start.Weekday()) != int(end.Weekday()) {
		err = errors.New("Utrka mora započeti i završiti isti dan!")
		return start, end, err
	} else if int(start.Weekday()) != 6 && int(start.Weekday()) != 0 {
		err = errors.New("Utrka se mora održavati vikendom!")
		return start, end, err
	} else if start.Before(time.Now()) || end.Before(time.Now()) {
		err = errors.New("Utrka se ne može održavati u prošlosti!")
		return start, end, err
	} else if start.After(end) {
		err = errors.New("Utrka se ne može prije završiti nego što je počela!")
	} else if (start.Hour() < 9 || start.Hour() > 18) || (end.Hour() < 9 || end.Hour() > 18) {
		err = errors.New("Utrka se mora održavati između 9 i 18 sati!")
	} else if start.Equal(end) {
		err = errors.New("Utrka ne može završiti i započeti u isto vrijeme!")
	}
	return start, end, err
}

// Priprema vremena za ažuriranje
func PrepareDateTime(start, end string) (string, string) {
	startOfDay, _ := dateparse.ParseLocal(start)
	endOfDay, _ := dateparse.ParseLocal(end)

	year, month, day := startOfDay.Date()
	newStart := time.Date(year, month, day, 9, 0, 0, 0, time.Local)

	year, month, day = endOfDay.Date()
	newEnd := time.Date(year, month, day, 18, 0, 0, 0, time.Local)

	return newStart.String(), newEnd.String()
}
