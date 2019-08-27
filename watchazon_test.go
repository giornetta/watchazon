package watchazon

import "testing"

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    string
		wantErr bool
	}{
		{
			name:    "GP",
			arg:     "https://www.amazon.it/gp/slredirect/picassoRedirect.html/ref=pa_sp_atf_aps_sr_pg1_1?ie=UTF8&adId=A07724242K61NDVW98IBC&url=%2FMoEx-Custodia-Samsung-Galaxy-S8%2Fdp%2FB071S3PX1Q%2Fref%3Dsr_1_1_sspa%3Fkeywords%3DSamsung%2BS8%26qid%3D1566902010%26s%3Dgateway%26sr%3D8-1-spons%26psc%3D1&qualifier=1566902010&id=1479389534632242&widgetName=sp_atf",
			want:    "https://www.amazon.it/dp/B071S3PX1Q",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeURL(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SanitizeURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
