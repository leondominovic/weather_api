package api

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	// Driveri za PostgreSQL koji proširuju standardno database/sql sučelje
	"github.com/lib/pq"
	// Dodatna biblioteka koja nam omogućuje
	// bolje parsiranje stringova u time varijable
)

var db *sql.DB

// Ove varijable treba exportirati u sistemsko okruženje
// sa njima pripadajućim vrijednostima
const (
	dbUser = "DBUSER"
	dbPass = "DBPASS"
	dbHost = "DBHOST"
	dbName = "DBNAME"
	dbPort = "DBPORT"
)

// InitializeDb incijalizira konekciju na bazi
func InitializeDb() {

	// Poziv dbConfig funkcije da dohvati potrebne varijable
	config := dbConfig()
	var err error

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbHost], config[dbPort],
		config[dbUser], config[dbPass], config[dbName])

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Print("Ne može se uspostaviti konkcija prema bazi")
		panic(err)
	}

	// Provjeravamo da li postoji konekcija
	err = db.Ping()
	if err != nil {
		log.Print("Ne može se uspostaviti konkcija prema bazi")
		panic(err)
	}
	fmt.Println("Uspješno spojeno na bazu!")
	return
}

// Ova funkcija preuzima sistemske varijable i vraća ih u map conf varijabli
func dbConfig() map[string]string {
	conf := make(map[string]string)
	host, err := os.LookupEnv(dbHost)
	if !err {
		log.Print("DBHOST environment variable nije podešena!")
		panic("DBHOST environment variable nije podešena!")
	}
	port, err := os.LookupEnv(dbPort)
	if !err {
		log.Print()
		panic("DBPORT environment variable nije podešena!")
	}
	user, err := os.LookupEnv(dbUser)
	if !err {
		log.Print("DBUSER environment variable nije podešena!")
		panic("DBUSER environment variable nije podešena!")
	}
	password, err := os.LookupEnv(dbPass)
	if !err {
		log.Print("DBPASS environment variable nije podešena!")
		panic("DBPASS environment variable nije podešena!")
	}
	name, err := os.LookupEnv(dbName)
	if !err {
		log.Print("DBNAME environment variable nije podešena!")
		panic("DBNAME environment variable nije podešena!")
	}
	conf[dbHost] = host
	conf[dbPort] = port
	conf[dbUser] = user
	conf[dbPass] = password
	conf[dbName] = name
	return conf
}

// GetWeather dohvaća određene prognoze za utrku koja ima određeni id
func GetWeather(id int) (data WeatherData, err error) {

	// Ovom naredvom vratiti ćemo samo
	// prognoze koje nisu prošle.
	sqlStr := `WITH race AS (
					SELECT location_id,
					race_start,
					race_end 
				FROM 
					races 
				WHERE race_id=$1)

				SELECT icon, forecast_time, rain, snow, temperature, humidity, wind_speed
				FROM
					forecasts
				WHERE location_id = (SELECT location_id FROM race) 
				AND
					forecast_time >= (SELECT race_start FROM race)
				AND 
					forecast_time <= (SELECT race_end FROM race)
				AND
					forecast_time >= CURRENT_TIMESTAMP`

	// Dohvaćamo prognozu iz baze koji odgovaraju,
	// u suprotnom vraćamo grešku
	err = db.QueryRow(sqlStr, id).Scan(&data.WeatherIcon,
		&data.Date,
		&data.Rain,
		&data.Snow,
		&data.Temp,
		&data.Humidity,
		&data.WindSpeed)

	// Ovisno o postojanju ili nepostojanju greške
	// vračamo odgovarajući odgovor
	switch err {
	case sql.ErrNoRows:
		log.Println("nepostojeći id", err)
		err = errors.New("nepostojeći id ili je zadana utrka prošla")
		return data, err
	case nil:
		return data, err
	default:
		log.Println("greška pri dohvaćanju podataka", err)
		err = errors.New("greška pri dohvaćanju podataka")
		return data, err
	}
}

// UpdateWeather ažurira postojeće prognoze
func UpdateWeather(allData []AllData) (err error) {

	var listData [][]string
	var elem []string
	i := 0
	for _, data := range allData {
		for _, element := range data.Data {
			elem := append(elem,
				fmt.Sprintf("%v", data.Loc),
				element.WeatherIcon,
				element.Date,
				fmt.Sprintf("%v", element.Rain),
				fmt.Sprintf("%v", element.Snow),
				fmt.Sprintf("%v", element.Temp),
				fmt.Sprintf("%v", element.Humidity),
				fmt.Sprintf("%v", element.WindSpeed))

			listData = append(listData, elem)
		}
		i++
	}

	// Priprema SQL naredbe za ažuriranje.
	// Ova naredba ažurira podatke prognoze.
	sqlStr := `SELECT update_weather($1)`

	_, err = db.Exec(sqlStr, pq.Array(listData))

	return err
}

