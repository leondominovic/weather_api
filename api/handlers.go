package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	// Jednostavan i brz HTTP web framework
	"github.com/gin-gonic/gin"
)

func GetWeatherHandler(c *gin.Context) {

	// Dohvaćamo id u obliku stringa iz GET zahtjeva
	idString := c.Param("id")

	// Pretvaramo dohvaćeni u int, odmah i
	// provjeramo potencijalnu grešku.
	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Greska": "ID mora biti cijeli broj!"})
		log.Print(err)
		return
	}

	// Proslijeđujemo id funckiji koja dohvača
	// prognozu ili vraća grešku
	data, err := GetWeather(id)
	if err != nil {
		if fmt.Sprint(err) == "Greška pri dohvaćanju podataka!" {
			c.JSON(http.StatusInternalServerError, gin.H{"Greska": fmt.Sprint(err)})
		}
		c.JSON(http.StatusBadRequest, gin.H{"Greska": fmt.Sprint(err)})
		log.Print(err)
		return
	}

	c.JSON(http.StatusOK, data)
	return
}

func GetRaceHandler(c *gin.Context) {
	// Dohvaćamo id u obliku stringa iz GET zahtjeva
	idString := c.Param("id")

	// Pretvaramo dohvaćeni u int, odmah i
	// provjeramo potencijalnu grešku.
	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Greska": "ID mora biti cijeli broj!"})
		log.Print(err)
		return
	}

	// Proslijeđivamo id utrke funkcije za dohvat podataka iz baze.
	// Ovisno o rezultatu, da li ima greške, vraćamo ili podatke ili grešku.
	data, err := GetRace(id)
	if err != nil {
		if fmt.Sprint(err) == "Greška pri dohvaćanju podataka!" {
			c.JSON(http.StatusInternalServerError, gin.H{"Greska": fmt.Sprint(err)})
		}
		c.JSON(http.StatusBadRequest, gin.H{"Greska": fmt.Sprint(err)})
		log.Print(err)
		return
	}
	c.JSON(http.StatusOK, data)
	return
}

func CreateRaceHandler(c *gin.Context) {

	// Dohvat podataka iz POST zahtjeva
	naziv := c.PostForm("naziv")
	lat := c.PostForm("lat")
	lon := c.PostForm("lon")
	pocetak := c.PostForm("pocetak")
	kraj := c.PostForm("kraj")

	// Provjeravamo da li su svi zaprimiljeni podatci poslani,
	// ako nisu vraćamo odgovarajuču grešku.
	start, end, err := CheckData(naziv, lat, lon, pocetak, kraj)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Greska:": fmt.Sprint(err)})
		return
	}
	// Pozicanje funkcije za dodavanjem nove utrke
	id, err := CreateRace(naziv, lat, lon, start, end)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"Greska": fmt.Sprint(err)})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"Poruka": "Utrka je uspješno dodana!", "Id_utrke": id})
	}

	// Nakon dodavanje nove utrke potrebno
	// je dodati prognoze za novu utrku u bazu podataka.
	// Prvo dohvaćamo prognoze vremena sa Open Weather
	err, data := GetWeatherFromOpenWeather(lat, lon, pocetak, kraj)
	if err != nil {
		log.Print(err)
		fmt.Printf("Greška: %s", fmt.Sprint(err))
		return
	}

	// Potom podatke prognoze proslijeđujemo funkciji
	// koja će ih spremiti u bazu podataka
	err = InsertWeatherPodcast(data, id)
	if err != nil {
		log.Printf("Greška pri dodavanju prognoza: %v", err)
		fmt.Printf("Greška pri dodavanju prognoza: %s", err)
		return
	}
	return

}

func UpdateRaceHandler(c *gin.Context) {

	// Dohvat ID utrke koje treba izmijeniti
	idString := c.Param("id")

	// Pretvaramo dohvaćeni id u int, odmah i
	// provjeramo potencijalnu grešku.
	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Greska": "ID mora biti cijeli broj!"})
		log.Print(err)
		return
	}

	// Dohvat podataka iz PUT body zahtjeva
	naziv := c.PostForm("naziv")
	lat := c.PostForm("lat")
	lon := c.PostForm("lon")
	pocetak := c.PostForm("pocetak")
	kraj := c.PostForm("kraj")

	// Provjeravamo da li su svi zaprimiljeni podatci poslani,
	// ako nisu vraćamo odgovarajuču grešku.
	start, end, err := CheckData(naziv, lat, lon, pocetak, kraj)
	if err != nil {
		c.JSON(http.StatusBadRequest, fmt.Sprint(err))
	}

	// Pozivanje funkcije za ažuriranje utrke
	needUpdate, err := UpdateRace(id, naziv, lat, lon, start, end)

	// Ako postoji greška vraćamo je, ako ne postoji onda
	// provjeramo treba li ažurirati podatke vezane za prognozu.
	// Ako ne treba samo izlazimo iz funkcije i vraćamo odgovarajuću poruku.
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"Greska": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Poruka": "Utrka je uspješno ažurirana!", "Id": id})

	if !needUpdate {
		return
	}

	// Prvo pobrišemo stare progonoze, koje su možda nevažeće.
	err = DeleteWeatherPodcastByID(id)
	if err != nil {
		return
	}

	// Nakon ažuriranje nove utrke potrebno
	// je ažurirati prognoze.
	err, data := GetWeatherFromOpenWeather(lat, lon, pocetak, kraj)
	if err != nil {
		log.Print(err)
		fmt.Printf("Greška: %s", err)
		return
	}

	// Potom podatke prognoze proslijeđujemo funkciji
	// koja će ih spremiti u bazu podataka
	err = InsertWeatherPodcast(data, id)
	if err != nil {
		log.Printf("Greška pri ažuriranju prognoza: %v", err)
		fmt.Printf("Greška pri ažuriranju prognoza: %s", err)
		return
	}
	return
}

func DeleteRaceHandler(c *gin.Context) {
	// Dohvaćamo id u obliku stringa iz GET zahtjeva
	idString := c.Param("id")

	// Pretvaramo dohvaćeni u int, odmah i
	// provjeramo potencijalnu grešku.
	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Greska": "ID mora biti cijeli broj!"})
		log.Print(err)
		return
	}

	// Proslijeđivamo id utrke funkciji za brisanje utrke.
	// Ako je vraćena greška znači da brisanje nije uspjelo i šaljemo odgovor.
	err = DeleteRace(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Greska": fmt.Sprint(err)})
		log.Print(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"Odgovor": "Utrka uspješno izbrisana"})
	return
}
