package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	fieldservice "github.com/example/fieldservice-error-handoff"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	result, err := fieldservice.CaptureDispatchFailure(context.Background(), fieldservice.NewClient(key), fieldservice.WorkOrderFailure{
		WorkOrderID: "WO-1842", DispatchStatus: "photo_upload_failed", PhotoCount: 3,
		TechnicianID: "tech-17", Exception: "image checksum mismatch",
	})
	if err != nil {
		log.Fatal(err)
	}
	out, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}
