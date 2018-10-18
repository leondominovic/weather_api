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
	_ "github.com/lib/pq"

	// Dodatna biblioteka koja nam omogućuje
	// bolje parsiranje stringova u time varijable
	"github.com/araddon/dateparse"
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

// Incijalizacija konekcije na bazi
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

// Funkcija koja dohvaća određene prognoze za utrku koja ima određeni id
func GetWeather(id int) (data WeatherData, err error) {

	// Ovom naredvom vratiti ćemo samo
	// prognoze koje nisu prošle.
	sqlStr := `SELECT 
					ikona_stanja,
					vrijeme_prognoze,
					kisa, snijeg,
					temperatura,
					vlaznost,
					brzina_vjetra 
					FROM prognoze 
					WHERE 
					utrka_id=$1 
					AND 
					vrijeme_prognoze > CURRENT_TIMESTAMP`

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
		log.Println("Nepostojeći id!", err)
		err = errors.New("Nepostojeći id!")
		return data, err
	case nil:
		return data, err
	default:
		log.Println("Greška pri dohvaćanju podataka!", err)
		err = errors.New("Greška pri dohvaćanju podataka!")
		return data, err
	}
}

// Funkcija za ažuriranje postojećih prognoza
func UpdateWeather(data []WeatherData, lat, lon string) (err error) {

	// Priprema SQL naredbe za ažuriranje.
	// Ova naredba ažurira podatke prognoze.
	sqlStr := ` UPDATE prognoze SET
					ikona_stanja = $1,
					kisa = $2,
					snijeg =$3,
					temperatura = $4,
					vlaznost = $5,
					brzina_vjetra = $6
				WHERE
					vrijeme_prognoze = $7
				AND
				EXISTS (SELECT utrka_id FROM utrke 
						WHERE lat = $8 
						AND lon = $9 
						AND prognoze.utrka_id = utrka_id)`

	// Za svaki niz podataka vršimo ažuriranje podataka.
	// Ako nastane greška vraćamo je.
	for _, row := range data {
		_, err := db.Exec(sqlStr,
			row.WeatherIcon,
			row.Rain,
			row.Snow,
			row.Temp,
			row.Humidity,
			row.WindSpeed,
			row.Date,
			lat,
			lon)
		if err != nil {
			log.Println("Greška pri ažuriranju podataka!", err)
			return err
		}
	}

	return nil
}

