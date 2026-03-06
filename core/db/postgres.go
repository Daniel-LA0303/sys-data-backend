package dbpost

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Driver de Postgres
)

// ConnectPostgres abstrae toda la lógica de reintentos
func ConnectPostgres(dsn string) *sqlx.DB {
	counts := 0

	for {
		connection, err := sqlx.Open("postgres", dsn)
		if err != nil {
			log.Println("Postgres no está listo todavía...")
		} else {
			// Intentamos un Ping para asegurar que la conexión es real
			err = connection.Ping()
			if err == nil {
				log.Println("¡Conectado a Postgres con éxito!")
				return connection
			}
			log.Println("Error haciendo ping a Postgres:", err)
		}

		if counts > 10 {
			log.Println("No se pudo conectar a Postgres tras 10 intentos.")
			return nil
		}

		log.Println("Reintentando en 2 segundos...")
		counts++
		time.Sleep(2 * time.Second)
	}
}