// FirstTime služi za dodavanje prognoza za utrke
// koje ih još nemaju
func FirstTime(data []WeatherData, lat, lon string) error {

	// Naredba za dohvaćanje id-a utrka koje još nemaju prognozu.
	sqlStr1 := `SELECT utrka_id
					FROM utrke
					WHERE lat = $1 
					AND lon = $2
					AND raceStart > CURRENT_TIMESTAMP
					AND NOT EXISTS 
					(SELECT * FROM prognoze
							  WHERE utrke.utrka_id = utrka_id)`

	// Izvršavanje upita nad bazom, ako je neuspješan vraćamo grešku.
	rows, err := db.Query(sqlStr1, lat, lon)
	if err != nil {
		log.Panicln("Greška pri dohvaćanju ID-ova pri automaskom ažuriranju!", err)
		return err
	}

	var idList []uint
	var id uint

	//Čitamo dohvaćene id-ove
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			log.Panicln("Greška pri dohvaćanju ID-ova pri automaskom ažuriranju!", err)
			return err
		}
		idList = append(idList, id)
	}

	// Ako ne postoji utrka za koje ne postoje dohvaćene prognoze
	// prekidamo operaciju, jer nema potrebe nastavljati.
	// Za njih je več funkcija UpdateWeather izvršilo ažuriranje podataka
	if len(idList) == 0 {
		return nil
	}

	sqlStr := `	INSERT INTO prognoze (
							utrka_id,
							ikona_stanja,
							vrijeme_prognoze,
							kisa,
							snijeg,
							temperatura,
							vlaznost,
							brzina_vjetra)
				VALUES ($1,$2,$3,$4,$5,$6, $7, $8)`

	// Za svaku utrku za koju ne postoje još progonoze,
	// a nisu prošle dodajemo dohvaćane prognoze za nju.
	for _, oneID := range idList {
		for _, row := range data {
			_, err := db.Exec(sqlStr,
				oneID,
				row.WeatherIcon,
				row.Date,
				row.Rain,
				row.Snow,
				row.Temp,
				row.Humidity,
				row.WindSpeed)
			if err != nil {
				log.Println("Greška pri dodavanju prognoza!", err)
				return err
			}
		}
	}

	return nil
}

// CreateRace dodaje nove utrku u bazu
func CreateRace(name, lat, lon string, raceStart, raceEnd time.Time) (raceID, locID int64, err error) {

	// Ovdje koristimo funkciju za dodavanje nove utrke.
	// Ona provjera da li postoji lokacija nove utrke u bazi.
	sqlStr := `SELECT create_race($1, $2, $3, $4, $5)`

	var twoID []sql.NullInt64
	err = db.QueryRow(sqlStr, name, lat, lon, raceStart, raceEnd).Scan(pq.Array(&twoID))

	return twoID[0].Int64, twoID[1].Int64, err
}

// InsertWeatherPodcast za zadanu utrku ubacuje prognoze u bazu
func InsertWeatherPodcast(data []WeatherData, locID int64) (err error) {

	// Priprema SQL naredbe za insert
	sqlStr := `INSERT INTO 
					forecasts
				VALUES`

	// Incijalizacija varijable koja će biti
	// lista svih podataka koje treba unijeti u bazu.
	// Biblioteka database/sql pruža zaštitu protiv SQL injection-a,
	// pa nije potrebna dodatna prvojera.
	vals := []interface{}{}
	l := len(data)
	for i := 0; i < l; i++ {

		// Dodajemo potrebne argumente u SQL naredbu
		// za kasnije učitavanje podataka
		sqlStr += fmt.Sprintf(" ($%v, $%v, $%v, $%v, $%v, $%v, $%v, $%v),",
			i*8+1, i*8+2, i*8+3, i*8+4, i*8+5, i*8+6, i*8+7, i*8+8)

		// Učitavamo podatke u listu
		vals = append(vals,
			locID,
			data[i].WeatherIcon,
			data[i].Date,
			data[i].Rain,
			data[i].Snow,
			data[i].Temp,
			data[i].Humidity,
			data[i].WindSpeed)
	}

	// Uklanjamo posljednji zarez iz naredbe
	sqlStr = strings.TrimSuffix(sqlStr, ",")
	// Ako već postoji prognoza za određenu lokaciju i vrijeme onda
	// preskaćemo te dodavenj prognoze.
	sqlStr += " ON CONFLICT ON CONSTRAINT forecasts_location_id_forecast_time_key DO NOTHING"
	// Izvrašavamo sql naredbu i dodajemo podatke u bazu.
	// U suprotnom vraćmo grešku.
	_, err = db.Exec(sqlStr, vals...)
	return err
}

