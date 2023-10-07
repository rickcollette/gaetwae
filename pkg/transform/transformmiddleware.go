package transform

import (
	"log"
	"net/http"
)
type TransformationConfig struct {
    Headers []struct {
        Key   string `json:"key"`
        Value string `json:"value"`
    } `json:"headers"`
    Body struct {
        Type    string `json:"type"`
        Content string `json:"content"`
    } `json:"body"`
    URLParams []struct {  // Added this field to include URLParams
        Key   string `json:"key"`
        Value string `json:"value"`
    } `json:"urlParams"`  // Assuming that URL parameters are similar to headers, adjust as needed
}


type TransformationMiddleware struct {
    Config TransformationConfig
}

func (tm *TransformationMiddleware) ApplyTransformations(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Apply Header Transformations
        for _, header := range tm.Config.Headers {
            r.Header.Set(header.Key, header.Value)  // Modify as needed for add/modify/remove operations
        }

        // Apply Body Transformations
        if tm.Config.Body.Type == "jsonToXml" {
            err := ConvertJSONToXML(r)
            if err != nil {
                // Handle error - maybe log and return an error response, or continue without transformation
                log.Println("Error converting JSON to XML:", err)
            }
        }

        // Apply URL Transformations - you may extend this as needed
        for _, param := range tm.Config.URLParams {
            ModifyQueryParam(r, param.Key, param.Value)
        }
    
        // Call the next handler in the chain
        next.ServeHTTP(w, r)
    })
}

// Usage
// Apply middleware
// transformationMiddleware := &TransformationMiddleware{Config: yourTransformationConfig}
// http.Handle("/your-path", transformationMiddleware.ApplyTransformations(yourNextHandler))
