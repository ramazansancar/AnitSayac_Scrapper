package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

/*
'Name': name,
'Tarih': "",
'Maktülün yaşı': "",
'İl/ilçe': "",
'Neden öldürüldü': "",
'Kim tarafından öldürüldü': "",
'Korunma talebi': "",
'Öldürülme şekli': "",
'Failin durumu': "",
'Kaynak': "",
'image': img_src,
'url': href
*/

type Incident struct {
	Id         int      `json:"id"`
	Name       string   `json:"name"`
	FullName   string   `json:"fullname"`
	Age        string   `json:"age"`
	Location   string   `json:"location"`
	Date       string   `json:"date"`
	Reason     string   `json:"reason"`
	By         string   `json:"by"`
	Protection string   `json:"protection"`
	Method     string   `json:"method"`
	Status     string   `json:"status"`
	Source     []string `json:"source"`
	Image      string   `json:"image"`
	Url        string   `json:"url"`
}

type Detail struct {
	Name       string   `json:"name"`
	Age        string   `json:"age"`
	Location   string   `json:"location"`
	Date       string   `json:"date"`
	Reason     string   `json:"reason"`
	By         string   `json:"by"`
	Protection string   `json:"protection"`
	Method     string   `json:"method"`
	Status     string   `json:"status"`
	Source     []string `json:"source"`
	Image      string   `json:"image"`
}

// Static and dynamic Variables
var (
	baseUrl      = "https://anitsayac.com"
	jsonFileName = "data.json"
	csvFileName  = "data.csv"
)

func ReplaceAll(s, old, new string, n int) string {
	// Replace All for Regexp
	re := regexp.MustCompile(old)
	return re.ReplaceAllString(s, new)
}

// validateFiles checks if the generated files are valid and have reasonable content
func validateFiles(jsonFile, csvFile string, expectedCount int) bool {
	// Check if files exist
	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		log.Printf("JSON file %s does not exist", jsonFile)
		return false
	}

	if _, err := os.Stat(csvFile); os.IsNotExist(err) {
		log.Printf("CSV file %s does not exist", csvFile)
		return false
	}

	// Check file sizes
	jsonInfo, err := os.Stat(jsonFile)
	if err != nil {
		log.Printf("Failed to get JSON file info: %s", err)
		return false
	}

	csvInfo, err := os.Stat(csvFile)
	if err != nil {
		log.Printf("Failed to get CSV file info: %s", err)
		return false
	}

	// Files should not be empty
	if jsonInfo.Size() == 0 {
		log.Printf("JSON file is empty")
		return false
	}

	if csvInfo.Size() == 0 {
		log.Printf("CSV file is empty")
		return false
	}

	// Basic JSON validation - try to parse
	jsonFileContent, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Printf("Failed to read JSON file: %s", err)
		return false
	}

	var incidents []Incident
	if err := json.Unmarshal(jsonFileContent, &incidents); err != nil {
		log.Printf("Invalid JSON format: %s", err)
		return false
	}

	// Check if we have reasonable number of incidents
	if len(incidents) < expectedCount/2 {
		log.Printf("Too few incidents in JSON: got %d, expected at least %d", len(incidents), expectedCount/2)
		return false
	}

	// Basic CSV validation - count lines
	csvFileContent, err := os.ReadFile(csvFile)
	if err != nil {
		log.Printf("Failed to read CSV file: %s", err)
		return false
	}

	lines := strings.Split(string(csvFileContent), "\n")
	// Should have header + data lines (allowing for empty last line)
	if len(lines) < expectedCount {
		log.Printf("Too few lines in CSV: got %d, expected at least %d", len(lines), expectedCount)
		return false
	}

	return true
}

