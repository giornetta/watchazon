package scraper

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/giornetta/watchazon"
)

func Test_convertPrice(t *testing.T) {
	type args struct {
		text   string
		domain watchazon.Domain
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "Empty",
			args: args{
				text:   "",
				domain: "",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Float",
			args: args{
				text:   "12,99 €",
				domain: "it",
			},
			want:    12.99,
			wantErr: false,
		},
		{
			name: "Thousands",
			args: args{
				text:   "1.099,89 €",
				domain: "es",
			},
			want:    1099.89,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertPrice(tt.args.text, tt.args.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertPrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("convertPrice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScraper_Scrape(t *testing.T) {
	type fields struct {
		AllowedDomains []string
	}
	itFields := fields{AllowedDomains: []string{"www.amazon.it", "www.amazon.com"}}

	tests := []struct {
		name    string
		fields  fields
		arg     string
		want    *watchazon.Product
		wantErr bool
	}{
		{
			name:   "Echo Dot",
			fields: itFields,
			arg:    "https://www.amazon.it/amazon-echo-dot-3-generazione-altoparlante-intelligente-con-integrazione-alexa-tessuto-antracite/dp/B07PHPXHQS?pf_rd_p=0126cc1b-63c4-49e1-9993-8a52c0ac2cfb&pd_rd_wg=enBag&pf_rd_r=KDKKPBE97TFWNZQ3KZZG&ref_=pd_gw_cr_simh&pd_rd_w=4yt8m&pd_rd_r=df8dcc6a-d24d-418c-8fae-306022d242ec",
			want: &watchazon.Product{
				Title: "Echo Dot (3ª generazione) - Altoparlante intelligente con integrazione Alexa - Tessuto antracite",
				Image: "",
				Link:  "https://www.amazon.it/amazon-echo-dot-3-generazione-altoparlante-intelligente-con-integrazione-alexa-tessuto-antracite/dp/B07PHPXHQS?pf_rd_p=0126cc1b-63c4-49e1-9993-8a52c0ac2cfb&pd_rd_wg=enBag&pf_rd_r=KDKKPBE97TFWNZQ3KZZG&ref_=pd_gw_cr_simh&pd_rd_w=4yt8m&pd_rd_r=df8dcc6a-d24d-418c-8fae-306022d242ec",
				Price: 59.99,
			},
			wantErr: false,
		},
		{
			name:   "Mi Band 3",
			fields: itFields,
			arg:    "https://www.amazon.com/Activity-Waterproof-Bracelet-Wristband-Pedometer/dp/B07GNGJK97/ref=sr_1_1?keywords=mi+band&qid=1566988072&s=gateway&sr=8-1",
			want: &watchazon.Product{
				Title: "Xiaomi Fitness Tracker, Mi Band 3 Heart Rate Monitor Activity Tracker Watch 50M Waterproof Smart Bracelet 0.78 OLED Display Weather Forecast Wristband Pedometer Calories Burned Sleep Monitor Black",
				Image: "",
				Link:  "https://www.amazon.com/Activity-Waterproof-Bracelet-Wristband-Pedometer/dp/B07GNGJK97/ref=sr_1_1?keywords=mi+band&qid=1566988072&s=gateway&sr=8-1",
				Price: 28.98,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scraper{
				AllowedDomains: tt.fields.AllowedDomains,
			}
			got, err := s.Scrape(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scrape() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scrape() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScraper_Search(t *testing.T) {
	s := New("www.amazon.it", "www.amazon.com")

	got, err := s.Search("Samsung S8", "it")
	if err != nil {
		t.Fatalf("could not search: %v", err)
	}

	for _, p := range got {
		fmt.Printf("%s: %.2f\n", p.Link, p.Price)
	}
}
