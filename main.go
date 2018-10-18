package main

import (
	"io"
	"log"
	"os"

	"weather_api/api"
	// Jednostavan i brz HTTP web framework
	"github.com/gin-gonic/gin"
	// Biblioteka koja nam omogućuje zadavanje
	// vremnski određenih automatskih zadaća
	"github.com/jasonlvhit/gocron"
)

// export DBUSER="weather_api_user"; export DBPASS=jud34DZ1; export DBHOST="localhost"; export DBNAME="weather_api_db"; export DBPORT="5432";

func initializeRoutes() *gin.Engine {
	router := gin.Default()
	router.Use(
		// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
		// By default gin.DefaultWriter = os.Stdout
		gin.Logger(),
		// Recovery middleware sprječava zastoj u slučaju panic-a i zapisuje 500 ako postoji jedan takav.
		gin.Recovery(),
	)
	// Grupiranje ruta pod dodatnu rutu api omogućava nam da izbjegnemo sukob
	// sa drugim rutama web aplikacije, a podruta v1 nam omogućava,
	// u slučaju kasnije nadogradnje api-ja, lakše prebacivanje na njegove različite verzije.
	v1 := router.Group("api/v1")
	{
		v1.GET("/utrka/:id/prognoza", api.GetWeatherHandler)
		v1.POST("/utrka", api.CreateRaceHandler)
		v1.GET("/utrka/:id", api.GetRaceHandler)
		v1.PUT("/utrka/:id", api.UpdateRaceHandler)
		v1.DELETE("/utrka/:id", api.DeleteRaceHandler)
	}

	return router
}

func main() {
	// Zapisivanje svih HTTP zahtjeca u gin.log
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)

	// Zapisivanje grešaka u datoteku greske.log
	f1, _ := os.Create("greske.log")
	log.SetOutput(io.MultiWriter(f1))

	// Inicijalizacija rutera
	router := initializeRoutes()
	// Incijalizacije konekcije prema bazi
	api.InitializeDb()

	// Pokretanje procesa koji je zadužen za automatsko
	// ažuriranje svakih 6 sati i procesa koji se brine o
	// brisanju prognoza za utrke koje su prošle svaki sat vremena.
	// Proces za ažuriranje pokrećemo u posebnoj goroutine, kako,
	// u sluačaju čekanja, kod dohvaćanja ažuriranja prognoza, cijeli
	// api ne bi postao nedostupan.
	go gocron.Every(6).Hours().Do(api.AutomaticUpdate)
	gocron.Every(1).Hours().Do(api.DeleteWeatherPodcast)
	gocron.Start()

	// Po default-u port je :8080 osim ako je
	// PORT sustavna varijabla definirana drukčije.
	router.Run()
}
