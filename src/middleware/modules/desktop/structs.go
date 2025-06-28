package desktop

import "encoding/json"

type OrderedMap []OrderedKV

type OrderedKV struct {
	Key   string
	Value any
}

// ------------ LOGIN STRUCTS ------------ \\
type CheckExistPayload struct {
	Email string `json:"email"`
	S     string `json:"s"`
	T     int64  `json:"t"`
}

type CheckResp struct {
	Code    string   `json:"code"`
	Data    RespData `json:"data"`
	Message string   `json:"message"`
	Now     int      `json:"now"`
	Ret     int      `json:"ret"`
}

type RespData struct {
	User              User `json:"user"`
	IsVerified        bool `json:"is_verified"`
	NeedResetPassword int  `json:"need_reset_password"`
}

type User struct {
	Gid        int    `json:"gid"`
	Nickname   string `json:"nickname"`
	Email      string `json:"email"`
	Country    string `json:"country"`
	Language   string `json:"language"`
	IsVerified bool   `json:"IsVerified"`
	IdentityId string `json:"identityId"`
	AuthType   string `json:"authType"`
	Account    string `json:"account"`
}

type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	S        string `json:"s"`
	T        int64  `json:"t"`
}

type LoginResp struct {
	Code    string    `json:"code"`
	Data    LoginData `json:"data"`
	Message string    `json:"message"`
	Now     int64     `json:"now"`
	Ret     int       `json:"ret"`
}

type LoginData struct {
	Token      string   `json:"token"`
	User       UserInfo `json:"user"`
	IsVerified bool     `json:"is_verified"`
}

type UserInfo struct {
	Gid        int    `json:"gid"`
	Nickname   string `json:"nickname"`
	Email      string `json:"email"`
	Country    string `json:"country"`
	Language   string `json:"language"`
	IsVerified bool   `json:"IsVerified"`
	IdentityId string `json:"identityId"`
	AuthType   string `json:"authType"`
	Account    string `json:"account"`
}

// ------------ PRODUCT STRUCTS ------------ \\
type ProductResp struct {
	Code    string      `json:"code"`
	Data    ProductData `json:"data"`
	Message string      `json:"message"`
	Now     int64       `json:"now"`
	Ret     int         `json:"ret"`
}

type ProductData struct {
	ID                    string            `json:"id"`
	Title                 string            `json:"title"`
	SubTitle              string            `json:"subTitle"`
	BrandID               int               `json:"brandId"`
	Type                  string            `json:"type"`
	Show                  bool              `json:"show"`
	IsPublish             bool              `json:"isPublish"`
	ShowTime              int64             `json:"showTime"`
	HideTime              int64             `json:"hideTime"`
	UpTime                int64             `json:"upTime"`
	DownTime              int64             `json:"downTime"`
	MainImage             string            `json:"mainImage"`
	Desc                  string            `json:"desc"`
	Parameters            []Parameter       `json:"parameters"`
	Skus                  []Sku             `json:"skus"`
	IsAvailable           bool              `json:"isAvailable"`
	LimitQuantity         int               `json:"limitQuantity"`
	PurchasedQuantity     int               `json:"purchasedQuantity"`
	ShareURL              string            `json:"shareURL"`
	LogicCategory         LogicCategory     `json:"logicCategory"`
	Qualify               bool              `json:"qualify"`
	ShowCategories        []Category        `json:"showCategories"`
	Banners               []Banner          `json:"banners"`
	Spec                  string            `json:"spec"`
	Brand                 Brand             `json:"brand"`
	SpuExtID              int               `json:"spuExtID"`
	UserQualificationInfo QualificationInfo `json:"userQualificationInfo"`
	SpuURL                SpuURL            `json:"spuURL"`
	ExpressType           string            `json:"expressType"`
	SaleChannel           string            `json:"saleChannel"`
}

type Parameter struct {
	ID          int    `json:"id"`
	Value       string `json:"value"`
	ParameterID int    `json:"parameterID"`
}

type Sku struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	MainImage           string `json:"mainImage"`
	Price               int    `json:"price"`
	Stock               Stock  `json:"stock"`
	SkuCode             string `json:"skuCode"`
	SendType            string `json:"sendType"`
	LimitQuantity       int    `json:"limitQuantity"`
	PurchasedQuantity   int    `json:"purchasedQuantity"`
	Platform            string `json:"platform"`
	SpecCount           int    `json:"specCount"`
	WishListID          int    `json:"wishListID"`
	SubscribeRestocking bool   `json:"subscribeRestocking"`
	SubscribeAvailable  bool   `json:"subscribeAvailableForSale"`
	DiscountPrice       int    `json:"discountPrice"`
	Currency            string `json:"currency"`
}

