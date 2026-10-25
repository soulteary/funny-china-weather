package weather

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"unicode"
	"unicode/utf8"
)

func TestParseLocation(t *testing.T) {
	cases := []struct{ input, province, city string }{
		{"北京市", "北京市", "北京市"}, {"北京", "北京市", "北京市"},
		{"上海市", "上海市", "上海市"}, {"天津市", "天津市", "天津市"}, {"重庆市", "重庆市", "重庆市"},
		{"广东省深圳市", "广东省", "深圳市"}, {" 广东省深圳市\n", "广东省", "深圳市"},
		{"广东省广州市", "广东省", "广州市"}, {"浙江省杭州市", "浙江省", "杭州市"},
		{"江苏省苏州市", "江苏省", "苏州市"}, {"福建省福州市", "福建省", "福州市"},
		{"广东深圳市", "广东省", "深圳市"}, {"黑龙江省哈尔滨市", "黑龙江省", "哈尔滨市"},
		{"广西壮族自治区南宁市", "广西壮族自治区", "南宁市"},
		{"内蒙古自治区呼和浩特市", "内蒙古自治区", "呼和浩特市"},
		{"宁夏回族自治区银川市", "宁夏回族自治区", "银川市"},
		{"西藏自治区拉萨市", "西藏自治区", "拉萨市"},
		{"新疆维吾尔自治区乌鲁木齐市", "新疆维吾尔自治区", "乌鲁木齐市"},
		{"广西省南宁市", "广西壮族自治区", "南宁市"},
		{"内蒙古省呼和浩特市", "内蒙古自治区", "呼和浩特市"},
		{"湖北省恩施土家族苗族自治州", "湖北省", "恩施土家族苗族自治州"},
		{"黑龙江省大兴安岭地区", "黑龙江省", "大兴安岭地区"},
		{"内蒙古自治区阿拉善盟", "内蒙古自治区", "阿拉善盟"},
		{"湖南省津市市", "湖南省", "津市市"},
		{"湖南省津市", "湖南省", "津市市"},
		{"黑龙江省大兴安岭", "黑龙江省", "大兴安岭地区"},
		{"内蒙古自治区阿拉善", "内蒙古自治区", "阿拉善盟"},
		{"内蒙古锡林郭勒", "内蒙古自治区", "锡林郭勒盟"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := parseLocation(tc.input)
			if err != nil || got.province != tc.province || got.city != tc.city {
				t.Fatalf("parseLocation(%q) = %#v, %v; want %q, %q", tc.input, got, err, tc.province, tc.city)
			}
			again, err := parseLocation(got.String())
			if err != nil || again != got {
				t.Fatalf("canonical location did not round trip: %#v, %v", again, err)
			}
		})
	}
}

func TestParseLocationRejectsInvalidInput(t *testing.T) {
	for _, input := range []string{"", " ", "invalid", "广东省", "广东省深圳", "错误省错误市", "北京市北京市", "广东省深圳市/extra", "广东省深 圳市", "广东省深圳市?x=1", "广东省深圳市\x00", "广东省江苏省南京市", "广东省深圳市东莞市", "广西自治区南宁市", "广东省龙华区", "海南省琼中黎族苗族自治县", "湖南省津市市市", "湖南省津市市长沙市", "湖南省市市"} {
		t.Run(input, func(t *testing.T) {
			if got, err := parseLocation(input); err == nil {
				t.Fatalf("parseLocation(%q) = %#v; want an error", input, got)
			}
		})
	}
}

func TestNormalizeLocationFieldsRejectsUnsupportedCities(t *testing.T) {
	for _, city := range []string{"省", "市", "区", "县", "旗", "州", "自治区", "自治州", "地区", "盟", "自治县", "自治旗", "特别行政区", "龙华区", "琼中黎族苗族自治县", "深圳市东莞市", "江苏省南京市"} {
		t.Run(city, func(t *testing.T) {
			got, err := normalizeLocationFields("广东", city)
			if err == nil || got != (administrativeLocation{}) {
				t.Fatalf("normalizeLocationFields() = %#v, %v; want zero location and error", got, err)
			}
		})
	}
}

