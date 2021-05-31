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
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"backend.juicedbot.io/m/v2/juiced.sitescripts/util"
)

func BecomeGuest(client *http.Client) bool {
	resp, err := util.MakeRequest(&util.Request{
		Client:     *client,
		Method:     "GET",
		URL:        BaseEndpoint,
		RawHeaders: DefaultRawHeaders,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()

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

// Creates the webhook depending on whether successful or not
func (task *Task) CreateBestbuyFields(success bool) []util.Field {
	// When monitoring in Fast mode there is no way to find the name so this field will be empty and the
	// webhook would fail to send. This makes it NaN if this is the case.
	if task.CheckoutInfo.ItemName == "" {
		task.CheckoutInfo.ItemName = "*NaN*"
	}
	return []util.Field{
		{
			Name:   "Site:",
			Value:  "Bestbuy",
			Inline: true,
		},
		{
			Name:   "Price:",
			Value:  "$" + fmt.Sprint(task.CheckoutInfo.Price),
			Inline: true,
		},
		{
			Name:   "Product SKU:",
			Value:  fmt.Sprintf("[%v](https://www.bestbuy.com/site/%v.p?skuId=%v)", task.CheckoutInfo.SKUInStock, task.CheckoutInfo.SKUInStock, task.CheckoutInfo.SKUInStock),
			Inline: true,
		},
		{
			Name:  "Product Name:",
			Value: task.CheckoutInfo.ItemName,
		},
		{
			Name:  "Task Type:",
			Value: string(task.TaskType),
		},
		{
			Name:  "Proxy:",
			Value: "||" + " " + util.ProxyCleaner(task.Task.Proxy) + " " + "||",
		},
	}
}

// CreateParams turns a string->string map into a URL parameter string
func CreateParams(paramsLong map[string]string) string {
	params := url.Values{}
	for key, value := range paramsLong {
		params.Add(key, value)
	}
	return params.Encode()
}