// Get article content
func getArticleContent(url string) Detail {
	/*
		! - Unknown Url :( (ID: 38934 - https://anitsayac.com/details.aspx?id=38934 - Commit: 38b5d7f6b113f4894c703624a15880ae0b7c0bb8 - https://github.com/ramazansancar/AnitSayac_Scrapper/commit/38b5d7f6b113f4894c703624a15880ae0b7c0bb8)
		<b>Ad Soyad:</b> Fidan Çakır<br><b>Maktülün yaşı: </b>Reşit<br><b>İl/ilçe: </b>İzmir<br><b>Tarih: </b>16/10/2024<br><b>Neden öldürüldü:</b>  Tespit Edilemeyen<br><b>Kim tarafından öldürüldü:</b>  Tespit Edilemeyen<br><b>Korunma talebi:</b>  Yok<br><b>Öldürülme şekli:</b>  Kesici Alet<br><b>Failin durumu: </b>Soruşturma Sürüyor<br><b>Kaynak:</b>  <a target=_blank href='https://www.t24.com.tr/haber/supheli-bir-kadin-olumu-daha-izmir-de-bir-kadin-evinde-vucudunda-kesi-izleriyle-olu-bulundu,1190726'><u>https://www.t24.com.tr/haber/supheli-bir-kadin-olumu-daha-izmir-de-bir-kadin-evinde-vucudunda-kesi-izleriyle-olu-bulundu,1190726</u></a><br><img width=750 style='margin-top:10px' src=ii/3202024.jpg>

		https://anitsayac.com/details.aspx?id=151
		<b>Ad Soyad:</b> Beyhan Yavuz<br><b>Tarih: </b>05/02/2008<br><b>Neden öldürüldü:</b>  Reddetme<br><b>Kim tarafından öldürüldü:</b>  Dini nikahlı kocası<br><b>Korunma talebi:</b>  Tespit Edilemeyen<br><b>Öldürülme şekli:</b>  Ateşli Silah<br><b>Kaynak:</b>  <a target=_blank href='http://hurarsiv.hurriyet.com.tr/goster/haber.aspx?id=8170561&tarih=2008-02-05'><u>http://hurarsiv.hurriyet.com.tr/goster/haber.aspx?id=8170561&tarih=2008-02-05</u></a>

		https://anitsayac.com/details.aspx?id=37902
		<b>Ad Soyad:</b> Ayşenur Halil<br><b>Maktülün yaşı: </b>Reşit<br><b>İl/ilçe: </b>İstanbul<br><b>Tarih: </b>04/10/2024<br><b>Neden öldürüldü:</b>  Tespit Edilemeyen<br><b>Kim tarafından öldürüldü:</b>  Eski Sevgilisi<br><b>Korunma talebi:</b>  Yok<br><b>Öldürülme şekli:</b>  Kesici Alet<br><b>Failin durumu: </b>İntihar<br><b>Kaynak:</b>  <br><a target=_blank href='https://www.t24.com.tr/haber/fatih-te-vahset-kadinin-kafasini-kesip-surlardan-atladi,1187812'><u>https://www.t24.com.tr/haber/fatih-te-vahset-kadinin-kafasini-kesip-surlardan-atladi,1187812</u></a><br><a target=_blank href='https://www.t24.com.tr/haber/yarim-saat-arayla-iki-kadini-katleden-semih-celik-5-kez-psikolojik-tedavi-gormus,1187904'><u>https://www.t24.com.tr/haber/yarim-saat-arayla-iki-kadini-katleden-semih-celik-5-kez-psikolojik-tedavi-gormus,1187904</u></a><br><img width=750 style='margin-top:10px' src=ii/2892024.jpg>

		https://anitsayac.com/details.aspx?id=37903
		<b>Ad Soyad:</b> İkbal Uzuner<br><b>Maktülün yaşı: </b>Reşit<br><b>İl/ilçe: </b>İstanbul<br><b>Tarih: </b>04/10/2024<br><b>Neden öldürüldü:</b>  Tespit Edilemeyen<br><b>Kim tarafından öldürüldü:</b>  Tanımadığı Birisi <br><b>Korunma talebi:</b>  Yok<br><b>Öldürülme şekli:</b>  Kesici Alet<br><b>Failin durumu: </b>İntihar<br><b>Kaynak:</b>  <br><a target=_blank href='https://www.t24.com.tr/haber/fatih-te-vahset-kadinin-kafasini-kesip-surlardan-atladi,1187812'><u>https://www.t24.com.tr/haber/fatih-te-vahset-kadinin-kafasini-kesip-surlardan-atladi,1187812</u></a><br><a target=_blank href='https://www.t24.com.tr/haber/yarim-saat-arayla-iki-kadini-katleden-semih-celik-5-kez-psikolojik-tedavi-gormus,1187904'><u>https://www.t24.com.tr/haber/yarim-saat-arayla-iki-kadini-katleden-semih-celik-5-kez-psikolojik-tedavi-gormus,1187904</u></a><br><img width=750 style='margin-top:10px' src=ii/2902024.jpg>

		New Version:

		https://anitsayac.com/details.aspx?id=35697
		<b>Ad Soyad:</b> Selma Çiftçi<br><b>Maktülün yaşı: </b>Reşit<br><b>İl/ilçe: </b>Mersin<br><b>Tarih: </b>04/04/2024<br><b>Neden öldürüldü:</b>  Tartışma<br><b>Kim tarafından öldürüldü:</b>  Oğlu<br><b>Korunma talebi:</b>  Yok<br><b>Öldürülme şekli:</b>  Kesic Alet, Ateşli Silah<br><b>Failin durumu: </b>Tutuklu<br><b>Kaynak:</b>  <a target=_blank href='https://www.hurriyet.com.tr/gundem/mersinde-hastanede-kan-donduran-cinayet-annesinin-bogazini-ve-ayaklarini-kesip-tek-kursunla-oldurdu-42441309'><u>https://www.hurriyet.com.tr/gundem/mersinde-hastanede-kan-donduran-cinayet-annesinin-bogazini-ve-ayaklarini-kesip-tek-kursunla-oldurdu-42441309</u></a><br><img width=750    onerror='this.style.display="none";'  style='margin-top:10px' src='//i.anitsayac.com/ii/972024.jpg'>


	*/

	detail := Detail{}
	c := colly.NewCollector(
		// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
		colly.AllowedDomains("anitsayac.com"),

		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./anitsayac_cache"),
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting Detail: ", r.URL.String())
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		allFieldCorrections := map[string]string{
			"Tespit Edİlemeyen":    "Tespit Edilemeyen",
			"Tespi Edilemeyen":     "Tespit Edilemeyen",
			"Tespite Edilemeyen":   "Tespit Edilemeyen",
			"Tespit Edlemeyen":     "Tespit Edilemeyen",
			"Tespite Edlielemeyen": "Tespit Edilemeyen",
			"Tespit Edielmeyen":    "Tespit Edilemeyen",
			"Tespit Edilemyen":     "Tespit Edilemeyen",
			"Tespit Edilmeyen":     "Tespit Edilemeyen",
			"Tesbit Edilemeyen":    "Tespit Edilemeyen",
			"Tespit Edielemeyen":   "Tespit Edilemeyen",
			"Tespite Edilemeye":    "Tespit Edilemeyen",
			"Tespit Edilmeyem":     "Tespit Edilemeyen",
			"Tespit Edilemeye":     "Tespit Edilemeyen",

			"Erkek Karkeşleri":          "Erkek Kardeşleri",
			"Eniltesi":                  "Eniştesi",
			"Kocacı":                    "Kocası",
			"Kocasi":                    "Kocası",
			"Tanımdıkları Birileri":     "Tanımadıkları Birileri",
			"Tespit edilemeyen":         "Tespit Edilemeyen",
			"Tanımadığı BIrileri":       "Tanımadığı Birileri",
			"Tanıdıği Birileri":         "Tanıdığı Birileri",
			"Tanımadıği Birisi":         "Tanımadığı Birisi",
			"Tanımadığı biri":           "Tanımadığı Birisi",
			"Babasi":                    "Babası",
			"Yo":                        "Yok",
			"YoK":                       "Yok",
			"yok":                       "Yok",
			"Var Uzaklaştırma Kararı":   "Var (Uzaklaştırma Kararı)",
			"Var(Uzaklaştırma Kararı)":  "Var (Uzaklaştırma Kararı)",
			"Var - Uzaklaştırma Kararı": "Var (Uzaklaştırma Kararı)",
			"Uzaklaştırma Kararı Var":   "Var (Uzaklaştırma Kararı)",
			"Uzaklaştırma Kararı":       "Var (Uzaklaştırma Kararı)",
			"Var (Uzaklaştıma Kararı)":  "Var (Uzaklaştırma Kararı)",
			"Var. Uzaklaştırma Kararı.": "Var (Uzaklaştırma Kararı)",
			"Uzaklaştırma kararı var":   "Var (Uzaklaştırma Kararı)",
			"Var -Uzaklaştırma Kararı":  "Var (Uzaklaştırma Kararı)",
			"Var (Uzaklaştırma kararı)": "Var (Uzaklaştırma Kararı)",
			"Korunma Talebi Var":        "Var (Korunma Talebi)",
			"Korunma talebi var":        "Var (Korunma Talebi)",
			"Var(Korunma Talebi)":       "Var (Korunma Talebi)",
			"İnithar":                   "İntihar",
			"İntiihar Teşebbüsü":        "İntihar Teşebbüsü",
			"İnthihara Teşebbüs":        "İntihara Teşebbüs",
			"Aranııyor":                 "Aranıyor",
			"Soruştutma Sürüyor":        "Soruşturma Sürüyor",
			"Tutuklu Değik":             "Tutuklu Değil",
			"Tutuklul":                  "Tutuklu",
			"Ateşl Silah":               "Ateşli Silah",
			"Ateşli Slah":               "Ateşli Silah",
			"Atesli Silah":              "Ateşli Silah",
			"Ateşli Sialh":              "Ateşli Silah",
			"Ateşlil Silah":             "Ateşli Silah",
			"Ateşli SIlah":              "Ateşli Silah",
			"Kesici Aelt":               "Kesici Alet",
			"Kesici alet":               "Kesici Alet",
			"-":                         "",
		}

		nameMatches := regexp.MustCompile(`(?i)<b>Ad Soyad:\s*</b>\s*(.+?)<br>`).FindStringSubmatch(string(e.Response.Body))
		if len(nameMatches) > 1 {
			detail.Name = strings.TrimSpace(nameMatches[1])
		}

		ageMatches := regexp.MustCompile(`(?i)<b>Maktülün yaşı:\s*</b>\s*(.+?)<br>`).FindStringSubmatch(string(e.Response.Body))
		if len(ageMatches) > 1 {
			age := strings.TrimSpace(ageMatches[1])

			// Normalize age values - use exact matching
			ageCorrections := map[string]string{
				"Reşi":        "Reşit",
				"ReşİT":       "Reşit",
				"Re\u0013şit": "Reşit",
				"Re\u0001şit": "Reşit",
			}

			// Apply age corrections with exact matching
			for incorrect, correct := range ageCorrections {
				if age == incorrect {
					age = correct
					break
				}
			}
			age = strings.TrimSpace(age)
			// Apply all field corrections
			for incorrect, correct := range allFieldCorrections {
				if age == incorrect {
					age = correct
					break
				}
			}
			age = strings.TrimSpace(age)

			detail.Age = age
		}

		locationMatches := regexp.MustCompile(`(?i)<b>İl/ilçe:\s*</b>\s*(.+?)<br>`).FindStringSubmatch(string(e.Response.Body))
		if len(locationMatches) > 1 {
			location := strings.TrimSpace(locationMatches[1])

			// Common location corrections (alphabetically sorted)
			locationCorrections := map[string]string{
				"Adapazarı":         "Adapazarı/Sakarya",
				"Afyonharahisar":    "Afyonkarahisar",
				"Agrı":              "Ağrı",
				"Akhisar":           "Akhisar/Manisa",
				"Aksu":              "Aksu/Antalya",
				"Akyazı":            "Akyazı/Sakarya",
				"ankara":            "Ankara",
				"Aralık":            "Aralık/Iğdır",
				"Arnavutköy":        "Arnavutköy/Istanbul",
				"Ayvalık":           "Ayvalık/Balıkesir",
				"Buca":              "Buca/İzmir",
				"Çiğli":             "Çiğli/İzmir",
				"Çine":              "Çine/Aydın",
				"Datça":             "Datça/Muğla",
				"Dersim":            "Dersim/Tunceli",
				"Devrek":            "Devrek/Zonguldak",
				"Didim":             "Didim/Aydın",
				"Diyarbaır":         "Diyarbakır",
				"Doğu Beyazıt":      "Doğubayazıt/Ağrı",
				"Edremit":           "Edremit/Balıkesir",
				"Eğirdir":           "Eğirdir/Isparta",
				"Ekazığ":            "Elazığ",
				"Ereğli":            "Ereğli/Zonguldak",
				"Ergani":            "Ergani/Diyarbakır",
				"Fatsa":             "Fatsa/Ordu",
				"Fetihiye":          "Fethiye/Muğla",
				"Fetyhiye":          "Fethiye/Muğla",
				"Gazipaşa":          "Gazipaşa/Antalya",
				"Gazianteop":        "Gaziantep",
				"Gebze":             "Gebze/Kocaeli",
				"Gemlik":            "Gemlik/Bursa",
				"Girne":             "Girne/Kıbrıs",
				"Harran":            "Harran/Şanlıurfa",
				"Iğdır":            "Iğdır",
				"İğdır":             "Iğdır",
				"İsparta":           "Isparta",
				"istanbul":          "İstanbul",
				"İsstanbul":         "İstanbul",
				"İstanbu":           "İstanbul",
				"izmir":             "İzmir",
				"İzmİr":             "İzmir",
				"İzmit":             "İzmit/Kocaeli",
				"Kahrmanmaraş":      "Kahramanmaraş",
				"Karamürsel":        "Karamürsel/Kocaeli",
				"kars":              "Kars",
				"Kaş":               "Kaş/Antalya",
				"Kastomonu":         "Kastamonu",
				"kayseri":           "Kayseri",
				"Keşan":             "Keşan/Edirne",
				"Kırlareli":         "Kırklareli",
				"Kırııkkale":        "Kırıkkale",
				"KocaeLİ":           "Kocaeli",
				"konya":             "Konya",
				"Küçükçekmece":      "Küçükçekmece/Istanbul",
				"Kuşadası":          "Kuşadası/Aydın",
				"Lapseki":           "Lapseki/Çanakkale",
				"Lefkoşa":           "Lefkoşa/Kıbrıs",
				"Maltepe":           "Maltepe/Istanbul",
				"Mardın":            "Mardin",
				"Marmaris":          "Marmaris/Muğla",
				"Mazıdağı":          "Mazıdağı/Mardin",
				"nevşehir":          "Nevşehir",
				"Nusaybin":          "Nusaybin/Mardin",
				"Ödemiş":            "Ödemiş/İzmir",
				"Orhaniye":          "Orhaniye/Muğla",
				"Osmancık":          "Osmancık/Çorum",
				"Polatlı":           "Polatlı/Ankara",
				"Safranbolu":        "Safranbolu/Karabük",
				"samsun":            "Samsun",
				"ŞanlıUrfa":         "Şanlıurfa",
				"Saruhan":           "Saruhanlı/Manisa",
				"Sincan":            "Sincan/Ankara",
				"Siverek":           "Siverek/Şanlıurfa",
				"Sultangazi":        "Sultangazi/Istanbul",
				"Torbalı":           "Torbalı/İzmir",
				"Tuzla":             "Tuzla/Istanbul",
				"Urfa":              "Şanlıurfa",
				"urfa":              "Şanlıurfa",
				"Zonguldak Ereğli":  "Ereğli/Zonguldak",
				"Tespit Edilemeyen": "",
			}

			// Apply corrections - use exact word matching to avoid partial replacements
			for incorrect, correct := range locationCorrections {
				// Only replace if the location exactly matches the incorrect value
				if location == incorrect {
					location = correct
					break // Exit loop once a match is found
				}
			}

			detail.Location = location
		}

		dateMatches := regexp.MustCompile(`(?i)<b>Tarih:\s*</b>\s*(.+?)<br>`).FindStringSubmatch(string(e.Response.Body))
		if len(dateMatches) > 1 {
			detail.Date = strings.TrimSpace(dateMatches[1])
		}

		reasonMatches := regexp.MustCompile(`(?i)<b>Neden öldürüldü:\s*</b>\s*(.*?)(?:<br><b>|<br>|$)`).FindStringSubmatch(string(e.Response.Body))
		if len(reasonMatches) > 1 {
			reason := strings.TrimSpace(reasonMatches[1])
			// Clean reason field from HTML tags that might have been captured
			reason = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(reason, "")
			reason = strings.TrimSpace(reason)

			// Apply corrections - use exact word matching to avoid partial replacements
			for incorrect, correct := range allFieldCorrections {
				if reason == incorrect {
					reason = correct
					break // Exit loop once a match is found
				}
			}
			reason = strings.TrimSpace(strings.ReplaceAll(reason, ",", " "))

			detail.Reason = reason
		}

		byMatches := regexp.MustCompile(`(?i)<b>Kim tarafından öldürüldü:\s*</b>\s*(.*?)(?:<br><b>|<br>|$)`).FindStringSubmatch(string(e.Response.Body))
		if len(byMatches) > 1 {
			by := strings.TrimSpace(byMatches[1])
			// Clean by field from HTML tags that might have been captured
			by = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(by, "")
			by = strings.TrimSpace(by)

			byCorrections := map[string]string{
				"Dini Nikahlı E\u0013şi": "Dini Nikahlı Eşi",
			}

			for incorrect, correct := range byCorrections {
				if by == incorrect {
					by = correct
					break // Exit loop once a match is found
				}
			}

			for incorrect, correct := range allFieldCorrections {
				if by == incorrect {
					by = correct
					break // Exit loop once a match is found
				}
			}

			detail.By = by
		}

		protectionMatches := regexp.MustCompile(`(?i)<b>Korunma talebi:\s*</b>\s*(.*?)(?:<br><b>|<br>|$)`).FindStringSubmatch(string(e.Response.Body))
		if len(protectionMatches) > 1 {
			protection := strings.TrimSpace(protectionMatches[1])
			// Clean protection field from HTML tags that might have been captured
			protection = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(protection, "")
			// Remove any content that looks like it belongs to another field
			protection = regexp.MustCompile(`(?i).*?Öldürülme şekli:\s*`).ReplaceAllString(protection, "")
			protection = strings.TrimSpace(protection)

			// Apply corrections - use exact word matching to avoid partial replacements
			for incorrect, correct := range allFieldCorrections {
				if protection == incorrect {
					protection = correct
					break // Exit loop once a match is found
				}
			}
			protection = strings.TrimSpace(protection)
			detail.Protection = protection
		}

		methodMatches := regexp.MustCompile(`(?i)<b>Öldürülme şekli:\s*</b>\s*(.*?)(?:<br><b>|<br>|$)`).FindStringSubmatch(string(e.Response.Body))
		if len(methodMatches) > 1 {
			method := strings.TrimSpace(methodMatches[1])
			// Clean method field from HTML tags that might have been captured
			method = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(method, "")
			// Remove any content that looks like it belongs to another field
			method = regexp.MustCompile(`(?i).*?Failin durumu:\s*`).ReplaceAllString(method, "")
			method = regexp.MustCompile(`(?i).*?Kaynak:\s*`).ReplaceAllString(method, "")
			method = strings.TrimSpace(method)

			// Apply corrections - use exact word matching to avoid partial replacements
			for incorrect, correct := range allFieldCorrections {
				if method == incorrect {
					method = correct
					break // Exit loop once a match is found
				}
			}
			method = strings.TrimSpace(method)
			detail.Method = method
		}

		statusMatches := regexp.MustCompile(`(?i)<b>Failin durumu:\s*</b>\s*(.+?)(?:<br>|$)`).FindStringSubmatch(string(e.Response.Body))
		if len(statusMatches) > 1 {
			// Clean status from HTML tags and unwanted content
			status := statusMatches[1]
			// Remove HTML tags like <br>, <b>, <a>, etc.
			status = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(status, "")
			// Remove "Kaynak:" prefix and any URLs that might be left
			status = regexp.MustCompile(`(?i).*?Kaynak:\s*`).ReplaceAllString(status, "")
			// Clean extra whitespace
			status = strings.TrimSpace(status)
			// If status is empty or just contains URL fragments, set to empty
			if status == "" || strings.Contains(status, "http") {
				status = ""
			}
			// Apply corrections - use exact word matching to avoid partial replacements
			for incorrect, correct := range allFieldCorrections {
				if status == incorrect {
					status = correct
					break // Exit loop once a match is found
				}
			}
			status = strings.TrimSpace(status)
			detail.Status = status
		}

		sources := e.ChildAttrs("a", "href")
		detail.Source = sources

		// Check if image exists and has valid src attribute
		imgSrc := e.ChildAttr("img", "src")
		if imgSrc != "" {
			// Handle protocol-relative URLs (//domain.com/path)
			if strings.HasPrefix(imgSrc, "//") {
				detail.Image = "https:" + imgSrc
			} else if strings.HasPrefix(imgSrc, "http://") || strings.HasPrefix(imgSrc, "https://") {
				// Full URL, use as is
				detail.Image = imgSrc
			} else {
				// Relative path, prepend base URL
				detail.Image = baseUrl + "/" + imgSrc
			}
		} else {
			detail.Image = ""
		}
	})

	c.Visit(url)

	// Return detail object
	return detail
}

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
		colly.AllowedDomains("anitsayac.com"),

		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./anitsayac_cache"),
	)

	// Create another collector to scrape additional details
	// detailCollector := c.Clone()

	incidents := make([]Incident, 0, 100000)

	/*
		<span class='xxy'>
			<a href='details.aspx?id=38931'  data-width='800' data-height='380'  class='html5lightbox'  adata-group='mygroup' >Burçin Sevgi T.</a>
			</span>
		<span class='xxy'>
			<a href='details.aspx?id=38933'  data-width='800' data-height='380'  class='html5lightbox'  adata-group='mygroup' >Rojin Kabaiş</a>
		</span>
	*/
	// Blog post page scraper
	c.OnHTML("div#divcounter", func(e *colly.HTMLElement) {
		e.ForEach("span.xxy", func(i int, e *colly.HTMLElement) {
			// Detail url: https://anitsayac.com/details.aspx?id=38931
			// Inside ForEach("span.xxy"), e IS the span, so target its child <a> directly.
			detail := getArticleContent(baseUrl + "/" + e.ChildAttr("a", "href"))
			incident := Incident{
				Id: func() int {
					/*
						<span class="xxy bgyear2025"> <a href="details.aspx?id=50364" data-width="800" data-height="380" class="html5lightbox" adata-group="mygroup">Keziban Pars</a></span>
					*/
					id, _ := strconv.Atoi(strings.Split(e.ChildAttr("a", "href"), "=")[1])
					return id
				}(),
				Name:       e.ChildText("a"),
				FullName:   detail.Name,
				Age:        detail.Age,
				Location:   detail.Location,
				Date:       detail.Date,
				Reason:     detail.Reason,
				By:         detail.By,
				Protection: detail.Protection,
				Method:     detail.Method,
				Status:     detail.Status,
				Source:     detail.Source,
				Image:      detail.Image,
				Url:        baseUrl + "/" + e.ChildAttr("a", "href"),
			}

			incidents = append(incidents, incident)
		})
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL.String())
	})

	// Url: https://anitsayac.com/?year=2000
	c.Visit(baseUrl + "/?year=2000")

	// Check if we have valid data before writing
	if len(incidents) == 0 {
		log.Println("No incidents found, not updating files")
		return
	}

	// Write to temporary files first to avoid data loss
	tempJsonFile := jsonFileName + ".tmp"
	tempCsvFile := csvFileName + ".tmp"

	// Write JSON to temporary file
	file, err := os.Create(tempJsonFile)
	if err != nil {
		log.Fatalf("Cannot create temp file %q: %s\n", tempJsonFile, err)
		return
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	// Encode json to temporary file
	if err := enc.Encode(incidents); err != nil {
		log.Fatalf("Failed to encode JSON: %s\n", err)
		return
	}
	file.Close()

	// Write CSV to temporary file
	csvTempFile, err := os.Create(tempCsvFile)
	if err != nil {
		log.Fatalf("Cannot create temp CSV file %q: %s\n", tempCsvFile, err)
		return
	}
	defer csvTempFile.Close()

	csvTempFile.WriteString("Id,Name,FullName,Age,Location,Date,Reason,By,Protection,Method,Status,Source,Image,Url\n")
	for _, incident := range incidents {
		csvTempFile.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			incident.Id,
			incident.Name,
			incident.FullName,
			incident.Age,
			incident.Location,
			incident.Date,
			strings.Join(strings.Split(incident.Reason, ","), " "),
			strings.Join(strings.Split(incident.By, ","), " "),
			strings.Join(strings.Split(incident.Protection, ","), " "),
			strings.Join(strings.Split(incident.Method, ","), " "),
			strings.Join(strings.Split(incident.Status, ","), " "),
			ReplaceAll(ReplaceAll(strings.Join(incident.Source, " "), ",", "%2C", 1), "\n", "", 1),
			incident.Image,
			incident.Url,
		))
	}
	csvTempFile.Close()

	// Validate temporary files before replacing originals
	if !validateFiles(tempJsonFile, tempCsvFile, len(incidents)) {
		log.Println("Validation failed, not updating original files")
		// Clean up temporary files
		os.Remove(tempJsonFile)
		os.Remove(tempCsvFile)
		return
	}

	// If validation passes, replace original files
	if err := os.Rename(tempJsonFile, jsonFileName); err != nil {
		log.Fatalf("Failed to replace JSON file: %s\n", err)
		return
	}

	if err := os.Rename(tempCsvFile, csvFileName); err != nil {
		log.Fatalf("Failed to replace CSV file: %s\n", err)
		return
	}

	fmt.Printf("Successfully updated files with %d incidents\n", len(incidents))
}
