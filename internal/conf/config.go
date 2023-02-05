package conf

import (
	"flag"
	"os"
)

var SMTPHost string
var SMTPUsername string
var SMTPPassword string
var NATSURI string
var PostgresUsername string
var PostgresPassword string
var PostgresDBName string

func init() {
	flag.StringVar(&SMTPHost, "smtp-host", os.Getenv("SMTP_HOST"), "smtp host")
	flag.StringVar(&SMTPUsername, "smtp-username", os.Getenv("SMTP_EMAIL"), "smtp username")
	flag.StringVar(&SMTPPassword, "smtp-password", os.Getenv("SMTP_PASS"), "smtp password")
	flag.StringVar(&NATSURI, "nats-uri", os.Getenv("NATS_URI"), "nats uri")
	flag.StringVar(&PostgresUsername, "postgres-username", os.Getenv("POSTGRES_USER"), "postgres username")
	flag.StringVar(&PostgresPassword, "postgres-password", os.Getenv("POSTGRES_PASS"), "postgres password")
	flag.StringVar(&PostgresDBName, "postgres-dbname", os.Getenv("POSTGRES_DB"), "postgres dbname")

	flag.Parse()
}
