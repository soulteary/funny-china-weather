package weather

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
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
	for _, input := range []string{"", " ", "invalid", "广东省", "广东省深圳", "错误省错误市", "北京市北京市", "广东省深圳市/extra", "广东省深 圳市", "广东省深圳市?x=1", "广东省深圳市\x00", "广东省江苏省南京市", "广东省深圳市东莞市", "广西自治区南宁市", "广东省龙华区", "海南省琼中黎族苗族自治县"} {
		t.Run(input, func(t *testing.T) {
			if got, err := parseLocation(input); err == nil {
				t.Fatalf("parseLocation(%q) = %#v; want an error", input, got)
			}
		})
	}
}

func TestNormalizeLocationFieldsRejectsUnsupportedCities(t *testing.T) {
	for _, city := range []string{"市", "自治州", "地区", "盟", "龙华区", "琼中黎族苗族自治县", "深圳市东莞市", "江苏省南京市"} {
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

func FuzzParseLocation(f *testing.F) {
	for _, value := range []string{"北京市", "广东省深圳市", "广西壮族自治区南宁市", "", "广东省深圳市?x=1"} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		got, err := parseLocation(value)
		if err != nil {
			return
		}
		again, err := parseLocation(got.String())
		if err != nil || again != got {
			t.Fatalf("canonical location did not round trip: %#v => %#v, %v", got, again, err)
		}
	})
}
