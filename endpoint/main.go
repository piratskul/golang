package main

import (
	"database/sql"
	"fmt"
	"math"

	_ "github.com/lib/pq"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/ewkb"
)

type Endpoint struct {
	id          string
	name        string
	website     sql.NullString
	coordinates orb.Point
	description sql.NullString
	rating      float64
}

func main() {
	Latitude := 51.089932
	Longitude := -0.5215325000000001
	Radius := 5000

	eps := endpointFinder(Latitude, Longitude, Radius)

	for _, ep := range eps {
		fmt.Printf("%v, %s, %v, %v, %v, %f\n", ep.id, ep.name, ep.website, ep.coordinates, ep.description, ep.rating)
	}
}

func endpointFinder(lat float64, lng float64, radius int) []Endpoint {
	coef := float64(radius) * 0.000008983
	new_lat_max := lat + coef
	new_lat_min := lat - coef
	new_lng_max := lng + coef/math.Cos(lat*0.018)
	new_lng_min := lng - coef/math.Cos(lat*0.018)

	db, err := sql.Open("postgres", "postgres://postgres:1510@localhost/endpoint?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		panic(err)
	}

	// use this for shpere radius.
	// q := `
	// 	SELECT * FROM "MY_TABLE" WHERE ST_DistanceSphere(coordinates::geometry, ST_MakePoint($1, $2)) <= $3
	// `

	q := `
		SELECT * FROM "MY_TABLE" WHERE st_astext(coordinates) @ ST_MakeEnvelope($1, $2, $3, $4) ORDER BY st_astext(coordinates) <-> ST_MakePoint($5, $6) 
	`
	rows, err := db.Query(q, new_lng_min, new_lat_min, new_lng_max, new_lat_max, lng, lat)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	eps := make([]Endpoint, 0)
	for rows.Next() {
		ep := Endpoint{}
		err := rows.Scan(&ep.id, &ep.name, &ep.website, ewkb.Scanner(&ep.coordinates), &ep.description, &ep.rating)
		if err != nil {
			panic(err)
		}
		eps = append(eps, ep)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}

	return eps
}
