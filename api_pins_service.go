package sips

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// PinsApiService is a service that implents the logic for the PinsApiServicer
// This service should implement the business logic for every endpoint for the PinsApi API.
// Include any external packages or services that will be required by this service.
type PinsApiService struct {
}

// NewPinsApiService creates a default api service
func NewPinsApiService() PinsApiServicer {
	return &PinsApiService{}
}

// PinsGet - List pin objects
func (s *PinsApiService) PinsGet(ctx context.Context, cid []string, name string, match TextMatchingStrategy, status []Status, before time.Time, after time.Time, limit int32, meta map[string]string) (ImplResponse, error) {
	// TODO - update PinsGet with the required logic for this service method.
	// Add api_pins_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, PinResults{}) or use other options such as http.Ok ...
	//return Response(200, PinResults{}), nil

	//TODO: Uncomment the next line to return response Response(400, Failure{}) or use other options such as http.Ok ...
	//return Response(400, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(401, Failure{}) or use other options such as http.Ok ...
	//return Response(401, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(404, Failure{}) or use other options such as http.Ok ...
	//return Response(404, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(409, Failure{}) or use other options such as http.Ok ...
	//return Response(409, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(4XX, Failure{}) or use other options such as http.Ok ...
	//return Response(4XX, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(5XX, Failure{}) or use other options such as http.Ok ...
	//return Response(5XX, Failure{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("PinsGet method not implemented")
}

// PinsPost - Add pin object
func (s *PinsApiService) PinsPost(ctx context.Context, pin Pin) (ImplResponse, error) {
	// TODO - update PinsPost with the required logic for this service method.
	// Add api_pins_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(202, PinStatus{}) or use other options such as http.Ok ...
	//return Response(202, PinStatus{}), nil

	//TODO: Uncomment the next line to return response Response(400, Failure{}) or use other options such as http.Ok ...
	//return Response(400, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(401, Failure{}) or use other options such as http.Ok ...
	//return Response(401, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(404, Failure{}) or use other options such as http.Ok ...
	//return Response(404, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(409, Failure{}) or use other options such as http.Ok ...
	//return Response(409, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(4XX, Failure{}) or use other options such as http.Ok ...
	//return Response(4XX, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(5XX, Failure{}) or use other options such as http.Ok ...
	//return Response(5XX, Failure{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("PinsPost method not implemented")
}

// PinsRequestidDelete - Remove pin object
func (s *PinsApiService) PinsRequestidDelete(ctx context.Context, requestid string) (ImplResponse, error) {
	// TODO - update PinsRequestidDelete with the required logic for this service method.
	// Add api_pins_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(202, {}) or use other options such as http.Ok ...
	//return Response(202, nil),nil

	//TODO: Uncomment the next line to return response Response(400, Failure{}) or use other options such as http.Ok ...
	//return Response(400, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(401, Failure{}) or use other options such as http.Ok ...
	//return Response(401, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(404, Failure{}) or use other options such as http.Ok ...
	//return Response(404, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(409, Failure{}) or use other options such as http.Ok ...
	//return Response(409, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(4XX, Failure{}) or use other options such as http.Ok ...
	//return Response(4XX, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(5XX, Failure{}) or use other options such as http.Ok ...
	//return Response(5XX, Failure{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("PinsRequestidDelete method not implemented")
}

// PinsRequestidGet - Get pin object
func (s *PinsApiService) PinsRequestidGet(ctx context.Context, requestid string) (ImplResponse, error) {
	// TODO - update PinsRequestidGet with the required logic for this service method.
	// Add api_pins_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, PinStatus{}) or use other options such as http.Ok ...
	//return Response(200, PinStatus{}), nil

	//TODO: Uncomment the next line to return response Response(400, Failure{}) or use other options such as http.Ok ...
	//return Response(400, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(401, Failure{}) or use other options such as http.Ok ...
	//return Response(401, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(404, Failure{}) or use other options such as http.Ok ...
	//return Response(404, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(409, Failure{}) or use other options such as http.Ok ...
	//return Response(409, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(4XX, Failure{}) or use other options such as http.Ok ...
	//return Response(4XX, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(5XX, Failure{}) or use other options such as http.Ok ...
	//return Response(5XX, Failure{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("PinsRequestidGet method not implemented")
}

// PinsRequestidPost - Replace pin object
func (s *PinsApiService) PinsRequestidPost(ctx context.Context, requestid string, pin Pin) (ImplResponse, error) {
	// TODO - update PinsRequestidPost with the required logic for this service method.
	// Add api_pins_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(202, PinStatus{}) or use other options such as http.Ok ...
	//return Response(202, PinStatus{}), nil

	//TODO: Uncomment the next line to return response Response(400, Failure{}) or use other options such as http.Ok ...
	//return Response(400, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(401, Failure{}) or use other options such as http.Ok ...
	//return Response(401, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(404, Failure{}) or use other options such as http.Ok ...
	//return Response(404, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(409, Failure{}) or use other options such as http.Ok ...
	//return Response(409, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(4XX, Failure{}) or use other options such as http.Ok ...
	//return Response(4XX, Failure{}), nil

	//TODO: Uncomment the next line to return response Response(5XX, Failure{}) or use other options such as http.Ok ...
	//return Response(5XX, Failure{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("PinsRequestidPost method not implemented")
}
