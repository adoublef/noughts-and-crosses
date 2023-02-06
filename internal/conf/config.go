package conf

import (
	"flag"
	"os"
	"strconv"
)

var (
	PORT      int
	ClientURI string

	NATSURI   string
	NATSToken string
	NATSSeed  string

	SMTPHost     string
	SMTPUsername string
	SMTPPassword string

	PostgresUsername string
	PostgresPassword string
	PostgresDBName   string
	PostgresHost     string
)

func init() {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	flag.IntVar(&PORT, "port", port, "port")

	flag.StringVar(&ClientURI, "client-uri", os.Getenv("CLIENT_URI"), "client uri")

	flag.StringVar(&SMTPHost, "smtp-host", os.Getenv("SMTP_HOST"), "smtp host")
	flag.StringVar(&SMTPUsername, "smtp-username", os.Getenv("SMTP_EMAIL"), "smtp username")
	flag.StringVar(&SMTPPassword, "smtp-password", os.Getenv("SMTP_PASS"), "smtp password")

	flag.StringVar(&PostgresHost, "postgres-host", os.Getenv("POSTGRES_HOST"), "postgres host")
	flag.StringVar(&PostgresUsername, "postgres-username", os.Getenv("POSTGRES_USER"), "postgres username")
	flag.StringVar(&PostgresPassword, "postgres-password", os.Getenv("POSTGRES_PASS"), "postgres password")
	flag.StringVar(&PostgresDBName, "postgres-dbname", os.Getenv("POSTGRES_DB"), "postgres dbname")

	flag.StringVar(&NATSURI, "nats-uri", os.Getenv("NATS_URI"), "nats uri")
	flag.StringVar(&NATSToken, "nats-token", os.Getenv("NATS_TOKEN"), "nats token")
	flag.StringVar(&NATSSeed, "nats-seed", os.Getenv("NATS_SEED"), "nats seed")

	flag.Parse()
}
