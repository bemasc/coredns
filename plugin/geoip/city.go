package geoip

import (
	"context"
	"strconv"

	"github.com/coredns/coredns/plugin/metadata"

	"github.com/oschwald/geoip2-golang"
)

const defaultLang = "en"

func (g GeoIP) setCityMetadata(ctx context.Context, data *geoip2.City, ednsScope int) {

	set := func(suffix string, value string) {
		metadata.SetValueFunc(ctx, pluginName+suffix, func() string {
			return value
		})
		metadata.SetValueFunc(ctx, pluginName+suffix+"/_ecs-scope", func() string {
			return strconv.Itoa(ednsScope)
		})
	}

	// Set labels for city, country and continent names.
	set("/city/name", data.City.Names[defaultLang])
	set("/country/name", data.Country.Names[defaultLang])
	set("/continent/name", data.Continent.Names[defaultLang])
	set("/country/code", data.Country.IsoCode)
	set("/country/is_in_european_union", strconv.FormatBool(data.Country.IsInEuropeanUnion))
	set("/continent/code", data.Continent.Code)
	set("/latitude", strconv.FormatFloat(data.Location.Latitude, 'f', -1, 64))
	set("/longitude", strconv.FormatFloat(data.Location.Longitude, 'f', -1, 64))
	set("/timezone", data.Location.TimeZone)
	set("/postalcode", data.Postal.Code)
}