type Stock struct {
	OnlineStock     int `json:"onlineStock"`
	OnlineLockStock int `json:"onlineLockStock"`
}

type LogicCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Category struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	Recommended bool   `json:"recommended"`
	Alias       string `json:"alias"`
}

type Banner struct {
	Type   string       `json:"type"`
	Values []BannerItem `json:"values"`
}

type BannerItem struct {
	URL   string         `json:"url"`
	Cover map[string]any `json:"cover"`
}

type Brand struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type QualificationInfo struct {
	ID      int `json:"id"`
	StartAt int `json:"startAt"`
	EndAt   int `json:"endAt"`
}

type SpuURL struct {
	AppPath   string `json:"appPath"`
	WebPath   string `json:"webPath"`
	Type      int    `json:"type"`
	ExtraDesc string `json:"extraDesc"`
}

type ProductDetails struct {
	ProductName string
	SpuId       string
	SkuId       string
	SkuTitle    string
	MainImage   string
	Price       int
	Quantity    int
}

// ------------ ADD TO CART STRUCTS ------------ \\
type AtcPayload struct {
	SkuId       int    `json:"skuId"`
	SpuId       int    `json:"spuId"`
	OffsetCount int    `json:"offsetCount"`
	GID         string `json:"GID"`
	S           string `json:"s"`
	T           int64  `json:"t"`
}

type AtcResp struct {
	Code    string  `json:"code"`
	Data    AtcData `json:"data"`
	Message string  `json:"message"`
	Now     int     `json:"now"`
	Ret     int     `json:"ret"`
}

type AtcData struct {
	Success bool `json:"success"`
}

// ------------ GET ADDRESS STRUCTS ------------ \\
type DefaultResp struct {
	Code    string             `json:"code"`
	Data    DefaultAddressData `json:"data"`
	Message string             `json:"message"`
	Now     int64              `json:"now"`
	Ret     int                `json:"ret"`
}

type DefaultAddressData struct {
	Count int       `json:"count"`
	List  []Address `json:"list"`
}

type Address struct {
	ID           int    `json:"id"`
	ProvinceName string `json:"provinceName"`
	CityName     string `json:"cityName"`
	CountryName  string `json:"countryName"`
	Subdistrict  string `json:"subdistrict"`
	DetailInfo   string `json:"detailInfo"`
	ExtraAddress string `json:"extraAddress"`
	FamilyName   string `json:"familyName"`
	GivenName    string `json:"givenName"`
	MiddleName   string `json:"middleName"`
	TelNumber    string `json:"telNumber"`
	UserName     string `json:"userName"`
	PostalCode   string `json:"postalCode"`
	NationalCode string `json:"nationalCode"`
	Note         string `json:"note"`
	IsDefault    bool   `json:"isDefault"`
	Extra        string `json:"extra"`
	UserID       int    `json:"userID"`
	IdentityNum  string `json:"identityNum"`
	ProvinceCode string `json:"provinceCode"`
	TaxCode      string `json:"taxCode"`
}

// ------------ ADD ADDRESS STRUCTS ------------ \\
type AddressPayload struct {
	Address AddressData `json:"address"`
	S       string      `json:"s"`
	T       int64       `json:"t"`
}

type AddressData struct {
	FirstName    string `json:"givenName"`
	LastName     string `json:"familyName"`
	Phone        string `json:"telNumber"`
	Line1        string `json:"detailInfo"`
	Line2        string `json:"extraAddress"`
	City         string `json:"cityName"`
	PostalCode   string `json:"postalCode"`
	IsDefault    bool   `json:"isDefault"`
	NationalCode string `json:"nationalCode"`
	FullName     string `json:"userName"`
	Country      string `json:"countryName"`
	ProvinceName string `json:"provinceName"`
	ProvinceCode string `json:"provinceCode"`
}

type AddressResp struct {
	Code    string        `json:"code"`
	Data    AddressStruct `json:"data"`
	Message string        `json:"message"`
	Now     int64         `json:"now"`
	Ret     int           `json:"ret"`
}

type AddressStruct struct {
	Address AddressDetail `json:"address"`
}

