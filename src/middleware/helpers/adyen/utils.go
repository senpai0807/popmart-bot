package adyen

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v3"
)

var (
	errNoKey = errors.New("key field is empty")
	liveKey  = "live_T4D4ECRSB5G3DHDXMJHYRUDRP4ER4U52"
	adyenKey = "10001|96D77478B000A309DF65F302F25CD951BA2865BF8A5FB5F06D82CC6D2C7F449509F8F1CE70B528D6FAAD71D910DA098CC0E293D7A32E6E3B0DDD8B7D757EE9607CBC4B54C08CCF48EF2D459A8D5619B443EEE9870BDAAFD0E9257240878C458C0F78FC329145760C67662A82FD786703977849DD83DDE27FBDDDF956CCD580BCE766186AECB88C2FC93F6DA186DFE84769ADDF254EF0DD0F64314CC8413A00628264607DB87E4299C626A62920D2E355F023378B99D7B06747A98439B9C23AAD1809BB453971E61BA35D0A3B020755B66C6486F693D51362FB8A7B635D76623018707AB5F3527BDDCB9AF7166D0E83DFC4AFE071D6874F1649601BB2AA2B31D5"
)

var (
	chunkSize = 32768
	re        = regexp.MustCompile(`(\d{4})(\d{4})(\d{4})(\d{4})`)
)

var (
	errSplit = errors.New("invalid key format: missing '|'")
)

func NewEncryptor(key string) *Encryptor {
	return &Encryptor{}
}

func (enc *Encryptor) SetKey(key string) {
	enc.Key = key
}

func (enc *Encryptor) SetOriginKey(originkey string) {
	enc.OriginKey = originkey
}

func (enc *Encryptor) SetDomain(domain string) {
	enc.Domain = domain
}

func (enc *Encryptor) ParseKey() (err error) {
	if enc.Key == "" {
		return errNoKey
	}
	jwk := DefaultJWK()
	if err = jwk.ParseAdyenKey(enc.Key); err != nil {
		return
	}
	enc.RsaPubKey = jwk.JWKToPem()
	return
}

func DefaultJWK() *JWK {
	return &JWK{
		Kty: "RSA",
		Kid: "asf-key",
		Alg: "RSA-OAEP",
		Use: "sig",
	}
}

func (jwk *JWK) Marshal() []byte {
	m, _ := json.Marshal(jwk)
	return m
}

func (jwk *JWK) ParseAdyenKey(key string) error {
	parts := strings.Split(key, "|")
	if len(parts) < 2 {
		return errSplit
	}
	decodedExponent := HexDecode(parts[0])
	decodedKey := HexDecode(parts[1])
	encodedExponent := EncodeToBase64(decodedExponent)
	encodedKey := EncodeToBase64(decodedKey)
	jwk.E = encodedExponent
	jwk.N = encodedKey
	return nil
}

func (jwk *JWK) JWKToPem() *rsa.PublicKey {
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil
	}
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil
	}

	rsaPub := &rsa.PublicKey{
		N: big.NewInt(0).SetBytes(nBytes),
		E: int(big.NewInt(0).SetBytes(eBytes).Uint64()),
	}
	return rsaPub
}

func NowTimeISO() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.999") + "Z"
}

func FormatCardNumber(cardNumber string) string {
	formatted := re.ReplaceAllString(cardNumber, "$1 $2 $3 $4")
	return formatted
}

func EncodeToBase64(input any) string {
	var inputData []byte

	switch input := input.(type) {
	case string:
		inputData = []byte(input)
	case []byte:
		inputData = input
	default:
		fmt.Println("Adyen Encrypt: [EncodeToBase64: unsupported input type]")
		return ""
	}

	var chunks []string
	for n := 0; n < len(inputData); n += chunkSize {
		end := n + chunkSize
		if end > len(inputData) {
			end = len(inputData)
		}
		chunks = append(chunks, string(inputData[n:end]))
	}

	combinedData := []byte{}
	for _, chunk := range chunks {
		combinedData = append(combinedData, chunk...)
	}

	encoded := base64.StdEncoding.EncodeToString(combinedData)
	encoded = strings.ReplaceAll(encoded, "=", "")
	encoded = strings.ReplaceAll(encoded, "+", "-")
	encoded = strings.ReplaceAll(encoded, "/", "_")
	return encoded
}

