package iplookup

import (
	"net"

	"github.com/oschwald/maxminddb-golang"
)

type MaxMindLookup struct {
	db *maxminddb.Reader
}

// NewMaxMindLookup opens a MaxMind MMDB database file.
func NewMaxMindLookup(dbPath string) (*MaxMindLookup, error) {
	db, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &MaxMindLookup{db: db}, nil
}

func (m *MaxMindLookup) Close() error {
	return m.db.Close()
}

type geoNames map[string]string

type geoCity struct {
	Names geoNames `maxminddb:"names"`
}

type geoSubdivision struct {
	Names geoNames `maxminddb:"names"`
}

type geoCountry struct {
	Names geoNames `maxminddb:"names"`
}

type geoRecord struct {
	City         geoCity         `maxminddb:"City"`
	Subdivisions []geoSubdivision `maxminddb:"Subdivisions"`
	Country      geoCountry      `maxminddb:"Country"`
}

func (m *MaxMindLookup) Lookup(ip string) (Region, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return Region{}, nil
	}

	var record geoRecord
	if err := m.db.Lookup(parsed, &record); err != nil {
		return Region{}, err
	}

	r := Region{
		Country: localizedName(record.Country.Names),
	}
	if len(record.Subdivisions) > 0 {
		r.Province = localizedName(record.Subdivisions[0].Names)
	}
	r.City = localizedName(record.City.Names)
	return r, nil
}

// localizedName returns the Chinese name if available, otherwise English, otherwise any.
func localizedName(names geoNames) string {
	if names == nil {
		return ""
	}
	if v, ok := names["zh-CN"]; ok {
		return v
	}
	if v, ok := names["en"]; ok {
		return v
	}
	for _, v := range names {
		return v
	}
	return ""
}