type AddressDetail struct {
	ID           int    `json:"id"`
	ProvinceName string `json:"provinceName"`
	CityName     string `json:"cityName"`
	CountryName  string `json:"countryName"`
	Subdistrict  string `json:"subdistrict"`
	DetailInfo   string `json:"detailInfo"`
	ExtraAddress string `json:"extraAddress"`
	FamilyName   string `json:"familyName"`
	GivenName    string `json:"givenName"`
	MiddleName   string `json:"middleName"`
	TelNumber    string `json:"telNumber"`
	UserName     string `json:"userName"`
	PostalCode   string `json:"postalCode"`
	NationalCode string `json:"nationalCode"`
	Note         string `json:"note"`
	IsDefault    bool   `json:"isDefault"`
	Extra        string `json:"extra"`
	UserID       int    `json:"userID"`
	IdentityNum  string `json:"identityNum"`
	ProvinceCode string `json:"provinceCode"`
	TaxCode      string `json:"taxCode"`
}

// ------------ CUSTOMER ADDRESS STRUCTS ------------ \\
type CustomerAddress struct {
	AddressId int
	UserId    int
	State     string
	Line1     string
	Line2     string
	City      string
	PostCode  string
	Phone     string
	FirstName string
	LastName  string
}

// ------------ SHIPPING RATE STRUCTS ------------ \\
type RatePayload struct {
	PlaceOrderReq PlaceOrderReq `json:"placeOrderReq"`
	S             string        `json:"s"`
	T             int64         `json:"t"`
}

type PlaceOrderReq struct {
	UserId           int        `json:"userId"`
	PaymentChannel   int        `json:"paymentChannel"`
	SkuItem          []RateItem `json:"skuItem"`
	MpUserCouponId   *string    `json:"mpUserCouponID"`
	UserCouponId     *string    `json:"userCouponID"`
	DiscountCode     *string    `json:"DiscountCode"`
	OrderTotalAmount int        `json:"orderTotalAmount"`
	TotalAmount      int        `json:"totalAmount"`
	Currency         string     `json:"currency"`
}

type RateItem struct {
	SpuId           int64  `json:"spuId"`
	SkuId           int64  `json:"skuId"`
	Count           int    `json:"count"`
	SkuCount        int    `json:"skuCount"`
	Price           int    `json:"price"`
	Title           string `json:"title"`
	SpuTitle        string `json:"spuTitle"`
	DiscountedPrice int    `json:"discountPrice"`
	Cart            int    `json:"currentSKUInCartNum"`
}

type RateResp struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Now     int64    `json:"now"`
	Ret     int      `json:"ret"`
	Data    RateData `json:"data"`
}

type RateData struct {
	ExpressList      []ExpressOption               `json:"expressList"`
	FreeShippingList []FreeShippingOption          `json:"freeShippingList"`
	DiscountList     map[string]DiscountListDetail `json:"discountList"`
}

type ExpressOption struct {
	ExpressCode          string `json:"expressCode"`
	ExpressName          string `json:"expressName"`
	ExpressPrice         int    `json:"expressPrice"`
	ExpressOriginalPrice int    `json:"expressOriginalPrice"`
	Currency             string `json:"currency"`
}

type FreeShippingOption struct {
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	DiscountAmount int        `json:"discountAmount"`
	OriginalAmount int        `json:"originalAmount"`
	Type           string     `json:"type"`
	Step           StepDetail `json:"step"`
	Currency       string     `json:"currency"`
}

type DiscountListDetail struct {
	List []DiscountItem `json:"list"`
}

type DiscountItem struct {
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	DiscountAmount int        `json:"discountAmount"`
	OriginalAmount int        `json:"originalAmount"`
	Type           string     `json:"type"`
	Step           StepDetail `json:"step"`
	Currency       string     `json:"currency"`
}

type StepDetail struct {
	Type     string `json:"type"`
	Must     int    `json:"must"`
	Currency string `json:"currency"`
}

// ------------ CALCULATE TAXES STRUCTS ------------ \\
type TaxesPayload struct {
	UserId     int         `json:"userId"`
	AddressId  int         `json:"AddressId"`
	SkuItem    []TaxesItem `json:"skuItem"`
	Activities []any       `json:"activities"`
	Currency   string      `json:"currency"`
	S          string      `json:"s"`
	T          int64       `json:"t"`
}

type TaxesItem struct {
	SpuId           int64  `json:"spuId"`
	SkuId           int64  `json:"skuId"`
	Count           int    `json:"count"`
	SkuCount        int    `json:"skuCount"`
	Price           int    `json:"price"`
	Title           string `json:"title"`
	SpuTitle        string `json:"spuTitle"`
	DiscountedPrice int    `json:"discountPrice"`
	Cart            int    `json:"currentSKUInCartNum"`
}