func HexDecode(e string) []byte {
	if e == "" {
		return []byte{}
	}

	if len(e)%2 == 1 {
		e = "0" + e
	}

	t := len(e) / 2
	r := make([]byte, t)

	for n := 0; n < t; n++ {
		hexByte := e[2*n : 2*n+2]
		byteValue, err := hex.DecodeString(hexByte)
		if err != nil {
			fmt.Println("Error decoding hex string:", err)
			return nil
		}
		r[n] = byteValue[0]
	}

	return r
}

func PrepareEncryptor(key, originkey, domain string) (enc *Encryptor, err error) {
	if originkey == "" {
		originkey = "live_YCN5QJ4BXJHSTL24DUQMIHO4JQP2XDLK"
	}
	if domain == "" {
		domain = "https://www.bstn.com"
	}
	enc = &Encryptor{
		Key:       key,
		OriginKey: originkey,
		Domain:    domain,
	}
	err = enc.ParseKey()
	return
}

func (enc *Encryptor) EncryptSingle(data []byte) (ret string, err error) {
	rcpt := jose.Recipient{
		Algorithm: jose.KeyAlgorithm(jose.RSA_OAEP),
		Key:       enc.RsaPubKey,
	}

	opts := &jose.EncrypterOptions{
		ExtraHeaders: map[jose.HeaderKey]any{
			"version": "1",
		},
	}

	jwe, err := jose.NewEncrypter(jose.ContentEncryption(jose.A256CBC_HS512), rcpt, opts)
	if err != nil {
		fmt.Println(err)
		return
	}

	cipherText, err := jwe.Encrypt(data)
	if err != nil {
		return
	}
	return cipherText.CompactSerialize()
}

func (enc *Encryptor) EncryptData(number, expMonth, expYear, cvc string) (ret *AdyenData, err error) {
	ret = &AdyenData{}

	ts := NowTimeISO()
	ref := "https://checkoutshopper-live.adyen.com/checkoutshopper/securedfields/" + enc.OriginKey + "/4.5.0/securedFields.html?type=card&d=" + base64.StdEncoding.EncodeToString([]byte(enc.Domain))
	number = FormatCardNumber(number)

	numberPayload := []byte(`{"number":"` + number + `","generationtime":"` + ts + `","numberBind":"1","activate":"3","referrer":"` + ref + `","numberFieldFocusCount":"3","numberFieldLog":"fo@44070,cl@44071,KN@44082,fo@44324,cl@44325,cl@44333,KN@44346,KN@44347,KN@44348,KN@44350,KN@44351,KN@44353,KN@44354,KN@44355,KN@44356,KN@44358,fo@44431,cl@44432,KN@44434,KN@44436,KN@44438,KN@44440,KN@44440","numberFieldClickCount":"4","numberFieldKeyCount":"16"}`)
	expMonthPayload := []byte(`{"expiryMonth":"` + expMonth + `","generationtime":"` + ts + `"}`)
	expYearPayload := []byte(`{"expiryYear":"` + expYear + `","generationtime":"` + ts + `"}`)
	cvcPayload := []byte(`{"cvc":"` + cvc + `","generationtime":"` + ts + `","cvcBind":"1","activate":"4","referrer":"` + ref + `","cvcFieldFocusCount":"4","cvcFieldLog":"fo@122,cl@123,KN@136,KN@138,KN@140,fo@11204,cl@11205,ch@11221,bl@11221,fo@33384,bl@33384,fo@50318,cl@50319,cl@50321,KN@50334,KN@50336,KN@50336","cvcFieldClickCount":"4","cvcFieldKeyCount":"6","cvcFieldChangeCount":"1","cvcFieldBlurCount":"2","deactivate":"2"}`)

	ret.EncryptedCardNumber, err = enc.EncryptSingle(numberPayload)
	if err != nil {
		return
	}
	ret.EncryptedExpiryMonth, err = enc.EncryptSingle(expMonthPayload)
	if err != nil {
		return
	}
	ret.EncryptedExpiryYear, err = enc.EncryptSingle(expYearPayload)
	if err != nil {
		return
	}
	ret.EncryptedSecurityCode, err = enc.EncryptSingle(cvcPayload)
	if err != nil {
		return
	}
	return
}

// ----------------- RISK DATA FUNCTIONS ----------------- \\
func CalculateMd5_b64(input string) string {
	bin := md5_s2b(input)
	hash := md5_cmc5(bin, len(input)*8)
	return md5_binl2b64(hash)
}