// Funkcija za dodavanje prognoza za utrke
// koje ih još nemaju
func FirstTime(data []WeatherData, lat, lon string) error {

	// Naredba za dohvaćanje id-a utrka koje još nemaju prognozu.
	sqlStr1 := `SELECT utrka_id
					FROM utrke
					WHERE lat = $1 
					AND lon = $2
					AND pocetak > CURRENT_TIMESTAMP
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
	for _, oneId := range idList {
		for _, row := range data {
			_, err := db.Exec(sqlStr,
				oneId,
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

// Funkcija koja dodaje nove utrku u bazu
func CreateRace(naziv, lat, lon string, pocetak, kraj time.Time) (id int, err error) {
	// Ovaj način dodovanje u bazu, preko database/sql funkcije,
	// pruža zaštitu protiv SQL injection-a.
	// Vraćamo id utrke kako bi dodali potrebne prognoze za nju.
	sqlStr := `INSERT INTO utrke (naziv, lat, lon, pocetak, kraj)
			  VALUES ($1,$2,$3,$4,$5)
			  RETURNING utrka_id`

	row := db.QueryRow(sqlStr, naziv, lat, lon, pocetak, kraj)
	err = row.Scan(&id)
	if err != nil {
		log.Printf("Greška pri dodavanju nove utrke u bazu:%v", err)
		fmt.Printf("Greška: %v", err)
	}

	return id, err
}

// Funkcija koja za zadanu utrku ubacuje prognoze u bazu
func InsertWeatherPodcast(data []WeatherData, id int) (err error) {
	// Priprema SQL naredbe za insert
	sqlStr := `INSERT INTO prognoze (
									utrka_id,
									ikona_stanja,
									vrijeme_prognoze,
									kisa,
									snijeg,
									temperatura,
									vlaznost,
									brzina_vjetra
									) VALUES`

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
			id,
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
	// Izvrašavamo sql naredbu i dodajemo podatke u bazu.
	// U suprotnom vraćmo grešku.
	_, err = db.Exec(sqlStr, vals...)
	return err
}

// Dohvaćanje podataka o utrkama
func GetRace(id int) (data Race, err error) {
	// Ovom naredvom vratiti ćemo samo
	// prognoze koje nisu prošle.
	sqlStr := `SELECT * FROM utrke WHERE utrka_id=$1 `

	// Dohvaćamo retke iz baze koji odgovaraju,
	// u suprotnom vraćamo grešku
	err = db.QueryRow(sqlStr, id).Scan(&data.Id, &data.Name, &data.Lat, &data.Lon, &data.Begin, &data.End)

	// Ovisno o postojanju ili nepostojanju greške
	// vračamo odgovarajući odgovor
	switch err {
	case sql.ErrNoRows:
		log.Println("Nepostojeći id!", err)
		err = errors.New("Nepostojeći id!")
		return data, err
	case nil:
		return data, err
	default:
		log.Println("Greška pri dohvaćanju podataka!", err)
		err = errors.New("Greška pri dohvaćanju podataka!")
		return data, err
	}
}

// Brisanje utrke
func DeleteRace(id int) (err error) {
	// Ovom naredvom vratiti ćemo samo
	// prognoze koje nisu prošle.
	sqlStr := `DELETE FROM utrke WHERE utrka_id=$1 `

	// Brišemo utrku,
	// u suprotnom vraćamo grešku
	res, err := db.Exec(sqlStr, id)

	// Ovisno o postojanju ili nepostojanju greške
	// vračamo odgovarajući odgovor
	if err != nil {
		log.Println("Problem pri brisanju utrke!", err)
		err = errors.New("Problem pri brisanju utrke!")
		return err
	}

	// Ako je broj obrisanih redaka nula
	// to znači da ne postoji utrka sa zadanim id-om.
	if num, _ := res.RowsAffected(); num == 0 {
		log.Println("Nepostojeći id!", err)
		err = errors.New("Nepostojeći id!")
		return err
	}

	return err
}

// Ažuriranje podataka o utrci
func UpdateRace(id int, naziv, lat, lon string, start, end time.Time) (needForUpdate bool, err error) {

	// Samo ako su geografska lokacija i
	// vrijeme održavanje utrke izmijenjene
	// potrebno je i ažurirati prognozu za tu utrku.
	data, err := GetRace(id)
	oldStart, _ := dateparse.ParseLocal(data.Begin)
	oldEnd, _ := dateparse.ParseLocal(data.End)

	if data.Lat != lat && data.Lon != lon {
		needForUpdate = true
	} else {
		needForUpdate = false
	}
	if start.Equal(oldStart) && end.Equal(oldEnd) {
		needForUpdate = false
	} else {
		needForUpdate = true
	}
	sqlStr := `UPDATE utrke 
				SET naziv = $1,
					lat = $2,
					lon = $3,
					pocetak = $4,
					kraj = $5
				WHERE utrka_id = $6`
	_, err = db.Exec(sqlStr, naziv, lat, lon, start, end, id)
	if err != nil {
		er := errors.New("Neuspjelo ažuriranje utrke")
		log.Println(er, err)
		return false, er
	}

	return needForUpdate, err
}

// Funkcija za dohvat nezavršenih utrka iz baze
func GetRaces() (races []Race, err error) {
	sqlStr := `SELECT * FROM utrke WHERE kraj > CURRENT_TIMESTAMP`

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
	var row Race
	for rows.Next() {
		err = rows.Scan(&row.Id, &row.Name, &row.Lat, &row.Lon, &row.Begin, &row.End)
		if err != nil {
			log.Println(err)
			return races, err
		}
		races = append(races, row)
	}

	return races, err
}

// Funckija za brisanje prognoza za utrke koje su prošle
func DeleteWeatherPodcast() {

	// Naredba za dohvaćanje id-a utrka koje imaju prognozu, a prošle su.
	sqlStr1 := `SELECT utrka_id
					FROM utrke
					WHERE 
					kraj < CURRENT_TIMESTAMP
					AND EXISTS 
					(SELECT * FROM prognoze
							  WHERE utrke.utrka_id = utrka_id)`

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

// Funkcija kojom brišemo nevažeče prognoze
func DeleteWeatherPodcastByID(id int) (err error) {
	sqlStr := `DELETE FROM prognoze WHERE utrka_id = $1`
	_, err = db.Exec(sqlStr, id)
	if err != nil {
		log.Println("Problem pri brisanju prognoza:", err)
		return err
	}

	return nil
}