// GetRace dohvaća podataka o utrkama
func GetRace(id int64) (data Race, err error) {
	// Ovom naredvom vratiti ćemo samo
	// prognoze koje nisu prošle.
	sqlStr := `SELECT 
					race_id,
					name,
					race_start,
					race_end,
					locations.lat,
					locations.lon 
			   FROM 
					races  
				NATURAL INNER JOIN 
					locations 
				WHERE 
					races.race_id=$1`

	// Dohvaćamo retke iz baze koji odgovaraju,
	// u suprotnom vraćamo grešku
	err = db.QueryRow(sqlStr, id).Scan(&data.ID, &data.Name, &data.Begin, &data.End, &data.Lat, &data.Lon)

	// Ovisno o postojanju ili nepostojanju greške
	// vračamo odgovarajući odgovor
	switch err {
	case sql.ErrNoRows:
		log.Println("nepostojeći id", err)
		err = errors.New("nepostojeći id")
		return data, err
	case nil:
		return data, err
	default:
		log.Println("greška pri dohvaćanju podataka", err)
		err = errors.New("greška pri dohvaćanju podataka")
		return data, err
	}
}

// DeleteRace briše utrke
func DeleteRace(id int) (err error) {
	// Ovom naredvom vratiti ćemo samo
	// prognoze koje nisu prošle.
	sqlStr := `SELECT delete_race($1)`

	var find bool
	// Brišemo utrku,
	// u suprotnom vraćamo grešku
	err = db.QueryRow(sqlStr, id).Scan(&find)

	// Ovisno o postojanju ili nepostojanju greške
	// vračamo odgovarajući odgovor
	if err != nil {
		log.Println("problem pri brisanju utrke", err)
		err = errors.New("problem pri brisanju utrke")
		return err
	}

	// Provjeravamo da li je Id utrke bio važeći.
	if !find {
		log.Println("Nepostojeći id!", err)
		err = errors.New("nepostojeći id")
		return err
	}

	return nil
}

// UpdateRace ažurira podataka o utrci
func UpdateRace(id int64, name, lat, lon string, start, end time.Time) (returnValue int, err error) {

	sqlStr := `SELECT update_race($1, $2, $3, $4, $5, $6)`

	err = db.QueryRow(sqlStr, id, name, start, end, lat, lon).Scan(&returnValue)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return returnValue, err
}

// GetNotFinishedRaces služi za dohvat nezavršenih utrka iz baze
func GetNotFinishedRaces() (races []NotFinishedRace, err error) {
	sqlStr := `SELECT 
					*
				FROM 
					races  
				NATURAL INNER JOIN 
					locations 
				WHERE 
					race_end > CURRENT_TIMESTAMP`

	// Dohvaćamo prognozu iz baze koji odgovaraju,
	// u suprotnom vraćamo grešku
	rows, err := db.Query(sqlStr)
	if err != nil {
		log.Println(err)
		return races, err
	}

	// Potrebno je zatvoriti konekciju kada se podatci učitaju
	defer rows.Close()

	// Čitanje dohvaćenih podataka
	var row NotFinishedRace
	for rows.Next() {
		err = rows.Scan(&row.LocID, &row.ID, &row.Name, &row.Begin, &row.End, &row.Lat, &row.Lon)
		if err != nil {
			log.Println(err)
			return races, err
		}
		races = append(races, row)
	}

	return races, err
}

//DeleteWeatherPodcast služi za brisanje prognoza za utrke koje su prošle
func DeleteWeatherPodcast() {

	// Naredba za dohvaćanje id-a utrka koje imaju prognozu, a prošle su.
	sqlStr1 := `DELETE FROM 
							forecasts
						WHERE 
							forecast_time < CURRENT_TIMESTAMP`

	// Izvršavanje upita nad bazom, ako je neuspješan vraćamo grešku.
	rows, err := db.Query(sqlStr1)
	if err != nil {
		log.Panicln("Greška pri dohvaćanju ID-ova pri automaskom brisanju!", err)
		return
	}

	var idList []uint
	var id uint

	//Čitamo dohvaćene id-ove
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			log.Panicln("Greška pri dohvaćanju ID-ova pri automaskom brisanju!", err)
			return
		}
		idList = append(idList, id)
	}

	// Ako ne postoje prognoze za utrke koje su prošle
	// prekidamo operaciju, jer nema prognoza koje treba izbrisati.
	if len(idList) == 0 {
		return
	}

	// Upit za brisanje prognoze
	sqlStr2 := `DELETE FROM prognoze
					WHERE 
					utrka_id = $1`

	// Brisanje prognozi za prošle utrke.
	for _, oneID := range idList {
		_, err := db.Exec(sqlStr2, oneID)
		if err != nil {
			log.Println("Problem pri brisanju prošlih utrka!", err)
			return
		}
	}
	return
}
