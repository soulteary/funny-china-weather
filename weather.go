// Package weather fetches current weather for Chinese administrative locations
// and provides animated SVG icons for weather codes.
//
// GetMyIPLocation resolves the outgoing IP address through IPIP.
// GetWeatherByLocation fetches a Tencent observation, returning the weather code,
// temperature in degrees Celsius, relative humidity percentage, and unchanged
// upstream update timestamp. Both functions require network access and return
// an error when a provider response is unsuccessful or incomplete.
//
// GetWeatherIconByCode uses the current local time. GetWeatherIconAt uses the
// time zone of the supplied time.Time. Daytime is 05:00 inclusive to 19:00
// exclusive. Unknown codes, including -1, use the day icon and the name 未知.
// GetIconSVG retrieves an embedded asset by its exact icon key. These icon
// functions do not access the network.
package weather
