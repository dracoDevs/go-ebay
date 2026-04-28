package trading

// UploadSiteHostedPicture exists as a Trading XML shim because the Sell
// Inventory API has no REST equivalent for eBay-hosted image uploads (EPS).
// Callers should pass either ExternalPictureURL (eBay fetches the image
// from your CDN) or PictureData (base64-encoded bytes inline).
type UploadSiteHostedPicture struct {
	ExternalPictureURL string `xml:"ExternalPictureURL,omitempty"`
	PictureName        string `xml:"PictureName,omitempty"`
	PictureSet         string `xml:"PictureSet,omitempty"`
	PictureUploadPolicy string `xml:"PictureUploadPolicy,omitempty"`
}

func (c UploadSiteHostedPicture) CallName() string { return "UploadSiteHostedPicture" }

func (c UploadSiteHostedPicture) Body() interface{} {
	return c
}

func (c UploadSiteHostedPicture) ParseResponse(r []byte) (Response, error) {
	return ParseXMLResponse[UploadSiteHostedPictureResponse](r)
}

type UploadSiteHostedPictureResponse struct {
	BaseResponse
	SiteHostedPictureDetails *SiteHostedPictureDetails `xml:"SiteHostedPictureDetails,omitempty"`
}

type SiteHostedPictureDetails struct {
	PictureFormat string         `xml:"PictureFormat,omitempty"`
	FullURL       string         `xml:"FullURL"`
	BaseURL       string         `xml:"BaseURL,omitempty"`
	PictureSetMember []PictureSetMember `xml:"PictureSetMember,omitempty"`
}

type PictureSetMember struct {
	MemberURL    string `xml:"MemberURL,omitempty"`
	PictureWidth int    `xml:"PictureWidth,omitempty"`
	PictureHeight int   `xml:"PictureHeight,omitempty"`
}
