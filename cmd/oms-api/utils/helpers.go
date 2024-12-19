package utils

import (
	"context"
	"log"

	"google.golang.org/grpc/metadata"
)

func LogError(err error) {
	if err != nil {
		log.Println("Error: ", err)
	}
}

// AddJWTToContext adds the JWT token to the outgoing gRPC context metadata.
func AddJWTToContext(ctx context.Context, token string) context.Context {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	return metadata.NewOutgoingContext(ctx, md)
}