func md5_cmc5(g []int, a int) []int {
	requiredLen := ((((a + 64) >> 9) << 4) + 16)
	if len(g) < requiredLen {
		g = append(g, make([]int, requiredLen-len(g))...)
	}
	g[a>>5] |= 0x80 << (a % 32)
	g[(((a+64)>>9)<<4)+14] = a

	h := 1732584193
	i := -271733879
	j := -1732584194
	y := 271733878

	for e := 0; e < len(g); e += 16 {
		b := h
		c := i
		d := j
		f := y

		h = md5_ff(h, i, j, y, g[e+0], 7, -680876936)
		y = md5_ff(y, h, i, j, g[e+1], 12, -389564586)
		j = md5_ff(j, y, h, i, g[e+2], 17, 606105819)
		i = md5_ff(i, j, y, h, g[e+3], 22, -1044525330)
		h = md5_ff(h, i, j, y, g[e+4], 7, -176418897)
		y = md5_ff(y, h, i, j, g[e+5], 12, 1200080426)
		j = md5_ff(j, y, h, i, g[e+6], 17, -1473231341)
		i = md5_ff(i, j, y, h, g[e+7], 22, -45705983)
		h = md5_ff(h, i, j, y, g[e+8], 7, 1770035416)
		y = md5_ff(y, h, i, j, g[e+9], 12, -1958414417)
		j = md5_ff(j, y, h, i, g[e+10], 17, -42063)
		i = md5_ff(i, j, y, h, g[e+11], 22, -1990404162)
		h = md5_ff(h, i, j, y, g[e+12], 7, 1804603682)
		y = md5_ff(y, h, i, j, g[e+13], 12, -40341101)
		j = md5_ff(j, y, h, i, g[e+14], 17, -1502002290)
		i = md5_ff(i, j, y, h, g[e+15], 22, 1236535329)

		h = md5_gg(h, i, j, y, g[e+1], 5, -165796510)
		y = md5_gg(y, h, i, j, g[e+6], 9, -1069501632)
		j = md5_gg(j, y, h, i, g[e+11], 14, 643717713)
		i = md5_gg(i, j, y, h, g[e+0], 20, -373897302)
		h = md5_gg(h, i, j, y, g[e+5], 5, -701558691)
		y = md5_gg(y, h, i, j, g[e+10], 9, 38016083)
		j = md5_gg(j, y, h, i, g[e+15], 14, -660478335)
		i = md5_gg(i, j, y, h, g[e+4], 20, -405537848)
		h = md5_gg(h, i, j, y, g[e+9], 5, 568446438)
		y = md5_gg(y, h, i, j, g[e+14], 9, -1019803690)
		j = md5_gg(j, y, h, i, g[e+3], 14, -187363961)
		i = md5_gg(i, j, y, h, g[e+8], 20, 1163531501)
		h = md5_gg(h, i, j, y, g[e+13], 5, -1444681467)
		y = md5_gg(y, h, i, j, g[e+2], 9, -51403784)
		j = md5_gg(j, y, h, i, g[e+7], 14, 1735328473)
		i = md5_gg(i, j, y, h, g[e+12], 20, -1926607734)

		h = md5_hh(h, i, j, y, g[e+5], 4, -378558)
		y = md5_hh(y, h, i, j, g[e+8], 11, -2022574463)
		j = md5_hh(j, y, h, i, g[e+11], 16, 1839030562)
		i = md5_hh(i, j, y, h, g[e+14], 23, -35309556)
		h = md5_hh(h, i, j, y, g[e+1], 4, -1530992060)
		y = md5_hh(y, h, i, j, g[e+4], 11, 1272893353)
		j = md5_hh(j, y, h, i, g[e+7], 16, -155497632)
		i = md5_hh(i, j, y, h, g[e+10], 23, -1094730640)
		h = md5_hh(h, i, j, y, g[e+13], 4, 681279174)
		y = md5_hh(y, h, i, j, g[e+0], 11, -358537222)
		j = md5_hh(j, y, h, i, g[e+3], 16, -722521979)
		i = md5_hh(i, j, y, h, g[e+6], 23, 76029189)
		h = md5_hh(h, i, j, y, g[e+9], 4, -640364487)
		y = md5_hh(y, h, i, j, g[e+12], 11, -421815835)
		j = md5_hh(j, y, h, i, g[e+15], 16, 530742520)
		i = md5_hh(i, j, y, h, g[e+2], 23, -995338651)

		h = md5_ii(h, i, j, y, g[e+0], 6, -198630844)
		y = md5_ii(y, h, i, j, g[e+7], 10, 1126891415)
		j = md5_ii(j, y, h, i, g[e+14], 15, -1416354905)
		i = md5_ii(i, j, y, h, g[e+5], 21, -57434055)
		h = md5_ii(h, i, j, y, g[e+12], 6, 1700485571)
		y = md5_ii(y, h, i, j, g[e+3], 10, -1894986606)
		j = md5_ii(j, y, h, i, g[e+10], 15, -1051523)
		i = md5_ii(i, j, y, h, g[e+1], 21, -2054922799)
		h = md5_ii(h, i, j, y, g[e+8], 6, 1873313359)
		y = md5_ii(y, h, i, j, g[e+15], 10, -30611744)
		j = md5_ii(j, y, h, i, g[e+6], 15, -1560198380)
		i = md5_ii(i, j, y, h, g[e+13], 21, 1309151649)
		h = md5_ii(h, i, j, y, g[e+4], 6, -145523070)
		y = md5_ii(y, h, i, j, g[e+11], 10, -1120210379)
		j = md5_ii(j, y, h, i, g[e+2], 15, 718787259)
		i = md5_ii(i, j, y, h, g[e+9], 21, -343485551)

		h = md5_safe_add(h, b)
		i = md5_safe_add(i, c)
		j = md5_safe_add(j, d)
		y = md5_safe_add(y, f)
	}

	return []int{h, i, j, y}
}