type TaxResp struct {
	Code    string  `json:"code"`
	Data    TaxData `json:"data"`
	Message string  `json:"message"`
	Now     int64   `json:"now"`
	Ret     int     `json:"ret"`
}

type TaxData struct {
	TotalAmount    int    `json:"totalAmount"`
	TaxAmount      int    `json:"taxAmount"`
	CouponDiscount int    `json:"couponDiscount"`
	ShowTax        bool   `json:"showTax"`
	RateDiscount   int    `json:"rateDiscount"`
	Currency       string `json:"currency"`
}

// ------------ CREATE ORDER STRUCTS ------------ \\
type CreatePayload struct {
	UserId              int          `json:"userId"`
	AddressId           int          `json:"addressId"`
	TotalAmount         int          `json:"totalAmount"`
	OrderTotalAmount    int          `json:"orderTotalAmount"`
	SkuItem             []CreateItem `json:"skuItem"`
	DiscountCode        *string      `json:"discountCode"`
	UserCouponId        *string      `json:"userCouponID"`
	MpUserCouponId      *string      `json:"mpUserCouponID"`
	ActivityId          *string      `json:"activityId"`
	GiftId              *string      `json:"giftId"`
	Express             ExpressData  `json:"express"`
	BillAddressId       int          `json:"billAddressId"`
	OrderCreatePage     int          `json:"orderCreatePage"`
	SnapshotId          string       `json:"snapshotID"`
	TaxAmount           int          `json:"taxAmount"`
	GwcClickID          string       `json:"gwcClickID"`
	GwcProvider         string       `json:"gwcProvider"`
	Activities          []string     `json:"activities"`
	TrafficSource       string       `json:"trafficSource"`
	TrafficPlatform     string       `json:"trafficPlatform"`
	MegaClotSpecialType string       `json:"megaClotSpecialType"`
	Currency            string       `json:"currency"`
	IsBox               bool         `json:"isBox"`
	Captcha             *string      `json:"captcha_data"`
	S                   string       `json:"s"`
	T                   int64        `json:"t"`
}

type ExpressData struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type CreateItem struct {
	SpuId         int64  `json:"spuId"`
	SkuId         int64  `json:"skuId"`
	Count         int    `json:"count"`
	SkuCount      int    `json:"skuCount"`
	Price         int    `json:"price"`
	Title         string `json:"title"`
	SpuTitle      string `json:"spuTitle"`
	DiscountPrice int    `json:"discountPrice"`
	CurrentCart   int    `json:"currentSKUInCartNum"`
}

type CreateResp struct {
	Code    string     `json:"code"`
	Data    CreateData `json:"data"`
	Message string     `json:"message"`
	Now     int64      `json:"now"`
	Ret     int        `json:"ret"`
}

type CreateData struct {
	OrderNo            string     `json:"orderNo"`
	TradeOrderNum      string     `json:"tradeOrderNum"`
	OrderCreatedTime   string     `json:"orderCreatedTime"`
	PayStatus          int        `json:"payStatus"`
	PayPrice           int        `json:"payPrice"`
	PaymentMethods     any        `json:"paymentMethods"`
	Amount             AmountInfo `json:"amount"`
	ActivityGoods      any        `json:"activityGoods"`
	OrderType          string     `json:"orderType"`
	AutoCloseTimestamp int64      `json:"autoCloseTimestamp"`
	ExpressType        string     `json:"expressType"`
}

type AmountInfo struct {
	Currency string `json:"currency"`
	Value    int    `json:"value"`
}

type OrderDetails struct {
	ProductName  string
	ProductImage string
	SkuId        string
	SpuId        string
	ProductPrice int64
	TotalAmount  int64
	OrderNumber  string
}

// ------------ PAYMENT STRUCTS ------------ \\
type PaymentPayload struct {
	PayType  string   `json:"payType"`
	OrderNo  string   `json:"orderNo"`
	PayMark  string   `json:"payMark"`
	Platform string   `json:"platform"`
	Info     CardInfo `json:"cardInfo"`
	Adyen    string   `json:"adyen"`
	S        string   `json:"s"`
	T        int64    `json:"t"`
}

type CardInfo struct {
	LastFour   string `json:"lastFour"`
	CardBin    string `json:"cardBin"`
	HolderName string `json:"holderName"`
}

