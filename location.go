package weather

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type administrativeLocation struct {
	province string
	city     string
}

var provinceAliases = map[string]string{
	"北京": "北京市", "北京市": "北京市",
	"上海": "上海市", "上海市": "上海市",
	"天津": "天津市", "天津市": "天津市",
	"重庆": "重庆市", "重庆市": "重庆市",
	"内蒙古": "内蒙古自治区", "内蒙古自治区": "内蒙古自治区", "内蒙古省": "内蒙古自治区",
	"广西": "广西壮族自治区", "广西壮族自治区": "广西壮族自治区", "广西省": "广西壮族自治区",
	"宁夏": "宁夏回族自治区", "宁夏回族自治区": "宁夏回族自治区", "宁夏省": "宁夏回族自治区",
	"西藏": "西藏自治区", "西藏自治区": "西藏自治区", "西藏省": "西藏自治区",
	"新疆": "新疆维吾尔自治区", "新疆维吾尔自治区": "新疆维吾尔自治区", "新疆省": "新疆维吾尔自治区",
}

var ordinaryProvinces = []string{
	"安徽省", "福建省", "甘肃省", "广东省", "贵州省", "海南省", "河北省", "河南省",
	"黑龙江省", "湖北省", "湖南省", "吉林省", "江苏省", "江西省", "辽宁省", "青海省",
	"山东省", "山西省", "陕西省", "四川省", "云南省", "浙江省", "台湾省",
}

// These explicit aliases preserve the names of known administrative areas.
// This is a small exception list, not a registry of every Chinese location.
var knownCityAliases = []struct{ province, shortName, fullName string }{
	{"黑龙江省", "大兴安岭", "大兴安岭地区"},
	{"内蒙古自治区", "阿拉善", "阿拉善盟"},
	{"内蒙古自治区", "锡林郭勒", "锡林郭勒盟"},
	{"湖南省", "津市", "津市市"},
}

func canonicalKnownCity(province, city string) (string, error) {
	for _, alias := range knownCityAliases {
		if city == alias.shortName || city == alias.fullName {
			if province != alias.province {
				return "", fmt.Errorf("city %q does not match province %q", city, province)
			}
			return alias.fullName, nil
		}
	}
	return city, nil
}

func canonicalProvince(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if province, ok := provinceAliases[value]; ok {
		return province, true
	}
	for _, province := range ordinaryProvinces {
		if value == province || value == strings.TrimSuffix(province, "省") {
			return province, true
		}
	}
	return "", false
}

func isMunicipality(province string) bool {
	return province == "北京市" || province == "上海市" || province == "天津市" || province == "重庆市"
}

func validLocationPart(value string) bool {
	if value == "" || !utf8.ValidString(value) || strings.ContainsAny(value, "/\\?&#=%") {
		return false
	}
	for _, char := range value {
		if unicode.IsSpace(char) || unicode.IsControl(char) || unicode.Is(unicode.Cf, char) {
			return false
		}
	}
	return true
}

func hasCitySuffix(city string) bool {
	for _, suffix := range []string{"市", "自治州", "地区", "盟"} {
		if strings.HasSuffix(city, suffix) && len(city) > len(suffix) {
			return true
		}
	}
	return false
}

func validCityName(city string, allowShort bool) bool {
	if !validLocationPart(city) {
		return false
	}
	switch city {
	case "省", "市", "区", "县", "旗", "州", "自治区", "自治州", "地区", "盟", "自治县", "自治旗", "特别行政区":
		return false
	}
	// In this real Hunan city, the first 市 belongs to the place name.
	if city == "津市市" {
		return true
	}
	stem := city
	for _, suffix := range []string{"自治州", "地区", "市", "盟"} {
		if strings.HasSuffix(city, suffix) {
			stem = strings.TrimSuffix(city, suffix)
			break
		}
	}
	if stem == city && !allowShort {
		return false
	}
	// Reject incomplete suffixes, unsupported district/county names, and
	// duplicated administrative levels instead of inventing a city name.
	return stem != "" && !strings.ContainsAny(stem, "省市区县旗盟") && !strings.Contains(stem, "自治州")
}

func normalizeLocationFields(province, city string) (administrativeLocation, error) {
	normalized, ok := canonicalProvince(province)
	if !ok {
		return administrativeLocation{}, fmt.Errorf("unsupported province %q", strings.TrimSpace(province))
	}
	city = strings.TrimSpace(city)
	if !validCityName(city, true) {
		return administrativeLocation{}, fmt.Errorf("invalid city %q", city)
	}
	municipality, cityIsProvince := canonicalProvince(city)
	if isMunicipality(normalized) {
		if !cityIsProvince || municipality != normalized {
			return administrativeLocation{}, fmt.Errorf("city %q does not match municipality %q", city, normalized)
		}
		return administrativeLocation{normalized, normalized}, nil
	}
	if cityIsProvince && isMunicipality(municipality) {
		return administrativeLocation{}, fmt.Errorf("municipality %q does not belong to province %q", city, normalized)
	}
	city, err := canonicalKnownCity(normalized, city)
	if err != nil {
		return administrativeLocation{}, err
	}
	if !hasCitySuffix(city) {
		city += "市"
		// Adding a suffix can expose a known alias, such as 津市.
		city, err = canonicalKnownCity(normalized, city)
		if err != nil {
			return administrativeLocation{}, err
		}
	}
	return administrativeLocation{normalized, city}, nil
}

func parseLocation(value string) (administrativeLocation, error) {
	value = strings.TrimSpace(value)
	if !validLocationPart(value) {
		return administrativeLocation{}, fmt.Errorf("invalid location %q", value)
	}
	if province, ok := canonicalProvince(value); ok && isMunicipality(province) {
		return administrativeLocation{province, province}, nil
	}

	prefix, province := locationProvincePrefix(value)
	if prefix == "" || isMunicipality(province) {
		return administrativeLocation{}, fmt.Errorf("location must contain a supported province and city: %q", value)
	}
	city, err := canonicalKnownCity(province, strings.TrimPrefix(value, prefix))
	if err != nil {
		return administrativeLocation{}, err
	}
	if !hasCitySuffix(city) || !validCityName(city, false) {
		return administrativeLocation{}, fmt.Errorf("location must end in a city, prefecture, or league name: %q", value)
	}
	return normalizeLocationFields(province, city)
}

func (location administrativeLocation) String() string {
	if isMunicipality(location.province) {
		return location.province
	}
	return location.province + location.city
}

func locationProvincePrefix(value string) (prefix, province string) {
	// Prefer the longest recognized prefix, including historical aliases.
	for alias, canonical := range provinceAliases {
		if strings.HasPrefix(value, alias) && len(alias) > len(prefix) {
			prefix, province = alias, canonical
		}
	}
	for _, canonical := range ordinaryProvinces {
		for _, alias := range []string{canonical, strings.TrimSuffix(canonical, "省")} {
			if strings.HasPrefix(value, alias) && len(alias) > len(prefix) {
				prefix, province = alias, canonical
			}
		}
	}
	return prefix, province
}