func TestWeatherLocationQueryParameters(t *testing.T) {
	cases := []struct{ input, province, city string }{
		{"北京市", "北京市", "北京市"}, {"上海市", "上海市", "上海市"},
		{"天津市", "天津市", "天津市"}, {"重庆市", "重庆市", "重庆市"},
		{"广东省深圳市", "广东省", "深圳市"},
		{"广西壮族自治区南宁市", "广西壮族自治区", "南宁市"},
		{"内蒙古自治区呼和浩特市", "内蒙古自治区", "呼和浩特市"},
		{"宁夏回族自治区银川市", "宁夏回族自治区", "银川市"},
		{"西藏自治区拉萨市", "西藏自治区", "拉萨市"},
		{"新疆维吾尔自治区乌鲁木齐市", "新疆维吾尔自治区", "乌鲁木齐市"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			calls := 0
			client := fixtureClient(weatherFixture(nil))
			responseTransport := client.Transport
			client.Transport = fixtureTransport(func(request *http.Request) (*http.Response, error) {
				calls++
				query := request.URL.Query()
				if query.Get("province") != tc.province || query.Get("city") != tc.city || query.Get("weather_type") != "observe" || query.Get("source") != "pc" {
					t.Fatalf("unexpected decoded query parameters: %#v", query)
				}
				return responseTransport.RoundTrip(request)
			})
			if _, _, _, _, err := getWeatherByLocation(context.Background(), tc.input, client); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatalf("made %d requests, want one", calls)
			}
		})
	}
}

func TestIPLocationNormalization(t *testing.T) {
	cases := []struct{ province, city, want string }{
		{"广东", "深圳", "广东省深圳市"}, {"广东省", "深圳市", "广东省深圳市"},
		{"北京", "北京", "北京市"}, {"上海市", "上海市", "上海市"},
		{"天津", "天津", "天津市"}, {"重庆", "重庆", "重庆市"},
		{"广西", "南宁", "广西壮族自治区南宁市"},
		{"内蒙古", "呼和浩特", "内蒙古自治区呼和浩特市"},
		{"宁夏", "银川", "宁夏回族自治区银川市"},
		{"西藏", "拉萨", "西藏自治区拉萨市"},
		{"新疆", "乌鲁木齐", "新疆维吾尔自治区乌鲁木齐市"},
		{"湖北省", "恩施土家族苗族自治州", "湖北省恩施土家族苗族自治州"},
		{"湖南", "津市", "湖南省津市市"},
		{"湖南", "津市市", "湖南省津市市"},
		{"黑龙江", "大兴安岭", "黑龙江省大兴安岭地区"},
		{"内蒙古", "阿拉善", "内蒙古自治区阿拉善盟"},
		{"内蒙古", "锡林郭勒", "内蒙古自治区锡林郭勒盟"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			body, err := json.Marshal(map[string]interface{}{
				"ret":  "ok",
				"data": map[string]interface{}{"ip": "192.0.2.1", "location": []string{"中国", tc.province, tc.city}},
			})
			if err != nil {
				t.Fatal(err)
			}
			got, err := getMyIPLocation(context.Background(), fixtureClient(string(body)))
			if err != nil || got != tc.want {
				t.Fatalf("getMyIPLocation() = %q, %v; want %q", got, err, tc.want)
			}
			if _, err := parseLocation(got); err != nil {
				t.Fatalf("IP output cannot be used as weather input: %v", err)
			}
		})
	}
}

func TestMunicipalitiesMustMatchBothLocationFields(t *testing.T) {
	municipalities := []string{"北京", "上海", "天津", "重庆"}
	for _, province := range []string{"浙江", "广东", "内蒙古"} {
		for _, municipality := range municipalities {
			for _, city := range []string{municipality, municipality + "市"} {
				t.Run(province+"/"+city, func(t *testing.T) {
					if got, err := normalizeLocationFields(province, city); err == nil || got != (administrativeLocation{}) {
						t.Fatalf("non-municipal province accepted municipality: %#v, %v", got, err)
					}
					if got, err := parseLocation(province + city); err == nil {
						t.Fatalf("combined address accepted mismatched municipality: %#v", got)
					}
				})
			}
		}
	}
	for _, province := range municipalities {
		for _, city := range municipalities {
			if province == city {
				continue
			}
			if got, err := normalizeLocationFields(province, city); err == nil || got != (administrativeLocation{}) {
				t.Errorf("municipality %q accepted city %q: %#v, %v", province, city, got, err)
			}
		}
		if got, err := normalizeLocationFields(province, "深圳"); err == nil || got != (administrativeLocation{}) {
			t.Errorf("municipality %q accepted ordinary city: %#v, %v", province, got, err)
		}
	}
}

