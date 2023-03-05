package conf

import (
	"flag"
	"os"
	"strconv"
)

var (
	PORT         int
	ClientURI    string
	SMTPHost     string
	SMTPUsername string
	SMTPPassword string
	NATSURI      string
	NATSToken    string
	NATSSeed     string
	DBURL        string
	JWTSecret    string
)

func init() {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	flag.IntVar(&PORT, "port", port, "port")

	flag.StringVar(&ClientURI, "client-uri", os.Getenv("CLIENT_URI"), "client uri")

	flag.StringVar(&DBURL, "database-uri", os.Getenv("DATABASE_URL"), "database uri")

	flag.StringVar(&JWTSecret, "jwt-secret", os.Getenv("JWT_SECRET"), "jwt secret")

	flag.StringVar(&SMTPHost, "smtp-host", os.Getenv("SMTP_HOST"), "smtp host")
	flag.StringVar(&SMTPUsername, "smtp-username", os.Getenv("SMTP_EMAIL"), "smtp username")
	flag.StringVar(&SMTPPassword, "smtp-password", os.Getenv("SMTP_PASS"), "smtp password")

	flag.StringVar(&NATSURI, "nats-uri", os.Getenv("NATS_URI"), "nats uri")
	flag.StringVar(&NATSToken, "nats-token", os.Getenv("NATS_TOKEN"), "nats token")
	flag.StringVar(&NATSSeed, "nats-seed", os.Getenv("NATS_SEED"), "nats seed")

	flag.Parse()
}