func md5_cmn(a, j, k, m, b, i int) int {
	return md5_safe_add(md5_bit_rol(md5_safe_add(md5_safe_add(j, a), md5_safe_add(m, i)), b), k)
}

func md5_ff(m, o, a, b, p, c, d int) int {
	return md5_cmn((o&a)|(^o&b), m, o, p, c, d)
}

func md5_gg(m, o, a, b, p, c, d int) int {
	return md5_cmn((o&b)|(a&^b), m, o, p, c, d)
}

func md5_hh(m, o, a, b, p, c, d int) int {
	return md5_cmn(o^a^b, m, o, p, c, d)
}

func md5_ii(m, o, a, b, p, c, d int) int {
	return md5_cmn(a^(o|^b), m, o, p, c, d)
}

func md5_safe_add(x, y int) int {
	l := (x & 0xffff) + (y & 0xffff)
	h := (x >> 16) + (y >> 16) + (l >> 16)
	return (h << 16) | (l & 0xffff)
}

func md5_bit_rol(num, cnt int) int {
	return (num << cnt) | (int(uint32(num) >> uint32(32-cnt)))
}

func md5_s2b(input string) []int {
	n := ((len(input) + 3) >> 2) + 1
	out := make([]int, n)
	for i := 0; i < len(input)*8; i += 8 {
		out[i>>5] |= int(input[i/8]) << (i % 32)
	}
	return out
}

func md5_binl2b64(bin []int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result strings.Builder

	for i := 0; i < len(bin)*4; i += 3 {
		val := ((bin[i>>2] >> (8 * (i % 4))) & 0xFF) << 16
		if (i+1)>>2 < len(bin) {
			val |= ((bin[(i+1)>>2] >> (8 * ((i + 1) % 4))) & 0xFF) << 8
		}
		if (i+2)>>2 < len(bin) {
			val |= (bin[(i+2)>>2] >> (8 * ((i + 2) % 4))) & 0xFF
		}
		for j := 0; j < 4; j++ {
			if (i*8 + j*6) > len(bin)*32 {
				result.WriteByte('=')
			} else {
				result.WriteByte(chars[(val>>(6*(3-j)))&0x3F])
			}
		}
	}

	return result.String()
}

type Uint64 [2]uint32

func x64Add(m, n Uint64) Uint64 {
	mParts := [4]uint32{m[0] >> 16, m[0] & 0xffff, m[1] >> 16, m[1] & 0xffff}
	nParts := [4]uint32{n[0] >> 16, n[0] & 0xffff, n[1] >> 16, n[1] & 0xffff}
	o := [4]uint32{}

	o[3] += mParts[3] + nParts[3]
	o[2] += o[3] >> 16
	o[3] &= 0xffff
	o[2] += mParts[2] + nParts[2]
	o[1] += o[2] >> 16
	o[2] &= 0xffff
	o[1] += mParts[1] + nParts[1]
	o[0] += o[1] >> 16
	o[1] &= 0xffff
	o[0] += mParts[0] + nParts[0]
	o[0] &= 0xffff

	return Uint64{(o[0] << 16) | o[1], (o[2] << 16) | o[3]}
}

