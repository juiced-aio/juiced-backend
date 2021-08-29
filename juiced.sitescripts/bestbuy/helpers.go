package bestbuy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func BecomeGuest(client http.Client) bool {
	resp, _, err := util.MakeRequest(&util.Request{
		Client:     client,
		Method:     "GET",
		URL:        BaseEndpoint,
		RawHeaders: DefaultRawHeaders,
	})
	if err != nil || resp.StatusCode != 200 {
		if err != nil {
			log.Println(err.Error())
		}
		return false
	}

	return true
}

func encrypt(encrypt []byte, key string) string {
	pubKeyBlock, _ := pem.Decode([]byte(key))

	random := rand.Reader

	var pub *rsa.PublicKey
	pubInterface, parseErr := x509.ParsePKIXPublicKey(pubKeyBlock.Bytes)
	ok := util.HandleErrors(parseErr, util.EncryptionParsingError)
	if !ok {
		return ""
	}

	pub = pubInterface.(*rsa.PublicKey)
	encryptedData, encryptErr := rsa.EncryptOAEP(sha1.New(), random, pub, encrypt, nil)
	ok = util.HandleErrors(encryptErr, util.EncryptionEncryptingError)
	if !ok {
		return ""
	}

	encodedData := base64.StdEncoding.EncodeToString(encryptedData)

	return encodedData
}

func CheckTime(uuid string) (float64, error) {
	// 1. Split on dash
	sections := strings.Split(uuid, "-")

	// 2. Parse each base16 string to base10
	sectionsMappedToBase10 := []int64{}
	for _, section := range sections {
		sectionMappedToBase10, err := strconv.ParseInt(section, 16, 64)
		if err != nil {
			log.Println(err.Error())
			return -1, err
		}
		sectionsMappedToBase10 = append(sectionsMappedToBase10, sectionMappedToBase10)
	}

	// o.wIuzV(10 * o[t("0x9d", "R6oB")](o[r("0x52")](parseInt, o[r("0x21")](i[2], i[3]), 16), u[1]), 100);
	// --> 100 * 10 * o[t("0x9d", "R6oB")](o[r("0x52")](parseInt, o[r("0x21")](i[2], i[3]), 16), u[1])
	// --> 1000 * o[xJJTb](o[xSemx](parseInt, o[OFIaw](i[2], i[3]), 16), u[1])
	// --> 1000 * (o[xSemx](parseInt, o[OFIaw](i[2], i[3]), 16) / u[1]
	// --> 1000 * (parseInt(o[OFIaw](i[2], i[3]), 16) / u[1])
	// --> 1000 * (parseInt(i[2] + i[3], 16) / u[1])
	// --> 1000 * (parseInt(sections[2] + sections[3], 16) / sectionsMappedToBase10[1])

	// 3. Do that^
	middleSectionsToBase10, err := strconv.ParseInt(sections[2]+sections[3], 16, 64)
	if err != nil {
		log.Println(err.Error())
		return -1, err
	}
	//milliseconds := int64(1000) * middleSectionsToBase10 / sectionsMappedToBase10[1]
	return float64(1000) * float64(middleSectionsToBase10) / float64(sectionsMappedToBase10[1]) / 60000, nil
}

// Creates a embed for the DiscordWebhook function
func (task *Task) CreateBestbuyEmbed(status enums.OrderStatus, imageURL string) []sec.DiscordEmbed {
	embeds := []sec.DiscordEmbed{
		{
			Fields: []sec.DiscordField{
				{
					Name:   "Site:",
					Value:  "BestBuy",
					Inline: true,
				},
				{
					Name:   "Price:",
					Value:  "$" + fmt.Sprint(task.StockData.Price),
					Inline: true,
				},
				{
					Name:   "Product SKU:",
					Value:  fmt.Sprintf("[%v](https://www.bestbuy.com/site/%v.p?skuId=%v)", task.StockData.SKU, task.StockData.SKU, task.StockData.SKU),
					Inline: true,
				},
				{
					Name:  "Product Name:",
					Value: task.StockData.ProductName,
				},
				{
					Name:  "Task Type:",
					Value: string(task.TaskType),
				},
				{
					Name:  "Proxy:",
					Value: "||" + " " + util.ProxyCleaner(task.Task.Proxy) + " " + "||",
				},
			},
			Footer: sec.DiscordFooter{
				Text:    "Juiced AIO",
				IconURL: "https://media.discordapp.net/attachments/849430464036077598/855979506204278804/Icon_1.png?width=128&height=128",
			},
			Timestamp: time.Now(),
		},
	}

	switch status {
	case enums.OrderStatusSuccess:
		embeds[0].Title = ":tangerine: Checkout! :tangerine:"
		embeds[0].Color = 16742912
		embeds[0].Thumbnail = sec.DiscordThumbnail{
			URL: imageURL,
		}
	case enums.OrderStatusDeclined:
		embeds[0].Title = ":lemon: Card Declined :lemon:"
		embeds[0].Color = 16766464
		embeds[0].Thumbnail = sec.DiscordThumbnail{
			URL: imageURL,
		}

	case enums.OrderStatusFailed:
		embeds[0].Title = ":apple: Failed to Place Order :apple:"
		embeds[0].Color = 14495044
		embeds[0].Thumbnail = sec.DiscordThumbnail{
			URL: imageURL,
		}

	default:
		embeds[0].Title = fmt.Sprintf(":apple: Failed to Place Order: %s :apple:", status)
		embeds[0].Color = 14495044
		embeds[0].Thumbnail = sec.DiscordThumbnail{
			URL: imageURL,
		}
	}

	return embeds

}

// CreateParams turns a string->string map into a URL parameter string
func CreateParams(paramsLong map[string]string) string {
	params := url.Values{}
	for key, value := range paramsLong {
		params.Add(key, value)
	}
	return params.Encode()
}