func TestLocationRejectsMalformedText(t *testing.T) {
	invalidParts := []string{
		"深\xff圳市", "深圳市\xe4\xb8", "深\u200b圳市", "深圳\u200c市",
		"深\u200d圳市", "\ufeff深圳市", "深\u202e圳市", "深圳市\u2066",
	}
	for _, city := range invalidParts {
		t.Run(city, func(t *testing.T) {
			if got, err := normalizeLocationFields("广东", city); err == nil || got != (administrativeLocation{}) {
				t.Fatalf("malformed city accepted: %#v, %v", got, err)
			}
			if got, err := parseLocation("广东省" + city); err == nil {
				t.Fatalf("malformed combined address accepted: %#v", got)
			}
		})
	}
	for _, province := range []string{"广\xff东", "广\u200b东", "\ufeff广东"} {
		if got, err := normalizeLocationFields(province, "深圳"); err == nil || got != (administrativeLocation{}) {
			t.Errorf("malformed province %q accepted: %#v, %v", province, got, err)
		}
	}
}

func TestKnownCityAliasesDoNotInventMunicipalSuffixes(t *testing.T) {
	for _, city := range []string{"大兴安岭", "大兴安岭地区", "阿拉善", "阿拉善盟", "锡林郭勒", "锡林郭勒盟", "津市", "津市市"} {
		if got, err := normalizeLocationFields("广东", city); err == nil || got != (administrativeLocation{}) {
			t.Errorf("known city %q accepted in the wrong province: %#v, %v", city, got, err)
		}
	}
	for _, city := range []string{"津市市市", "津市市长沙市", "津市省", "市市"} {
		if got, err := normalizeLocationFields("湖南", city); err == nil || got != (administrativeLocation{}) {
			t.Errorf("repeated administrative levels %q accepted: %#v, %v", city, got, err)
		}
	}
}

func assertCanonicalLocation(t *testing.T, location administrativeLocation) {
	t.Helper()
	if location.province == "" || location.city == "" || !utf8.ValidString(location.String()) ||
		!utf8.ValidString(location.province) || !utf8.ValidString(location.city) {
		t.Fatalf("successful parsing produced empty or invalid UTF-8 fields: %#v", location)
	}
	for _, char := range location.String() {
		if unicode.IsSpace(char) || unicode.IsControl(char) || unicode.Is(unicode.Cf, char) {
			t.Fatalf("successful parsing retained invisible/control character U+%04X", char)
		}
	}
	again, err := parseLocation(location.String())
	if err != nil || again != location {
		t.Fatalf("canonical location did not round trip: %#v => %#v, %v", location, again, err)
	}
}

func FuzzParseLocation(f *testing.F) {
	for _, value := range []string{
		"北京市", "广东省深圳市", "广西壮族自治区南宁市", "", "广东省深圳市?x=1",
		"湖南省津市市", "黑龙江省大兴安岭", "内蒙古阿拉善", "内蒙古锡林郭勒",
		"浙江省北京市", "广东省深\xff圳市", "广东省深\u200b圳市", "广东省州",
	} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		got, err := parseLocation(value)
		if err != nil {
			return
		}
		if !utf8.ValidString(value) {
			t.Fatalf("invalid UTF-8 input %q was accepted", value)
		}
		assertCanonicalLocation(t, got)
	})
}

func FuzzNormalizeLocationFields(f *testing.F) {
	for _, value := range []struct{ province, city string }{
		{"广东", "深圳"}, {"北京", "北京"}, {"广西", "南宁"},
		{"湖南", "津市市"}, {"湖南", "津"}, {"黑龙江", "大兴安岭"}, {"内蒙古", "阿拉善"},
		{"内蒙古", "锡林郭勒"}, {"浙江", "北京市"}, {"", ""},
		{"广东", "深\xff圳"}, {"广东", "深\u200b圳"}, {"广东", "州"},
	} {
		f.Add(value.province, value.city)
	}
	f.Fuzz(func(t *testing.T, province, city string) {
		got, err := normalizeLocationFields(province, city)
		if err != nil {
			return
		}
		if !utf8.ValidString(province) || !utf8.ValidString(city) {
			t.Fatalf("invalid UTF-8 input (%q, %q) was accepted", province, city)
		}
		assertCanonicalLocation(t, got)
	})
}