func x64Multiply(m, n Uint64) Uint64 {
	mParts := [4]uint32{m[0] >> 16, m[0] & 0xffff, m[1] >> 16, m[1] & 0xffff}
	nParts := [4]uint32{n[0] >> 16, n[0] & 0xffff, n[1] >> 16, n[1] & 0xffff}
	o := [4]uint32{}

	o[3] += mParts[3] * nParts[3]
	o[2] += o[3] >> 16
	o[3] &= 0xffff
	o[2] += mParts[2]*nParts[3] + mParts[3]*nParts[2]
	o[1] += o[2] >> 16
	o[2] &= 0xffff
	o[1] += mParts[1]*nParts[3] + mParts[2]*nParts[2] + mParts[3]*nParts[1]
	o[0] += o[1] >> 16
	o[1] &= 0xffff
	o[0] += mParts[0]*nParts[3] + mParts[1]*nParts[2] + mParts[2]*nParts[1] + mParts[3]*nParts[0]
	o[0] &= 0xffff

	return Uint64{(o[0] << 16) | o[1], (o[2] << 16) | o[3]}
}

func x64Rotl(m Uint64, n uint) Uint64 {
	n %= 64
	if n == 32 {
		return Uint64{m[1], m[0]}
	} else if n < 32 {
		return Uint64{
			(m[0] << n) | (m[1] >> (32 - n)),
			(m[1] << n) | (m[0] >> (32 - n)),
		}
	} else {
		n -= 32
		return Uint64{
			(m[1] << n) | (m[0] >> (32 - n)),
			(m[0] << n) | (m[1] >> (32 - n)),
		}
	}
}

func x64LeftShift(m Uint64, n uint) Uint64 {
	n %= 64
	if n == 0 {
		return m
	} else if n < 32 {
		return Uint64{
			(m[0] << n) | (m[1] >> (32 - n)),
			m[1] << n,
		}
	} else {
		return Uint64{
			m[1] << (n - 32),
			0,
		}
	}
}

func x64Xor(m, n Uint64) Uint64 {
	return Uint64{m[0] ^ n[0], m[1] ^ n[1]}
}

func x64Fmix(h Uint64) Uint64 {
	h = x64Xor(h, Uint64{0, h[0] >> 1})
	h = x64Multiply(h, Uint64{0xff51afd7, 0xed558ccd})
	h = x64Xor(h, Uint64{0, h[0] >> 1})
	h = x64Multiply(h, Uint64{0xc4ceb9fe, 0x1a85ec53})
	h = x64Xor(h, Uint64{0, h[0] >> 1})
	return h
}

