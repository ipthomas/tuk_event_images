package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ipthomas/tukcnst"
	"github.com/ipthomas/tukdbint"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var dbconn = tukdbint.TukDBConnection{DBUser: os.Getenv(tukcnst.ENV_DB_USER), DBPassword: os.Getenv(tukcnst.ENV_DB_PASSWORD), DBHost: os.Getenv(tukcnst.ENV_DB_HOST), DBPort: os.Getenv(tukcnst.ENV_DB_PORT), DBName: os.Getenv(tukcnst.ENV_DB_NAME)}
var initstate bool

func main() {
	lambda.Start(Handle_Request)
}
func Handle_Request(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.SetFlags(log.Lshortfile)

	log.Printf("Processing API Gateway %s Request Path %s", req.HTTPMethod, req.Path)
	if !initstate {
		if err := tukdbint.NewDBEvent(&dbconn); err != nil {
			return queryResponse(http.StatusInternalServerError, err.Error(), tukcnst.TEXT_PLAIN)
		}
		initstate = true
	}

	switch req.HTTPMethod {
	case http.MethodGet:
		statics := tukdbint.Statics{Action: tukcnst.SELECT}
		static := tukdbint.Static{Name: req.QueryStringParameters[tukcnst.TUK_EVENT_QUERY_PARAM_NAME]}
		statics.Static = append(statics.Static, static)
		if err := tukdbint.NewDBEvent(&statics); err != nil {
			log.Println(err.Error())
			return queryResponse(http.StatusInternalServerError, err.Error(), tukcnst.TEXT_PLAIN)
		}
		if statics.Count == 1 {
			return queryResponse(http.StatusOK, statics.Static[1].Content, tukcnst.TEXT_HTML)
		}
	case http.MethodPost:
		if len(req.Body) != 0 {
			statics := tukdbint.Statics{Action: tukcnst.INSERT}
			static := tukdbint.Static{Name: req.QueryStringParameters[tukcnst.TUK_EVENT_QUERY_PARAM_NAME], Content: req.Body}
			statics.Static = append(statics.Static, static)
			if err := tukdbint.NewDBEvent(&statics); err != nil {
				log.Println(err.Error())
				return queryResponse(http.StatusInternalServerError, err.Error(), tukcnst.TEXT_PLAIN)
			}
			log.Printf("Persisted Image %s", req.QueryStringParameters[tukcnst.TUK_EVENT_QUERY_PARAM_NAME])
		}
	}
	return queryResponse(http.StatusOK, "", tukcnst.TEXT_HTML)
}
func setAwsResponseHeaders(contentType string) map[string]string {
	awsHeaders := make(map[string]string)
	awsHeaders["Server"] = "Event_Image_Server"
	awsHeaders["Access-Control-Allow-Origin"] = "*"
	awsHeaders["Access-Control-Allow-Headers"] = "accept, Content-Type"
	awsHeaders["Access-Control-Allow-Methods"] = "GET, OPTIONS"
	awsHeaders[tukcnst.CONTENT_TYPE] = contentType
	return awsHeaders
}
func queryResponse(statusCode int, body string, contentType string) (*events.APIGatewayProxyResponse, error) {
	return &events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    setAwsResponseHeaders(contentType),
		Body:       body,
	}, nil
}