type AdyenData struct {
	PaymentMethod  AdyenPayment `json:"paymentMethod"`
	BrowserInfo    AdyenBrowser `json:"browserInfo"`
	StorePayment   bool         `json:"storePaymentMethod"`
	Risk           RiskData     `json:"riskData"`
	AdditionalData Additional   `json:"additionalData"`
	Channel        string       `json:"channel"`
	Origin         string       `json:"origin"`
	ReturnUrl      string       `json:"returnURL"`
}

type AdyenPayment struct {
	Type              string `json:"type"`
	HolderName        string `json:"holderName"`
	CardNumber        string `json:"encryptedCardNumber"`
	ExpiryMonth       string `json:"encryptedExpiryMonth"`
	ExpiryYear        string `json:"encryptedExpiryYear"`
	SecurityCode      string `json:"encryptedSecurityCode"`
	CardBrand         string `json:"brand"`
	CheckoutAttemptId string `json:"checkoutAttemptId"`
}

type AdyenBrowser struct {
	TimezoneOffset    int    `json:"timeZoneOffset"`
	AcceptHeader      string `json:"acceptHeader"`
	JavascriptEnabled bool   `json:"javaScriptEnabled"`
	Language          string `json:"language"`
	JavaEnabled       bool   `json:"javaEnabled"`
	ScreenHeight      int    `json:"screenHeight"`
	ScreenWidth       int    `json:"screenWidth"`
	ColorDepth        int    `json:"colorDepth"`
	UserAgent         string `json:"userAgent"`
}

type RiskData struct {
	ClientData string `json:"clientData"`
}

type Additional struct {
	Allow3DS2 string `json:"allow3DS2"`
}

type ProcessResp struct {
	Code    string         `json:"code"`
	Data    map[string]any `json:"data"`
	Message string         `json:"message"`
	Now     int64          `json:"now"`
	Ret     int            `json:"ret"`
}

type AdyenResponse struct {
	ID string `json:"id"`
}

// ------------ PAYPAL STRUCTS ------------ \\
type PaypalPayload struct {
	OrderNo   string `json:"orderNo"`
	SaveCard  bool   `json:"saveCard"`
	ReturnUrl string `json:"returnURL"`
	CancelUrl string `json:"cancelURL"`
	S         string `json:"s"`
	T         int64  `json:"t"`
}

type PaypalResp struct {
	Code    string     `json:"code"`
	Data    PaypalData `json:"data"`
	Message string     `json:"message"`
	Now     int64      `json:"now"`
	Ret     int        `json:"ret"`
}

type PaypalData struct {
	OrderNo          string       `json:"orderNo"`
	TradeOrderNum    string       `json:"tradeOrderNum"`
	PlatformOrderNum string       `json:"platformOrderNum"`
	Amount           PaypalAmount `json:"amount"`
}

type PaypalAmount struct {
	Currency string  `json:"currency"`
	Value    float64 `json:"value"`
}

// ------------ 3DS STRUCTS ------------ \\
type ThreeDsPayload struct {
	ClientKey         string `json:"clientKey"`
	FingerprintResult string `json:"fingerprintResult"`
	PaymentData       string `json:"paymentData"`
}

type ThreeDsAction struct {
	AuthorisationToken string `json:"authorisationToken"`
	Subtype            string `json:"subtype"`
	Token              string `json:"token"`
	Type               string `json:"type"`
}

type ThreeDsDetails struct {
	ThreeDSResult string `json:"threeDSResult"`
}

type ThreeDsResp struct {
	Action  *json.RawMessage `json:"action,omitempty"`
	Details *json.RawMessage `json:"details,omitempty"`
	Type    string           `json:"type,omitempty"`
}

type ParsedThreeDsResult struct {
	IsAction      bool
	ActionData    *ThreeDsAction
	ThreeDSResult string
}

type CheckPayload struct {
	TradeOrderNum  string         `json:"tradeOrderNum"`
	S              string         `json:"s"`
	DetailsRequest DetailsRequest `json:"detailsRequest"`
	OrderNo        string         `json:"orderNo"`
	T              string         `json:"t"`
}

type DetailsRequest struct {
	Details Details `json:"details"`
}

type Details struct {
	ThreeDsResult string `json:"threeDSResult"`
}

type Check3DsResp struct {
	Data    map[string]any `json:"data"`
	Message string         `json:"message"`
	Now     int64          `json:"now"`
	Ret     int            `json:"ret"`
}