func X64Hash128(key string, seed uint32) string {
	length := len(key)
	bytes := length &^ 15
	remainder := length - bytes

	h1 := Uint64{0, seed}
	h2 := Uint64{0, seed}
	var k1, k2 Uint64
	c1 := Uint64{0x87c37b91, 0x114253d5}
	c2 := Uint64{0x4cf5ad43, 0x2745937f}

	for i := 0; i+15 < length; i += 16 {
		k1 = Uint64{
			uint32(key[i+4]) | uint32(key[i+5])<<8 | uint32(key[i+6])<<16 | uint32(key[i+7])<<24,
			uint32(key[i]) | uint32(key[i+1])<<8 | uint32(key[i+2])<<16 | uint32(key[i+3])<<24,
		}
		k2 = Uint64{
			uint32(key[i+12]) | uint32(key[i+13])<<8 | uint32(key[i+14])<<16 | uint32(key[i+15])<<24,
			uint32(key[i+8]) | uint32(key[i+9])<<8 | uint32(key[i+10])<<16 | uint32(key[i+11])<<24,
		}

		k1 = x64Multiply(k1, c1)
		k1 = x64Rotl(k1, 31)
		k1 = x64Multiply(k1, c2)
		h1 = x64Xor(h1, k1)
		h1 = x64Rotl(h1, 27)
		h1 = x64Add(h1, h2)
		h1 = x64Add(x64Multiply(h1, Uint64{0, 5}), Uint64{0, 0x52dce729})

		k2 = x64Multiply(k2, c2)
		k2 = x64Rotl(k2, 33)
		k2 = x64Multiply(k2, c1)
		h2 = x64Xor(h2, k2)
		h2 = x64Rotl(h2, 31)
		h2 = x64Add(h2, h1)
		h2 = x64Add(x64Multiply(h2, Uint64{0, 5}), Uint64{0, 0x38495ab5})
	}

	i := bytes
	k1 = Uint64{}
	k2 = Uint64{}

	for j := 0; j < remainder; j++ {
		if i+j >= length {
			break
		}
		b := uint32(key[i+j])
		shift := uint(j * 8)
		if j < 8 {
			k1 = x64Xor(k1, x64LeftShift(Uint64{0, b}, shift))
		} else {
			k2 = x64Xor(k2, x64LeftShift(Uint64{0, b}, shift-64))
		}
	}

	if remainder > 0 {
		k1 = x64Multiply(k1, c1)
		k1 = x64Rotl(k1, 31)
		k1 = x64Multiply(k1, c2)
		h1 = x64Xor(h1, k1)
	}
	if remainder > 8 {
		k2 = x64Multiply(k2, c2)
		k2 = x64Rotl(k2, 33)
		k2 = x64Multiply(k2, c1)
		h2 = x64Xor(h2, k2)
	}

	h1 = x64Xor(h1, Uint64{0, uint32(length)})
	h2 = x64Xor(h2, Uint64{0, uint32(length)})
	h1 = x64Add(h1, h2)
	h2 = x64Add(h2, h1)
	h1 = x64Fmix(h1)
	h2 = x64Fmix(h2)
	h1 = x64Add(h1, h2)
	h2 = x64Add(h2, h1)

	return fmt.Sprintf("%08x%08x%08x%08x", h1[0], h1[1], h2[0], h2[1])
}

// --------------------- RISK DATA GENERATION FUNCTION --------------------- \\
func NewRiskData(userAgent, language string, colorDepth, deviceMemory, hardwareConcurrency, screenWidth, screenHeight, availWidth, availHeight, timezoneOffset int, timezone, platform string, cpuClass, doNotTrack *string) *RiskData {
	return &RiskData{
		UserAgent:           userAgent,
		Language:            language,
		ColorDepth:          colorDepth,
		DeviceMemory:        deviceMemory,
		HardwareConcurrency: hardwareConcurrency,
		ScreenWidth:         screenWidth,
		ScreenHeight:        screenHeight,
		AvailScreenWidth:    availWidth,
		AvailScreenHeight:   availHeight,
		TimezoneOffset:      timezoneOffset,
		Timezone:            timezone,
		Platform:            platform,
		CpuClass:            cpuClass,
		DoNotTrack:          doNotTrack,
	}
}

func (r *RiskData) Generate() string {
	payload := map[string]any{
		"version":           "1.0.0",
		"deviceFingerprint": r.dfValue(),
		"persistentCookie":  []string{},
		"components":        r.generateComponents(),
	}
	jsonBytes, _ := json.Marshal(payload)
	return base64.StdEncoding.EncodeToString(jsonBytes)
}

func (r *RiskData) dfValue() string {
	return r.generateFingerprint() + ":40"
}

