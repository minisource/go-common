package helper

import (
	"encoding/json"
	"fmt"
)

// DeserializeResponse takes in the base response and the target struct type.
func DeserializeResponse(data []byte, target interface{}) error {
    // First, deserialize the full base response (e.g., BaseHttpResponse).
    var baseResponse BaseHttpResponse
    err := json.Unmarshal(data, &baseResponse)
    if err != nil {
        return fmt.Errorf("error deserializing base response: %v", err)
    }

    // Check if the response was successful.
    if !baseResponse.Success {
        return fmt.Errorf("response not successful: %v", baseResponse.Error)
    }

    // Now, deserialize the 'Result' field into the target type (like User).
    resultData, _ := json.Marshal(baseResponse.Result) // Marshal Result field into JSON.
    err = json.Unmarshal(resultData, target)
    if err != nil {
        return fmt.Errorf("error deserializing result data: %v", err)
    }

    return nil
}