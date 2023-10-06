package transform

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Struct for holding the JSON data as a map
type JSONObject map[string]interface{}

// Struct for holding the XML data as a string
type XMLObject struct {
	XMLName xml.Name
	Content []XMLField `xml:",any"`
}

// Struct for each field in the XML data
type XMLField struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}
// Header transformations
func AddHeader(r *http.Request, key, value string) {
	r.Header.Add(key, value)
}

func RemoveHeader(r *http.Request, key string) {
	r.Header.Del(key)
}

func ModifyHeader(r *http.Request, key, value string) {
	r.Header.Set(key, value)
}

// Body transformations
func ModifyBody(r *http.Request, newBody []byte) {
	r.Body = io.NopCloser(bytes.NewReader(newBody))
	r.ContentLength = int64(len(newBody))
}

func CompressBody(r *http.Request) error {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := io.Copy(gz, r.Body); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	r.Body = io.NopCloser(&b)
	r.Header.Set("Content-Encoding", "gzip")
	r.ContentLength = int64(b.Len())
	return nil
}
// DecompressBody decompresses the gzip-compressed request body
func DecompressBody(r *http.Request) error {
	reader, err := gzip.NewReader(r.Body)
	if err != nil {
		return err
	}
	defer reader.Close()

	var b bytes.Buffer
	if _, err := io.Copy(&b, reader); err != nil {
		return err
	}

	r.Body = io.NopCloser(&b)
	r.Header.Del("Content-Encoding")
	r.ContentLength = int64(b.Len())
	return nil
}
// ConvertXMLToJSON converts the XML body of a request to JSON
func ConvertXMLToJSON(r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var xmlObject XMLObject
	if err := xml.Unmarshal(body, &xmlObject); err != nil {
		return err
	}

	jsonObject := make(JSONObject)
	for _, field := range xmlObject.Content {
		jsonObject[field.XMLName.Local] = field.Value
	}

	jsonBytes, err := json.MarshalIndent(jsonObject, "", "  ")
	if err != nil {
		return err
	}

	jsonString := string(jsonBytes)
	r.Body = io.NopCloser(strings.NewReader(jsonString))
	r.ContentLength = int64(len(jsonString))
	r.Header.Set("Content-Type", "application/json")
	
	return nil
}

func convertMapToXML(jsonData map[string]interface{}) (XMLObject, error) {
    xmlData := XMLObject{XMLName: xml.Name{Local: "root"}}
    
    for key, value := range jsonData {
        switch v := value.(type) {
        case map[string]interface{}:
            nestedXML, err := convertMapToXML(v)
            if err != nil {
                return XMLObject{}, err
            }
            xmlData.Content = append(xmlData.Content, XMLField{XMLName: xml.Name{Local: key}, Nested: &nestedXML})
        case []interface{}:
            // Handle array - this might need further enhancement to handle array of objects
            for _, item := range v {
                itemStr := fmt.Sprintf("%v", item)
                xmlData.Content = append(xmlData.Content, XMLField{XMLName: xml.Name{Local: key}, Value: itemStr})
            }
        default:
            xmlData.Content = append(xmlData.Content, XMLField{XMLName: xml.Name{Local: key}, Value: fmt.Sprintf("%v", v)})
        }
    }

    return xmlData, nil
}

func ConvertJSONToXML(r *http.Request) error {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        return err
    }

    var jsonData map[string]interface{}
    if err := json.Unmarshal(body, &jsonData); err != nil {
        return err
    }

    xmlData, err := convertMapToXML(jsonData)
    if err != nil {
        return err
    }

    xmlBytes, err := xml.MarshalIndent(xmlData, "", "  ")
    if err != nil {
        return err
    }

    r.Body = io.NopCloser(bytes.NewReader(xmlBytes))
    r.ContentLength = int64(len(xmlBytes))
    r.Header.Set("Content-Type", "application/xml")

    return nil
}

// URL transformations
func RewriteURL(r *http.Request, newURL string) error {
	parsedURL, err := url.Parse(newURL)
	if err != nil {
		return err
	}
	r.URL = parsedURL
	return nil
}

func ModifyQueryParam(r *http.Request, param, value string) {
	q := r.URL.Query()
	q.Set(param, value)
	r.URL.RawQuery = q.Encode()
}