func (r *RiskData) generateComponents() map[string]any {
	components := map[string]any{
		"userAgent":                 r.UserAgent,
		"webdriver":                 false,
		"language":                  r.Language,
		"colorDepth":                r.ColorDepth,
		"deviceMemory":              r.DeviceMemory,
		"pixelRatio":                2,
		"hardwareConcurrency":       r.HardwareConcurrency,
		"screenResolution":          []int{r.ScreenWidth, r.ScreenHeight},
		"availableScreenResolution": []int{r.AvailScreenWidth, r.AvailScreenHeight},
		"timezoneOffset":            r.TimezoneOffset,
		"timezone":                  r.Timezone,
		"sessionStorage":            true,
		"localStorage":              true,
		"indexedDb":                 true,
		"addBehavior":               false,
		"openDatabase":              false,
		"cpuClass":                  "not available",
		"platform":                  r.Platform,
		"doNotTrack":                "not available",
		"plugins": []any{
			[]any{"PDF Viewer", "Portable Document Format", [][]string{{"application/pdf", "pdf"}, {"text/pdf", "pdf"}}},
			[]any{"Chrome PDF Viewer", "Portable Document Format", [][]string{{"application/pdf", "pdf"}, {"text/pdf", "pdf"}}},
			[]any{"Chromium PDF Viewer", "Portable Document Format", [][]string{{"application/pdf", "pdf"}, {"text/pdf", "pdf"}}},
			[]any{"Microsoft Edge PDF Viewer", "Portable Document Format", [][]string{{"application/pdf", "pdf"}, {"text/pdf", "pdf"}}},
			[]any{"WebKit built-in PDF", "Portable Document Format", [][]string{{"application/pdf", "pdf"}, {"text/pdf", "pdf"}}},
		},
		"canvas":                 []string{"canvas winding:yes", `canvas fp:data:image/png;base64,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QIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIBA3wgI0PvmVhqEAAECBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECBJoICNCb6KklQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAgb4REKD3za00CAECBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECBAg0ERCgN9FTS4AAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQJ9IyBA75tbaRACBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECBAgQaCIgQG+ip5YAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIE+kZAgN43t9IgBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECBAgQINBEQIDeRE8tAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECPSNgAC9b26lQQgQIECAAAECBAgQIECAAAECBAgQIECAAAECBAgQIECgiYAAvYmeWgIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBDoGwEBet/cSoMQIECAAAECBAgQIECAAAECBAgQIECAAAECBAgQIECAQBMBAXoTPbUECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBAg0DcCAvS+uZUGIUCAAAECBAgQIECAAAECBAgQIECAAAECBAgQIECAAIEmAgL0JnpqCRAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQKBvBATofXMrDUKAAAECBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECTQQE6E301BIgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIBA3wgI0PvmVhqEAAECBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECBJoICNCb6KklQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAgb4REKD3za00CAECBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECBAg0ERCgN9FTS4AAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQJ9IyBA75tbaRACBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECBAgQaCIgQG+ip5YAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIE+kZAgN43t9IgBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECBAgQINBEQIDeRE8tAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECPSNgAC9b26lQQgQIECAAAECBAgQIECAAAECBAgQIECAAAECBAgQIECgiYAAvYmeWgIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBDoGwEBet/cSoMQIECAAAECBAgQIECAAAECBAgQIECAAAECBAgQIECAQBMBAXoTPbUECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBAg0DcCAvS+uZUGIUCAAAECBAgQIECAAAECBAgQIECAAAECBAgQIECAAIEmAgL0JnpqCRAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQKBvBATofXMrDUKAAAECBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECTQQE6E301BIgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIBA3wgI0PvmVhqEAAECBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECBJoICNCb6KklQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAgb4REKD3za00CAECBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECBAg0ERCgN9FTS4AAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQJ9IyBA75tbaRACBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECBAgQaCIgQG+ip5YAAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIE+kZAgN43t9IgBAgQIECAAAECBAgQIECAAAECBAgQIECAAAECBAgQINBEQIDeRE8tAQIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECPSNgAC9b26lQQgQIECAAAECBAgQIECAAAECBAgQIECAAAECBAgQIECgiYAAvYmeWgIECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBDoGwEBet/cSoMQIECAAAECBAgQIECAAAECBAgQIECAAAECBAgQIECAQBMBAXoTPbUECBAgQIAAAQIECBAgQIAAAQIECBAgQIAAAQIECBAg0DcCAvS+uZUGIUCAAAECBAgQIECAAAECBAgQIECAAAECBAgQIECAAIEmAv8fx7XEbaVyP34AAAAASUVORK5CYII=`},
		"webgl":                  "not available",
		"webglVendorAndRenderer": "",
		"adBlock":                false,
		"hasLiedLanguages":       false,
		"hasLiedResolution":      false,
		"hasLiedOs":              false,
		"hasLiedBrowser":         false,
		"fonts": []string{
			"Andale Mono", "Arial", "Arial Black", "Arial Hebrew", "Arial Narrow", "Arial Rounded MT Bold", "Arial Unicode MS", "Calibri",
			"Comic Sans MS", "Courier", "Courier New", "Geneva", "Georgia", "Helvetica", "Helvetica Neue", "Impact", "LUCIDA GRANDE",
			"Microsoft Sans Serif", "Monaco", "Palatino", "Tahoma", "Times", "Times New Roman", "Trebuchet MS", "Verdana",
			"Wingdings", "Wingdings 2", "Wingdings 3",
		},
		"audio":            "124.04344968475198",
		"enumerateDevices": []string{"id=;gid=;audioinput;", "id=;gid=;videoinput;", "id=;gid=;audiooutput;"},
	}

	parsed := make(map[string]any)
	for k, v := range components {
		switch k {
		case "screenResolution":
			if dims, ok := v.([]int); ok {
				parsed["screenWidth"] = dims[0]
				parsed["screenHeight"] = dims[1]
			}
		case "availableScreenResolution":
			if dims, ok := v.([]int); ok {
				parsed["availableScreenWidth"] = dims[0]
				parsed["availableScreenHeight"] = dims[1]
			}
		default:
			switch vv := v.(type) {
			case bool:
				if vv {
					parsed[k] = 1
				} else {
					parsed[k] = 0
				}
			case string:
				if vv != "not available" && vv != "error" && vv != "excluded" {
					parsed[k] = X64Hash128(vv, 0)
				}
			case int:
				parsed[k] = vv
			case []string:
				parsed[k] = X64Hash128(strings.Join(vv, ""), 0)
			default:
				parsed[k] = X64Hash128(fmt.Sprintf("%v", vv), 0)
			}
		}
	}

	return parsed
}

