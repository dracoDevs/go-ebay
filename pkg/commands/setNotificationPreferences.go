package commands

import (
	"encoding/xml"

	"github.com/dracoDevs/go-ebay/pkg/ebay"
)

type DeliveryURLDetail struct {
	DeliveryURL     string
	DeliveryURLName string
	Status          string
}

type ApplicationDeliveryPreferences struct {
	AlertEmail         string              `xml:",omitempty"`
	AlertEnable        string              `xml:",omitempty"`
	ApplicationEnable  string              `xml:",omitempty"`
	ApplicationURL     string              `xml:",omitempty"`
	DeliveryURLDetails []DeliveryURLDetail `xml:",omitempty"`
	DeviceType         string              `xml:",omitempty"`
	PayloadVersion     string              `xml:",omitempty"`
}

type EventProperty struct {
	EventType string
	Name      string
	Value     string
}

type UserData struct {
	ExternalUserData string
}

type NotificationEnable struct {
	EventEnable string
	EventType   string
}

type UserDeliveryPreferenceArray struct {
	NotificationEnable []NotificationEnable
}

type SetNotificationPreferences struct {
	ApplicationDeliveryPreferences *ApplicationDeliveryPreferences `xml:"ApplicationDeliveryPreferences,omitempty"`
	DeliveryURLName                string                         `xml:"DeliveryURLName,omitempty"`
	EventProperty                  []EventProperty                `xml:"EventProperty,omitempty"`
	UserData                       *UserData                       `xml:"UserData,omitempty"`
	UserDeliveryPreferenceArray    UserDeliveryPreferenceArray    `xml:"UserDeliveryPreferenceArray,omitempty"`
	ErrorLanguage                  string                         `xml:"ErrorLanguage,omitempty"`
	MessageID                      string                         `xml:"MessageID,omitempty"`
	Version                        string                         `xml:"Version,omitempty"`
	WarningLevel                   string                         `xml:"WarningLevel,omitempty"`
}

func (c SetNotificationPreferences) CallName() string {
	return "SetNotificationPreferences"
}

func (c SetNotificationPreferences) Body() interface{} {	
	return c
}

func (c SetNotificationPreferences) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	var xmlResponse SetNotificationPreferencesResponse
	err := xml.Unmarshal(r, &xmlResponse)

	return xmlResponse, err
}

type SetNotificationPreferencesResponse struct {
	ebay.OtherEbayResponse
}

func (r SetNotificationPreferencesResponse) ResponseErrors() ebay.EbayErrors {
	return r.OtherEbayResponse.Errors
}
