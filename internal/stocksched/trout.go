package stocksched

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type TroutCounts struct {
	Stockings []Stock
}

type Stock struct {
	WaterBody string    // DETROIT RES
	Starting  time.Time // Jun. 28, 2021
	Zone      string    // Willamette
	Office    string    // Corvallis
	Legals    int64     // 1,110
	Trophy    int64     // 3,800
	Brood     int64     // 0
	Total     int64     // 4,900
}

func (s Stock) String() string {
	return fmt.Sprintf("Body=%q Starting=%q Zone=%q Office=%q Legals=%d Trophy=%d Brood=%d Total=%d",
		s.WaterBody, s.Starting, s.Zone, s.Office, s.Legals, s.Trophy, s.Brood, s.Total)
}

//<table class="tablesaw tablesaw-stack views-table views-view-table cols-7" data-tablesaw-mode="stack" id="tablesaw-8147">
//
//<tr>
//<td><b>Week of</b><span class="tablesaw-cell-content"><time datetime="2021-06-28T12:00:00Z" class="datetime">Jun. 28, 2021</time> - <time datetime="2021-07-02T12:00:00Z" class="datetime">Jul. 02, 2021</time>
//</span></td>
//<td><b>Waterbody</b><span class="tablesaw-cell-content">FALL CR (WILLAMETTE R)          </span></td>
//<td><b>Zone/Office</b><span class="tablesaw-cell-content">Willamette / Springfield          </span></td>
//<td><b>Legals</b><span class="tablesaw-cell-content">1,200          </span></td>
//<td><b>Trophy</b><span class="tablesaw-cell-content">0          </span></td>
//<td><b>Brood</b><span class="tablesaw-cell-content">0          </span></td>
//<td><b>Total</b><span class="tablesaw-cell-content">1,200          </span></td>
//</tr>

type Zone string

const (
	UnknownZone Zone = "0"
	Central     Zone = "1"
	Northeast   Zone = "2"
	Northwest   Zone = "3"
	Southeast   Zone = "4"
	Southwest   Zone = "5"
	Willamette  Zone = "6"
)

func LoadTroutCounts(client *http.Client, zone Zone) (*TroutCounts, error) {
	resp, err := downloadHtml(client, zone)
	if err != nil {
		return nil, fmt.Errorf("error downloading html: %v", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	counts, err := parseTroutCounts(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing trout counts: %v", err)
	}
	return counts, nil
}

func downloadHtml(client *http.Client, zone Zone) (*http.Response, error) {
	location := fmt.Sprintf("https://myodfw.com/fishing/species/trout/stocking-schedule?field_zone_value=%v", zone)
	req, err := http.NewRequest("GET", location, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %v", err)
	}
	return client.Do(req)
}

func parseTroutCounts(data io.Reader) (*TroutCounts, error) {
	doc, err := htmlquery.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("unable to parse html: %v", err)
	}

	plannedStockDates, err := matchesClass(doc, "views-field-field-planned-stocking-date")
	if err != nil {
		return nil, fmt.Errorf("error with plannedStockDates xpath: %v", err)
	}
	waterbodyNames, err := matchesClass(doc, "views-field-field-waterbody-name")
	if err != nil {
		return nil, fmt.Errorf("error with waterbodyNames xpath: %v", err)
	}
	zones, err := matchesClass(doc, "views-field-field-zone")
	if err != nil {
		return nil, fmt.Errorf("error with zones xpath: %v", err)
	}
	legals, err := matchesClass(doc, "views-field-field-scheduling-life-stage")
	if err != nil {
		return nil, fmt.Errorf("error with legals xpath: %v", err)
	}
	trophies, err := matchesClass(doc, "views-field-field-scheduling-life-stage-1")
	if err != nil {
		return nil, fmt.Errorf("error with trophies xpath: %v", err)
	}
	broods, err := matchesClass(doc, "views-field-field-scheduling-life-stage-2")
	if err != nil {
		return nil, fmt.Errorf("error with broods xpath: %v", err)
	}

	fmt.Printf("DOWNLOAD: plannedStockDates=%d waterbodyNames=%d zones=%d legals=%d trophies=%d broods=%d\n",
		len(plannedStockDates), len(waterbodyNames), len(zones), len(legals), len(trophies), len(broods))

	total := len(plannedStockDates) + len(waterbodyNames) + len(zones) + len(legals) + len(trophies) + len(broods)
	if rem := total % 6; rem != 0 {
		return nil, fmt.Errorf("found uneven elements, remainder=%d", rem)
	}

	out := &TroutCounts{}
	for i := 0; i < len(plannedStockDates); i++ {
		var stocking Stock
		stocking.Starting = parseStartTime(plannedStockDates[i])
		stocking.WaterBody = parseWaterBody(waterbodyNames[i])
		stocking.Zone = parseZone(zones[i])
		stocking.Office = parseOffice(zones[i])
		stocking.Legals = parseCount(legals[i])
		stocking.Trophy = parseCount(trophies[i])
		stocking.Brood = parseCount(broods[i])
		out.Stockings = append(out.Stockings, stocking)
	}
	return out, nil
}

func matchesClass(doc *html.Node, className string) ([]*html.Node, error) {
	elems, err := htmlquery.QueryAll(doc, "//td[@class]")
	if err != nil {
		return nil, fmt.Errorf("error reading td/%s elements", className)
	}
	var out []*html.Node
	for i := range elems {
		for j := range elems[i].Attr {
			if elems[i].Attr[j].Key == "class" && elems[i].Attr[j].Val == fmt.Sprintf("views-field %s", className) {
				out = append(out, elems[i])
			}
		}
	}
	return out, nil
}

func parseStartTime(node *html.Node) time.Time {
	attrs := node.FirstChild.Attr
	for i := range attrs {
		if attrs[i].Key == "datetime" {
			tt, err := time.Parse("2006-01-02T15:04:05Z", attrs[i].Val)
			if err == nil {
				return tt
			}
			fmt.Printf("%s  error=%v\n", attrs[i].Val, err)
		}
	}
	return time.Time{}
}

func parseWaterBody(node *html.Node) string {
	return strings.TrimSpace(node.FirstChild.Data)
}

func parseZone(node *html.Node) string {
	return strings.TrimSpace(strings.Split(node.FirstChild.Data, "/")[0])
}

func parseOffice(node *html.Node) string {
	return strings.TrimSpace(strings.Split(node.FirstChild.Data, "/")[1])
}

func parseCount(node *html.Node) int64 {
	cleaned := strings.TrimSpace(node.FirstChild.Data)
	cleaned = strings.ReplaceAll(cleaned, ",", "")

	n, _ := strconv.ParseInt(cleaned, 10, 32)
	return n
}