func (r *RiskData) generateFingerprint() string {
	sizes := map[string]int{
		"plugins":       10,
		"nrOfPlugins":   3,
		"fonts":         10,
		"nrOfFonts":     3,
		"timeZone":      10,
		"video":         10,
		"superCookies":  10,
		"userAgent":     10,
		"mimeTypes":     10,
		"nrOfMimeTypes": 3,
		"canvas":        10,
		"cpuClass":      5,
		"platform":      5,
		"doNotTrack":    5,
		"webglFp":       10,
		"jsFonts":       10,
	}

	getPad := func(str string, size int) string {
		if len(str) >= size {
			return str[:size]
		}
		return strings.Repeat("0", size-len(str)) + str
	}

	E := make(map[string]string)
	E["plugins"] = getPad(CalculateMd5_b64("...plugin data..."), sizes["plugins"])
	E["nrOfPlugins"] = getPad("5", sizes["nrOfPlugins"])
	E["fonts"] = getPad("", sizes["fonts"])
	E["nrOfFonts"] = getPad("", sizes["nrOfFonts"])
	E["timeZone"] = getPad(CalculateMd5_b64("-300**-360"), sizes["timeZone"])
	E["video"] = getPad(fmt.Sprintf("%d", (r.ScreenWidth+7)*(r.ScreenHeight+7)*r.ColorDepth), sizes["video"])
	E["superCookies"] = getPad(CalculateMd5_b64("DOM-LS: No, DOM-SS: No"), sizes["superCookies"]/2) +
		getPad(CalculateMd5_b64(", IE-UD: No"), sizes["superCookies"]/2)
	E["userAgent"] = getPad(CalculateMd5_b64(r.UserAgent), sizes["userAgent"])
	E["mimeTypes"] = getPad(CalculateMd5_b64("undefinedPortable Document Formatapplication/pdfpdfPortable Document Formattext/pdfpdf"), sizes["mimeTypes"])
	E["nrOfMimeTypes"] = getPad("2", sizes["nrOfMimeTypes"])
	E["canvas"] = getPad(CalculateMd5_b64("data:image/png;base64,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"), sizes["canvas"])
	E["cpuClass"] = getPad("", sizes["cpuClass"])
	E["platform"] = getPad(CalculateMd5_b64(r.Platform), sizes["platform"])
	E["doNotTrack"] = getPad("", sizes["doNotTrack"])
	E["webglFp"] = getPad(CalculateMd5_b64("0000000000"), sizes["webglFp"])
	E["jsFonts"] = getPad(CalculateMd5_b64("Wingdings 3;Wingdings 2;Wingdings;Webdings;Verdana;..."), sizes["jsFonts"])

	var f strings.Builder
	for _, k := range []string{"plugins", "nrOfPlugins", "fonts", "nrOfFonts", "timeZone", "video", "superCookies", "userAgent", "mimeTypes", "nrOfMimeTypes", "canvas", "cpuClass", "platform", "doNotTrack", "webglFp", "jsFonts"} {
		f.WriteString(E[k])
	}
	return strings.NewReplacer("+", "G", "/", "D").Replace(f.String())
}
